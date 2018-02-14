// +build integration

// Space above here matters
package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//var logger, _ = lifecycle.NewStdErrLogger()
var logger = loggers.NewNoopInfoTraceLogger()

const monaxKeysBin = "monax-keys"
const keysHost = "localhost"
const keysPort = "56667"
const keysTimeoutSeconds = 3

var rpcString = fmt.Sprintf("http://%s:%s", keysHost, keysPort)

func TestMain(m *testing.M) {
	fmt.Fprint(os.Stderr, "Running monax-keys using test main\n")
	_, err := exec.LookPath(monaxKeysBin)
	if err != nil {
		fatalf("could not run keys integration tests because could not find keys binary: %v", err)
	}

	keysDir, err := ioutil.TempDir("", "key_client_test")
	if err != nil {
		fatalf("could not create temp dir: %v", err)
	}
	cmd := exec.Command(monaxKeysBin, "server", "--dir", keysDir, "--port", keysPort)
	err = cmd.Start()
	if err != nil {
		fatalf("could not start command: %v", err)
	}

	select {
	case <-waitKeysRunning():
		// A plain call to os.Exit will terminate before deferred calls run, so defer that too.
		defer os.Exit(m.Run())
	case <-time.After(keysTimeoutSeconds * time.Second):
		defer fatalf("timed out waiting for monax-keys to become live")
	}

	defer func() {
		err := cmd.Process.Kill()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error killing monax-keys from test main:%v\n", err)
		}
		fmt.Fprint(os.Stderr, "Killed monax-keys from test main\n")
	}()
}

func TestMonaxKeyClient_Generate(t *testing.T) {
	keyClient := keys.NewKeyClient(rpcString, logger)
	addr, err := keyClient.Generate("I'm a lovely hat", keys.KeyTypeEd25519Ripemd160)
	assert.NoError(t, err)
	assert.NotEqual(t, acm.ZeroAddress, addr)
}

func TestMonaxKeyClient_PublicKey(t *testing.T) {
	keyClient := keys.NewKeyClient(rpcString, logger)
	addr, err := keyClient.Generate("I'm a lovely hat", keys.KeyTypeEd25519Ripemd160)
	assert.NoError(t, err)
	pubKey, err := keyClient.PublicKey(addr)
	assert.Equal(t, addr, pubKey.Address())
}

func TestMonaxKeyClient_PublicKey_NonExistent(t *testing.T) {
	keyClient := keys.NewKeyClient(rpcString, logger)
	_, err := keyClient.PublicKey(acm.Address{8, 7, 6, 222})
	assert.Error(t, err)
}

func TestMonaxKeyClient_Sign(t *testing.T) {
	keyClient := keys.NewKeyClient(rpcString, logger)
	addr, err := keyClient.Generate("I'm a lovely hat", keys.KeyTypeEd25519Ripemd160)
	require.NoError(t, err)
	pubKey, err := keyClient.PublicKey(addr)
	assert.NoError(t, err)
	message := []byte("I'm a hat, a hat, a hat")
	signature, err := keyClient.Sign(addr, message)
	assert.NoError(t, err)
	assert.True(t, pubKey.VerifyBytes(message, signature), "signature should verify message")
}

func TestMonaxKeyClient_HealthCheck(t *testing.T) {
	deadKeyClient := keys.NewKeyClient("http://localhost:99999", logger)
	assert.NotNil(t, deadKeyClient.HealthCheck())
	keyClient := keys.NewKeyClient(rpcString, logger)
	assert.Nil(t, keyClient.HealthCheck())
}

func TestPublicKeyAddressAgreement(t *testing.T) {
	keyClient := keys.NewKeyClient(rpcString, logger)
	addr, err := keyClient.Generate("I'm a lovely hat", keys.KeyTypeEd25519Ripemd160)
	require.NoError(t, err)
	pubKey, err := keyClient.PublicKey(addr)
	addrOut := pubKey.Address()
	require.NoError(t, err)
	assert.Equal(t, addr, addrOut)
}

func fatalf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

func waitKeysRunning() chan bool {
	ch := make(chan bool)
	keyClient := keys.NewKeyClient(rpcString, logger)
	go func() {
		for {
			err := keyClient.HealthCheck()
			if err == nil {
				ch <- true
				return
			}
		}

	}()
	return ch
}
