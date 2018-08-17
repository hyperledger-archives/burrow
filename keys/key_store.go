package keys

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"

	"github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/scrypt"
)

const (
	scryptN       = 1 << 18
	scryptr       = 8
	scryptp       = 1
	scryptdkLen   = 32
	CryptoNone    = "none"
	CryptoAESGCM  = "scrypt-aes-gcm"
	HashEd25519   = "go-crypto-0.5.0"
	HashSecp256k1 = "btc"
)

//-----------------------------------------------------------------------------
// json encodings

// addresses should be hex encoded

type keyJSON struct {
	CurveType   string
	Address     string
	PublicKey   string
	AddressHash string
	PrivateKey  privateKeyJSON
}

type privateKeyJSON struct {
	Crypto     string
	Plain      string `json:",omitempty"`
	Salt       []byte `json:",omitempty"`
	Nonce      []byte `json:",omitempty"`
	CipherText []byte `json:",omitempty"`
}

func (k *Key) MarshalJSON() (j []byte, err error) {
	jStruct := keyJSON{
		CurveType:   k.CurveType.String(),
		Address:     hex.EncodeUpperToString(k.Address[:]),
		PublicKey:   hex.EncodeUpperToString(k.Pubkey()),
		AddressHash: k.PublicKey.AddressHashType(),
		PrivateKey:  privateKeyJSON{Crypto: CryptoNone, Plain: hex.EncodeUpperToString(k.PrivateKey.RawBytes())},
	}
	j, err = json.Marshal(jStruct)
	return j, err
}

func (k *Key) UnmarshalJSON(j []byte) (err error) {
	keyJ := new(keyJSON)
	err = json.Unmarshal(j, &keyJ)
	if err != nil {
		return err
	}
	if len(keyJ.PrivateKey.Plain) == 0 {
		return fmt.Errorf("no private key")
	}
	curveType, err := crypto.CurveTypeFromString(keyJ.CurveType)
	if err != nil {
		curveType = crypto.CurveTypeEd25519
	}
	privKey, err := hex.DecodeString(keyJ.PrivateKey.Plain)
	if err != nil {
		return err
	}
	k2, err := NewKeyFromPriv(curveType, privKey)
	if err != nil {
		return err
	}

	k.Address = k2.Address
	k.CurveType = curveType
	k.PublicKey = k2.PrivateKey.GetPublicKey()
	k.PrivateKey = k2.PrivateKey

	return nil
}

// returns the address if valid, nil otherwise
func IsValidKeyJson(j []byte) []byte {
	j1 := new(keyJSON)
	e1 := json.Unmarshal(j, &j1)
	if e1 == nil {
		addr, _ := hex.DecodeString(j1.Address)
		return addr
	}
	return nil
}

func NewKeyStore(dir string, AllowBadFilePermissions bool, logger *logging.Logger) *KeyStore {
	return &KeyStore{
		keysDirPath:             dir,
		AllowBadFilePermissions: AllowBadFilePermissions,
		logger:                  logger.With(structure.ComponentKey, "keys").WithScope("NewKeyStore"),
	}
}

type KeyStore struct {
	sync.Mutex
	AllowBadFilePermissions bool
	keysDirPath             string
	logger                  *logging.Logger
}

func (ks *KeyStore) Gen(passphrase string, curveType crypto.CurveType) (key *Key, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("GenerateNewKey error: %v", r)
		}
	}()
	key, err = NewKey(curveType)
	if err != nil {
		return nil, err
	}
	err = ks.StoreKey(passphrase, key)
	return key, err
}

func (ks *KeyStore) GetKey(passphrase string, keyAddr []byte) (*Key, error) {
	ks.Lock()
	defer ks.Unlock()
	dataDirPath, err := returnDataDir(ks.keysDirPath)
	if err != nil {
		return nil, err
	}
	fileContent, err := ks.GetKeyFile(dataDirPath, keyAddr)
	if err != nil {
		return nil, err
	}
	key := new(keyJSON)
	if err = json.Unmarshal(fileContent, key); err != nil {
		return nil, err
	}

	if len(key.PrivateKey.CipherText) > 0 {
		return DecryptKey(passphrase, key)
	} else {
		key := new(Key)
		err = key.UnmarshalJSON(fileContent)
		return key, err
	}
}

