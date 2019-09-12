package spec

import (
	"fmt"

	"github.com/hyperledger/burrow/acm/balance"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/permission"
)

func (ta TemplateAccount) Validator(keyStore *keys.KeyStore, index int, curve crypto.CurveType) (*genesis.Validator, error) {
	var err error
	gv := new(genesis.Validator)
	gv.PublicKey, gv.Address, err = ta.RealisePublicKeyAndAddress(keyStore, crypto.CurveTypeEd25519)
	if err != nil {
		return nil, err
	}
	gv.Amount = ta.Balances().GetPower(DefaultPower)
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

func (ta TemplateAccount) AccountPermissions() (permission.AccountPermissions, error) {
	basePerms, err := permission.BasePermissionsFromStringList(ta.Permissions)
	if err != nil {
		return permission.ZeroAccountPermissions, nil
	}
	return permission.AccountPermissions{
		Base:  basePerms,
		Roles: ta.Roles,
	}, nil
}

func (ta TemplateAccount) GenesisAccount(keyStore *keys.KeyStore, index int, curve crypto.CurveType) (*genesis.Account, error) {
	var err error
	ga := new(genesis.Account)
	ga.PublicKey, ga.Address, err = ta.RealisePublicKeyAndAddress(keyStore, curve)
	if err != nil {
		return nil, err
	}
	ga.Amount = ta.Balances().GetNative(DefaultAmount)
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
func (ta TemplateAccount) RealisePublicKeyAndAddress(keyStore *keys.KeyStore, curve crypto.CurveType) (pubKey crypto.PublicKey, address crypto.Address, err error) {
	if ta.PublicKey == nil {
		var key *keys.Key
		if ta.Address == nil {
			// If neither PublicKey or Address set then generate a new one
			key, err = keyStore.Gen("", curve)
			if err != nil {
				return
			}
			if ta.Name != "" {
				err = keyStore.AddName(ta.Name, key.Address)
				if err != nil {
					return
				}
			}
		} else {
			key, err = keyStore.GetKey("", *ta.Address)
			if err != nil {
				return
			}
		}
		address = key.GetAddress()
		pubKey = key.GetPublicKey()
	} else {
		address = (*ta.PublicKey).GetAddress()
		if ta.Address != nil && *ta.Address != address {
			err = fmt.Errorf("template address %s does not match public key derived address %s", ta.Address,
				ta.PublicKey)
		}
		pubKey = *ta.PublicKey
	}
	return
}

func (ta TemplateAccount) Balances() balance.Balances {
	return ta.Amounts
}
