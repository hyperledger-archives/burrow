// +build integration

package governance

import (
	"testing"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/stretchr/testify/require"
)

func TestBonding(t *testing.T) {
	genesisAccounts := integration.MakePrivateAccounts("accounts", 2)
	genesisKernels := make([]core.Kernel, len(genesisAccounts))
	genesisDoc := integration.TestGenesisDoc(genesisAccounts)

	// we need at least one validator to start
	// in this case genesisKernels[0]
	for i, acc := range genesisAccounts {
		err := startNode(&genesisKernels[i], genesisDoc, acc, genesisAccounts...)
		require.NoError(t, err)
		defer integration.Shutdown(&genesisKernels[i])
	}

	connectKernels(&genesisKernels[0], &genesisKernels[1])

	// lets do the bond tx from the non-validator
	grpcGenVal := genesisKernels[1].GRPCListenAddress().String()
	tcli := rpctest.NewTransactClient(t, grpcGenVal)
	qcli := rpctest.NewQueryClient(t, grpcGenVal)

	var power uint64 = 1000
	inputAddress := genesisAccounts[1].GetAddress()

	// make a new validator to grant power to
	val := acm.GeneratePrivateAccountFromSecret("validator")

	accBefore := getAccount(t, qcli, inputAddress)

	bondTx := createBondTx(inputAddress, power, val.GetPublicKey())
	_, err := sendPayload(tcli, bondTx)
	require.NoError(t, err)
	accAfter := getAccount(t, qcli, inputAddress)
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
	valKernel := &core.Kernel{}
	err = startNode(valKernel, genesisDoc, val, append(genesisAccounts, val)...)
	require.NoError(t, err)
	connectKernels(&genesisKernels[0], valKernel)

	// wait for new validator to see themself in set
	time.Sleep(2 * time.Second)
	grpcBondedVal := valKernel.GRPCListenAddress().String()
	qcli = rpctest.NewQueryClient(t, grpcBondedVal)
	vsOut = getValidators(t, qcli)
	require.Contains(t, vsOut, val.GetAddress())
	require.Equal(t, vsOut[val.GetAddress()].GetPower(), power)

	unbondTx := createUnbondTx(val.GetAddress(), inputAddress)
	tcli = rpctest.NewTransactClient(t, grpcBondedVal)
	_, err = sendPayload(tcli, unbondTx)
	require.NoError(t, err)

	tcli = rpctest.NewTransactClient(t, grpcGenVal)
	qcli = rpctest.NewQueryClient(t, grpcGenVal)
	vsOut = getValidators(t, qcli)
	require.NotContains(t, vsOut, val.GetAddress())
	accAfter = getAccount(t, qcli, inputAddress)
	require.Equal(t, accBefore.GetBalance(), accAfter.GetBalance())

	// TODO:
	// - ensure bonded validator can vote
	// - add / remove too quickly
	// - only validator can unbond themselves
	// - cannot bond more than one validator? / delegated bonding?
}
