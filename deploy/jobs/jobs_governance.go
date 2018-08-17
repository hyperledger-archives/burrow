package jobs

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/util"
	"github.com/hyperledger/burrow/execution/evm/abi"
)

func UpdateAccountJob(gov *def.UpdateAccount, do *def.Packages) (interface{}, []*abi.Variable, error) {
	gov.Source = useDefault(gov.Source, do.Package.Account)
	perms := make([]string, len(gov.Permissions))

	for i, p := range gov.Permissions {
		perms[i] = string(p)
	}
	arg := &def.GovArg{
		Input:       gov.Source,
		Sequence:    gov.Sequence,
		Power:       gov.Power,
		Native:      gov.Native,
		Roles:       gov.Roles,
		Permissions: perms,
	}
	newAccountMatch := def.NewKeyRegex.FindStringSubmatch(gov.Target)
	if len(newAccountMatch) > 0 {
		keyName, curveType := def.KeyNameCurveType(newAccountMatch)
		publicKey, err := do.CreateKey(keyName, curveType)
		if err != nil {
			return nil, nil, fmt.Errorf("could not create key for new account: %v", err)
		}
		arg.Address = publicKey.Address().String()
		arg.PublicKey = publicKey.String()
	} else if len(gov.Target) == crypto.AddressHexLength {
		arg.Address = gov.Target
	} else {
		arg.PublicKey = gov.Target
	}

	tx, err := do.UpdateAccount(arg)
	if err != nil {
		return nil, nil, err
	}

	txe, err := do.SignAndBroadcast(tx)
	if err != nil {
		return nil, nil, util.ChainErrorHandler(do, err)
	}

	util.ReadTxSignAndBroadcast(txe, err)
	if err != nil {
		return nil, nil, err
	}

	return txe, util.Variables(arg), nil
}
