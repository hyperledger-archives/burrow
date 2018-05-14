package spec

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission/types"
)

const DefaultAmount uint64 = 1000000
const DefaultAmountBonded uint64 = 10000

// A GenesisSpec is schematic representation of a genesis state, that is it is a template
// for a GenesisDoc excluding that which needs to be instantiated at the point of genesis
// so it describes the type and number of accounts, the genesis salt, but not the
// account keys or addresses, or the GenesisTime. It is responsible for generating keys
// by interacting with the KeysClient it is passed and other information not known at
// specification time
type GenesisSpec struct {
	GenesisTime       *time.Time        `json:",omitempty" toml:",omitempty"`
	ChainName         string            `json:",omitempty" toml:",omitempty"`
	Salt              []byte            `json:",omitempty" toml:",omitempty"`
	GlobalPermissions []string          `json:",omitempty" toml:",omitempty"`
	Accounts          []TemplateAccount `json:",omitempty" toml:",omitempty"`
}

type TemplateAccount struct {
	// Template accounts sharing a name will be merged when merging genesis specs
	Name string `json:",omitempty" toml:",omitempty"`
	// Address  is convenient to have in file for reference, but otherwise ignored since derived from PublicKey
	Address   *acm.Address   `json:",omitempty" toml:",omitempty"`
	PublicKey *acm.PublicKey `json:",omitempty" toml:",omitempty"`
	Amount    *uint64        `json:",omitempty" toml:",omitempty"`
	// If any bonded amount then this account is also a Validator
	AmountBonded *uint64  `json:",omitempty" toml:",omitempty"`
	Permissions  []string `json:",omitempty" toml:",omitempty"`
	Roles        []string `json:",omitempty" toml:",omitempty"`
}

func (ta TemplateAccount) Validator(keyClient keys.KeyClient, index int) (*genesis.Validator, error) {
	var err error
	gv := new(genesis.Validator)
	gv.PublicKey, gv.Address, err = ta.RealisePubKeyAndAddress(keyClient)
	if err != nil {
		return nil, err
	}
	if ta.AmountBonded == nil {
		gv.Amount = DefaultAmountBonded
	} else {
		gv.Amount = *ta.AmountBonded
	}
	if ta.Name == "" {
		gv.Name = accountNameFromIndex(index)
	} else {
		gv.Name = ta.Name
	}

	gv.UnbondTo = []genesis.BasicAccount{{
		Address:   gv.Address,
		PublicKey: gv.PublicKey,
		Amount:    gv.Amount,
	}}
	return gv, nil
}

func (ta TemplateAccount) AccountPermissions() (ptypes.AccountPermissions, error) {
	basePerms, err := permission.BasePermissionsFromStringList(ta.Permissions)
	if err != nil {
		return permission.ZeroAccountPermissions, nil
	}
	return ptypes.AccountPermissions{
		Base:  basePerms,
		Roles: ta.Roles,
	}, nil
}

func (ta TemplateAccount) Account(keyClient keys.KeyClient, index int) (*genesis.Account, error) {
	var err error
	ga := new(genesis.Account)
	ga.PublicKey, ga.Address, err = ta.RealisePubKeyAndAddress(keyClient)
	if err != nil {
		return nil, err
	}
	if ta.Amount == nil {
		ga.Amount = DefaultAmount
	} else {
		ga.Amount = *ta.Amount
	}
	if ta.Name == "" {
		ga.Name = accountNameFromIndex(index)
	} else {
		ga.Name = ta.Name
	}
	if ta.Permissions == nil {
		ga.Permissions = permission.DefaultAccountPermissions.Clone()
	} else {
		ga.Permissions, err = ta.AccountPermissions()
		if err != nil {
			return nil, err
		}
	}
	return ga, nil
}

