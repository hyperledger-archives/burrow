package governance

import (
	"bytes"
	"testing"

	"github.com/hyperledger/burrow/permission"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/stretchr/testify/require"
)

func TestBonding(t *testing.T) {
	genesisAccounts := integration.MakePrivateAccounts("accounts", 4)
	genesisKernels := make([]*core.Kernel, len(genesisAccounts))
	genesisDoc := integration.TestGenesisDoc(genesisAccounts, 0, 1)
	genesisDoc.GlobalPermissions = permission.NewAccountPermissions(permission.Input)
	genesisDoc.Accounts[3].Permissions = permission.ZeroAccountPermissions.Clone()
	var err error

	// we need at least one validator to start
	for i, acc := range genesisAccounts {
		genesisKernels[i], err = createKernel(genesisDoc, acc, genesisAccounts...)
		require.NoError(t, err)
		defer integration.Shutdown(genesisKernels[i])
	}

	connectAllKernels(genesisKernels)

	t.Run("NoPermission", func(t *testing.T) {
		val := acm.GeneratePrivateAccountFromSecret("validator_1")
		localAddress := genesisKernels[3].GRPCListenAddress().String()
		inputAccount := genesisAccounts[3].GetAddress()
		tcli := rpctest.NewTransactClient(t, localAddress)
		bondTx := createBondTx(inputAccount, uint64(1<<2), val.GetPublicKey())
		_, err = payloadSync(tcli, bondTx)
		require.Error(t, err)
	})

	t.Run("BondFromNonVal", func(t *testing.T) {
		// lets do the bond tx from a non-validator node
		localAddress := genesisKernels[2].GRPCListenAddress().String()
		inputAccount := genesisAccounts[2].GetAddress()
		tcli := rpctest.NewTransactClient(t, localAddress)
		qcli := rpctest.NewQueryClient(t, localAddress)

		// make a new validator to grant power to
		val := acm.GeneratePrivateAccountFromSecret("validator_2")
		accBefore := getAccount(t, qcli, inputAccount)
		var power uint64 = 1 << 16

		bondTx := createBondTx(inputAccount, power, val.GetPublicKey())
		_, err = payloadSync(tcli, bondTx)
		require.NoError(t, err)
		accAfter := getAccount(t, qcli, inputAccount)
		// ensure power is subtracted from original account balance
		require.Equal(t, accBefore.GetBalance()-power, accAfter.GetBalance())

		valAfter := getAccount(t, qcli, val.GetAddress())
		// validator must have associated account
		// typically without balance if just created
		require.NotEmpty(t, valAfter.GetAddress())
		require.Equal(t, uint64(0), valAfter.GetBalance())

		// make sure our new validator exists in the set
		vsOut := getValidators(t, qcli)
		require.Contains(t, vsOut, val.GetAddress())
		require.Equal(t, vsOut[val.GetAddress()].GetPower(), power)

		// start the new validator
		valKernel, err := createKernel(genesisDoc, val, append(genesisAccounts, val)...)
		require.NoError(t, err)
		connectKernels(genesisKernels[0], valKernel)

		// wait for new validator to see themself in set
		waitFor(3, valKernel.Blockchain)
		grpcBondedVal := valKernel.GRPCListenAddress().String()
		qcli = rpctest.NewQueryClient(t, grpcBondedVal)
		vsOut = getValidators(t, qcli)
		require.Contains(t, vsOut, val.GetAddress())
		require.Equal(t, vsOut[val.GetAddress()].GetPower(), power)

		// wait for validator to propose a block
		waitFor(7, valKernel.Blockchain)
		checkProposed(t, genesisKernels[0], val.GetPublicKey().GetAddress().Bytes())

		unbondTx := createUnbondTx(val.GetAddress(), inputAccount)
		tcli = rpctest.NewTransactClient(t, grpcBondedVal)
		_, err = payloadSync(tcli, unbondTx)
		require.NoError(t, err)

		waitFor(2, genesisKernels[0].Blockchain)
		tcli = rpctest.NewTransactClient(t, localAddress)
		qcli = rpctest.NewQueryClient(t, localAddress)
		vsOut = getValidators(t, qcli)
		require.NotContains(t, vsOut, val.GetAddress())
		accAfter = getAccount(t, qcli, inputAccount)
		require.Equal(t, accBefore.GetBalance(), accAfter.GetBalance())
	})

	// TODO:
	// - add / remove too quickly
	// - only validator can unbond themselves
}

func checkProposed(t *testing.T, kern *core.Kernel, exp []byte) {
	height := kern.Node.BlockStore().Height()
	t.Logf("current height is %d", height)
	for i := int64(1); i < height; i++ {
		bm := kern.Node.BlockStore().LoadBlockMeta(i)
		if bytes.Equal(bm.Header.ProposerAddress, exp) {
			t.Logf("%X proposed block %d", exp, i)
			return
		}
	}
	require.Fail(t, "bonded validator did not propose any blocks")
}

func waitFor(height uint64, blockchain *bcm.Blockchain) {
	until := blockchain.LastBlockHeight() + height
	for h := uint64(0); h < until; h = blockchain.LastBlockHeight() {
		continue
	}
}
