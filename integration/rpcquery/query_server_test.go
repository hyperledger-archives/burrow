// +build integration

package rpcquery

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAccount(t *testing.T) {
	cli := rpctest.NewQueryClient(t, testConfig.RPC.GRPC.ListenAddress)
	ca, err := cli.GetAccount(context.Background(), &rpcquery.GetAccountParam{
		Address: rpctest.PrivateAccounts[2].Address(),
	})
	require.NoError(t, err)
	genAcc := rpctest.GenesisDoc.Accounts[2]
	genAccOut := genesis.GenesisAccountFromAccount(rpctest.GenesisDoc.Accounts[2].Name, ca.Account())
	// Normalise
	genAcc.Permissions.Roles = nil
	genAccOut.Permissions.Roles = nil
	assert.Equal(t, genAcc, genAccOut)
}

func TestListAccounts(t *testing.T) {
	cli := rpctest.NewQueryClient(t, testConfig.RPC.GRPC.ListenAddress)
	stream, err := cli.ListAccounts(context.Background(), &rpcquery.ListAccountsParam{})
	require.NoError(t, err)
	var accs []acm.Account
	acc, err := stream.Recv()
	for err == nil {
		accs = append(accs, acc.Account())
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
			rpctest.UpdateName(t, tcli, rpctest.PrivateAccounts[0].Address(), name, dataA, 200)
		} else {
			rpctest.UpdateName(t, tcli, rpctest.PrivateAccounts[1].Address(), name, dataB, 200)
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
