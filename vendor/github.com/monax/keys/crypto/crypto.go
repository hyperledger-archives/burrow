package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"

	"github.com/monax/keys/crypto/helpers"
	"github.com/monax/keys/crypto/sha3"

	"golang.org/x/crypto/ripemd160"
)

func Ripemd160(data ...[]byte) []byte {
	d := ripemd160.New()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func Sha256(data ...[]byte) []byte {
	d := sha256.New()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func Sha3(data ...[]byte) []byte {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func Sha3Hash(data ...[]byte) (h helpers.Hash) {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}

// From https://leanpub.com/gocrypto/read#leanpub-auto-block-cipher-modes
func PKCS7Pad(in []byte) []byte {
	padding := 16 - (len(in) % 16)
	if padding == 0 {
		padding = 16
	}
	for i := 0; i < padding; i++ {
		in = append(in, byte(padding))
	}
	return in
}

func PKCS7Unpad(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}

	padding := in[len(in)-1]
	if int(padding) > len(in) || padding > aes.BlockSize {
		return nil
	} else if padding == 0 {
		return nil
	}

	for i := len(in) - 1; i > len(in)-int(padding)-1; i-- {
		if in[i] != padding {
			return nil
		}
	}
	return in[:len(in)-int(padding)]
}

func aesGCMDecrypt(key []byte, cipherText []byte, nonce []byte) (plainText []byte, err error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, err
	}

	paddedPlainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	plainText = PKCS7Unpad(paddedPlainText)
	if plainText == nil {
		err = fmt.Errorf("Decryption failed: PKCS7Unpad failed after decryption")
	}
	return plainText, err
}