// Adds a public key and address to the template. If PublicKey will try to fetch it by Address.
// If both PublicKey and Address are not set will use the keyClient to generate a new keypair
func (ta TemplateAccount) RealisePubKeyAndAddress(keyClient keys.KeyClient) (pubKey acm.PublicKey, address acm.Address, err error) {
	if ta.PublicKey == nil {
		if ta.Address == nil {
			// If neither PublicKey or Address set then generate a new one
			address, err = keyClient.Generate(ta.Name, keys.KeyTypeEd25519Ripemd160)
			if err != nil {
				return
			}
		} else {
			address = *ta.Address
		}
		// Get the (possibly existing) key
		pubKey, err = keyClient.PublicKey(address)
		if err != nil {
			return
		}
	} else {
		address = ta.PublicKey.Address()
		if ta.Address != nil && *ta.Address != address {
			err = fmt.Errorf("template address %s does not match public key derived address %s", ta.Address,
				ta.PublicKey)
		}
		pubKey = *ta.PublicKey
	}
	return
}

// Produce a fully realised GenesisDoc from a template GenesisDoc that may omit values
func (gs *GenesisSpec) GenesisDoc(keyClient keys.KeyClient) (*genesis.GenesisDoc, error) {
	genesisDoc := new(genesis.GenesisDoc)
	if gs.GenesisTime == nil {
		genesisDoc.GenesisTime = time.Now()
	} else {
		genesisDoc.GenesisTime = *gs.GenesisTime
	}

	if gs.ChainName == "" {
		genesisDoc.ChainName = fmt.Sprintf("BurrowChain_%X", gs.ShortHash())
	} else {
		genesisDoc.ChainName = gs.ChainName
	}

	if len(gs.GlobalPermissions) == 0 {
		genesisDoc.GlobalPermissions = permission.DefaultAccountPermissions.Clone()
	} else {
		basePerms, err := permission.BasePermissionsFromStringList(gs.GlobalPermissions)
		if err != nil {
			return nil, err
		}
		genesisDoc.GlobalPermissions = ptypes.AccountPermissions{
			Base: basePerms,
		}
	}

	templateAccounts := gs.Accounts
	if len(gs.Accounts) == 0 {
		amountBonded := DefaultAmountBonded
		templateAccounts = append(templateAccounts, TemplateAccount{
			AmountBonded: &amountBonded,
		})
	}

	for i, templateAccount := range templateAccounts {
		account, err := templateAccount.Account(keyClient, i)
		if err != nil {
			return nil, fmt.Errorf("could not create Account from template: %v", err)
		}
		genesisDoc.Accounts = append(genesisDoc.Accounts, *account)
		// Create a corresponding validator
		if templateAccount.AmountBonded != nil {
			// Note this does not modify the input template
			templateAccount.Address = &account.Address
			validator, err := templateAccount.Validator(keyClient, i)
			if err != nil {
				return nil, fmt.Errorf("could not create Validator from template: %v", err)
			}
			genesisDoc.Validators = append(genesisDoc.Validators, *validator)
		}
	}

	return genesisDoc, nil
}

func (gs *GenesisSpec) JSONBytes() ([]byte, error) {
	bs, err := json.Marshal(gs)
	if err != nil {
		return nil, err
	}
	// rewrite buffer with indentation
	indentedBuffer := new(bytes.Buffer)
	if err := json.Indent(indentedBuffer, bs, "", "\t"); err != nil {
		return nil, err
	}
	return indentedBuffer.Bytes(), nil
}

func (gs *GenesisSpec) Hash() []byte {
	gsBytes, err := gs.JSONBytes()
	if err != nil {
		panic(fmt.Errorf("could not create hash of GenesisDoc: %v", err))
	}
	hasher := sha256.New()
	hasher.Write(gsBytes)
	return hasher.Sum(nil)
}

func (gs *GenesisSpec) ShortHash() []byte {
	return gs.Hash()[:genesis.ShortHashSuffixBytes]
}

func GenesisSpecFromJSON(jsonBlob []byte) (*GenesisSpec, error) {
	genDoc := new(GenesisSpec)
	err := json.Unmarshal(jsonBlob, genDoc)
	if err != nil {
		return nil, fmt.Errorf("couldn't read GenesisSpec: %v", err)
	}
	return genDoc, nil
}

func accountNameFromIndex(index int) string {
	return fmt.Sprintf("Account_%v", index)
}
