package genesis

import (
	"math/rand"
	"sort"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/permission"
	"github.com/tendermint/ed25519"
	"github.com/tendermint/go-crypto"
	tm_types "github.com/tendermint/tendermint/types"
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
	randBonded bool, minBonded int64) (*GenesisDoc, []acm.PrivateAccount, []*tm_types.PrivValidatorFS) {

	accounts := make([]GenesisAccount, numAccounts)
	privAccounts := make([]acm.PrivateAccount, numAccounts)
	defaultPerms := permission.DefaultAccountPermissions
	for i := 0; i < numAccounts; i++ {
		account, privAccount := dg.Account(randBalance, minBalance)
		accounts[i] = GenesisAccount{
			BasicAccount: BasicAccount{
				Address: account.Address(),
				Amount:  account.Balance(),
			},
			Permissions: defaultPerms.Clone(), // This will get copied into each state.Account.
		}
		privAccounts[i] = privAccount
	}
	validators := make([]GenesisValidator, numValidators)
	privValidators := make([]*tm_types.PrivValidatorFS, numValidators)
	for i := 0; i < numValidators; i++ {
		valInfo, privVal := dg.Validator(randBonded, minBonded)
		validators[i] = GenesisValidator{
			PubKey: valInfo.PubKey,
			Amount: uint64(valInfo.VotingPower),
			UnbondTo: []BasicAccount{
				{
					Address: acm.MustAddressFromBytes(valInfo.PubKey.Address()),
					Amount:  uint64(valInfo.VotingPower),
				},
			},
		}
		privValidators[i] = privVal
	}
	sort.Sort(tm_types.PrivValidatorsByAddress(privValidators))
	return &GenesisDoc{
		ChainName:   "TestChain",
		GenesisTime: time.Unix(1506172037, 0),
		Accounts:    accounts,
		Validators:  validators,
	}, privAccounts, privValidators

}

func (dg *deterministicGenesis) Account(randBalance bool, minBalance uint64) (acm.Account, acm.PrivateAccount) {
	privKey := dg.PrivKey()
	pubKey := privKey.PubKey()
	privAccount := &acm.ConcretePrivateAccount{
		PubKey:  pubKey,
		PrivKey: privKey.Wrap(),
		Address: acm.MustAddressFromBytes(pubKey.Address()),
	}
	perms := permission.DefaultAccountPermissions
	acc := &acm.ConcreteAccount{
		Address:     privAccount.Address,
		PubKey:      privAccount.PubKey,
		Sequence:    uint64(dg.random.Int()),
		Balance:     minBalance,
		Permissions: perms,
	}
	if randBalance {
		acc.Balance += uint64(dg.random.Int())
	}
	return acc.Account(), privAccount.PrivateAccount()
}

func (dg *deterministicGenesis) Validator(randPower bool, minPower int64) (*tm_types.Validator, *tm_types.PrivValidatorFS) {
	privKey := dg.PrivKey()
	pubKey := privKey.PubKey()
	privVal := &tm_types.PrivValidatorFS{
		Address: pubKey.Address(),
		PubKey:  pubKey,
		PrivKey: privKey.Wrap(),
		Signer:  tm_types.NewDefaultSigner(privKey.Wrap()),
	}
	votePower := minPower
	if randPower {
		votePower += dg.random.Int63()
	}
	val := tm_types.NewValidator(privVal.PubKey, votePower)
	return val, privVal
}

func (dg *deterministicGenesis) PrivKey() crypto.PrivKeyEd25519 {
	privKeyBytes := new([64]byte)
	for i := 0; i < 32; i++ {
		privKeyBytes[i] = byte(dg.random.Int() % 256)
	}
	ed25519.MakePublicKey(privKeyBytes)
	return crypto.PrivKeyEd25519(*privKeyBytes)
}
