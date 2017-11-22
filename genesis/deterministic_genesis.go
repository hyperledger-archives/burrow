package genesis

import (
	"math/rand"
	"time"

	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/permission"
	"github.com/tendermint/ed25519"
	"github.com/tendermint/go-crypto"
)

type deterministicGenesis struct {
	random *rand.Rand
}

// Generates deterministic pseudo-random genesis state
func NewDeterministicGenesis(seed int64) *deterministicGenesis {
	return &deterministicGenesis{
		random: rand.New(rand.NewSource(31230398587433)),
	}
}

func (dg *deterministicGenesis) GenesisDoc(numAccounts int, randBalance bool, minBalance uint64, numValidators int,
	randBonded bool, minBonded int64) (*GenesisDoc, []acm.PrivateAccount) {

	accounts := make([]Account, numAccounts)
	privAccounts := make([]acm.PrivateAccount, numAccounts)
	defaultPerms := permission.DefaultAccountPermissions
	for i := 0; i < numAccounts; i++ {
		account, privAccount := dg.Account(randBalance, minBalance)
		accounts[i] = Account{
			BasicAccount: BasicAccount{
				Address: account.Address(),
				Amount:  account.Balance(),
			},
			Permissions: defaultPerms.Clone(), // This will get copied into each state.Account.
		}
		privAccounts[i] = privAccount
	}
	validators := make([]Validator, numValidators)
	for i := 0; i < numValidators; i++ {
		validator := acm.GeneratePrivateAccountFromSecret(fmt.Sprintf("val_%v", i))
		validators[i] = Validator{
			BasicAccount: BasicAccount{
				Address:   validator.Address(),
				PublicKey: validator.PublicKey(),
				Amount:    uint64(dg.random.Int63()),
			},
			UnbondTo: []BasicAccount{
				{
					Address: validator.Address(),
					Amount:  uint64(dg.random.Int63()),
				},
			},
		}
	}
	return &GenesisDoc{
		ChainName:   "TestChain",
		GenesisTime: time.Unix(1506172037, 0),
		Accounts:    accounts,
		Validators:  validators,
	}, privAccounts

}

func (dg *deterministicGenesis) Account(randBalance bool, minBalance uint64) (acm.Account, acm.PrivateAccount) {
	privKey := dg.PrivateKey()
	pubKey := acm.PublicKeyFromPubKey(privKey.PubKey())
	privAccount := &acm.ConcretePrivateAccount{
		PublicKey:  pubKey,
		PrivateKey: privKey,
		Address:    acm.MustAddressFromBytes(pubKey.Address()),
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

func (dg *deterministicGenesis) PrivateKey() acm.PrivateKey {
	privKeyBytes := new([64]byte)
	for i := 0; i < 32; i++ {
		privKeyBytes[i] = byte(dg.random.Int() % 256)
	}
	ed25519.MakePublicKey(privKeyBytes)
	return acm.PrivateKeyFromPrivKey(crypto.PrivKeyEd25519(*privKeyBytes).Wrap())
}
