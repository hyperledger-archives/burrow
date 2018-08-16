// Copyright 2017 Monax Industries Limited
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

package rpc

import (
	"encoding/json"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmEd25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tmTypes "github.com/tendermint/tendermint/types"
)

func TestResultListAccounts(t *testing.T) {
	concreteAcc := acm.AsConcreteAccount(acm.FromAddressable(
		acm.GeneratePrivateAccountFromSecret("Super Semi Secret")))
	acc := concreteAcc
	res := ResultAccounts{
		Accounts:    []*acm.ConcreteAccount{acc},
		BlockHeight: 2,
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)
	resOut := new(ResultAccounts)
	json.Unmarshal(bs, resOut)
	bsOut, err := json.Marshal(resOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
}

func TestResultGetBlock(t *testing.T) {
	res := &ResultBlock{
		Block: &Block{&tmTypes.Block{
			LastCommit: &tmTypes.Commit{
				Precommits: []*tmTypes.Vote{
					{
						Signature: tmEd25519.SignatureEd25519{1, 2, 3},
					},
				},
			},
		},
		},
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)
	resOut := new(ResultBlock)
	require.NoError(t, json.Unmarshal([]byte(bs), resOut))
	bsOut, err := json.Marshal(resOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
}
