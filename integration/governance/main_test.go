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

package governance

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/permission"
	"github.com/tendermint/tendermint/p2p"
)

var privateAccounts = integration.MakePrivateAccounts(10) // make keys
var genesisDoc = integration.TestGenesisDoc(privateAccounts)
var _ = integration.ClaimPorts()
var testConfigs []*config.BurrowConfig
var kernels []*core.Kernel

// Needs to be in a _test.go file to be picked up
func TestMain(m *testing.M) {
	cleanup := integration.EnterTestDirectory()
	defer cleanup()
	testConfigs = make([]*config.BurrowConfig, len(privateAccounts))
	kernels = make([]*core.Kernel, len(privateAccounts))
	genesisDoc.Accounts[4].Permissions = permission.NewAccountPermissions(permission.Send | permission.Call)
	for i, acc := range privateAccounts {
		testConfig := integration.NewTestConfig(genesisDoc)
		testConfigs[i] = testConfig
		logconf := logconfig.New().Root(func(sink *logconfig.SinkConfig) *logconfig.SinkConfig {
			return sink.SetTransform(logconfig.FilterTransform(logconfig.IncludeWhenAllMatch,
				"total_validator")).SetOutput(logconfig.StdoutOutput())
		})
		//logconf = logconfig.New()
		//logconf := logconfig.New().Root(func(sink *logconfig.SinkConfig) *logconfig.SinkConfig {
		//	return sink.SetOutput(logconfig.FileOutput(fmt.Sprintf("burrow_%d.log", i)))
		//})
		kernels[i] = integration.TestKernel(acc, privateAccounts, testConfigs[i],
			logconf)
		err := kernels[i].Boot()
		if err != nil {
			panic(err)
		}
		// Sometimes better to not shutdown as logging errors on shutdown may obscure real issue
		defer func() {
			kernels[i].Shutdown(context.Background())
		}()
	}
	time.Sleep(1 * time.Second)
	for i := 0; i < len(kernels); i++ {
		for j := i + 1; j < len(kernels); j++ {
			connectKernels(kernels[i], kernels[j])
		}
	}
	os.Exit(m.Run())
}

func connectKernels(k1, k2 *core.Kernel) {
	k1Address := k1.Node.NodeInfo().NetAddress()
	k2Address := k2.Node.NodeInfo().NetAddress()
	fmt.Printf("Connecting %v -> %v\n", k1Address, k2Address)
	err := k1.Node.Switch().DialPeerWithAddress(k2Address, false)
	if err != nil {
		switch e := err.(type) {
		case p2p.ErrRejected:
			panic(fmt.Errorf("connection between test kernels was rejected: %v", e))
		default:
			panic(fmt.Errorf("could not connect test kernels: %v", err))
		}
	}
}
