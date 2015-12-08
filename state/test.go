package state

import (
	"sort"
	"time"

	. "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/go-common"
	dbm "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/go-db"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
	acm "github.com/eris-ltd/eris-db/account"
	ptypes "github.com/eris-ltd/eris-db/permission/types"
	. "github.com/eris-ltd/eris-db/state/types"
)

func RandAccount(randBalance bool, minBalance int64) (*acm.Account, *acm.PrivAccount) {
	privAccount := acm.GenPrivAccount()
	perms := ptypes.DefaultAccountPermissions
	acc := &acm.Account{
		Address:     privAccount.PubKey.Address(),
		PubKey:      privAccount.PubKey,
		Sequence:    RandInt(),
		Balance:     minBalance,
		Permissions: perms,
	}
	if randBalance {
		acc.Balance += int64(RandUint32())
	}
	return acc, privAccount
}

func RandGenesisDoc(numAccounts int, randBalance bool, minBalance int64, numValidators int, randBonded bool, minBonded int64) (*GenesisDoc, []*acm.PrivAccount, []*types.PrivValidator) {
	accounts := make([]GenesisAccount, numAccounts)
	privAccounts := make([]*acm.PrivAccount, numAccounts)
	defaultPerms := ptypes.DefaultAccountPermissions
	for i := 0; i < numAccounts; i++ {
		account, privAccount := RandAccount(randBalance, minBalance)
		accounts[i] = GenesisAccount{
			Address:     account.Address,
			Amount:      account.Balance,
			Permissions: &defaultPerms, // This will get copied into each state.Account.
		}
		privAccounts[i] = privAccount
	}
	validators := make([]GenesisValidator, numValidators)
	privValidators := make([]*types.PrivValidator, numValidators)
	for i := 0; i < numValidators; i++ {
		val, privVal := types.RandValidator(randBonded, minBonded)
		validators[i] = GenesisValidator{
			PubKey: val.PubKey,
			Amount: val.VotingPower,
			UnbondTo: []BasicAccount{
				{
					Address: val.PubKey.Address(),
					Amount:  val.VotingPower,
				},
			},
		}
		privValidators[i] = privVal
	}
	sort.Sort(types.PrivValidatorsByAddress(privValidators))
	return &GenesisDoc{
		GenesisTime: time.Now(),
		ChainID:     "tendermint_test",
		Accounts:    accounts,
		Validators:  validators,
	}, privAccounts, privValidators

}

func RandGenesisState(numAccounts int, randBalance bool, minBalance int64, numValidators int, randBonded bool, minBonded int64) (*State, []*acm.PrivAccount, []*types.PrivValidator) {
	db := dbm.NewMemDB()
	genDoc, privAccounts, privValidators := RandGenesisDoc(numAccounts, randBalance, minBalance, numValidators, randBonded, minBonded)
	s0 := MakeGenesisState(db, genDoc)
	s0.Save()
	return s0, privAccounts, privValidators
}
