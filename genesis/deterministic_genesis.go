package genesis

import (
	"fmt"
	"math/rand"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/permission"
)

type deterministicGenesis struct {
	random *rand.Rand
}

// Generates deterministic pseudo-random genesis state
func NewDeterministicGenesis(seed int64) *deterministicGenesis {
	return &deterministicGenesis{
		random: rand.New(rand.NewSource(seed)),
	}
}

func (dg *deterministicGenesis) GenesisDoc(numAccounts int, randBalance bool, minBalance uint64, numValidators int,
	randBonded bool, minBonded int64) (*GenesisDoc, []acm.AddressableSigner, []acm.AddressableSigner) {

	accounts := make([]Account, numAccounts)
	privAccounts := make([]acm.AddressableSigner, numAccounts)
	defaultPerms := permission.DefaultAccountPermissions
	for i := 0; i < numAccounts; i++ {
		account, privAccount := dg.Account(randBalance, minBalance)
		accounts[i] = Account{
			BasicAccount: BasicAccount{
				Address:   account.Address(),
				PublicKey: account.PublicKey(),
				Amount:    account.Balance(),
			},
			Permissions: defaultPerms.Clone(), // This will get copied into each state.Account.
		}
		privAccounts[i] = privAccount
	}
	validators := make([]Validator, numValidators)
	privValidators := make([]acm.AddressableSigner, numValidators)
	for i := 0; i < numValidators; i++ {
		validator := acm.GeneratePrivateAccountFromSecret(fmt.Sprintf("val_%v", i))
		privValidators[i] = validator
		validators[i] = Validator{
			BasicAccount: BasicAccount{
				Address:   validator.Address(),
				PublicKey: validator.PublicKey(),
				Amount:    uint64(dg.random.Int63()),
			},
			UnbondTo: []BasicAccount{
				{
					Address:   validator.Address(),
					PublicKey: validator.PublicKey(),
					Amount:    uint64(dg.random.Int63()),
				},
			},
		}
	}
	return &GenesisDoc{
		ChainName:   "TestChain",
		GenesisTime: time.Unix(1506172037, 0),
		Accounts:    accounts,
		Validators:  validators,
	}, privAccounts, privValidators

}

func (dg *deterministicGenesis) Account(randBalance bool, minBalance uint64) (*acm.Account, acm.AddressableSigner) {
	privateKey, err := crypto.GeneratePrivateKey(dg.random, crypto.CurveTypeEd25519)
	if err != nil {
		panic(fmt.Errorf("could not generate private key deterministically"))
	}
	privAccount := &acm.ConcretePrivateAccount{
		PublicKey:  privateKey.GetPublicKey(),
		PrivateKey: privateKey,
		Address:    privateKey.GetPublicKey().Address(),
	}
	perms := permission.DefaultAccountPermissions
	balance := minBalance
	if randBalance {
		balance += uint64(dg.random.Int())
	}
	acc := acm.NewAccount(privAccount.PublicKey, perms)
	acc.AddToBalance(balance)

	return acc, privAccount.PrivateAccount()
}
