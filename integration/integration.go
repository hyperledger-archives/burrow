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
)

const (
	ChainName         = "Integration_Test_Chain"
	scratchDir        = "test_scratch"
	runningInCIEnvVar = "CI"
)

// Enable logger output during tests

var node uint64 = 0

func RunningInCI() bool {
	_, ok := os.LookupEnv(runningInCIEnvVar)
	return ok
}

func NoConsensus(conf *config.BurrowConfig) {
	conf.Tendermint.Enabled = false
}

func RunNode(t testing.TB, genesisDoc *genesis.GenesisDoc, privateAccounts []*acm.PrivateAccount,
	options ...func(*config.BurrowConfig)) (kern *core.Kernel, shutdown func()) {
	var err error
	var loggingConfig *logconfig.LoggingConfig
	testConfig, cleanup := NewTestConfig(genesisDoc)
	for _, opt := range options {
		opt(testConfig)
	}
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
		kern.Shutdown(context.Background())
		cleanup()
	}
}

func NewTestConfig(genesisDoc *genesis.GenesisDoc) (conf *config.BurrowConfig, cleanup func()) {
	nodeNumber := atomic.AddUint64(&node, 1)
	name := fmt.Sprintf("node_%03d", nodeNumber)
	conf = config.DefaultBurrowConfig()
	testDir, cleanup := EnterTestDirectory()
	conf.BurrowDir = path.Join(testDir, fmt.Sprintf(".burrow_%s", name))
	conf.GenesisDoc = genesisDoc
	conf.Tendermint.Moniker = name
	conf.Keys.RemoteAddress = ""
	// Assign run of ports
	const localhostFreePort = "tcp://localhost:0"
	conf.Tendermint.ListenAddress = localhostFreePort
	conf.RPC.GRPC.ListenAddress = localhostFreePort
	conf.RPC.Metrics.ListenAddress = localhostFreePort
	conf.RPC.Info.ListenAddress = localhostFreePort
	conf.Execution.TimeoutFactor = 0.3
	conf.Execution.VMOptions = []execution.VMOption{execution.DebugOpcodes}
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

func TestGenesisDoc(addressables []*acm.PrivateAccount) *genesis.GenesisDoc {
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
	return genesis.MakeGenesisDocFromAccounts(ChainName, nil, genesisTime, accounts,
		map[string]*validator.Validator{
			"genesis_validator": validator.FromAccount(accounts["user_0"], 1<<16),
		})
}

// Deterministic account generation helper. Pass number of accounts to make
func MakePrivateAccounts(n int) []*acm.PrivateAccount {
	accounts := make([]*acm.PrivateAccount, n)
	for i := 0; i < n; i++ {
		accounts[i] = acm.GeneratePrivateAccountFromSecret("mysecret" + strconv.Itoa(i))
	}
	return accounts
}
