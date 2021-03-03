// +build integration,ethereum

package service_test

import (
	"testing"

	"github.com/hyperledger/burrow/rpc/web3/ethclient"
	"github.com/hyperledger/burrow/tests/web3/web3test"
	"github.com/hyperledger/burrow/vent/test"
	"github.com/stretchr/testify/require"
)

func TestEthereumConsumer(t *testing.T) {
	pk := web3test.GetPrivateKey(t)
	tcli := ethclient.NewTransactClient(web3test.GetChainRPCClient())
	chainID, err := tcli.GetChainID()
	require.NoError(t, err)
	testConsumer(t, chainID, test.PostgresVentConfig(web3test.GetChainRemote()), tcli, pk.GetAddress())
}