func (ks *KeyStore) AllKeys() ([]*Key, error) {

	dataDirPath, err := returnDataDir(ks.keysDirPath)
	if err != nil {
		return nil, err
	}
	addrs, err := GetAllAddresses(dataDirPath)
	if err != nil {
		return nil, err
	}

	var list []*Key

	for _, addr := range addrs {
		k, err := ks.GetKey("", addr)
		if err != nil {
			return nil, err
		}
		list = append(list, k)
	}

	return list, nil
}

func DecryptKey(passphrase string, keyProtected *keyJSON) (*Key, error) {
	salt := keyProtected.PrivateKey.Salt
	nonce := keyProtected.PrivateKey.Nonce
	cipherText := keyProtected.PrivateKey.CipherText

	curveType, err := crypto.CurveTypeFromString(keyProtected.CurveType)
	if err != nil {
		return nil, err
	}
	authArray := []byte(passphrase)
	derivedKey, err := scrypt.Key(authArray, salt, scryptN, scryptr, scryptp, scryptdkLen)
	if err != nil {
		return nil, err
	}
	aesBlock, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, err
	}
	pubKey, err := hex.DecodeString(keyProtected.PublicKey)
	if err != nil {
		return nil, err
	}
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		pkey, _ := NewKeyFromPub(curveType, pubKey)
		return pkey, err
	}
	address, err := crypto.AddressFromHexString(keyProtected.Address)
	if err != nil {
		return nil, err
	}
	k, err := NewKeyFromPriv(curveType, plainText)
	if err != nil {
		return nil, err
	}
	if address != k.Address {
		return nil, fmt.Errorf("address does not match")
	}
	return k, nil
}

func (ks *KeyStore) GetAllAddresses() (addresses [][]byte, err error) {
	ks.Lock()
	defer ks.Unlock()
	return GetAllAddresses(ks.keysDirPath)
}

func (ks *KeyStore) StoreKey(passphrase string, key *Key) error {
	ks.Lock()
	defer ks.Unlock()
	if passphrase != "" {
		return ks.StoreKeyEncrypted(passphrase, key)
	} else {
		return ks.StoreKeyPlain(key)
	}
}

func (ks *KeyStore) StoreKeyPlain(key *Key) (err error) {
	keyJSON, err := json.Marshal(key)
	if err != nil {
		return err
	}
	dataDirPath, err := returnDataDir(ks.keysDirPath)
	if err != nil {
		return err
	}
	err = WriteKeyFile(key.Address[:], dataDirPath, keyJSON)
	return err
}

func (ks *KeyStore) StoreKeyEncrypted(passphrase string, key *Key) error {
	authArray := []byte(passphrase)
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return err
	}

	derivedKey, err := scrypt.Key(authArray, salt, scryptN, scryptr, scryptp, scryptdkLen)
	if err != nil {
		return err
	}

	toEncrypt := key.PrivateKey.RawBytes()

	AES256Block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(AES256Block)
	if err != nil {
		return err
	}

	// XXX: a GCM nonce may only be used once per key ever!
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return err
	}

	// (dst, nonce, plaintext, extradata)
	cipherText := gcm.Seal(nil, nonce, toEncrypt, nil)

	cipherStruct := privateKeyJSON{
		Crypto: CryptoAESGCM, Salt: salt, Nonce: nonce, CipherText: cipherText,
	}
	keyStruct := keyJSON{
		CurveType:   key.CurveType.String(),
		Address:     hex.EncodeUpperToString(key.Address[:]),
		PublicKey:   hex.EncodeUpperToString(key.Pubkey()),
		AddressHash: key.PublicKey.AddressHashType(),
		PrivateKey:  cipherStruct,
	}
	keyJSON, err := json.Marshal(keyStruct)
	if err != nil {
		return err
	}
	dataDirPath, err := returnDataDir(ks.keysDirPath)
	if err != nil {
		return err
	}

	return WriteKeyFile(key.Address[:], dataDirPath, keyJSON)
}

