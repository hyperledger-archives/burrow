// +build integration

package governance

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/balance"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/genesis/spec"
	"github.com/hyperledger/burrow/governance"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/p2p"
	tmcore "github.com/tendermint/tendermint/rpc/core"
	rpctypes "github.com/tendermint/tendermint/rpc/lib/types"
)

type foo struct {
	i int
}

func (f *foo) Speak() {
	fmt.Println(f.i)
}

func TestDefer(t *testing.T) {
	var foos []*foo
	for i := 0; i < 10; i++ {
		foos = append(foos, &foo{i})
	}
	for i := 0; i < 10; i++ {
		defer foos[i].Speak()
	}
}

func TestGovernance(t *testing.T) {
	privateAccounts := integration.MakePrivateAccounts(10) // make keys
	genesisDoc := integration.TestGenesisDoc(privateAccounts)
	kernels := make([]*core.Kernel, len(privateAccounts))
	genesisDoc.Accounts[4].Permissions = permission.NewAccountPermissions(permission.Send | permission.Call)

	for i, acc := range privateAccounts {
		testConfig, cleanup := integration.NewTestConfig(genesisDoc)
		defer cleanup()

		logconf := logconfig.New().Root(func(sink *logconfig.SinkConfig) *logconfig.SinkConfig {
			return sink.SetTransform(logconfig.FilterTransform(logconfig.IncludeWhenAllMatch,
				"total_validator")).SetOutput(logconfig.StdoutOutput())
		})

		// Try and grab a free port - this is not foolproof since there is race between other concurrent tests after we close
		// the listener and start the node
		l, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)
		testConfig.Tendermint.ListenAddress = fmt.Sprintf("tcp://%v", l.Addr().String())

		kernels[i], err = integration.TestKernel(acc, privateAccounts, testConfig, logconf)
		require.NoError(t, err)

		err = l.Close()
		require.NoError(t, err)

		err = kernels[i].Boot()
		require.NoError(t, err)

		defer kernels[i].Shutdown(context.Background())
	}

	time.Sleep(1 * time.Second)
	for i := 0; i < len(kernels); i++ {
		for j := i + 1; j < len(kernels); j++ {
			connectKernels(kernels[i], kernels[j])
		}
	}

	t.Run("Group", func(t *testing.T) {
		t.Run("AlterValidators", func(t *testing.T) {
			inputAddress := privateAccounts[0].GetAddress()
			grpcAddress := kernels[0].GRPCListenAddress().String()
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
			require.NoError(t, err)

			// Mirror in our check set
			changePower(vs, 3, 0)
			vsOut = getValidatorSet(t, qcli)
			assertValidatorsEqual(t, vs, vsOut)

			// Now check Tendermint
			rpctest.WaitNBlocks(t, ecli, 5)
			height := int64(kernels[0].Blockchain.LastBlockHeight())
			kernels[0].Node.ConfigureRPC()
			tmVals, err := tmcore.Validators(&rpctypes.Context{}, &height)
			require.NoError(t, err)
			vsOut = validator.NewTrimSet()

			for _, v := range tmVals.Validators {
				publicKey, err := crypto.PublicKeyFromTendermintPubKey(v.PubKey)
				require.NoError(t, err)
				vsOut.ChangePower(publicKey, big.NewInt(v.VotingPower))
			}
			assertValidatorsEqual(t, vs, vsOut)
		})

		t.Run("WaitBlocks", func(t *testing.T) {
			grpcAddress := kernels[0].GRPCListenAddress().String()
			ecli := rpctest.NewExecutionEventsClient(t, grpcAddress)
			rpctest.WaitNBlocks(t, ecli, 2)
		})

		t.Run("AlterValidatorsTooQuickly", func(t *testing.T) {
			grpcAddress := kernels[0].GRPCListenAddress().String()
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
		})

		t.Run("NoRootPermission", func(t *testing.T) {
			grpcAddress := kernels[0].GRPCListenAddress().String()
			tcli := rpctest.NewTransactClient(t, grpcAddress)
			// Account does not have Root permission
			inputAddress := privateAccounts[4].GetAddress()
			_, err := govSync(tcli, governance.AlterPowerTx(inputAddress, account(5), 3433))
			require.Error(t, err)
			assert.Contains(t, err.Error(), errors.PermissionDenied{Address: inputAddress, Perm: permission.Root}.Error())
		})

		t.Run("AlterAmount", func(t *testing.T) {
			inputAddress := privateAccounts[0].GetAddress()
			grpcAddress := kernels[0].GRPCListenAddress().String()
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
		})

		t.Run("AlterPermissions", func(t *testing.T) {
			inputAddress := privateAccounts[0].GetAddress()
			grpcAddress := kernels[0].GRPCListenAddress().String()
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
		})

		t.Run("CreateAccount", func(t *testing.T) {
			inputAddress := privateAccounts[0].GetAddress()
			grpcAddress := kernels[0].GRPCListenAddress().String()
			tcli := rpctest.NewTransactClient(t, grpcAddress)
			qcli := rpctest.NewQueryClient(t, grpcAddress)
			var amount uint64 = 18889
			acc := acm.GeneratePrivateAccountFromSecret("we almost certainly don't exist")
			_, err := govSync(tcli, governance.AlterBalanceTx(inputAddress, acc, balance.New().Native(amount)))
			require.NoError(t, err)
			ca, err := qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: acc.GetAddress()})
			require.NoError(t, err)
			assert.Equal(t, amount, ca.Balance)
		})

		t.Run("ChangePowerByAddress", func(t *testing.T) {
			// Should use the key client to look up public key
			inputAddress := privateAccounts[0].GetAddress()
			grpcAddress := kernels[0].GRPCListenAddress().String()
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
		})

		t.Run("InvalidSequenceNumber", func(t *testing.T) {
			inputAddress := privateAccounts[0].GetAddress()
			tcli1 := rpctest.NewTransactClient(t, kernels[0].GRPCListenAddress().String())
			tcli2 := rpctest.NewTransactClient(t, kernels[4].GRPCListenAddress().String())
			qcli := rpctest.NewQueryClient(t, kernels[0].GRPCListenAddress().String())

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
			_, err := localSignAndBroadcastSync(t, tcli1, genesisDoc.ChainID(), privateAccounts[0], tx)
			require.NoError(t, err)

			// Make it a different Tx hash so it can enter cache but keep sequence number
			tx.AccountUpdates[0].Amounts = balance.New().Power(power).Native(1)
			_, err = localSignAndBroadcastSync(t, tcli2, genesisDoc.ChainID(), privateAccounts[0], tx)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid sequence")
		})
	})
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

func localSignAndBroadcastSync(t testing.TB, tcli rpctransact.TransactClient, chainID string,
	signer acm.AddressableSigner, tx payload.Payload) (*exec.TxExecution, error) {
	txEnv := txs.Enclose(chainID, tx)
	err := txEnv.Sign(signer)
	require.NoError(t, err)

	return tcli.BroadcastTxSync(context.Background(), &rpctransact.TxEnvelopeParam{Envelope: txEnv})
}

func connectKernels(k1, k2 *core.Kernel) {
	k1Address := k1.Node.NodeInfo().NetAddress()
	k2Address := k2.Node.NodeInfo().NetAddress()
	fmt.Printf("Connecting %v -> %v\n", k1Address, k2Address)
	err := k1.Node.Switch().DialPeerWithAddress(k2Address, false)
	if err != nil {
		switch e := err.(type) {
		case p2p.ErrRejected:
			panic(fmt.Errorf("connection between test kernels was rejected: %v", e))
		default:
			panic(fmt.Errorf("could not connect test kernels: %v", err))
		}
	}
}
