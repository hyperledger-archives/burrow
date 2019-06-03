// Copyright 2019 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The governance package contains functionality for altering permissions, token distribution, consensus parameters,
// validators, and network forks.
package governance

import (
	"github.com/hyperledger/burrow/acm/balance"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/genesis/spec"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs/payload"
)

// TODO:
// - Set validator power
// - Set account amount(s)
// - Set account permissions
// - Set global permissions
// - Set ConsensusParams
// Future considerations:
// - Handle network forks/termination/merging/replacement ?
// - Provide transaction in stasis/sudo (voting?)
// - Handle bonding by other means (e.g. pre-shared key permitting n bondings)
// - Network administered proxies (i.e. instead of keys have password authentication for identities - allow calls to originate as if from address without key?)
// Subject to:
// - Less than 1/3 validator power change per block

// Creates a GovTx that alters the validator power of id to the power passed
func AlterPowerTx(inputAddress crypto.Address, id crypto.Addressable, power uint64) *payload.GovTx {
	return AlterBalanceTx(inputAddress, id, balance.New().Power(power))
}

func AlterBalanceTx(inputAddress crypto.Address, id crypto.Addressable, bal balance.Balances) *payload.GovTx {
	publicKey := id.GetPublicKey()
	return UpdateAccountTx(inputAddress, &spec.TemplateAccount{
		PublicKey: &publicKey,
		Amounts:   bal,
	})
}

func AlterPermissionsTx(inputAddress crypto.Address, id crypto.Addressable, perms permission.PermFlag) *payload.GovTx {
	address := id.GetAddress()
	return UpdateAccountTx(inputAddress, &spec.TemplateAccount{
		Address:     &address,
		Permissions: permission.PermFlagToStringList(perms),
	})
}

func UpdateAccountTx(inputAddress crypto.Address, updates ...*spec.TemplateAccount) *payload.GovTx {
	return &payload.GovTx{
		Inputs: []*payload.TxInput{{
			Address: inputAddress,
		}},
		AccountUpdates: updates,
	}
}
