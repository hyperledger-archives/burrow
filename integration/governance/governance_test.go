// +build integration

package governance

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/balance"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/governance"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/rpc/core"
)

func TestAlterValidators(t *testing.T) {
	inputAddress := privateAccounts[0].Address()
	grpcAddress := testConfigs[0].RPC.GRPC.ListenAddress
	tcli := rpctest.NewTransactClient(t, grpcAddress)
	qcli := rpctest.NewQueryClient(t, grpcAddress)
	ecli := rpctest.NewExecutionEventsClient(t, grpcAddress)

	// Build a batch of validator alterations to make
	vs := validator.NewTrimSet()
	alterPower(vs, 3, 2131)
	alterPower(vs, 2, 4561)
	alterPower(vs, 5, 7831)
	alterPower(vs, 8, 9931)

	vs.Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		_, err := govSync(t, tcli, governance.AlterPowerTx(inputAddress, id, power.Uint64()))
		require.NoError(t, err)
		return
	})

	vsOut := getValidatorHistory(t, qcli)
	// Include the genesis validator and compare the sets
	alterPower(vs, 0, genesisDoc.Validators[0].Amount)
	assertValidatorsEqual(t, vs, vsOut)

	// Remove validator from chain
	txe, err := govSync(t, tcli, governance.AlterPowerTx(inputAddress, account(3), 0))
	// Mirror in our check set
	alterPower(vs, 3, 0)
	fmt.Println(txe.Events)
	vsOut = getValidatorHistory(t, qcli)
	assertValidatorsEqual(t, vs, vsOut)

	waitNBlocks(t, ecli, 3)
	height := int64(kernels[4].Blockchain.LastBlockHeight())
	kernels[4].Node.ConfigureRPC()
	tmVals, err := core.Validators(&height)
	require.NoError(t, err)
	vsOut = validator.NewTrimSet()

	for _, v := range tmVals.Validators {
		publicKey, err := crypto.PublicKeyFromTendermintPubKey(v.PubKey)
		require.NoError(t, err)
		vsOut.AlterPower(publicKey, big.NewInt(v.VotingPower))
	}
	assertValidatorsEqual(t, vs, vsOut)
}

func TestNoRootPermission(t *testing.T) {
	grpcAddress := testConfigs[0].RPC.GRPC.ListenAddress
	tcli := rpctest.NewTransactClient(t, grpcAddress)
	// Account does not have Root permission
	inputAddress := privateAccounts[4].Address()
	_, err := govSync(t, tcli, governance.AlterPowerTx(inputAddress, account(5), 3433))
	require.Error(t, err)
	assert.Contains(t, err.Error(), errors.ErrorCodePermissionDenied.Error())
}

func TestAlterAmount(t *testing.T) {
	inputAddress := privateAccounts[0].Address()
	grpcAddress := testConfigs[0].RPC.GRPC.ListenAddress
	tcli := rpctest.NewTransactClient(t, grpcAddress)
	qcli := rpctest.NewQueryClient(t, grpcAddress)
	var amount uint64 = 18889
	acc := account(5)
	_, err := govSync(t, tcli, governance.AlterBalanceTx(inputAddress, acc, balance.New().Native(amount)))
	require.NoError(t, err)
	ca, err := qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: acc.Address()})
	require.NoError(t, err)
	assert.Equal(t, amount, ca.Balance)
}

func TestAlterPermissions(t *testing.T) {
	inputAddress := privateAccounts[0].Address()
	grpcAddress := testConfigs[0].RPC.GRPC.ListenAddress
	tcli := rpctest.NewTransactClient(t, grpcAddress)
	qcli := rpctest.NewQueryClient(t, grpcAddress)
	acc := account(5)
	_, err := govSync(t, tcli, governance.AlterPermissionsTx(inputAddress, acc, permission.Send))
	require.NoError(t, err)
	ca, err := qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: acc.Address()})
	require.NoError(t, err)
	assert.Equal(t, permission.AccountPermissions{
		Base: permission.BasePermissions{
			Perms:  permission.Send,
			SetBit: permission.Send,
		},
	}, ca.Permissions)
}

func TestCreateAccount(t *testing.T) {
	inputAddress := privateAccounts[0].Address()
	grpcAddress := testConfigs[0].RPC.GRPC.ListenAddress
	tcli := rpctest.NewTransactClient(t, grpcAddress)
	qcli := rpctest.NewQueryClient(t, grpcAddress)
	var amount uint64 = 18889
	acc := acm.GeneratePrivateAccountFromSecret("we almost certainly don't exist")
	_, err := govSync(t, tcli, governance.AlterBalanceTx(inputAddress, acc, balance.New().Native(amount)))
	require.NoError(t, err)
	ca, err := qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: acc.Address()})
	require.NoError(t, err)
	assert.Equal(t, amount, ca.Balance)
}

func getValidatorHistory(t testing.TB, qcli rpcquery.QueryClient) *validator.Set {
	history, err := qcli.GetValidatorSet(context.Background(), &rpcquery.GetValidatorSetParam{
		IncludeHistory: true,
	})
	require.NoError(t, err)

	// Include the genesis validator and compare the sets
	return validator.UnpersistSet(history.Set)
}

func account(i int) *acm.PrivateAccount {
	return rpctest.PrivateAccounts[i]
}

func govSync(t testing.TB, cli rpctransact.TransactClient, tx *payload.GovTx) (*exec.TxExecution, error) {
	return cli.BroadcastTxSync(context.Background(), &rpctransact.TxEnvelopeParam{
		Payload: tx.Any(),
	})
}

func assertValidatorsEqual(t testing.TB, expected, actual *validator.Set) {
	if !assert.True(t, expected.Equal(actual), "sets should be equal") {
		fmt.Printf("Expected:\n%v\nActual:\n%v\n", expected, actual)
	}
}

func alterPower(vs *validator.Set, i int, power uint64) {
	vs.AlterPower(account(i), new(big.Int).SetUint64(power))
}

func waitNBlocks(t testing.TB, ecli rpcevents.ExecutionEventsClient, n int) {
	stream, err := ecli.GetBlocks(context.Background(), &rpcevents.BlocksRequest{
		BlockRange: rpcevents.NewBlockRange(rpcevents.LatestBound(), rpcevents.StreamBound()),
	})
	defer stream.CloseSend()
	for i := 0; i < n; i++ {
		require.NoError(t, err)
		_, err = stream.Recv()
	}
}
