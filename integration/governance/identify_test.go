// +build integration

package governance

import (
	"context"
	"testing"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/rpc/rpcquery"

	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/execution/registry"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/require"
)

func TestIdentify(t *testing.T) {
	accounts := integration.MakePrivateAccounts("accounts", 2)
	kernels := make([]*core.Kernel, len(accounts))
	configs := make([]*config.BurrowConfig, len(accounts))
	genesisDoc := integration.TestGenesisDoc(accounts, 0)
	var err error

	for i, acc := range accounts {
		configs[i], err = newConfig(genesisDoc, acc, accounts...)
		require.NoError(t, err)
		configs[i].Tendermint.IdentifyPeers = true
	}

	// start first validator
	kernels[0], err = newKernelAndBoot(configs[0], accounts[0], accounts...)
	require.NoError(t, err)
	defer integration.Shutdown(kernels[0])

	// identify first validator (self)
	node := nodeFromConf(t,
		configs[0],
		configs[0].Tendermint.ListenAddress(),
		accounts[0].ConcretePrivateAccount().PrivateKey)
	identifyTx := payload.NewIdentifyTx(accounts[0].GetAddress(), node)
	tcli := rpctest.NewTransactClient(t, kernels[0].GRPCListenAddress().String())
	_, err = payloadSync(tcli, identifyTx)
	require.NoError(t, err)

	// start second node
	kernels[1], err = newKernelAndBoot(configs[1], accounts[1], accounts...)
	require.NoError(t, err)
	defer integration.Shutdown(kernels[1])

	// should not connect before identified
	err = connectKernels(kernels[1], kernels[0])
	require.Error(t, err)

	// identify second node (from first)
	node = nodeFromConf(t,
		configs[1],
		configs[1].Tendermint.ListenHost,
		accounts[1].ConcretePrivateAccount().PrivateKey)
	identifyTx = payload.NewIdentifyTx(accounts[1].GetAddress(), node)
	_, err = payloadSync(tcli, identifyTx)
	require.NoError(t, err)

	// once identified, proceed
	err = connectKernels(kernels[1], kernels[0])
	require.NoError(t, err)

	// query first validator for identities
	qcli := rpctest.NewQueryClient(t, kernels[0].GRPCListenAddress().String())
	nr, err := qcli.GetNetworkRegistry(context.TODO(), &rpcquery.GetNetworkRegistryParam{})
	require.NoError(t, err)
	netset := nr.GetSet()
	require.Len(t, netset, 2)
	addrs := make([]crypto.Address, len(netset))
	for _, node := range netset {
		addrs = append(addrs, node.Address)
	}
	require.Contains(t, addrs, accounts[0].GetAddress())
	require.Contains(t, addrs, accounts[1].GetAddress())

	// re-register node with different moniker
	configs[1].Tendermint.Moniker = "foobar"
	node = nodeFromConf(t,
		configs[1],
		configs[1].Tendermint.ListenHost,
		accounts[1].ConcretePrivateAccount().PrivateKey)
	identifyTx = payload.NewIdentifyTx(accounts[1].GetAddress(), node)
	_, err = payloadSync(tcli, identifyTx)
	require.NoError(t, err)

	// should update second node
	nr, err = qcli.GetNetworkRegistry(context.TODO(), &rpcquery.GetNetworkRegistryParam{})
	require.NoError(t, err)
	netset = nr.GetSet()
	require.Len(t, netset, 2)
	names := make([]string, len(netset))
	for _, node := range netset {
		names = append(names, node.Node.Moniker)
	}
	require.Contains(t, names, configs[1].Tendermint.Moniker)
}

func nodeFromConf(t *testing.T, conf *config.BurrowConfig, host string, val crypto.PrivateKey) *registry.NodeIdentity {
	tmConf, err := conf.TendermintConfig()
	require.NoError(t, err)
	nodeKey, err := tendermint.EnsureNodeKey(tmConf.NodeKeyFile())
	require.NoError(t, err)
	return registry.NewNodeIdentity(string(nodeKey.ID()), conf.Tendermint.Moniker, host, val)
}
