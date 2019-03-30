package keys

import (
	"net"
	"os"
	"testing"

	"github.com/hyperledger/burrow/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitKeyClient(t *testing.T) {
	dirTest := "test_scratch/.keys"
	os.RemoveAll(dirTest)
	server := keys.StandAloneServer(dirTest, true)
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	address := listener.Addr().String()
	go server.Serve(listener)
	localKeyClient, err := InitKeyClient(address)
	require.NoError(t, err)
	err = localKeyClient.HealthCheck()
	assert.NoError(t, err)
}
