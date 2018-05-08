package keys

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/hyperledger/burrow/crypto"

	"golang.org/x/crypto/scrypt"
)

const (
	scryptN     = 1 << 18
	scryptr     = 8
	scryptp     = 1
	scryptdkLen = 32
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
		k.CurveType.String(),
		fmt.Sprintf("%X", k.Address),
		k.Pubkey(),
		"go-crypto-0.5.0",
		privateKeyJSON{Crypto: "none", Plain: k.PrivateKey.RawBytes()},
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
	// TODO: remove this
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

type keyStoreFile struct {
	sync.Mutex
	keysDirPath string
}

func NewKeyStoreFile(path string) KeyStore {
	return &keyStoreFile{keysDirPath: path}
}

func (ks keyStoreFile) GenerateKey(passphrase string, curveType crypto.CurveType) (key *Key, err error) {
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

func (ks keyStoreFile) GetKey(passphrase string, keyAddr []byte) (*Key, error) {
	ks.Lock()
	defer ks.Unlock()
	fileContent, err := GetKeyFile(ks.keysDirPath, keyAddr)
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

func (ks keyStoreFile) GetAllAddresses() (addresses [][]byte, err error) {
	ks.Lock()
	defer ks.Unlock()
	return GetAllAddresses(ks.keysDirPath)
}

func (ks keyStoreFile) StoreKey(passphrase string, key *Key) error {
	ks.Lock()
	defer ks.Unlock()
	if passphrase != "" {
		return ks.StoreKeyEncrypted(passphrase, key)
	} else {
		return ks.StoreKeyPlain(key)
	}
}

func (ks keyStoreFile) StoreKeyPlain(key *Key) (err error) {
	keyJSON, err := json.Marshal(key)
	if err != nil {
		return err
	}
	err = WriteKeyFile(key.Address[:], ks.keysDirPath, keyJSON)
	return err
}

func (ks keyStoreFile) StoreKeyEncrypted(passphrase string, key *Key) error {
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
		Crypto: "scrypt-aes-gcm", Salt: salt, Nonce: nonce, CipherText: cipherText,
	}
	keyStruct := keyJSON{
		key.CurveType.String(),
		strings.ToUpper(hex.EncodeToString(key.Address[:])),
		key.Pubkey(),
		"go-crypto-0.5.0",
		cipherStruct,
	}
	keyJSON, err := json.Marshal(keyStruct)
	if err != nil {
		return err
	}

	return WriteKeyFile(key.Address[:], ks.keysDirPath, keyJSON)
}

func (ks keyStoreFile) DeleteKey(passphrase string, keyAddr []byte) (err error) {
	keyDirPath := path.Join(ks.keysDirPath, strings.ToUpper(hex.EncodeToString(keyAddr)))
	err = os.RemoveAll(keyDirPath)
	return err
}

func GetKeyFile(keysDirPath string, keyAddr []byte) (fileContent []byte, err error) {
	fileName := strings.ToUpper(hex.EncodeToString(keyAddr))
	return ioutil.ReadFile(path.Join(keysDirPath, fileName, fileName))
}

func WriteKeyFile(addr []byte, keysDirPath string, content []byte) (err error) {
	addrHex := strings.ToUpper(hex.EncodeToString(addr))
	keyDirPath := path.Join(keysDirPath, addrHex)
	keyFilePath := path.Join(keyDirPath, addrHex)
	err = os.MkdirAll(keyDirPath, 0700) // read, write and dir search for user
	if err != nil {
		return err
	}
	return ioutil.WriteFile(keyFilePath, content, 0600) // read, write for user
}

func GetAllAddresses(keysDirPath string) (addresses [][]byte, err error) {
	fileInfos, err := ioutil.ReadDir(keysDirPath)
	if err != nil {
		return nil, err
	}
	addresses = make([][]byte, len(fileInfos))
	for i, fileInfo := range fileInfos {
		address, err := hex.DecodeString(fileInfo.Name())
		if err != nil {
			continue
		}
		addresses[i] = address
	}
	return addresses, err
}
