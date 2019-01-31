// +build integration

package governance

import (
	"context"
	"math/big"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/balance"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/genesis/spec"
	"github.com/hyperledger/burrow/governance"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/rpc/core"
)

func TestAlterValidators(t *testing.T) {
	inputAddress := privateAccounts[0].GetAddress()
	grpcAddress := testConfigs[0].RPC.GRPC.ListenAddress
	tcli := rpctest.NewTransactClient(t, grpcAddress)
	qcli := rpctest.NewQueryClient(t, grpcAddress)
	ecli := rpctest.NewExecutionEventsClient(t, grpcAddress)

	// Build a batch of validator alterations to make
	vs := validator.NewTrimSet()
	changePower(vs, 3, 2131)
	changePower(vs, 2, 4561)
	changePower(vs, 5, 7831)
	changePower(vs, 8, 9931)

	err := vs.IterateValidators(func(id crypto.Addressable, power *big.Int) error {
		_, err := govSync(tcli, governance.AlterPowerTx(inputAddress, id, power.Uint64()))
		return err
	})
	require.NoError(t, err)

	vsOut := getValidatorSet(t, qcli)
	// Include the genesis validator and compare the sets
	changePower(vs, 0, genesisDoc.Validators[0].Amount)
	assertValidatorsEqual(t, vs, vsOut)

	// Remove validator from chain
	_, err = govSync(tcli, governance.AlterPowerTx(inputAddress, account(3), 0))
	// Mirror in our check set
	changePower(vs, 3, 0)
	vsOut = getValidatorSet(t, qcli)
	assertValidatorsEqual(t, vs, vsOut)

	// Now check Tendermint
	waitNBlocks(t, ecli, 4)
	height := int64(kernels[0].Blockchain.LastBlockHeight())
	kernels[0].Node.ConfigureRPC()
	tmVals, err := core.Validators(&height)
	require.NoError(t, err)
	vsOut = validator.NewTrimSet()

	for _, v := range tmVals.Validators {
		publicKey, err := crypto.PublicKeyFromTendermintPubKey(v.PubKey)
		require.NoError(t, err)
		vsOut.ChangePower(publicKey, big.NewInt(v.VotingPower))
	}
	assertValidatorsEqual(t, vs, vsOut)
}

func TestAlterValidatorsTooQuickly(t *testing.T) {
	grpcAddress := testConfigs[0].RPC.GRPC.ListenAddress
	inputAddress := privateAccounts[0].GetAddress()
	tcli := rpctest.NewTransactClient(t, grpcAddress)
	qcli := rpctest.NewQueryClient(t, grpcAddress)

	maxFlow := getMaxFlow(t, qcli)
	acc1 := acm.GeneratePrivateAccountFromSecret("Foo1")
	t.Logf("Changing power of new account %v to MaxFlow = %d that should succeed", acc1.GetAddress(), maxFlow)

	_, err := govSync(tcli, governance.AlterPowerTx(inputAddress, acc1, maxFlow))
	require.NoError(t, err)

	maxFlow = getMaxFlow(t, qcli)
	power := maxFlow + 1
	acc2 := acm.GeneratePrivateAccountFromSecret("Foo2")
	t.Logf("Changing power of new account %v to MaxFlow + 1 = %d that should fail", acc2.GetAddress(), power)

	_, err = govSync(tcli, governance.AlterPowerTx(inputAddress, acc2, power))
	require.Error(t, err)
}

func TestNoRootPermission(t *testing.T) {
	grpcAddress := testConfigs[0].RPC.GRPC.ListenAddress
	tcli := rpctest.NewTransactClient(t, grpcAddress)
	// Account does not have Root permission
	inputAddress := privateAccounts[4].GetAddress()
	_, err := govSync(tcli, governance.AlterPowerTx(inputAddress, account(5), 3433))
	require.Error(t, err)
	assert.Contains(t, err.Error(), errors.PermissionDenied{Address: inputAddress, Perm: permission.Root}.Error())
}

func TestAlterAmount(t *testing.T) {
	inputAddress := privateAccounts[0].GetAddress()
	grpcAddress := testConfigs[0].RPC.GRPC.ListenAddress
	tcli := rpctest.NewTransactClient(t, grpcAddress)
	qcli := rpctest.NewQueryClient(t, grpcAddress)
	var amount uint64 = 18889
	acc := account(5)
	_, err := govSync(tcli, governance.AlterBalanceTx(inputAddress, acc, balance.New().Native(amount)))
	require.NoError(t, err)
	ca, err := qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: acc.GetAddress()})
	require.NoError(t, err)
	assert.Equal(t, amount, ca.Balance)
	// Check we haven't altered permissions
	assert.Equal(t, genesisDoc.Accounts[5].Permissions, ca.Permissions)
}

func TestAlterPermissions(t *testing.T) {
	inputAddress := privateAccounts[0].GetAddress()
	grpcAddress := testConfigs[0].RPC.GRPC.ListenAddress
	tcli := rpctest.NewTransactClient(t, grpcAddress)
	qcli := rpctest.NewQueryClient(t, grpcAddress)
	acc := account(5)
	_, err := govSync(tcli, governance.AlterPermissionsTx(inputAddress, acc, permission.Send))
	require.NoError(t, err)
	ca, err := qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: acc.GetAddress()})
	require.NoError(t, err)
	assert.Equal(t, permission.AccountPermissions{
		Base: permission.BasePermissions{
			Perms:  permission.Send,
			SetBit: permission.Send,
		},
	}, ca.Permissions)
}

