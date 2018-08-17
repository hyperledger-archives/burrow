package keys

import (
	"os"
	"testing"

	burrow_keys "github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitKeyClient(t *testing.T) {
	dirTest := "test_scratch/.keys"
	os.RemoveAll(dirTest)
	go burrow_keys.StartStandAloneServer(dirTest, "localhost", "10997", true, logging.NewNoopLogger())

	localKeyClient, err := InitKeyClient(DefaultKeysURL())
	require.NoError(t, err)
	err = localKeyClient.HealthCheck()
	assert.NoError(t, err)
}
