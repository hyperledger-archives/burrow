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

	"time"

	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	goCrypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/tendermint/consensus/types"
	tmTypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tmlibs/common"
)

func TestResultListAccounts(t *testing.T) {
	concreteAcc := acm.AsConcreteAccount(acm.FromAddressable(
		acm.GeneratePrivateAccountFromSecret("Super Semi Secret")))
	acc := concreteAcc
	res := ResultListAccounts{
		Accounts:    []*acm.ConcreteAccount{acc},
		BlockHeight: 2,
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)
	resOut := new(ResultListAccounts)
	json.Unmarshal(bs, resOut)
	bsOut, err := json.Marshal(resOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
}

func TestResultGetBlock(t *testing.T) {
	res := &ResultGetBlock{
		Block: &Block{&tmTypes.Block{
			LastCommit: &tmTypes.Commit{
				Precommits: []*tmTypes.Vote{
					{
						Signature: goCrypto.SignatureEd25519{1, 2, 3},
					},
				},
			},
		},
		},
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)
	resOut := new(ResultGetBlock)
	require.NoError(t, json.Unmarshal([]byte(bs), resOut))
	bsOut, err := json.Marshal(resOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
}

func TestResultDumpConsensusState(t *testing.T) {
	res := &ResultDumpConsensusState{
		RoundState: types.RoundStateSimple{
			HeightRoundStep: "34/0/3",
			Votes:           json.RawMessage(`[{"i'm a json": "32"}]`),
			LockedBlockHash: common.HexBytes{'b', 'y', 't', 'e', 's'},
		},
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)
	resOut := new(ResultDumpConsensusState)
	require.NoError(t, json.Unmarshal([]byte(bs), resOut))
	bsOut, err := json.Marshal(resOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
}

func TestResultLastBlockInfo(t *testing.T) {
	res := &ResultLastBlockInfo{
		LastBlockTime:   time.Now(),
		LastBlockHash:   binary.HexBytes{3, 4, 5, 6},
		LastBlockHeight: 2343,
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)
	fmt.Println(string(bs))

}
