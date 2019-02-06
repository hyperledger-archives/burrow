// +build integration

package rpcquery

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/tendermint/tendermint/crypto/tmhash"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	cli := rpctest.NewQueryClient(t, testConfig.RPC.GRPC.ListenAddress)
	status, err := cli.Status(context.Background(), &rpcquery.StatusParam{})
	assert.Equal(t, rpctest.PrivateAccounts[0].GetPublicKey(), status.ValidatorInfo.PublicKey)
	assert.Equal(t, rpctest.GenesisDoc.ChainID(), status.ChainID)
	for i := 0; i < 3; i++ {
		// Unless we get lucky this is an error
		_, err = cli.Status(context.Background(), &rpcquery.StatusParam{
			BlockTimeWithin: "1ns",
		})
		if err != nil {
			break
		}
	}
	require.Error(t, err)
	assert.Contains(t, err.Error(), "have not committed block with sufficiently recent timestamp")

	for i := 0; i < 3; i++ {
		// Unless we get lucky this is an error
		_, err = cli.Status(context.Background(), &rpcquery.StatusParam{
			BlockSeenTimeWithin: "1ns",
		})
		if err != nil {
			break
		}
	}
	require.Error(t, err)
	assert.Contains(t, err.Error(), "have not committed a block sufficiently recently")
}

func TestGetAccount(t *testing.T) {
	cli := rpctest.NewQueryClient(t, testConfig.RPC.GRPC.ListenAddress)
	acc, err := cli.GetAccount(context.Background(), &rpcquery.GetAccountParam{
		Address: rpctest.PrivateAccounts[2].GetAddress(),
	})
	require.NoError(t, err)
	genAcc := rpctest.GenesisDoc.Accounts[2]
	genAccOut := genesis.GenesisAccountFromAccount(rpctest.GenesisDoc.Accounts[2].Name, acc)
	// Normalise
	genAcc.Permissions.Roles = nil
	genAccOut.Permissions.Roles = nil
	assert.Equal(t, genAcc, genAccOut)
}

func TestListAccounts(t *testing.T) {
	cli := rpctest.NewQueryClient(t, testConfig.RPC.GRPC.ListenAddress)
	stream, err := cli.ListAccounts(context.Background(), &rpcquery.ListAccountsParam{})
	require.NoError(t, err)
	var accs []*acm.Account
	acc, err := stream.Recv()
	for err == nil {
		accs = append(accs, acc)
		acc, err = stream.Recv()
	}
	if err != nil && err != io.EOF {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Len(t, accs, len(rpctest.GenesisDoc.Accounts)+1)
}

func TestListNames(t *testing.T) {
	tcli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	dataA, dataB := "NO TAMBOURINES", "ELEPHANTS WELCOME"
	n := 8
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("Flub/%v", i)
		if i%2 == 0 {
			rpctest.UpdateName(t, tcli, rpctest.PrivateAccounts[0].GetAddress(), name, dataA, 200)
		} else {
			rpctest.UpdateName(t, tcli, rpctest.PrivateAccounts[1].GetAddress(), name, dataB, 200)
		}
	}
	qcli := rpctest.NewQueryClient(t, testConfig.RPC.GRPC.ListenAddress)
	entries := receiveNames(t, qcli, "")
	assert.Len(t, entries, n)
	entries = receiveNames(t, qcli, query.NewBuilder().AndEquals("Data", dataA).String())
	if assert.Len(t, entries, n/2) {
		assert.Equal(t, dataA, entries[0].Data)
	}
}

func TestGetBlockHeader(t *testing.T) {
	qcli := rpctest.NewQueryClient(t, testConfig.RPC.GRPC.ListenAddress)
	ecli := rpctest.NewExecutionEventsClient(t, testConfig.RPC.GRPC.ListenAddress)
	height := 1
	rpctest.WaitNBlocks(t, ecli, 1)
	header, err := qcli.GetBlockHeader(context.Background(), &rpcquery.GetBlockParam{Height: uint64(height)})
	require.NoError(t, err)
	assert.Equal(t, testConfig.GenesisDoc.ChainID(), header.ChainID)
	assert.Equal(t, int64(height), header.Height)
	assert.Len(t, header.AppHash, tmhash.Size)
}

func receiveNames(t testing.TB, qcli rpcquery.QueryClient, query string) []*names.Entry {
	stream, err := qcli.ListNames(context.Background(), &rpcquery.ListNamesParam{
		Query: query,
	})
	require.NoError(t, err)
	var entries []*names.Entry
	entry, err := stream.Recv()
	for err == nil {
		entries = append(entries, entry)
		entry, err = stream.Recv()
	}
	if err != nil && err != io.EOF {
		t.Fatalf("unexpected error: %v", err)
	}
	return entries
}
