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
	PublicKey   []byte
	AddressHash string
	PrivateKey  privateKeyJSON
}

type privateKeyJSON struct {
	Crypto     string
	Plain      []byte `json:",omitempty"`
	Salt       []byte `json:",omitempty"`
	Nonce      []byte `json:",omitempty"`
	CipherText []byte `json:",omitempty"`
}

func (k *Key) MarshalJSON() (j []byte, err error) {
	jStruct := keyJSON{
		CurveType:   k.CurveType.String(),
		Address:     hex.EncodeToString(k.Address[:]),
		PublicKey:   k.Pubkey(),
		AddressHash: k.PublicKey.AddressHashType(),
		PrivateKey:  privateKeyJSON{Crypto: CryptoNone, Plain: k.PrivateKey.RawBytes()},
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
	k2, err := NewKeyFromPriv(curveType, keyJ.PrivateKey.Plain)
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

type KeyStore struct {
	sync.Mutex
	keysDirPath string
}

func (ks KeyStore) Gen(passphrase string, curveType crypto.CurveType) (key *Key, err error) {
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

func (ks KeyStore) GetKey(passphrase string, keyAddr []byte) (*Key, error) {
	ks.Lock()
	defer ks.Unlock()
	dataDirPath, err := returnDataDir(ks.keysDirPath)
	if err != nil {
		return nil, err
	}
	fileContent, err := GetKeyFile(dataDirPath, keyAddr)
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

func (ks KeyStore) AllKeys() ([]*Key, error) {

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
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		pkey, _ := NewKeyFromPub(curveType, keyProtected.PublicKey)
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

func (ks KeyStore) GetAllAddresses() (addresses [][]byte, err error) {
	ks.Lock()
	defer ks.Unlock()
	return GetAllAddresses(ks.keysDirPath)
}

func (ks KeyStore) StoreKey(passphrase string, key *Key) error {
	ks.Lock()
	defer ks.Unlock()
	if passphrase != "" {
		return ks.StoreKeyEncrypted(passphrase, key)
	} else {
		return ks.StoreKeyPlain(key)
	}
}

func (ks KeyStore) StoreKeyPlain(key *Key) (err error) {
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

func (ks KeyStore) StoreKeyEncrypted(passphrase string, key *Key) error {
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
		Address:     strings.ToUpper(hex.EncodeToString(key.Address[:])),
		PublicKey:   key.Pubkey(),
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

func (ks KeyStore) DeleteKey(passphrase string, keyAddr []byte) (err error) {
	dataDirPath, err := returnDataDir(ks.keysDirPath)
	if err != nil {
		return err
	}
	keyDirPath := path.Join(dataDirPath, strings.ToUpper(hex.EncodeToString(keyAddr))+".json")
	return os.Remove(keyDirPath)
}

func GetKeyFile(dataDirPath string, keyAddr []byte) (fileContent []byte, err error) {
	fileName := strings.ToUpper(hex.EncodeToString(keyAddr))
	return ioutil.ReadFile(path.Join(dataDirPath, fileName+".json"))
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

func GetAllAddresses(dataDirPath string) (addresses [][]byte, err error) {
	fileInfos, err := ioutil.ReadDir(dataDirPath)
	if err != nil {
		return nil, err
	}
	addresses = make([][]byte, len(fileInfos))
	for i, fileInfo := range fileInfos {
		addr := strings.TrimSuffix(fileInfo.Name(), "json")
		address, err := hex.DecodeString(addr)
		if err != nil {
			continue
		}
		addresses[i] = address
	}
	return addresses, err
}
