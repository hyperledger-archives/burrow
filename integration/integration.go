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
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/execution/evm/sha3"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys/mock"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/lifecycle"
	lConfig "github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/permission"
)

const (
	ChainName = "Integration_Test_Chain"
	testDir   = "./test_scratch/tm_test"
)

// Enable logger output during tests

// Starting point for assigning range of ports for tests
// Start at unprivileged port (hoping for the best)
const startingPort uint16 = 1024

// For each port claimant assign a bucket
const startingPortSeparation uint16 = 10
const startingPortBuckets = 1000

// Mutable port to assign to next claimant
var port = uint32(startingPort)

var node uint64 = 0

// We use this to wrap tests
func TestKernel(validatorAccount *acm.PrivateAccount, keysAccounts []*acm.PrivateAccount,
	testConfig *config.BurrowConfig, loggingConfig *lConfig.LoggingConfig) *core.Kernel {
	fmt.Println("Creating integration test Kernel...")

	logger := logging.NewNoopLogger()
	if loggingConfig != nil {
		var err error
		// Change config as needed
		logger, err = lifecycle.NewLoggerFromLoggingConfig(loggingConfig)
		if err != nil {
			panic(err)
		}
	}

	privValidator := tendermint.NewPrivValidatorMemory(validatorAccount, validatorAccount)
	keyClient := mock.NewKeyClient(keysAccounts...)
	kernel, err := core.NewKernel(context.Background(), keyClient, privValidator,
		testConfig.GenesisDoc,
		testConfig.Tendermint.TendermintConfig(),
		testConfig.RPC,
		testConfig.Keys,
		nil,
		[]execution.ExecutionOption{execution.VMOptions(evm.DebugOpcodes)},
		logger)
	if err != nil {
		panic(err)
	}

	return kernel
}

func EnterTestDirectory() (cleanup func()) {
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.Chdir(testDir)
	os.MkdirAll("config", 0777)
	return func() { os.RemoveAll(testDir) }
}

func TestGenesisDoc(addressables []*acm.PrivateAccount) *genesis.GenesisDoc {
	accounts := make(map[string]acm.Account, len(addressables))
	for i, pa := range addressables {
		account := acm.FromAddressable(pa)
		account.AddToBalance(1 << 32)
		account.SetPermissions(permission.AllAccountPermissions.Clone())
		accounts[fmt.Sprintf("user_%v", i)] = account
	}
	genesisTime, err := time.Parse("02-01-2006", "27-10-2017")
	if err != nil {
		panic("could not parse test genesis time")
	}
	return genesis.MakeGenesisDocFromAccounts(ChainName, nil, genesisTime, accounts,
		map[string]validator.Validator{
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

// Some helpers for setting Burrow's various ports in non-colliding ranges for tests
func ClaimPorts() uint16 {
	_, file, _, _ := runtime.Caller(1)
	startIndex := uint16(binary.LittleEndian.Uint16(sha3.Sha3([]byte(file)))) % startingPortBuckets
	newPort := startingPort + startIndex*startingPortSeparation
	// In case overflow
	if newPort < startingPort {
		newPort += startingPort
	}
	if !atomic.CompareAndSwapUint32(&port, uint32(startingPort), uint32(newPort)) {
		panic("GetPort() called before ClaimPorts() or ClaimPorts() called twice")
	}
	return uint16(atomic.LoadUint32(&port))
}

func GetPort() uint16 {
	return uint16(atomic.AddUint32(&port, 1))
}

// Gets an name based on an incrementing counter for running multiple nodes
func GetName() string {
	nodeNumber := atomic.AddUint64(&node, 1)
	return fmt.Sprintf("node_%03d", nodeNumber)
}

func GetLocalAddress() string {
	return fmt.Sprintf("127.0.0.1:%v", GetPort())
}

func GetTCPLocalAddress() string {
	return fmt.Sprintf("tcp://127.0.0.1:%v", GetPort())
}

func NewTestConfig(genesisDoc *genesis.GenesisDoc) *config.BurrowConfig {
	name := GetName()
	cnf := config.DefaultBurrowConfig()
	cnf.GenesisDoc = genesisDoc
	cnf.Tendermint.Moniker = name
	cnf.Tendermint.TendermintRoot = fmt.Sprintf(".burrow_%s", name)
	cnf.Tendermint.ListenAddress = GetTCPLocalAddress()
	cnf.Tendermint.ExternalAddress = cnf.Tendermint.ListenAddress
	cnf.RPC.GRPC.ListenAddress = GetLocalAddress()
	cnf.RPC.Metrics.ListenAddress = GetTCPLocalAddress()
	cnf.RPC.Info.ListenAddress = GetTCPLocalAddress()
	cnf.Keys.RemoteAddress = ""
	return cnf
}
