package state

import (
	"testing"

	"fmt"

	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/core/types"
	"github.com/hyperledger/burrow/genesis"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-db"
)

func TestBlockCache_Sync_Accounts(t *testing.T) {
	blockCache := NewBlockCache(stateForBlockCache(t, 10))
	blockCache.RemoveAccount(accountAddress(1))
	blockCache.Sync()
	// Expect account cache objects to be reaped so that an attempt is not made
	// to remove them twice from merkle tree
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Consecutive calls to BlockCache.Sync() failed. " +
					"Is cache dirty?\n Error: %s", r)
		}
	}()
	blockCache.Sync()
}

func TestBlockCache_Sync_NameReg(t *testing.T) {
	blockCache := NewBlockCache(stateForBlockCache(t, 10))
	blockCache.RemoveAccount(accountAddress(1))
	nameRegName := "foobs"
	nameRegEntry := &types.NameRegEntry{
		Name:    nameRegName,
		Owner:   accountAddress(1),
		Data:    "Dates",
		Expires: 434,
	}
	blockCache.UpdateNameRegEntry(nameRegEntry)
	blockCache.Sync()
	blockCache.RemoveNameRegEntry(nameRegName)
	// Expect name reg cache objects to be reaped so that an attempt is not made
	// to remove them twice from merkle tree
	blockCache.Sync()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Consecutive calls to BlockCache.Sync() failed. " +
					"Is cache dirty?\n Error: %s", r)
		}
	}()
	blockCache.Sync()
}

func stateForBlockCache(t *testing.T, numAccounts byte) *State {
	genAccounts := make([]*genesis.GenesisAccount, numAccounts)
	genValidators := make([]*genesis.GenesisValidator, 1)
	for i := byte(0); i < numAccounts; i++ {
		genAccounts[i] = genesisAccount("account", i)
	}
	genValidators[0] = genesisValidator("validator", 0)
	genDoc, err := genesis.MakeGenesisDocFromAccounts("BlockChainTest",
		genAccounts, genValidators)
	assert.NoError(t, err)
	return MakeGenesisState(db.NewMemDB(), &genDoc)
}

func genesisAccount(prefix string, index byte) *genesis.GenesisAccount {
	address := accountAddress(index)
	return &genesis.GenesisAccount{
		Address:     address,
		Amount:      10000 + int64(index),
		Name:        accountName(prefix, index, address),
		Permissions: &ptypes.DefaultAccountPermissions,
	}
}

func genesisValidator(prefix string, index byte) *genesis.GenesisValidator {
	privAccount := account.GenPrivAccountFromSecret(
		fmt.Sprintf("%s_%v", prefix, index))
	return &genesis.GenesisValidator{
		PubKey: privAccount.PubKey,
		Amount: 1000000 + int64(index),
		Name:   accountName(prefix, index, privAccount.Address),
		UnbondTo: []genesis.BasicAccount{
			{
				Address: privAccount.Address,
				Amount:  100,
			},
		},
	}
}

func accountName(prefix string, index byte, address []byte) string {
	return fmt.Sprintf("%s-%v_%X", prefix, index, address)
}

func accountAddress(index byte) []byte {
	var address [20]byte
	address[19] = index
	return address[:]
}
