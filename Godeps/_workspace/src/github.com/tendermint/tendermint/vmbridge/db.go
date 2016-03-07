package vmbridge

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethvm "github.com/ethereum/go-ethereum/core/vm"

	mintcommon "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/common"
	mintvm "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/vm"
)

// Implements github.com/ethereum/go-ethereum/core/vm.Database
type Database struct {
	appState mintvm.AppState

	refund *big.Int
}

func NewDatabase(appState mintvm.AppState) *Database {
	return &Database{
		appState: appState,
	}
}

func addrToWord(addr common.Address) mintcommon.Word256 {
	return mintcommon.LeftPadWord256(addr.Bytes())
}

func hashToWord(hash common.Hash) mintcommon.Word256 {
	return mintcommon.Word256(hash)
}

func (db *Database) getAccount(addr common.Address) *mintvm.Account {
	return db.appState.GetAccount(addrToWord(addr))
}

func (db *Database) GetAccount(addr common.Address) ethvm.Account {
	return NewAccount(db.getAccount(addr))
}

func (db *Database) CreateAccount(addr common.Address) ethvm.Account {
	// XXX: this should be someone else's responsibility and not called
	// see Create
	return nil
}

func (db *Database) AddBalance(addr common.Address, amount *big.Int) {
	acc := db.getAccount(addr)
	acc.Balance += amount.Int64() // XXX: overflow!
	db.appState.UpdateAccount(acc)
}

func (db *Database) GetBalance(addr common.Address) *big.Int {
	acc := db.getAccount(addr)
	return big.NewInt(acc.Balance)
}

func (db *Database) GetNonce(addr common.Address) uint64 {
	acc := db.getAccount(addr)
	return uint64(acc.Nonce)
}

func (db *Database) SetNonce(addr common.Address, nonce uint64) {
	acc := db.getAccount(addr)
	acc.Nonce = int64(nonce) // XXX: overflow
	db.appState.UpdateAccount(acc)
}

func (db *Database) GetCode(addr common.Address) []byte {
	acc := db.getAccount(addr)
	return acc.Code
}

func (db *Database) SetCode(addr common.Address, code []byte) {
	acc := db.getAccount(addr)
	acc.Code = code
	db.appState.UpdateAccount(acc)
}

// TODO: is there more to these refunds than these two funcs?
func (db *Database) AddRefund(refund *big.Int) {
	db.refund.Add(db.refund, refund)
}

func (db *Database) GetRefund() *big.Int {
	return db.refund
}

// Get a value out of a contract's storage
func (db *Database) GetState(addr common.Address, key common.Hash) common.Hash {
	return common.Hash(db.appState.GetStorage(addrToWord(addr), hashToWord(key)))
}

// Set a value in a contract's storage
func (db *Database) SetState(addr common.Address, key, value common.Hash) {
	db.appState.SetStorage(addrToWord(addr), hashToWord(key), hashToWord(value))
}

func (db *Database) Delete(addr common.Address) bool {
	acc := db.getAccount(addr)
	db.appState.RemoveAccount(acc)
	return true // XXX !
}

func (db *Database) Exist(addr common.Address) bool {
	acc := db.getAccount(addr)
	if acc != nil {
		return true
	}
	return false
}

func (db *Database) IsDeleted(addr common.Address) bool {
	acc := db.getAccount(addr)
	// NOTE: no way to tell between non-existant and removed without introspecting further!
	if acc == nil {
		return true
	}
	return false
}