func (ks *KeyStore) DeleteKey(passphrase string, keyAddr []byte) (err error) {
	dataDirPath, err := returnDataDir(ks.keysDirPath)
	if err != nil {
		return err
	}
	keyDirPath := path.Join(dataDirPath, strings.ToUpper(hex.EncodeToString(keyAddr))+".json")
	return os.Remove(keyDirPath)
}

func (ks *KeyStore) GetKeyFile(dataDirPath string, keyAddr []byte) (fileContent []byte, err error) {
	filename := path.Join(dataDirPath, strings.ToUpper(hex.EncodeToString(keyAddr))+".json")
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	if (uint32(fileInfo.Mode()) & 0077) != 0 {
		ks.logger.InfoMsg("file should be accessible by user only", "file", filename)
		if !ks.AllowBadFilePermissions {
			return nil, fmt.Errorf("file %s should be accessible by user only", filename)
		}
	}
	return ioutil.ReadFile(filename)
}

func WriteKeyFile(addr []byte, dataDirPath string, content []byte) (err error) {
	addrHex := strings.ToUpper(hex.EncodeToString(addr))
	keyFilePath := path.Join(dataDirPath, addrHex+".json")
	err = os.MkdirAll(dataDirPath, 0700) // read, write and dir search for user
	if err != nil {
		return err
	}
	return ioutil.WriteFile(keyFilePath, content, 0600) // read, write for user
}

func (ks *KeyStore) GetAllNames() (map[string]string, error) {
	return coreNameList(ks.keysDirPath)
}

func GetAllAddresses(dataDirPath string) (addresses [][]byte, err error) {
	fileInfos, err := ioutil.ReadDir(dataDirPath)
	if err != nil {
		return nil, err
	}
	addresses = make([][]byte, len(fileInfos))
	for i, fileInfo := range fileInfos {
		addr := strings.TrimSuffix(fileInfo.Name(), ".json")
		address, err := hex.DecodeString(addr)
		if err != nil {
			continue
		}
		addresses[i] = address
	}
	return addresses, err
}

type Key struct {
	CurveType  crypto.CurveType
	Address    crypto.Address
	PublicKey  crypto.PublicKey
	PrivateKey crypto.PrivateKey
}

func NewKey(typ crypto.CurveType) (*Key, error) {
	privKey, err := crypto.GeneratePrivateKey(nil, typ)
	if err != nil {
		return nil, err
	}
	pubKey := privKey.GetPublicKey()
	return &Key{
		CurveType:  typ,
		PublicKey:  pubKey,
		Address:    pubKey.Address(),
		PrivateKey: privKey,
	}, nil
}

func (k *Key) Pubkey() []byte {
	return k.PublicKey.PublicKey
}

func NewKeyFromPub(curveType crypto.CurveType, PubKeyBytes []byte) (*Key, error) {
	pubKey, err := crypto.PublicKeyFromBytes(PubKeyBytes, curveType)
	if err != nil {
		return nil, err
	}

	return &Key{
		CurveType: curveType,
		PublicKey: pubKey,
		Address:   pubKey.Address(),
	}, nil
}

func NewKeyFromPriv(curveType crypto.CurveType, PrivKeyBytes []byte) (*Key, error) {
	privKey, err := crypto.PrivateKeyFromRawBytes(PrivKeyBytes, curveType)

	if err != nil {
		return nil, err
	}

	pubKey := privKey.GetPublicKey()

	return &Key{
		CurveType:  curveType,
		Address:    pubKey.Address(),
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}, nil
}

func (k *Key) Sign(hash []byte) ([]byte, error) {
	signature, err := k.PrivateKey.Sign(hash)
	if err != nil {
		return nil, err
	}
	return signature.RawBytes(), nil
}
