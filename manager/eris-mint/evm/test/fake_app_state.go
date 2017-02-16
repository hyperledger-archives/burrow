package vm

import (
	"fmt"

	. "github.com/eris-ltd/eris-db/manager/eris-mint/evm"
	"github.com/eris-ltd/eris-db/manager/eris-mint/evm/sha3"
	. "github.com/eris-ltd/eris-db/word256"
)

type FakeAppState struct {
	accounts map[string]*Account
	storage  map[string]Word256
}

func (fas *FakeAppState) GetAccount(addr Word256) *Account {
	account := fas.accounts[addr.String()]
	return account
}

func (fas *FakeAppState) UpdateAccount(account *Account) {
	fas.accounts[account.Address.String()] = account
}

func (fas *FakeAppState) RemoveAccount(account *Account) {
	_, ok := fas.accounts[account.Address.String()]
	if !ok {
		panic(fmt.Sprintf("Invalid account addr: %X", account.Address))
	} else {
		// Remove account
		delete(fas.accounts, account.Address.String())
	}
}

func (fas *FakeAppState) CreateAccount(creator *Account) *Account {
	addr := createAddress(creator)
	account := fas.accounts[addr.String()]
	if account == nil {
		return &Account{
			Address: addr,
			Balance: 0,
			Code:    nil,
			Nonce:   0,
		}
	} else {
		panic(fmt.Sprintf("Invalid account addr: %X", addr))
	}
}

func (fas *FakeAppState) GetStorage(addr Word256, key Word256) Word256 {
	_, ok := fas.accounts[addr.String()]
	if !ok {
		panic(fmt.Sprintf("Invalid account addr: %X", addr))
	}

	value, ok := fas.storage[addr.String()+key.String()]
	if ok {
		return value
	} else {
		return Zero256
	}
}

func (fas *FakeAppState) SetStorage(addr Word256, key Word256, value Word256) {
	_, ok := fas.accounts[addr.String()]
	if !ok {
		panic(fmt.Sprintf("Invalid account addr: %X", addr))
	}

	fas.storage[addr.String()+key.String()] = value
}

// Creates a 20 byte address and bumps the nonce.
func createAddress(creator *Account) Word256 {
	nonce := creator.Nonce
	creator.Nonce += 1
	temp := make([]byte, 32+8)
	copy(temp, creator.Address[:])
	PutInt64BE(temp[32:], nonce)
	return LeftPadWord256(sha3.Sha3(temp)[:20])
}
