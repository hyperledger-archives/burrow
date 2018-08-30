package rpctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/rpcinfo/infoclient"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// Recursive call count for UpsieDownsie() function call from strange_loop.sol
// Equals initial call, then depth from 17 -> 34, one for the bounce, then depth from 34 -> 23,
// so... (I didn't say it had to make sense):
const UpsieDownsieCallCount = 1 + (34 - 17) + 1 + (34 - 23)

var i = UpsieDownsieCallCount

var PrivateAccounts = integration.MakePrivateAccounts(10) // make keys
var GenesisDoc = integration.TestGenesisDoc(PrivateAccounts)

// Helpers
func NewTransactClient(t testing.TB, listenAddress string) rpctransact.TransactClient {
	conn, err := grpc.Dial(listenAddress, grpc.WithInsecure())
	require.NoError(t, err)
	return rpctransact.NewTransactClient(conn)
}

func NewExecutionEventsClient(t testing.TB, listenAddress string) rpcevents.ExecutionEventsClient {
	conn, err := grpc.Dial(listenAddress, grpc.WithInsecure())
	require.NoError(t, err)
	return rpcevents.NewExecutionEventsClient(conn)
}

func NewQueryClient(t testing.TB, listenAddress string) rpcquery.QueryClient {
	conn, err := grpc.Dial(listenAddress, grpc.WithInsecure())
	require.NoError(t, err)
	return rpcquery.NewQueryClient(conn)
}

func CommittedTxCount(t *testing.T, em event.Emitter) chan int {
	var numTxs int32
	emptyBlocks := 0
	maxEmptyBlocks := 2
	outCh := make(chan int)
	subID := event.GenSubID()
	ch, err := em.Subscribe(context.Background(), subID, exec.QueryForBlockExecution(), 1)
	require.NoError(t, err)

	go func() {
		defer em.UnsubscribeAll(context.Background(), subID)
		for msg := range ch {
			be := msg.(*exec.BlockExecution)
			if be.BlockHeader.NumTxs == 0 {
				emptyBlocks++
			} else {
				emptyBlocks = 0
			}
			if emptyBlocks > maxEmptyBlocks {
				break
			}
			numTxs += be.BlockHeader.NumTxs
			fmt.Printf("Total TXs committed at block %v: %v (+%v)\n", be.Height, numTxs, be.BlockHeader.NumTxs)
		}
		outCh <- int(numTxs)
	}()
	return outCh
}

func CreateContract(t testing.TB, cli rpctransact.TransactClient, inputAddress crypto.Address, bytecode []byte) *exec.TxExecution {
	txe, err := cli.CallTxSync(context.Background(), &payload.CallTx{
		Input: &payload.TxInput{
			Address: inputAddress,
			Amount:  2,
		},
		Address:  nil,
		Data:     bytecode,
		Fee:      2,
		GasLimit: 10000,
	})
	require.NoError(t, err)
	return txe
}

func CallContract(t testing.TB, cli rpctransact.TransactClient, inputAddress, contractAddress crypto.Address,
	data []byte) *exec.TxExecution {
	txe, err := cli.CallTxSync(context.Background(), &payload.CallTx{
		Input: &payload.TxInput{
			Address: inputAddress,
			Amount:  2,
		},
		Address:  &contractAddress,
		Data:     data,
		Fee:      2,
		GasLimit: 1000000,
	})
	require.NoError(t, err)
	return txe
}

func UpdateName(t testing.TB, cli rpctransact.TransactClient, inputAddress crypto.Address, name, data string,
	expiresIn uint64) *exec.TxExecution {

	txe, err := cli.NameTxSync(context.Background(), &payload.NameTx{
		Input: &payload.TxInput{
			Address: inputAddress,
			Amount:  names.NameCostForExpiryIn(name, data, expiresIn),
		},
		Name: name,
		Data: data,
	})
	require.NoError(t, err)
	return txe
}

//-------------------------------------------------------------------------------
// some default transaction functions

func MakeDefaultCallTx(t *testing.T, client infoclient.RPCClient, addr *crypto.Address, code []byte, amt, gasLim,
	fee uint64) *txs.Envelope {
	sequence := GetSequence(t, client, PrivateAccounts[0].Address())
	tx := payload.NewCallTxWithSequence(PrivateAccounts[0].PublicKey(), addr, code, amt, gasLim, fee, sequence+1)
	txEnv := txs.Enclose(GenesisDoc.ChainID(), tx)
	require.NoError(t, txEnv.Sign(PrivateAccounts[0]))
	return txEnv
}

//-------------------------------------------------------------------------------
// rpc call wrappers (fail on err)

// get an account's sequence number
func GetSequence(t *testing.T, client infoclient.RPCClient, addr crypto.Address) uint64 {
	acc, err := infoclient.Account(client, addr)
	if err != nil {
		t.Fatal(err)
	}
	if acc == nil {
		return 0
	}
	return acc.Sequence()
}

// get the account
func GetAccount(t *testing.T, client infoclient.RPCClient, addr crypto.Address) acm.Account {
	ac, err := infoclient.Account(client, addr)
	if err != nil {
		t.Fatal(err)
	}
	return ac
}

// dump all storage for an account. currently unused
func DumpStorage(t *testing.T, client infoclient.RPCClient, addr crypto.Address) *rpc.ResultDumpStorage {
	resp, err := infoclient.DumpStorage(client, addr)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func GetStorage(t *testing.T, client infoclient.RPCClient, addr crypto.Address, key []byte) []byte {
	resp, err := infoclient.Storage(client, addr, key)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

//--------------------------------------------------------------------------------
// utility verification function

// simple call contract calls another contract
