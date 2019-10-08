// +build integration

// Space above here matters
// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package rpcinfo

import (
	"context"
	"encoding/json"
	"sort"
	"testing"
	"time"

	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/txs/payload"

	"github.com/hyperledger/burrow/core"

	"github.com/hyperledger/burrow/rpc/lib/client"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/rpcinfo/infoclient"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctypes "github.com/tendermint/tendermint/consensus/types"
)

const timeout = 5 * time.Second

func TestInfoServer(t *testing.T) {
	kern, shutdown := integration.RunNode(t, rpctest.GenesisDoc, rpctest.PrivateAccounts)
	defer shutdown()
	inputAddress := rpctest.PrivateAccounts[0].GetAddress()
	infoAddress := kern.InfoListenAddress().String()
	var clients = map[string]infoclient.RPCClient{
		"JSON RPC": client.NewJSONRPCClient(infoAddress),
		"URI":      client.NewURIClient(infoAddress),
	}
	cli := rpctest.NewTransactClient(t, kern.GRPCListenAddress().String())
	for clientName, rpcClient := range clients {
		t.Run(clientName, func(t *testing.T) {
			t.Run("Status", func(t *testing.T) {
				t.Parallel()
				resp, err := infoclient.Status(rpcClient)
				require.NoError(t, err)
				assert.Contains(t, resp.GetNodeInfo().GetMoniker(), "node")
				assert.Equal(t, rpctest.GenesisDoc.ChainID(), resp.NodeInfo.Network,
					"ChainID should match NodeInfo.Network")
			})

			t.Run("Account", func(t *testing.T) {
				t.Parallel()
				acc := rpctest.GetAccount(t, rpcClient, rpctest.PrivateAccounts[0].GetAddress())
				if acc == nil {
					t.Fatal("Account was nil")
				}
				if acc.GetAddress() != rpctest.PrivateAccounts[0].GetAddress() {
					t.Fatalf("Failed to get correct account. Got %s, expected %s", acc.GetAddress(),
						rpctest.PrivateAccounts[0].GetAddress())
				}
			})

			t.Run("Storage", func(t *testing.T) {
				t.Parallel()
				amt, gasLim, fee := uint64(1100), uint64(1000), uint64(1000)
				code := []byte{0x60, 0x5, 0x60, 0x1, 0x55}
				// Call with nil address will create a contract
				txe, err := cli.CallTxSync(context.Background(), &payload.CallTx{
					Input: &payload.TxInput{
						Address: inputAddress,
						Amount:  amt,
					},
					Data:     code,
					GasLimit: gasLim,
					Fee:      fee,
				})
				require.NoError(t, err)
				assert.Equal(t, true, txe.Receipt.CreatesContract, "This transaction should"+
					" create a contract")
				assert.NotEqual(t, 0, len(txe.TxHash), "Receipt should contain a"+
					" transaction hash")
				contractAddr := txe.Receipt.ContractAddress
				assert.NotEqual(t, 0, len(contractAddr), "Transactions claims to have"+
					" created a contract but the contract address is empty")

				v := rpctest.GetStorage(t, rpcClient, contractAddr, []byte{0x1})
				got := binary.LeftPadWord256(v)
				expected := binary.LeftPadWord256([]byte{0x5})
				if got.Compare(expected) != 0 {
					t.Fatalf("Wrong storage value. Got %x, expected %x", got.Bytes(),
						expected.Bytes())
				}
			})

			t.Run("Block", func(t *testing.T) {
				t.Parallel()
				waitNBlocks(t, kern, 1)
				res, err := infoclient.Block(rpcClient, 1)
				require.NoError(t, err)
				assert.Equal(t, int64(1), res.Block.Height)
			})

			t.Run("WaitBlocks", func(t *testing.T) {
				t.Parallel()
				waitNBlocks(t, kern, 5)
			})

			t.Run("BlockchainInfo", func(t *testing.T) {
				t.Parallel()
				// wait a mimimal number of blocks to ensure that the later query for block
				// headers has a non-trivial length
				nBlocks := 4
				waitNBlocks(t, kern, nBlocks)

				resp, err := infoclient.Blocks(rpcClient, 1, 0)
				if err != nil {
					t.Fatalf("Failed to get blockchain info: %v", err)
				}
				lastBlockHeight := resp.LastHeight
				nMetaBlocks := len(resp.BlockMetas)
				assert.True(t, uint64(nMetaBlocks) <= lastBlockHeight,
					"Logically number of block metas should be equal or less than block height.")
				assert.True(t, nBlocks <= len(resp.BlockMetas),
					"Should see at least %v BlockMetas after waiting for %v blocks but saw %v",
					nBlocks, nBlocks, len(resp.BlockMetas))
				// For the maximum number (default to 20) of retrieved block headers,
				// check that they correctly chain to each other.
				lastBlockHash := resp.BlockMetas[0].Header.Hash()
				for i := 1; i < nMetaBlocks-1; i++ {
					// the blockhash in header of height h should be identical to the hash
					// in the LastBlockID of the header of block height h+1.
					assert.Equal(t, lastBlockHash, resp.BlockMetas[i].Header.LastBlockID.Hash,
						"Blockchain should be a hash tree!")
					lastBlockHash = resp.BlockMetas[i].Header.Hash()
				}

				// Now retrieve only two blockheaders (h=1, and h=2) and check that we got
				// two results.
				resp, err = infoclient.Blocks(rpcClient, 1, 2)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(resp.BlockMetas),
					"Should see 2 BlockMetas after extracting 2 blocks")
			})

			t.Run("UnconfirmedTxs", func(t *testing.T) {
				amt, gasLim, fee := uint64(1100), uint64(1000), uint64(1000)
				code := []byte{0x60, 0x5, 0x60, 0x1, 0x55}
				// Call with nil address will create a contract
				txEnv := rpctest.MakeDefaultCallTx(t, rpcClient, nil, code, amt, gasLim, fee)
				txChan := make(chan []*txs.Envelope)

				// We want to catch the Tx in mempool before it gets reaped by tendermint
				// consensus. We should be able to do this almost always if we broadcast our
				// transaction immediately after a block has been committed. There is about
				// 1 second between blocks, and we will have the lock on Reap
				// So we wait for a block here
				waitNBlocks(t, kern, 1)

				go func() {
					for {
						resp, err := infoclient.UnconfirmedTxs(rpcClient, -1)
						if err != nil {
							// We get an error on exit
							return
						}
						if resp.NumTxs > 0 {
							txChan <- resp.Txs
						}
					}
				}()

				broadcastTxSync(t, cli, txEnv)
				select {
				case <-time.After(time.Second * timeout):
					t.Fatal("Timeout out waiting for unconfirmed transactions to appear")
				case transactions := <-txChan:
					assert.Len(t, transactions, 1, "There should only be a single transaction in the "+
						"mempool during this test (previous txs should have made it into a block)")
					assert.Contains(t, transactions, txEnv, "Transaction should be returned by ListUnconfirmedTxs")
				}
			})

			t.Run("Validators", func(t *testing.T) {
				t.Parallel()
				resp, err := infoclient.Validators(rpcClient)
				assert.NoError(t, err)
				assert.Len(t, resp.BondedValidators, 1)
				validator := resp.BondedValidators[0]
				assert.Equal(t, rpctest.GenesisDoc.Validators[0].PublicKey, validator.PublicKey)
			})

			t.Run("Consensus", func(t *testing.T) {
				t.Parallel()
				resp, err := infoclient.Consensus(rpcClient)
				require.NoError(t, err)

				// Now I do a special dance... because the votes section of RoundState has will Marshal but not Unmarshal yet
				// TODO: put in a PR in tendermint to fix thiss
				rawMap := make(map[string]json.RawMessage)
				err = json.Unmarshal(resp.RoundState, &rawMap)
				require.NoError(t, err)
				delete(rawMap, "votes")

				bs, err := json.Marshal(rawMap)
				require.NoError(t, err)

				cdc := rpc.NewAminoCodec()
				rs := new(ctypes.RoundState)
				err = cdc.UnmarshalJSON(bs, rs)
				require.NoError(t, err)

				assert.Equal(t, rs.Validators.Validators[0].Address, rs.Validators.Proposer.Address)
			})

			t.Run("Names", func(t *testing.T) {
				t.Parallel()
				names := []string{"bib", "flub", "flib"}
				sort.Strings(names)
				for _, name := range names {
					_, err := rpctest.UpdateName(cli, inputAddress, name, name, 99999)
					require.NoError(t, err)
				}

				entry, err := infoclient.Name(rpcClient, names[0])
				require.NoError(t, err)
				assert.Equal(t, names[0], entry.Name)
				assert.Equal(t, names[0], entry.Data)

				entry, err = infoclient.Name(rpcClient, "asdasdas")
				require.NoError(t, err)
				require.Nil(t, entry)

				var namesOut []string
				entries, err := infoclient.Names(rpcClient, "")
				require.NoError(t, err)
				for _, entry := range entries {
					namesOut = append(namesOut, entry.Name)
				}
				require.Equal(t, names, namesOut)

				namesOut = namesOut[:0]
				entries, err = infoclient.Names(rpcClient, "fl")
				require.NoError(t, err)
				for _, entry := range entries {
					namesOut = append(namesOut, entry.Name)
				}
				require.Equal(t, []string{"flib", "flub"}, namesOut)
			})
		})
	}

}

func waitNBlocks(t testing.TB, kern *core.Kernel, n int) {
	subID := event.GenSubID()
	ch, err := kern.Emitter.Subscribe(context.Background(), subID, exec.QueryForBlockExecution(), 10)
	require.NoError(t, err)
	defer kern.Emitter.UnsubscribeAll(context.Background(), subID)
	for i := 0; i < n; i++ {
		<-ch
	}
}

func broadcastTxSync(t testing.TB, cli rpctransact.TransactClient, txEnv *txs.Envelope) *exec.TxExecution {
	txe, err := cli.BroadcastTxSync(context.Background(), &rpctransact.TxEnvelopeParam{
		Envelope: txEnv,
	})
	require.NoError(t, err)
	return txe
}
