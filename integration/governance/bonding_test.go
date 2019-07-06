// +build integration

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
	genesisAccounts := integration.MakePrivateAccounts("accounts", 6)
	genesisKernels := make([]*core.Kernel, len(genesisAccounts))
	genesisDoc := integration.TestGenesisDoc(genesisAccounts, 0, 1, 2, 3)
	genesisDoc.GlobalPermissions = permission.NewAccountPermissions(permission.Input)
	genesisDoc.Accounts[4].Permissions = permission.ZeroAccountPermissions.Clone()
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
		localAddress := genesisKernels[4].GRPCListenAddress().String()
		inputAccount := genesisAccounts[4].GetAddress()
		tcli := rpctest.NewTransactClient(t, localAddress)
		bondTx := createBondTx(inputAccount, val.GetPublicKey(), uint64(1<<2))
		_, err = payloadSync(tcli, bondTx)
		require.Error(t, err)
	})

	t.Run("BondFromNonVal", func(t *testing.T) {
		// lets do the bond tx from a non-validator node
		valAccount := genesisAccounts[5]
		valKernel := genesisKernels[5]

		localAddress := valKernel.GRPCListenAddress().String()
		inputAccount := valAccount.GetAddress()
		tcli := rpctest.NewTransactClient(t, localAddress)
		qcli := rpctest.NewQueryClient(t, localAddress)

		accBefore := getAccount(t, qcli, inputAccount)
		var power uint64 = 1 << 16

		bondTx := createBondTx(inputAccount, valAccount.GetPublicKey(), power)
		_, err = payloadSync(tcli, bondTx)
		require.NoError(t, err)
		accAfter := getAccount(t, qcli, inputAccount)
		// ensure power is subtracted from original account balance
		require.Equal(t, accBefore.GetBalance()-power, accAfter.GetBalance())

		// make sure our new validator exists in the set
		vsOut := getValidators(t, qcli)
		require.Contains(t, vsOut, valAccount.GetAddress())
		require.Equal(t, vsOut[valAccount.GetAddress()].GetPower(), power)

		// wait for new validator to see themself in set
		waitFor(3, valKernel.Blockchain)
		vsOut = getValidators(t, qcli)
		require.Contains(t, vsOut, valAccount.GetAddress())
		require.Equal(t, vsOut[valAccount.GetAddress()].GetPower(), power)

		// wait for validator to propose a block
		waitFor(7, valKernel.Blockchain)
		checkProposed(t, genesisKernels[0], valAccount.GetPublicKey().GetAddress().Bytes())

		unbondTx := createUnbondTx(inputAccount, valAccount.GetPublicKey(), power)
		_, err = payloadSync(tcli, unbondTx)
		require.NoError(t, err)

		waitFor(2, valKernel.Blockchain)
		vsOut = getValidators(t, qcli)
		require.NotContains(t, vsOut, valAccount.GetAddress())
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
