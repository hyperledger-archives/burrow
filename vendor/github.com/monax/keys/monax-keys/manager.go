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

// NOTE: this file is copied from github.com/ethereum/go-ethereum
// and repurposed to suit the monax-keys daemon.
// We really only wanted the TimeoutUnlock feature

import (
	"log"
	"sync"
	"time"

	"github.com/monax/keys/crypto"
)

type Manager struct {
	keyStore crypto.KeyStore
	unlocked map[string]*unlocked
	mutex    sync.RWMutex
}

type unlocked struct {
	*crypto.Key
	abort chan struct{}
}

func NewManager(keyStore crypto.KeyStore) *Manager {
	return &Manager{
		keyStore: keyStore,
		unlocked: make(map[string]*unlocked),
	}
}

func (am *Manager) KeyStore() crypto.KeyStore {
	return am.keyStore
}

// GetKey only returns unlocked keys
func (am *Manager) GetKey(addr []byte) *crypto.Key {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	u, ok := am.unlocked[string(addr)]
	if !ok {
		return nil
	}
	return u.Key
}

// Unlock unlocks the given account indefinitely.
func (am *Manager) Unlock(addr []byte, keyAuth string) error {
	return am.TimedUnlock(addr, keyAuth, 0)
}

// TimedUnlock unlocks the account with the given address. The account
// stays unlocked for the duration of timeout. A timeout of 0 unlocks the account
// until the program exits.
//
// If the accout is already unlocked, TimedUnlock extends or shortens
// the active unlock timeout.
func (am *Manager) TimedUnlock(addr []byte, keyAuth string, timeout time.Duration) error {
	key, err := am.keyStore.GetKey(addr, keyAuth)
	if err != nil {
		return err
	}
	var u *unlocked
	am.mutex.Lock()
	defer am.mutex.Unlock()
	var found bool
	u, found = am.unlocked[string(addr)]
	if found {
		// terminate dropLater for this key to avoid unexpected drops.
		if u.abort != nil {
			close(u.abort)
		}
	}
	log.Printf("Unlocking key %X for %v\n", addr, timeout)
	if timeout > 0 {
		u = &unlocked{Key: key, abort: make(chan struct{})}
		go am.expire(addr, u, timeout)
	} else {
		u = &unlocked{Key: key}
	}
	am.unlocked[string(addr)] = u
	return nil
}

func (am *Manager) expire(addr []byte, u *unlocked, timeout time.Duration) {
	t := time.NewTimer(timeout)
	defer t.Stop()
	select {
	case <-u.abort:
		// just quit
	case <-t.C:
		am.mutex.Lock()

		log.Printf("Relocking %X\n", addr)
		// only drop if it's still the same key instance that dropLater
		// was launched with. we can check that using pointer equality
		// because the map stores a new pointer every time the key is
		// unlocked.
		if am.unlocked[string(addr)] == u {
			zeroKey(u.PrivateKey)
			delete(am.unlocked, string(addr))
		}
		am.mutex.Unlock()
	}
}

// zeroKey zeroes a private key in memory.
// TODO: is this good enough?
func zeroKey(k []byte) {
	for i := range k {
		k[i] = 0
	}
}

func (am *Manager) Update(addr []byte, authFrom, authTo string) (err error) {
	var key *crypto.Key
	key, err = am.keyStore.GetKey(addr, authFrom)

	if err == nil {
		if err = am.keyStore.StoreKey(key, authTo); err == nil {
			// TODO
			// am.keyStore.Cleanup(addr)
		}
	}
	return
}
