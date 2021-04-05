// +build integration

package governance

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/balance"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/genesis/spec"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmcore "github.com/tendermint/tendermint/rpc/core"
	"github.com/tendermint/tendermint/rpc/jsonrpc/types"
)

func TestGovernance(t *testing.T) {
	genesisAccounts := integration.MakePrivateAccounts("mysecret", 10) // make keys
	genesisConfigs := make([]*config.BurrowConfig, len(genesisAccounts))
	genesisKernels := make([]*core.Kernel, len(genesisAccounts))
	genesisDoc := integration.TestGenesisDoc(genesisAccounts, 0)
	genesisDoc.Accounts[4].Permissions = permission.NewAccountPermissions(permission.Send | permission.Call)
	var err error

	for i, acc := range genesisAccounts {
		genesisConfigs[i], err = newConfig(genesisDoc, acc, genesisAccounts...)
		require.NoError(t, err)

		genesisKernels[i], err = newKernelAndBoot(genesisConfigs[i], acc, genesisAccounts...)
		require.NoError(t, err)
		defer integration.Shutdown(genesisKernels[i])
	}

	time.Sleep(1 * time.Second)
	for i := 0; i < len(genesisKernels); i++ {
		for j := i + 1; j < len(genesisKernels); j++ {
			connectKernels(genesisKernels[i], genesisKernels[j])
		}
	}

	t.Run("Group", func(t *testing.T) {
		t.Run("AlterValidators", func(t *testing.T) {
			inputAddress := genesisAccounts[0].GetAddress()
			grpcAddress := genesisKernels[0].GRPCListenAddress().String()
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
				_, err := payloadSync(tcli, payload.AlterPowerTx(inputAddress, id, power.Uint64()))
				return err
			})
			require.NoError(t, err)

			vsOut := getValidatorSet(t, qcli)
			// Include the genesis validator and compare the sets
			changePower(vs, 0, genesisDoc.Validators[0].Amount)
			assertValidatorsEqual(t, vs, vsOut)

			// Remove validator from chain
			_, err = payloadSync(tcli, payload.AlterPowerTx(inputAddress, account(3), 0))
			require.NoError(t, err)

			// Mirror in our check set
			changePower(vs, 3, 0)
			vsOut = getValidatorSet(t, qcli)
			assertValidatorsEqual(t, vs, vsOut)

			// Now check Tendermint
			err = rpctest.WaitNBlocks(ecli, 6)
			require.NoError(t, err)
			height := int64(genesisKernels[0].Blockchain.LastBlockHeight())
			err = genesisKernels[0].Node.ConfigureRPC()
			require.NoError(t, err)
			tmVals, err := tmcore.Validators(&types.Context{}, &height, nil, nil)
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
			grpcAddress := genesisKernels[0].GRPCListenAddress().String()
			ecli := rpctest.NewExecutionEventsClient(t, grpcAddress)
			err := rpctest.WaitNBlocks(ecli, 2)
			require.NoError(t, err)
		})

		t.Run("AlterValidatorsTooQuickly", func(t *testing.T) {
			grpcAddress := genesisKernels[0].GRPCListenAddress().String()
			inputAddress := genesisAccounts[0].GetAddress()
			tcli := rpctest.NewTransactClient(t, grpcAddress)
			qcli := rpctest.NewQueryClient(t, grpcAddress)

			maxFlow := getMaxFlow(t, qcli)
			acc1 := acm.GeneratePrivateAccountFromSecret("Foo1")
			t.Logf("Changing power of new account %v to MaxFlow = %d that should succeed", acc1.GetAddress(), maxFlow)

			_, err := payloadSync(tcli, payload.AlterPowerTx(inputAddress, acc1, maxFlow))
			require.NoError(t, err)

			maxFlow = getMaxFlow(t, qcli)
			power := maxFlow + 1
			acc2 := acm.GeneratePrivateAccountFromSecret("Foo2")
			t.Logf("Changing power of new account %v to MaxFlow + 1 = %d that should fail", acc2.GetAddress(), power)

			_, err = payloadSync(tcli, payload.AlterPowerTx(inputAddress, acc2, power))
			require.Error(t, err)
		})

		t.Run("NoRootPermission", func(t *testing.T) {
			grpcAddress := genesisKernels[0].GRPCListenAddress().String()
			tcli := rpctest.NewTransactClient(t, grpcAddress)
			// Account does not have Root permission
			inputAddress := genesisAccounts[4].GetAddress()
			_, err := payloadSync(tcli, payload.AlterPowerTx(inputAddress, account(5), 3433))
			require.Error(t, err)
			assert.Contains(t, err.Error(), errors.PermissionDenied{Address: inputAddress, Perm: permission.Root}.Error())
		})

		t.Run("AlterAmount", func(t *testing.T) {
			inputAddress := genesisAccounts[0].GetAddress()
			grpcAddress := genesisKernels[0].GRPCListenAddress().String()
			tcli := rpctest.NewTransactClient(t, grpcAddress)
			qcli := rpctest.NewQueryClient(t, grpcAddress)
			var amount uint64 = 18889
			acc := account(5)
			_, err := payloadSync(tcli, payload.AlterBalanceTx(inputAddress, acc, balance.New().Native(amount)))
			require.NoError(t, err)
			ca, err := qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: acc.GetAddress()})
			require.NoError(t, err)
			assert.Equal(t, amount, ca.Balance)
			// Check we haven't altered permissions
			assert.Equal(t, genesisDoc.Accounts[5].Permissions, ca.Permissions)
		})

		t.Run("AlterPermissions", func(t *testing.T) {
			inputAddress := genesisAccounts[0].GetAddress()
			grpcAddress := genesisKernels[0].GRPCListenAddress().String()
			tcli := rpctest.NewTransactClient(t, grpcAddress)
			qcli := rpctest.NewQueryClient(t, grpcAddress)
			acc := account(5)
			_, err := payloadSync(tcli, payload.AlterPermissionsTx(inputAddress, acc, permission.Send))
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
			inputAddress := genesisAccounts[0].GetAddress()
			grpcAddress := genesisKernels[0].GRPCListenAddress().String()
			tcli := rpctest.NewTransactClient(t, grpcAddress)
			qcli := rpctest.NewQueryClient(t, grpcAddress)
			var amount uint64 = 18889
			acc := acm.GeneratePrivateAccountFromSecret("we almost certainly don't exist")
			govTx := payload.AlterBalanceTx(inputAddress, acc, balance.New().Native(amount))
			_, err := payloadSync(tcli, govTx)
			require.NoError(t, err)
			ca, err := qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: acc.GetAddress()})
			require.NoError(t, err)
			assert.Equal(t, amount, ca.Balance)
		})

		t.Run("ChangePowerByAddress", func(t *testing.T) {
			// Should use the key client to look up public key
			inputAddress := genesisAccounts[0].GetAddress()
			grpcAddress := genesisKernels[0].GRPCListenAddress().String()
			tcli := rpctest.NewTransactClient(t, grpcAddress)

			acc := account(2)
			address := acc.GetAddress()
			power := uint64(2445)
			_, err := payloadSync(tcli, payload.UpdateAccountTx(inputAddress, &spec.TemplateAccount{
				Address: &address,
				Amounts: balance.New().Power(power),
			}))
			require.Error(t, err, "Should not be able to set power without providing public key")
			assert.Contains(t, err.Error(), "GovTx: must be provided with public key when updating validator power")
		})

		t.Run("InvalidSequenceNumber", func(t *testing.T) {
			inputAddress := genesisAccounts[0].GetAddress()
			tcli1 := rpctest.NewTransactClient(t, genesisKernels[0].GRPCListenAddress().String())
			tcli2 := rpctest.NewTransactClient(t, genesisKernels[4].GRPCListenAddress().String())
			qcli := rpctest.NewQueryClient(t, genesisKernels[0].GRPCListenAddress().String())

			acc := account(2)
			address := acc.GetAddress()
			publicKey := acc.GetPublicKey()
			power := uint64(2445)
			tx := payload.UpdateAccountTx(inputAddress, &spec.TemplateAccount{
				Address:   &address,
				PublicKey: publicKey,
				Amounts:   balance.New().Power(power),
			})

			setSequence(t, qcli, tx)
			_, err := localSignAndBroadcastSync(t, tcli1, genesisDoc.GetChainID(), genesisAccounts[0], tx)
			require.NoError(t, err)

			// Make it a different Tx hash so it can enter cache but keep sequence number
			tx.AccountUpdates[0].Amounts = balance.New().Power(power).Native(1)
			_, err = localSignAndBroadcastSync(t, tcli2, genesisDoc.GetChainID(), genesisAccounts[0], tx)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid sequence")
		})
	})

	// tendermint AddPeer() runs asynchronously and needs to complete before we shutdown, else we get an exception like
	// goroutine 2181 [running]:
	// runtime/debug.Stack(0x12786c0, 0xc000085d70, 0xc000085c50)
	// /home/sean/go1.12.1/src/runtime/debug/stack.go:24 +0x9d
	// github.com/hyperledger/burrow/vendor/github.com/tendermint/tendermint/libs/db.(*GoLevelDB).Get(0xc005c5b318, 0xc01fd71840, 0x5, 0x8, 0x5, 0x8, 0x5)
	// /home/sean/go/src/github.com/hyperledger/burrow/vendor/github.com/tendermint/tendermint/libs/db/go_level_db.go:57 +0xaf
	// github.com/hyperledger/burrow/vendor/github.com/tendermint/tendermint/blockchain.(*BlockStore).LoadSeenCommit(0xc00bfd6120, 0x12, 0xc002b85c90)
	// /home/sean/go/src/github.com/hyperledger/burrow/vendor/github.com/tendermint/tendermint/blockchain/store.go:128 +0xf2
	// github.com/hyperledger/burrow/vendor/github.com/tendermint/tendermint/consensus.(*ConsensusState).LoadCommit(0xc002b85c00, 0x12, 0x0)
	// /home/sean/go/src/github.com/hyperledger/burrow/vendor/github.com/tendermint/tendermint/consensus/state.go:273 +0xb2
	// github.com/hyperledger/burrow/vendor/github.com/tendermint/tendermint/consensus.(*ConsensusReactor).queryMaj23Routine(0xc0008ec680, 0x12ad1a0, 0xc010f79800, 0xc009119520)
	// /home/sean/go/src/github.com/hyperledger/burrow/vendor/github.com/tendermint/tendermint/consensus/reactor.go:789 +0x291
	// created by github.com/hyperledger/burrow/vendor/github.com/tendermint/tendermint/consensus.(*ConsensusReactor).AddPeer
	// /home/sean/go/src/github.com/hyperledger/burrow/vendor/github.com/tendermint/tendermint/consensus/reactor.go:171 +0x23a

	time.Sleep(20 * time.Second)
}
