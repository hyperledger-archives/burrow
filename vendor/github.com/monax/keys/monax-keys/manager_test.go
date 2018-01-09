// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package keys

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/monax/keys/crypto"
)

var (
	testSigData = "1223344556677889"
	keyType     = "ed25519,ripemd160"
)

func TestTimedUnlock(t *testing.T) {
	_, ks := tmpKeyStore(t, crypto.NewKeyStorePassphrase)
	//	defer os.RemoveAll(dir)

	AccountManager = NewManager(ks)
	am := AccountManager
	pass := "foo"
	addr, err := coreKeygen(pass, keyType)
	if err != nil {
		t.Fatal(err)
	}
	addrHex := hex.EncodeToString(addr)

	// Signing without passphrase fails because account is locked
	_, err = coreSign(testSigData, addrHex)
	if err != ErrLocked {
		t.Fatal("Signing should've failed with ErrLocked before unlocking, got ", err)
	}

	// Signing with passphrase works
	if err = am.TimedUnlock(addr, pass, 100*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	// Signing without passphrase works because account is temp unlocked
	_, err = coreSign(testSigData, addrHex)
	if err != nil {
		t.Fatal("Signing shouldn't return an error after unlocking, got ", err)
	}

	// Signing fails again after automatic locking
	time.Sleep(150 * time.Millisecond)
	_, err = coreSign(testSigData, addrHex)
	if err != ErrLocked {
		t.Fatal("Signing should've failed with ErrLocked timeout expired, got ", err)
	}
}

func TestOverrideUnlock(t *testing.T) {
	_, ks := tmpKeyStore(t, crypto.NewKeyStorePassphrase)
	//defer os.RemoveAll(dir)

	AccountManager = NewManager(ks)
	am := AccountManager
	pass := "foo"
	addr, err := coreKeygen(pass, keyType)
	if err != nil {
		t.Fatal(err)
	}
	addrHex := hex.EncodeToString(addr)

	// Unlock indefinitely
	if err = am.Unlock(addr, pass); err != nil {
		t.Fatal(err)
	}

	// Signing without passphrase works because account is temp unlocked
	_, err = coreSign(testSigData, addrHex)
	if err != nil {
		t.Fatal("Signing shouldn't return an error after unlocking, got ", err)
	}

	// reset unlock to a shorter period, invalidates the previous unlock
	if err = am.TimedUnlock(addr, pass, 100*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	// Signing without passphrase still works because account is temp unlocked
	_, err = coreSign(testSigData, addrHex)
	if err != nil {
		t.Fatal("Signing shouldn't return an error after unlocking, got ", err)
	}

	// Signing fails again after automatic locking
	time.Sleep(150 * time.Millisecond)
	_, err = coreSign(testSigData, addrHex)
	if err != ErrLocked {
		t.Fatal("Signing should've failed with ErrLocked timeout expired, got ", err)
	}
}

// This test should fail under -race if signing races the expiration goroutine.
func TestSignRace(t *testing.T) {
	_, ks := tmpKeyStore(t, crypto.NewKeyStorePassphrase)
	//defer os.RemoveAll(dir)

	// Create a test account.
	am := NewManager(ks)
	pass := "foo"
	addr, err := coreKeygen(pass, keyType)
	if err != nil {
		t.Fatal("could not create the test account", err)
	}

	if err := am.TimedUnlock(addr, pass, 15*time.Millisecond); err != nil {
		t.Fatalf("could not unlock the test account: %v", err)
	}
	end := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(end) {
		if _, err := coreSign(testSigData, hex.EncodeToString(addr)); err == ErrLocked {
			return
		} else if err != nil {
			t.Errorf("Sign error: %v", err)
			return
		}
		time.Sleep(1 * time.Millisecond)
	}
	t.Errorf("Account did not lock within the timeout")
}

func tmpKeyStore(t *testing.T, new func(string) crypto.KeyStore) (string, crypto.KeyStore) {
	d, err := returnDataDir(KeysDir)
	if err != nil {
		t.Fatal(err)
	}
	return d, new(d)
}
