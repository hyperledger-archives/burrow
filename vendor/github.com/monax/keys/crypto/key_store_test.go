package crypto

import (
	"github.com/monax/keys/common"

	"reflect"
	"testing"
)

func TestKeyStorePlain(t *testing.T) {
	ks := NewKeyStorePlain(common.KeysPath)
	f := func(typ KeyType) {
		pass := "" // not used but required by API
		k1, err := ks.GenerateNewKey(typ, pass)
		if err != nil {
			t.Fatal(err)
		}

		k2 := new(Key)
		k2, err = ks.GetKey(k1.Address, pass)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(k1.Address, k2.Address) {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(k1.PrivateKey, k2.PrivateKey) {
			t.Fatal(err)
		}

		err = ks.DeleteKey(k2.Address, pass)
		if err != nil {
			t.Fatal(err)
		}
	}
	f(KeyType{CurveTypeSecp256k1, AddrTypeSha3})
	f(KeyType{CurveTypeEd25519, AddrTypeRipemd160})
}

func TestKeyStorePassphrase(t *testing.T) {
	ks := NewKeyStorePassphrase(common.KeysPath)
	f := func(typ KeyType) {
		pass := "foo"
		k1, err := ks.GenerateNewKey(typ, pass)
		if err != nil {
			t.Fatal(err)
		}
		k2 := new(Key)
		k2, err = ks.GetKey(k1.Address, pass)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(k1.Address, k2.Address) {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(k1.PrivateKey, k2.PrivateKey) {
			t.Fatal(err)
		}

		err = ks.DeleteKey(k2.Address, pass) // also to clean up created files
		if err != nil {
			t.Fatal(err)
		}
	}
	f(KeyType{CurveTypeSecp256k1, AddrTypeSha3})
	f(KeyType{CurveTypeEd25519, AddrTypeRipemd160})
}

func TestKeyStorePassphraseDecryptionFail(t *testing.T) {
	ks := NewKeyStorePassphrase(common.KeysPath)
	pass := "foo"
	f := func(typ KeyType) {
		k1, err := ks.GenerateNewKey(KeyType{CurveTypeSecp256k1, AddrTypeSha3}, pass)
		if err != nil {
			t.Fatal(err)
		}

		_, err = ks.GetKey(k1.Address, "bar") // wrong passphrase
		if err == nil {
			t.Fatal(err)
		}

		err = ks.DeleteKey(k1.Address, "bar") // wrong passphrase
		if err == nil {
			t.Fatal(err)
		}

		err = ks.DeleteKey(k1.Address, pass) // to clean up
		if err != nil {
			t.Fatal(err)
		}
	}
	f(KeyType{CurveTypeSecp256k1, AddrTypeSha3})
	f(KeyType{CurveTypeEd25519, AddrTypeRipemd160})
}