func TestCreateAccount(t *testing.T) {
	inputAddress := privateAccounts[0].GetAddress()
	grpcAddress := testConfigs[0].RPC.GRPC.ListenAddress
	tcli := rpctest.NewTransactClient(t, grpcAddress)
	qcli := rpctest.NewQueryClient(t, grpcAddress)
	var amount uint64 = 18889
	acc := acm.GeneratePrivateAccountFromSecret("we almost certainly don't exist")
	_, err := govSync(tcli, governance.AlterBalanceTx(inputAddress, acc, balance.New().Native(amount)))
	require.NoError(t, err)
	ca, err := qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: acc.GetAddress()})
	require.NoError(t, err)
	assert.Equal(t, amount, ca.Balance)
}

func TestChangePowerByAddress(t *testing.T) {
	// Should use the key client to look up public key
	inputAddress := privateAccounts[0].GetAddress()
	grpcAddress := testConfigs[0].RPC.GRPC.ListenAddress
	tcli := rpctest.NewTransactClient(t, grpcAddress)

	acc := account(2)
	address := acc.GetAddress()
	power := uint64(2445)
	_, err := govSync(tcli, governance.UpdateAccountTx(inputAddress, &spec.TemplateAccount{
		Address: &address,
		Amounts: balance.New().Power(power),
	}))
	require.Error(t, err, "Should not be able to set power without providing public key")
	assert.Contains(t, err.Error(), "GovTx must be provided with public key when updating validator power")
}

func TestInvalidSequenceNumber(t *testing.T) {
	inputAddress := privateAccounts[0].GetAddress()
	tcli1 := rpctest.NewTransactClient(t, testConfigs[0].RPC.GRPC.ListenAddress)
	tcli2 := rpctest.NewTransactClient(t, testConfigs[4].RPC.GRPC.ListenAddress)
	qcli := rpctest.NewQueryClient(t, testConfigs[0].RPC.GRPC.ListenAddress)

	acc := account(2)
	address := acc.GetAddress()
	publicKey := acc.GetPublicKey()
	power := uint64(2445)
	tx := governance.UpdateAccountTx(inputAddress, &spec.TemplateAccount{
		Address:   &address,
		PublicKey: &publicKey,
		Amounts:   balance.New().Power(power),
	})

	setSequence(t, qcli, tx)
	_, err := localSignAndBroadcastSync(t, tcli1, tx)
	require.NoError(t, err)

	// Make it a different Tx hash so it can enter cache but keep sequence number
	tx.AccountUpdates[0].Amounts = balance.New().Power(power).Native(1)
	_, err = localSignAndBroadcastSync(t, tcli2, tx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sequence")
}

// Helpers

func getMaxFlow(t testing.TB, qcli rpcquery.QueryClient) uint64 {
	vs, err := qcli.GetValidatorSet(context.Background(), &rpcquery.GetValidatorSetParam{})
	require.NoError(t, err)
	set := validator.UnpersistSet(vs.Set)
	totalPower := set.TotalPower()
	maxFlow := new(big.Int)
	return maxFlow.Sub(maxFlow.Div(totalPower, big.NewInt(3)), big.NewInt(1)).Uint64()
}

func getValidatorSet(t testing.TB, qcli rpcquery.QueryClient) *validator.Set {
	vs, err := qcli.GetValidatorSet(context.Background(), &rpcquery.GetValidatorSetParam{})
	require.NoError(t, err)
	// Include the genesis validator and compare the sets
	return validator.UnpersistSet(vs.Set)
}

func account(i int) *acm.PrivateAccount {
	return rpctest.PrivateAccounts[i]
}

func govSync(cli rpctransact.TransactClient, tx *payload.GovTx) (*exec.TxExecution, error) {
	return cli.BroadcastTxSync(context.Background(), &rpctransact.TxEnvelopeParam{
		Payload: tx.Any(),
	})
}

func assertValidatorsEqual(t testing.TB, expected, actual *validator.Set) {
	require.NoError(t, expected.Equal(actual), "validator sets should be equal\nExpected: %v\n\nActual: %v\n",
		expected, actual)
}

func changePower(vs *validator.Set, i int, power uint64) {
	vs.ChangePower(account(i).GetPublicKey(), new(big.Int).SetUint64(power))
}

func setSequence(t testing.TB, qcli rpcquery.QueryClient, tx payload.Payload) {
	for _, input := range tx.GetInputs() {
		ca, err := qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: input.Address})
		require.NoError(t, err)
		input.Sequence = ca.Sequence + 1
	}
}

func localSignAndBroadcastSync(t testing.TB, tcli rpctransact.TransactClient, tx payload.Payload) (*exec.TxExecution, error) {
	txEnv := txs.Enclose(genesisDoc.ChainID(), tx)
	err := txEnv.Sign(privateAccounts[0])
	require.NoError(t, err)

	return tcli.BroadcastTxSync(context.Background(), &rpctransact.TxEnvelopeParam{Envelope: txEnv})
}

func waitNBlocks(t testing.TB, ecli rpcevents.ExecutionEventsClient, n int) {
	stream, err := ecli.GetBlockEvents(context.Background(), &rpcevents.BlocksRequest{
		BlockRange: rpcevents.NewBlockRange(rpcevents.LatestBound(), rpcevents.StreamBound()),
	})
	defer require.NoError(t, stream.CloseSend())
	var ev *exec.BlockEvent
	for err == nil && n > 0 {
		ev, err = stream.Recv()
		if err == nil && ev.EndBlock != nil {
			n--
		}
	}
	require.NoError(t, err)
}
