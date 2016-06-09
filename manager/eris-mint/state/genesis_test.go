package state

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"testing"
	"time"

	acm "github.com/eris-ltd/eris-db/account"
	ptypes "github.com/eris-ltd/eris-db/permission/types"
	. "github.com/eris-ltd/eris-db/manager/eris-mint/state/types"

	. "github.com/tendermint/go-common"
	tdb "github.com/tendermint/go-db"
	"github.com/tendermint/tendermint/types"
)

var chain_id = "lone_ranger"
var addr1, _ = hex.DecodeString("964B1493BBE3312278B7DEB94C39149F7899A345")
var send1, name1, call1 = 1, 1, 0
var perms, setbit = 66, 70
var accName = "me"
var roles1 = []string{"master", "universal-ruler"}
var amt1 int64 = 1000000
var g1 = fmt.Sprintf(`
{
    "chain_id":"%s",
    "accounts": [
        {
            "address": "%X",
            "amount": %d,
	    "name": "%s",
            "permissions": {
		    "base": {
			    "perms": %d,
			    "set": %d
		    },
            	    "roles": [
			"%s",
			"%s"
            	]
	    }
        }
    ],
    "validators": [
        {
            "amount": 100000000,
            "pub_key": [1,"F6C79CF0CB9D66B677988BCB9B8EADD9A091CD465A60542A8AB85476256DBA92"],
            "unbond_to": [
                {
                    "address": "964B1493BBE3312278B7DEB94C39149F7899A345",
                    "amount": 10000000
                }
            ]
        }
    ]
}
`, chain_id, addr1, amt1, accName, perms, setbit, roles1[0], roles1[1])

func TestGenesisReadable(t *testing.T) {
	genDoc := GenesisDocFromJSON([]byte(g1))
	if genDoc.ChainID != chain_id {
		t.Fatalf("Incorrect chain id. Got %d, expected %d\n", genDoc.ChainID, chain_id)
	}
	acc := genDoc.Accounts[0]
	if bytes.Compare(acc.Address, addr1) != 0 {
		t.Fatalf("Incorrect address for account. Got %X, expected %X\n", acc.Address, addr1)
	}
	if acc.Amount != amt1 {
		t.Fatalf("Incorrect amount for account. Got %d, expected %d\n", acc.Amount, amt1)
	}
	if acc.Name != accName {
		t.Fatalf("Incorrect name for account. Got %s, expected %s\n", acc.Name, accName)
	}

	perm, _ := acc.Permissions.Base.Get(ptypes.Send)
	if perm != (send1 > 0) {
		t.Fatalf("Incorrect permission for send. Got %v, expected %v\n", perm, send1 > 0)
	}
}

func TestGenesisMakeState(t *testing.T) {
	genDoc := GenesisDocFromJSON([]byte(g1))
	db := tdb.NewMemDB()
	st := MakeGenesisState(db, genDoc)
	acc := st.GetAccount(addr1)
	v, _ := acc.Permissions.Base.Get(ptypes.Send)
	if v != (send1 > 0) {
		t.Fatalf("Incorrect permission for send. Got %v, expected %v\n", v, send1 > 0)
	}
}

//-------------------------------------------------------

func RandGenesisState(numAccounts int, randBalance bool, minBalance int64, numValidators int, randBonded bool, minBonded int64) (*State, []*acm.PrivAccount, []*types.PrivValidator) {
	db := tdb.NewMemDB()
	genDoc, privAccounts, privValidators := RandGenesisDoc(numAccounts, randBalance, minBalance, numValidators, randBonded, minBonded)
	s0 := MakeGenesisState(db, genDoc)
	s0.Save()
	return s0, privAccounts, privValidators
}

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
		valInfo, privVal := types.RandValidator(randBonded, minBonded)
		validators[i] = GenesisValidator{
			PubKey: valInfo.PubKey,
			Amount: valInfo.VotingPower,
			UnbondTo: []BasicAccount{
				{
					Address: valInfo.PubKey.Address(),
					Amount:  valInfo.VotingPower,
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
