package pipe

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/account"
	cmn "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/common"
	cs "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/consensus"
	mempl "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/mempool"
	"sync"
)

// The accounts struct has methods for working with accounts.
type accounts struct {
	consensusState *cs.ConsensusState
	mempoolReactor *mempl.MempoolReactor
	filterFactory  *FilterFactory
}

func newAccounts(consensusState *cs.ConsensusState, mempoolReactor *mempl.MempoolReactor) *accounts {
	ff := NewFilterFactory()

	ff.RegisterFilterPool("code", &sync.Pool{
		New: func() interface{} {
			return &AccountCodeFilter{}
		},
	})

	ff.RegisterFilterPool("balance", &sync.Pool{
		New: func() interface{} {
			return &AccountBalanceFilter{}
		},
	})

	return &accounts{consensusState, mempoolReactor, ff}

}

// Generate a new Private Key Account.
func (this *accounts) GenPrivAccount() (*account.PrivAccount, error) {
	pa := account.GenPrivAccount()
	return pa, nil
}

// Generate a new Private Key Account.
func (this *accounts) GenPrivAccountFromKey(privKey []byte) (*account.PrivAccount, error) {
	if len(privKey) != 64 {
		return nil, fmt.Errorf("Private key is not 64 bytes long.")
	}
	pa := account.GenPrivAccountFromPrivKeyBytes(privKey)
	return pa, nil
}

// Get all accounts.
func (this *accounts) Accounts(fda []*FilterData) (*AccountList, error) {
	accounts := make([]*account.Account, 0)
	state := this.consensusState.GetState()
	filter, err := this.filterFactory.NewFilter(fda)
	if err != nil {
		return nil, fmt.Errorf("Error in query: " + err.Error())
	}
	state.GetAccounts().Iterate(func(key interface{}, value interface{}) bool {
		acc := value.(*account.Account)
		if filter.Match(acc) {
			accounts = append(accounts, acc)
		}
		return false
	})
	return &AccountList{accounts}, nil
}

// Get an account.
func (this *accounts) Account(address []byte) (*account.Account, error) {
	cache := this.mempoolReactor.Mempool.GetCache()
	acc := cache.GetAccount(address)
	if acc == nil {
		acc = this.newAcc(address)
	}
	return acc, nil
}

// Get the value stored at 'key' in the account with address 'address'
// Both the key and value is returned.
func (this *accounts) StorageAt(address, key []byte) (*StorageItem, error) {

	state := this.consensusState.GetState()
	account := state.GetAccount(address)
	if account == nil {
		return &StorageItem{key, []byte{}}, nil
	}
	storageRoot := account.StorageRoot
	storageTree := state.LoadStorage(storageRoot)

	_, value := storageTree.Get(cmn.LeftPadWord256(key).Bytes())
	if value == nil {
		return &StorageItem{key, []byte{}}, nil
	}
	return &StorageItem{key, value.([]byte)}, nil
}

// Get the storage of the account with address 'address'.
func (this *accounts) Storage(address []byte) (*Storage, error) {

	state := this.consensusState.GetState()
	account := state.GetAccount(address)
	storageItems := make([]StorageItem, 0)
	if account == nil {
		return &Storage{nil, storageItems}, nil
	}
	storageRoot := account.StorageRoot
	storageTree := state.LoadStorage(storageRoot)

	storageTree.Iterate(func(key interface{}, value interface{}) bool {
		storageItems = append(storageItems, StorageItem{
			key.([]byte), value.([]byte)})
		return false
	})
	return &Storage{storageRoot, storageItems}, nil
}

// Create a new account.
func (this *accounts) newAcc(address []byte) *account.Account {
	return &account.Account{
		Address:     address,
		PubKey:      nil,
		Sequence:    0,
		Balance:     0,
		Code:        nil,
		StorageRoot: nil,
	}
}

// Filter for account code.
// Ops: == or !=
// Could be used to match against nil, to see if an account is a contract account.
type AccountCodeFilter struct {
	op    string
	value []byte
	match func([]byte, []byte) bool
}

func (this *AccountCodeFilter) Configure(fd *FilterData) error {
	op := fd.Op
	val, err := hex.DecodeString(fd.Value)

	if err != nil {
		return fmt.Errorf("Wrong value type.")
	}
	if op == "==" {
		this.match = func(a, b []byte) bool {
			return bytes.Equal(a, b)
		}
	} else if op == "!=" {
		this.match = func(a, b []byte) bool {
			return !bytes.Equal(a, b)
		}
	} else {
		return fmt.Errorf("Op: " + this.op + " is not supported for 'code' filtering")
	}
	this.op = op
	this.value = val
	return nil
}

func (this *AccountCodeFilter) Match(v interface{}) bool {
	acc, ok := v.(*account.Account)
	if !ok {
		return false
	}
	return this.match(acc.Code, this.value)
}

// Filter for account balance.
// Ops: All
type AccountBalanceFilter struct {
	op    string
	value int64
	match func(int64, int64) bool
}

func (this *AccountBalanceFilter) Configure(fd *FilterData) error {
	val, err := ParseNumberValue(fd.Value)
	if err != nil {
		return err
	}
	match, err2 := GetRangeFilter(fd.Op, "balance")
	if err2 != nil {
		return err2
	}
	this.match = match
	this.op = fd.Op
	this.value = val
	return nil
}

func (this *AccountBalanceFilter) Match(v interface{}) bool {
	acc, ok := v.(*account.Account)
	if !ok {
		return false
	}
	return this.match(int64(acc.Balance), this.value)
}
