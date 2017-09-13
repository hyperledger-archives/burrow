/*
	This file is part of go-ethereum

	go-ethereum is free software: you can redistribute it and/or modify
	it under the terms of the GNU Lesser General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	go-ethereum is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU Lesser General Public License
	along with go-ethereum.  If not, see <http://www.gnu.org/licenses/>.
*/
/**
 * @authors
 * 	Gustav Simonsson <gustav.simonsson@gmail.com>
 * @date 2015
 *
 */

/*

This key store behaves as KeyStorePlain with the difference that
the private key is encrypted and on disk uses another JSON encoding.

Cryptography:

1. Encryption key is scrypt derived key from user passphrase. Scrypt parameters
   (work factors) [1][2] are defined as constants below.
2. Scrypt salt is 32 random bytes from CSPRNG. It is appended to ciphertext.
3. Checksum is SHA3 of the private key bytes.
4. Plaintext is concatenation of private key bytes and checksum.
5. Encryption algo is AES 256 CBC [3][4]
6. CBC IV is 16 random bytes from CSPRNG. It is appended to ciphertext.
7. Plaintext padding is PKCS #7 [5][6]

Encoding:

1. On disk, ciphertext, salt and IV are encoded in a nested JSON object.
   cat a key file to see the structure.
2. byte arrays are base64 JSON strings.
3. The EC private key bytes are in uncompressed form [7].
   They are a big-endian byte slice of the absolute value of D [8][9].
4. The checksum is the last 32 bytes of the plaintext byte array and the
   private key is the preceeding bytes.

References:

1. http://www.tarsnap.com/scrypt/scrypt-slides.pdf
2. http://stackoverflow.com/questions/11126315/what-are-optimal-scrypt-work-factors
3. http://en.wikipedia.org/wiki/Advanced_Encryption_Standard
4. http://en.wikipedia.org/wiki/Block_cipher_mode_of_operation#Cipher-block_chaining_.28CBC.29
5. https://leanpub.com/gocrypto/read#leanpub-auto-block-cipher-modes
6. http://tools.ietf.org/html/rfc2315
7. http://bitcoin.stackexchange.com/questions/3059/what-is-a-compressed-bitcoin-key
8. http://golang.org/pkg/crypto/ecdsa/#PrivateKey
9. https://golang.org/pkg/math/big/#Int.Bytes

*/

/*
	Modifications:
		- Ethan Buchman <ethan@erisindustries.com>

	encryption has been modified to use GCM instead of CBC as it
	provides authenticated encryption, rather than managing the
	additional checksum ourselves. The CBC IV is replaced by a Nonce
	that may only be used once ever per key
*/

package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/monax/keys/crypto/randentropy"
	uuid "github.com/wayn3h0/go-uuid"
	"golang.org/x/crypto/scrypt" // 2^18 / 8 / 1 uses 256MB memory and approx 1s CPU time on a modern CPU.
)

const (
	scryptN     = 1 << 18
	scryptr     = 8
	scryptp     = 1
	scryptdkLen = 32
)

type keyStorePassphrase struct {
	keysDirPath string
}

func NewKeyStorePassphrase(path string) KeyStore {
	return &keyStorePassphrase{path}
}

func (ks keyStorePassphrase) GenerateNewKey(typ KeyType, auth string) (key *Key, err error) {
	return GenerateNewKeyDefault(ks, typ, auth)
}

func (ks keyStorePassphrase) GetKey(keyAddr []byte, auth string) (key *Key, err error) {
	key, err = DecryptKey(ks, keyAddr, auth)
	if err != nil {
		return nil, err
	}
	return key, err
}

func (ks keyStorePassphrase) GetAllAddresses() (addresses [][]byte, err error) {
	return GetAllAddresses(ks.keysDirPath)
}

func (ks keyStorePassphrase) StoreKey(key *Key, auth string) (err error) {
	authArray := []byte(auth)
	salt := randentropy.GetEntropyMixed(32)
	derivedKey, err := scrypt.Key(authArray, salt, scryptN, scryptr, scryptp, scryptdkLen)
	if err != nil {
		return err
	}

	keyBytes := key.PrivateKey
	toEncrypt := PKCS7Pad(keyBytes)

	AES256Block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(AES256Block)
	if err != nil {
		return err
	}

	// XXX: a GCM nonce may only be used once per key ever!
	nonce := randentropy.GetEntropyMixed(gcm.NonceSize())

	// (dst, nonce, plaintext, extradata)
	cipherText := gcm.Seal(nil, nonce, toEncrypt, nil)

	cipherStruct := cipherJSON{
		salt,
		nonce,
		cipherText,
	}
	keyStruct := encryptedKeyJSON{
		[]byte(key.Id.String()),
		key.Type.String(),
		strings.ToUpper(hex.EncodeToString(key.Address)),
		cipherStruct,
	}
	keyJSON, err := json.Marshal(keyStruct)
	if err != nil {
		return err
	}

	return WriteKeyFile(key.Address, ks.keysDirPath, keyJSON)
}

func (ks keyStorePassphrase) DeleteKey(keyAddr []byte, auth string) (err error) {
	// only delete if correct passphrase is given
	_, err = DecryptKey(ks, keyAddr, auth)
	if err != nil {
		return err
	}

	keyDirPath := path.Join(ks.keysDirPath, strings.ToUpper(hex.EncodeToString(keyAddr)))
	return os.RemoveAll(keyDirPath)
}

func IsEncryptedKey(ks KeyStore, keyAddr []byte) (bool, error) {
	kspp, ok := ks.(*keyStorePassphrase)
	if !ok {
		return false, fmt.Errorf("only keyStorePassphrase can handle encrypted key files")
	}

	fileContent, err := GetKeyFile(kspp.keysDirPath, keyAddr)
	if err != nil {
		return false, err
	}

	keyProtected := new(encryptedKeyJSON)
	if err = json.Unmarshal(fileContent, keyProtected); err != nil {
		return false, err
	}
	return len(keyProtected.Crypto.CipherText) > 0, nil
}

func DecryptKey(ks keyStorePassphrase, keyAddr []byte, auth string) (*Key, error) {
	fileContent, err := GetKeyFile(ks.keysDirPath, keyAddr)
	if err != nil {
		return nil, err
	}

	keyProtected := new(encryptedKeyJSON)
	if err = json.Unmarshal(fileContent, keyProtected); err != nil {
		return nil, err
	}

	keyId := keyProtected.Id
	keyType, err := KeyTypeFromString(keyProtected.Type)
	if err != nil {
		return nil, err
	}

	keyAddr2, err := hex.DecodeString(keyProtected.Address)
	if bytes.Compare(keyAddr, keyAddr2) != 0 {
		return nil, fmt.Errorf("address of key and address in file do not match. Got %x, expected %x", keyAddr2, keyAddr)
	}
	salt := keyProtected.Crypto.Salt
	nonce := keyProtected.Crypto.Nonce
	cipherText := keyProtected.Crypto.CipherText

	authArray := []byte(auth)
	derivedKey, err := scrypt.Key(authArray, salt, scryptN, scryptr, scryptp, scryptdkLen)
	if err != nil {
		return nil, err
	}
	plainText, err := aesGCMDecrypt(derivedKey, cipherText, nonce)
	if err != nil {
		return nil, err
	}

	// no need to use a checksum as done by gcm
	id, err := uuid.Parse(string(keyId))
	if err != nil {
		return nil, err
	}

	return &Key{
		Id:         id,
		Type:       keyType,
		Address:    keyAddr,
		PrivateKey: plainText,
	}, nil
}
