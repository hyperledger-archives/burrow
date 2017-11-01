package keys

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/monax/keys/monax-keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//var logger, _ = lifecycle.NewStdErrLogger()
var logger = loggers.NewNoopInfoTraceLogger()

const keysHost = "localhost"
const keysPort = "56757"

var rpcString = fmt.Sprintf("http://%s:%s", keysHost, keysPort)

func TestMain(m *testing.M) {
	var err error
	keys.KeysDir, err = ioutil.TempDir("", "key_client_test")
	if err != nil {
		fatalf("couldn't create temp dir: %v", err)
	}
	go keys.StartServer(keysHost, keysPort)
	m.Run()

}

func TestMonaxKeyClient_Generate(t *testing.T) {
	keyClient := NewBurrowKeyClient(rpcString, logger)
	addr, err := keyClient.Generate("I'm a lovely hat", KeyTypeEd25519Ripemd160)
	assert.NoError(t, err)
	assert.NotEqual(t, acm.ZeroAddress, addr)
}

func TestMonaxKeyClient_PublicKey(t *testing.T) {
	keyClient := NewBurrowKeyClient(rpcString, logger)
	addr, err := keyClient.Generate("I'm a lovely hat", KeyTypeEd25519Ripemd160)
	assert.NoError(t, err)
	pubKey, err := keyClient.PublicKey(addr)
	assert.Equal(t, addr[:], pubKey.Address())
}

func TestMonaxKeyClient_Sign(t *testing.T) {
	keyClient := NewBurrowKeyClient(rpcString, logger)
	addr, err := keyClient.Generate("I'm a lovely hat", KeyTypeEd25519Ripemd160)
	require.NoError(t, err)
	pubKey, err := keyClient.PublicKey(addr)
	assert.NoError(t, err)
	message := []byte("I'm a hat, a hat, a hat")
	signature, err := keyClient.Sign(addr, message)
	assert.NoError(t, err)
	assert.True(t, pubKey.VerifyBytes(message, signature), "signature should verify message")
}

func fatalf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
