package test

import (
	"context"
	"testing"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func NewTransactClient(t testing.TB, listenAddress string) rpctransact.TransactClient {
	t.Helper()

	conn, err := grpc.Dial(listenAddress, grpc.WithInsecure())
	require.NoError(t, err)
	return rpctransact.NewTransactClient(conn)
}

func CreateContract(t testing.TB, cli rpctransact.TransactClient, inputAddress crypto.Address) *exec.TxExecution {
	t.Helper()

	txe, err := cli.CallTxSync(context.Background(), &payload.CallTx{
		Input: &payload.TxInput{
			Address: inputAddress,
			Amount:  2,
		},
		Address:  nil,
		Data:     Bytecode_EventsTest,
		Fee:      2,
		GasLimit: 10000,
	})
	require.NoError(t, err)

	if txe.Exception != nil {
		t.Fatalf("call should not generate exception but returned: %v", txe.Exception.Error())
	}

	return txe
}

func CallRemoveEvent(t testing.TB, cli rpctransact.TransactClient, inputAddress, contractAddress crypto.Address,
	name string) *exec.TxExecution {
	return Call(t, cli, inputAddress, contractAddress, "removeThing", name)

}

func CallAddEvent(t testing.TB, cli rpctransact.TransactClient, inputAddress, contractAddress crypto.Address,
	name, description string) *exec.TxExecution {
	return Call(t, cli, inputAddress, contractAddress, "addThing", name, description)
}

func Call(t testing.TB, cli rpctransact.TransactClient, inputAddress, contractAddress crypto.Address,
	functionName string, args ...interface{}) *exec.TxExecution {
	t.Helper()

	spec, err := abi.ReadSpec(Abi_EventsTest)
	require.NoError(t, err)

	data, _, err := spec.Pack(functionName, args...)
	require.NoError(t, err)

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

	if txe.Exception != nil {
		t.Fatalf("call should not generate exception but returned: %v", txe.Exception.Error())
	}

	return txe
}
