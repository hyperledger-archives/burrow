// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys/mock"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/logconfig"
	lConfig "github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/rpc"
)

const (
	ChainName  = "Integration_Test_Chain"
	scratchDir = "test_scratch"
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
	var loggingConfig *logconfig.LoggingConfig

	testConfig, cleanup := NewTestConfig(genesisDoc, options...)
	// Uncomment for log output from tests
	//loggingConfig = logconfig.New().Root(func(sink *logconfig.SinkConfig) *logconfig.SinkConfig {
	//	return sink.SetOutput(logconfig.StderrOutput())
	//})
	kern, err = TestKernel(privateAccounts[0], privateAccounts, testConfig, loggingConfig)
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
	testDir, cleanup := EnterTestDirectory()
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
	testConfig *config.BurrowConfig, loggingConfig *lConfig.LoggingConfig) (*core.Kernel, error) {
	fmt.Println("Creating integration test Kernel...")

	kern, err := core.NewKernel(testConfig.BurrowDir)
	if err != nil {
		return nil, err
	}

	kern.SetLogger(logging.NewNoopLogger())
	if loggingConfig != nil {
		err := kern.LoadLoggerFromConfig(loggingConfig)
		if err != nil {
			return nil, err
		}
	}

	kern.SetKeyClient(mock.NewKeyClient(keysAccounts...))

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

func EnterTestDirectory() (testDir string, cleanup func()) {
	var err error
	testDir, err = ioutil.TempDir("", scratchDir)
	if err != nil {
		panic(fmt.Errorf("could not make temp dir for integration tests: %v", err))
	}
	// If you need to inspectdirs
	//testDir := scratchDir
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.Chdir(testDir)
	os.MkdirAll("config", 0777)
	return testDir, func() { os.RemoveAll(testDir) }
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
	}

	return genesis.MakeGenesisDocFromAccounts(ChainName, nil, genesisTime, accounts, validators)
}

// Deterministic account generation helper. Pass number of accounts to make
func MakePrivateAccounts(sec string, n int) []*acm.PrivateAccount {
	accounts := make([]*acm.PrivateAccount, n)
	for i := 0; i < n; i++ {
		accounts[i] = acm.GeneratePrivateAccountFromSecret(sec + strconv.Itoa(i))
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
