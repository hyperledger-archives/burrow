// +build integration

package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/integration"
)

var privateAccounts = integration.MakePrivateAccounts(5) // make keys
var genesisDoc = integration.TestGenesisDoc(privateAccounts)
var inputAccount = privateAccounts[0]
var testConfig = integration.NewTestConfig(genesisDoc)
var kern *core.Kernel

func TestMain(m *testing.M) {
	_, cleanup := integration.EnterTestDirectory()
	defer cleanup()

	kern = integration.TestKernel(inputAccount, privateAccounts, testConfig, nil)

	err := kern.Boot()
	if err != nil {
		panic(err)
	}
	// Sometimes better to not shutdown as logging errors on shutdown may obscure real issue
	defer func() {
		kern.Shutdown(context.Background())
	}()

	returnValue := m.Run()

	time.Sleep(3 * time.Second)
	os.Exit(returnValue)
}
