// +build integration

package governance

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/p2p"
)

func newConfig(genesisDoc *genesis.GenesisDoc, account *acm.PrivateAccount,
	keysAccounts ...*acm.PrivateAccount) (conf *config.BurrowConfig, err error) {

	// FIXME: some combination of cleanup and shutdown seems to make tests fail on CI
	// testConfig, cleanup := integration.NewTestConfig(genesisDoc)
	testConfig, _ := integration.NewTestConfig(genesisDoc)
	// defer cleanup()

	// comment to see all logging
	testConfig.Logging = logconfig.New().Root(func(sink *logconfig.SinkConfig) *logconfig.SinkConfig {
		return sink.SetTransform(logconfig.FilterTransform(logconfig.IncludeWhenAllMatch,
			"total_validator")).SetOutput(logconfig.StdoutOutput())
	})

	// Try and grab a free port - this is not foolproof since there is race between other concurrent tests after we close
	// the listener and start the node
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}
	host, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return nil, err
	}
	testConfig.Tendermint.ListenHost = host
	testConfig.Tendermint.ListenPort = port

	err = l.Close()
	if err != nil {
		return nil, err
	}

	return testConfig, nil
}

func newKernelAndBoot(conf *config.BurrowConfig, account *acm.PrivateAccount,
	keysAccounts ...*acm.PrivateAccount) (kernel *core.Kernel, err error) {

	kernel, err = integration.TestKernel(account, keysAccounts, conf)
	if err != nil {
		return nil, err
	}

	return kernel, kernel.Boot()
}

func signTx(t *testing.T, tx payload.Payload, chainID string, from acm.AddressableSigner) (txEnv *txs.Envelope) {
	txEnv = txs.Enclose(chainID, tx)
	require.NoError(t, txEnv.Sign(from))
	return
}

func getValidators(t testing.TB, qcli rpcquery.QueryClient) map[crypto.Address]*validator.Validator {
	vs, err := qcli.GetValidatorSet(context.Background(), &rpcquery.GetValidatorSetParam{})
	require.NoError(t, err)
	vals := make(map[crypto.Address]*validator.Validator, len(vs.Set))
	for _, v := range vs.Set {
		vals[v.PublicKey.GetAddress()] = v
	}
	return vals
}

func getValidatorSet(t testing.TB, qcli rpcquery.QueryClient) *validator.Set {
	vs, err := qcli.GetValidatorSet(context.Background(), &rpcquery.GetValidatorSetParam{})
	require.NoError(t, err)
	// Include the genesis validator and compare the sets
	return validator.UnpersistSet(vs.Set)
}

func getAccount(t testing.TB, qcli rpcquery.QueryClient, address crypto.Address) *acm.Account {
	acc, err := qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{
		Address: address,
	})
	require.NoError(t, err)
	return acc
}

func account(i int) *acm.PrivateAccount {
	return rpctest.PrivateAccounts[i]
}

func payloadSync(cli rpctransact.TransactClient, tx payload.Payload) (*exec.TxExecution, error) {
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

func connectKernels(k1, k2 *core.Kernel) error {
	k1Address, err := k1.Node.NodeInfo().NetAddress()
	if err != nil {
		return fmt.Errorf("could not get kernel address: %v", err)
	}
	k2Address, err := k2.Node.NodeInfo().NetAddress()
	if err != nil {
		return fmt.Errorf("could not get kernel address: %v", err)
	}
	fmt.Printf("Connecting %v -> %v\n", k1Address, k2Address)
	err = k1.Node.Switch().DialPeerWithAddress(k2Address)
	if err != nil {
		switch e := err.(type) {
		case p2p.ErrRejected:
			return fmt.Errorf("connection between test kernels was rejected: %v", e)
		default:
			return fmt.Errorf("could not connect test kernels: %v", err)
		}
	}
	return nil
}

func connectAllKernels(ks []*core.Kernel) error {
	source := ks[0]
	for _, dest := range ks[1:] {
		err := connectKernels(source, dest)
		if err != nil {
			return err
		}
	}
	return nil
}

func getMaxFlow(t testing.TB, qcli rpcquery.QueryClient) uint64 {
	vs, err := qcli.GetValidatorSet(context.Background(), &rpcquery.GetValidatorSetParam{})
	require.NoError(t, err)
	set := validator.UnpersistSet(vs.Set)
	totalPower := set.TotalPower()
	maxFlow := new(big.Int)
	return maxFlow.Sub(maxFlow.Div(totalPower, big.NewInt(3)), big.NewInt(1)).Uint64()
}

func setSequence(t testing.TB, qcli rpcquery.QueryClient, tx payload.Payload) {
	for _, input := range tx.GetInputs() {
		ca, err := qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: input.Address})
		require.NoError(t, err)
		input.Sequence = ca.Sequence + 1
	}
}

func localSignAndBroadcastSync(t testing.TB, tcli rpctransact.TransactClient, chainID string,
	signer acm.AddressableSigner, tx payload.Payload) (*exec.TxExecution, error) {
	txEnv := txs.Enclose(chainID, tx)
	err := txEnv.Sign(signer)
	require.NoError(t, err)

	return tcli.BroadcastTxSync(context.Background(), &rpctransact.TxEnvelopeParam{Envelope: txEnv})
}
