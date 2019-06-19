package rpctest

import (
	"context"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
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

func CreateContract(cli rpctransact.TransactClient, inputAddress crypto.Address, bytecode []byte) (*exec.TxExecution, error) {
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
	if err != nil {
		return nil, err
	}
	return txe, nil
}

func CallContract(cli rpctransact.TransactClient, inputAddress, contractAddress crypto.Address, data []byte) (*exec.TxExecution, error) {
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
	if err != nil {
		return nil, err
	}
	return txe, nil
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
	sequence := GetSequence(t, client, PrivateAccounts[0].GetAddress())
	tx := payload.NewCallTxWithSequence(PrivateAccounts[0].GetPublicKey(), addr, code, amt, gasLim, fee, sequence+1)
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
	return acc.Sequence
}

// get the account
func GetAccount(t *testing.T, client infoclient.RPCClient, addr crypto.Address) *acm.Account {
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

func WaitNBlocks(ecli rpcevents.ExecutionEventsClient, n int) (rerr error) {
	stream, err := ecli.Stream(context.Background(), &rpcevents.BlocksRequest{
		BlockRange: rpcevents.NewBlockRange(rpcevents.LatestBound(), rpcevents.StreamBound()),
	})
	if err != nil {
		return err
	}
	defer func() {
		rerr = stream.CloseSend()
	}()
	var ev *exec.StreamEvent
	for err == nil && n > 0 {
		ev, err = stream.Recv()
		if err == nil && ev.EndBlock != nil {
			n--
		}
	}
	if err != nil {
		return err
	}
	return
}
