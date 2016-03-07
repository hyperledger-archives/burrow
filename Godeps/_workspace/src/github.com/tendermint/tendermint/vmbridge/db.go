package vmbridge

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethvm "github.com/ethereum/go-ethereum/core/vm"

	mintvm "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/vm"
)

// Implements github.com/ethereum/go-ethereum/core/vm.Database
type Database struct {
	appState mintvm.AppState
}

func NewDatabase(appState mintvm.AppState) *Database {
	return &Database{
		appState: appState,
	}
}

func (db *Database) GetAccount(common.Address) ethvm.Account {
	return nil
}

func (db *Database) CreateAccount(common.Address) ethvm.Account {
	return nil
}
func (db *Database) AddBalance(common.Address, *big.Int) {
}
func (db *Database) GetBalance(common.Address) *big.Int {
	return nil
}
func (db *Database) GetNonce(common.Address) uint64 {
	return 0
}
func (db *Database) SetNonce(common.Address, uint64) {
}
func (db *Database) GetCode(common.Address) []byte {
	return nil
}
func (db *Database) SetCode(common.Address, []byte) {
}
func (db *Database) AddRefund(*big.Int) {
}
func (db *Database) GetRefund() *big.Int {
	return nil
}
func (db *Database) GetState(common.Address, common.Hash) common.Hash {
	return common.Hash{}
}

func (db *Database) SetState(common.Address, common.Hash, common.Hash) {
}
func (db *Database) Delete(common.Address) bool {
	return false
}
func (db *Database) Exist(common.Address) bool {
	return false
}
func (db *Database) IsDeleted(common.Address) bool {
	return false
}
