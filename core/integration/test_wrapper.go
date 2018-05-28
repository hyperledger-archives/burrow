// +build integration

// Space above here matters
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
	"os"
	"strconv"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/consensus/tendermint/validator"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys/mock"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/config"
	"github.com/hyperledger/burrow/logging/lifecycle"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/rpc"
	tm_config "github.com/tendermint/tendermint/config"
)

const (
	chainName = "Integration_Test_Chain"
	testDir   = "/tmp/test_scratch/tm_test"
)

// Enable logger output during tests
var debugLogging = false

// We use this to wrap tests
func TestWrapper(privateAccounts []acm.PrivateAccount, genesisDoc *genesis.GenesisDoc, runner func(*core.Kernel) int) int {
	fmt.Println("Running with integration TestWrapper (core/integration/test_wrapper.go)...")

	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.MkdirAll(testDir+"/config", 0777)
	os.Chdir(testDir)

	tmConf := tm_config.DefaultConfig()
	tmConf.RPC.ListenAddress = "tcp://localhost:0"
	tmConf.SetRoot(testDir)

	logger := logging.NewNoopLogger()
	if debugLogging {
		var err error
		// Change config as needed
		logger, err = lifecycle.NewLoggerFromLoggingConfig(&config.LoggingConfig{
			ExcludeTrace: false,
			RootSink: config.Sink().
				SetTransform(config.FilterTransform(config.IncludeWhenAnyMatches,
					structure.ComponentKey, "Tendermint",
					structure.ScopeKey, "executor.Execute\\(tx txs.Tx\\)",
				)).
				//AddSinks(config.Sink().SetTransform(config.FilterTransform(config.ExcludeWhenAnyMatches, "run_call", "false")).
				AddSinks(config.Sink().SetTransform(config.PruneTransform("log_channel", "trace", "scope", "returns", "run_id", "args")).
					AddSinks(config.Sink().SetTransform(config.SortTransform("tx_hash", "time", "message", "method")).
						SetOutput(config.StdoutOutput()))),
		})
		if err != nil {
			panic(err)
		}
	}

	validatorAccount := privateAccounts[0]
	privValidator := validator.NewPrivValidatorMemory(validatorAccount, validatorAccount)
	keyClient := mock.NewKeyClient(privateAccounts...)
	kernel, err := core.NewKernel(context.Background(), keyClient, privValidator, nil, genesisDoc, tmConf, rpc.DefaultRPCConfig(),
		nil, logger)
	if err != nil {
		panic(err)
	}
	// Sometimes better to not shutdown as logging errors on shutdown may obscure real issue
	defer func() {
		//kernel.Shutdown(context.Background())
	}()

	err = kernel.Boot()
	if err != nil {
		panic(err)
	}

	return runner(kernel)
}

func TestGenesisDoc(addressables []acm.PrivateAccount) *genesis.GenesisDoc {
	accounts := make([]acm.Account, len(addressables))
	for i, pa := range addressables {
		account := acm.FromAddressable(pa)
		account.AddToBalance(1 << 32)
		account.SetPermissions(permission.AllAccountPermissions.Clone())
		accounts[i] = account
	}
	validators := make([]acm.Validator, 1)
	validators[0] = acm.NewValidator(accounts[0].PublicKey(), 1000, 1)

	genesisTime, err := time.Parse("02-01-2006", "27-10-2017")
	if err != nil {
		panic("could not parse test genesis time")
	}
	genesisDoc := genesis.MakeGenesisDocFromAccounts(chainName, nil, genesisTime, permission.DefaultAccountPermissions.Clone(),
		accounts, validators)
	return &genesisDoc

}

// Deterministic account generation helper. Pass number of accounts to make
func MakePrivateAccounts(n int) []acm.PrivateAccount {
	accounts := make([]acm.PrivateAccount, n)
	for i := 0; i < n; i++ {
		accounts[i] = acm.GeneratePrivateAccountFromSecret("mysecret" + strconv.Itoa(i))
	}
	return accounts
}
