package genesis

import (
	"fmt"
	"math/rand"
	"time"

	acm "github.com/hyperledger/burrow/account"
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

	genAccounts := make([]genAccount, numAccounts)
	privAccounts := make([]acm.AddressableSigner, numAccounts)
	defaultPerms := permission.DefaultAccountPermissions
	for i := 0; i < numAccounts; i++ {
		account, privAccount := dg.Account(randBalance, minBalance)
		genAccounts[i] = genAccount{
			PublicKey:   account.PublicKey(),
			Address:     account.Address(),
			Amount:      account.Balance(),
			Permissions: defaultPerms.Clone(), // This will get copied into each state.Account.
		}
		privAccounts[i] = privAccount
	}
	genValidators := make([]genValidator, numValidators)
	privValidators := make([]acm.AddressableSigner, numValidators)
	for i := 0; i < numValidators; i++ {
		validator := acm.GeneratePrivateAccountFromSecret(fmt.Sprintf("val_%v", i))
		privValidators[i] = validator
		genValidators[i] = genValidator{
			PublicKey: validator.PublicKey(),
			Address:   validator.Address(),
			Stake:     uint64(dg.random.Int63()),
		}
	}
	return &GenesisDoc{
		ChainName:     "TestChain",
		GenesisTime:   time.Unix(1506172037, 0),
		GenAccounts:   genAccounts,
		GenValidators: genValidators,
		MaximumPower:  4,
	}, privAccounts, privValidators

}

func (dg *deterministicGenesis) Account(randBalance bool, minBalance uint64) (acm.Account, acm.AddressableSigner) {
	privateKey, err := acm.GeneratePrivateKey(dg.random)
	if err != nil {
		panic(fmt.Errorf("could not generate private key deterministically"))
	}
	privAccount := &acm.ConcretePrivateAccount{
		PublicKey:  privateKey.PublicKey(),
		PrivateKey: privateKey,
		Address:    privateKey.PublicKey().Address(),
	}
	perms := permission.DefaultAccountPermissions
	acc := &acm.ConcreteAccount{
		Address:     privAccount.Address,
		PublicKey:   privAccount.PublicKey,
		Sequence:    uint64(dg.random.Int()),
		Balance:     minBalance,
		Permissions: perms,
	}
	if randBalance {
		acc.Balance += uint64(dg.random.Int())
	}
	return acc.Account(), privAccount.PrivateAccount()
}
