// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/testutil"
	"github.com/stretchr/testify/require"
)

const (
	ChainName = "Integration_Test_Chain"
)

// Enable logger output during tests

var node uint64 = 0

func NoConsensus(conf *config.BurrowConfig) {
	conf.Tendermint.Enabled = false
}

func CommitImmediately(conf *config.BurrowConfig) {
	conf.Execution.TimeoutFactor = 0
}

func RunNode(t testing.TB, genesisDoc *genesis.GenesisDoc, privateAccounts []*acm.PrivateAccount,
	options ...func(*config.BurrowConfig)) (kern *core.Kernel, shutdown func()) {

	var err error
	testConfig, cleanup := NewTestConfig(genesisDoc, options...)
	// Uncomment for log output from tests
	// testConfig.Logging = logconfig.New().Root(func(sink *logconfig.SinkConfig) *logconfig.SinkConfig {
	//   return sink.SetOutput(logconfig.StderrOutput())
	// })
	testConfig.Logging = logconfig.New().Root(func(sink *logconfig.SinkConfig) *logconfig.SinkConfig {
		return sink.SetTransform(logconfig.FilterTransform(logconfig.IncludeWhenAllMatch,
			"total_validator")).SetOutput(logconfig.StdoutOutput())
	})
	kern, err = TestKernel(privateAccounts[0], privateAccounts, testConfig)
	require.NoError(t, err)
	err = kern.Boot()
	require.NoError(t, err)
	// Sometimes better to not shutdown as logging errors on shutdown may obscure real issue
	return kern, func() {
		Shutdown(kern)
		cleanup()
	}
}

func NewTestConfig(genesisDoc *genesis.GenesisDoc,
	options ...func(*config.BurrowConfig)) (conf *config.BurrowConfig, cleanup func()) {

	nodeNumber := atomic.AddUint64(&node, 1)
	name := fmt.Sprintf("node_%03d", nodeNumber)
	conf = config.DefaultBurrowConfig()
	conf.Logging = nil
	testDir, cleanup := testutil.EnterTestDirectory()
	conf.BurrowDir = path.Join(testDir, fmt.Sprintf(".burrow_%s", name))
	conf.GenesisDoc = genesisDoc
	conf.Tendermint.Moniker = name
	// Make blocks for purposes of tests
	conf.Tendermint.CreateEmptyBlocks = tendermint.AlwaysCreateEmptyBlocks
	conf.Keys.RemoteAddress = ""
	// Assign run of ports
	const freeport = "0"
	conf.Tendermint.ListenHost = rpc.LocalHost
	conf.Tendermint.ListenPort = freeport
	conf.RPC.GRPC.ListenHost = rpc.LocalHost
	conf.RPC.GRPC.ListenPort = freeport
	conf.RPC.Metrics.ListenHost = rpc.LocalHost
	conf.RPC.Metrics.ListenPort = freeport
	conf.RPC.Info.ListenHost = rpc.LocalHost
	conf.RPC.Info.ListenPort = freeport
	conf.RPC.Web3.ListenHost = rpc.LocalHost
	conf.RPC.Web3.ListenPort = freeport
	conf.Execution.TimeoutFactor = 0.5
	conf.Execution.VMOptions = []execution.VMOption{}
	for _, opt := range options {
		if opt != nil {
			opt(conf)
		}
	}
	return conf, cleanup
}

// We use this to wrap tests
func TestKernel(validatorAccount *acm.PrivateAccount, keysAccounts []*acm.PrivateAccount,
	testConfig *config.BurrowConfig) (*core.Kernel, error) {

	fmt.Println("Creating integration test Kernel...")

	kern, err := core.NewKernel(testConfig.BurrowDir)
	if err != nil {
		return nil, err
	}

	logger := logging.NewNoopLogger()
	kern.SetLogger(logger)
	if testConfig.Logging != nil {
		err := kern.LoadLoggerFromConfig(testConfig.Logging)
		if err != nil {
			return nil, err
		}
	}

	kern.SetKeyClient(keys.NewLocalKeyClient(keys.NewMemoryKeyStore(keysAccounts...), logger))

	err = kern.LoadExecutionOptionsFromConfig(testConfig.Execution)
	if err != nil {
		return nil, err
	}

	err = kern.LoadState(testConfig.GenesisDoc)
	if err != nil {
		return nil, err
	}

	privVal := tendermint.NewPrivValidatorMemory(validatorAccount, validatorAccount)

	err = kern.LoadTendermintFromConfig(testConfig, privVal)
	if err != nil {
		return nil, err
	}

	kern.AddProcesses(core.DefaultProcessLaunchers(kern, testConfig.RPC, testConfig.Keys)...)
	return kern, nil
}

// TestGenesisDoc creates genesis from a set of accounts
// and validators from indices within that slice
func TestGenesisDoc(addressables []*acm.PrivateAccount, vals ...int) *genesis.GenesisDoc {
	accounts := make(map[string]*acm.Account, len(addressables))
	for i, pa := range addressables {
		account := acm.FromAddressable(pa)
		account.Balance += 1 << 32
		account.Permissions = permission.AllAccountPermissions.Clone()
		accounts[fmt.Sprintf("user_%v", i)] = account
	}
	genesisTime, err := time.Parse("02-01-2006", "27-10-2017")
	if err != nil {
		panic("could not parse test genesis time")
	}

	validators := make(map[string]*validator.Validator)
	for _, i := range vals {
		name := fmt.Sprintf("user_%d", i)
		validators[name] = validator.FromAccount(accounts[name], 1<<16)
		// Tendermint validators use a different addressing scheme for secp256k1
		accounts[name].Address = validators[name].GetAddress()
	}

	return genesis.MakeGenesisDocFromAccounts(ChainName, nil, genesisTime, accounts, validators)
}

// Default deterministic account generation helper, pass number of accounts to make
func MakePrivateAccounts(sec string, n int) []*acm.PrivateAccount {
	accounts := make([]*acm.PrivateAccount, n)
	for i := 0; i < n; i++ {
		accounts[i] = acm.GeneratePrivateAccountFromSecret(sec + strconv.Itoa(i))
	}
	return accounts
}

func MakeEthereumAccounts(sec string, n int) []*acm.PrivateAccount {
	accounts := make([]*acm.PrivateAccount, n)
	for i := 0; i < n; i++ {
		accounts[i] = acm.GenerateEthereumAccountFromSecret(sec + strconv.Itoa(i))
	}
	return accounts
}

func Shutdown(kern *core.Kernel) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := kern.Shutdown(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error shutting down test kernel %v: %v", kern, err)
	}
}
