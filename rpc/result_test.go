// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"encoding/json"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmTypes "github.com/tendermint/tendermint/types"
)

func TestResultListAccounts(t *testing.T) {
	concreteAcc := acm.FromAddressable(acm.GeneratePrivateAccountFromSecret("Super Semi Secret"))
	acc := concreteAcc
	res := ResultAccounts{
		Accounts:    []*acm.Account{acc},
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
				Signatures: []tmTypes.CommitSig{
					{
						Signature: []byte{1, 2, 3},
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
