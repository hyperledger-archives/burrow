// +build integration

package keys

import (
	"context"
	"fmt"
	"github.com/hyperledger/burrow/keys"
	"os"
	"testing"
	"time"

	"github.com/hyperledger/burrow/integration"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/crypto/sha3"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type hashInfo struct {
	data     string
	expected string
}

var hashData = map[string]hashInfo{
	"sha256":    {"hi", "8F434346648F6B96DF89DDA901C5176B10A6D83961DD3C1AC88B59B2DC327AA4"},
	"ripemd160": {"hi", "242485AB6BFD3502BCB3442EA2E211687B8E4D89"},
}

var (
	KEY_TYPES  = []string{"ed25519", "secp256k1"}
	HASH_TYPES = []string{"sha256", "ripemd160"}
)

func TestMain(m *testing.M) {
	failedCh := make(chan error)
	testDir, cleanup := integration.EnterTestDirectory()
	defer cleanup()
	// start the server
	go func() {
		failedCh <- keys.StartStandAloneServer(testDir, keys.DefaultHost, keys.TestPort, false, logging.NewNoopLogger())
	}()
	ret := m.Run()
	select {
	case err := <-failedCh:
		if err != nil {
			panic(err)
		}
	default:
		os.Exit(ret)
	}
}

func grpcKeysClient() keys.KeysClient {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(keys.DefaultHost+":"+keys.TestPort, opts...)
	if err != nil {
		fmt.Printf("Failed to connect to grpc server: %v\n", err)
		os.Exit(1)
	}
	return keys.NewKeysClient(conn)
}

func testServerKeygenAndPub(t *testing.T, typ string) {
	c := grpcKeysClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	genresp, err := c.GenerateKey(ctx, &keys.GenRequest{CurveType: typ})
	require.NoError(t, err)

	addr := genresp.Address
	resp, err := c.PublicKey(ctx, &keys.PubRequest{Address: addr})
	require.NoError(t, err)

	addrB, err := crypto.AddressFromHexString(addr)
	require.NoError(t, err)

	curveType, err := crypto.CurveTypeFromString(typ)
	require.NoError(t, err)

	publicKey, err := crypto.PublicKeyFromBytes(resp.GetPublicKey(), curveType)
	require.NoError(t, err)
	assert.Equal(t, addrB, publicKey.GetAddress())
}

func TestServerKeygenAndPub(t *testing.T) {
	for _, typ := range KEY_TYPES {
		testServerKeygenAndPub(t, typ)
	}
}

func testServerSignAndVerify(t *testing.T, typ string) {
	c := grpcKeysClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	genresp, err := c.GenerateKey(ctx, &keys.GenRequest{CurveType: typ})
	require.NoError(t, err)

	addr := genresp.Address
	resp, err := c.PublicKey(ctx, &keys.PubRequest{Address: addr})
	require.NoError(t, err)

	hash := sha3.Sha3([]byte("the hash of something!"))

	sig, err := c.Sign(ctx, &keys.SignRequest{Address: addr, Message: hash})
	require.NoError(t, err)

	_, err = c.Verify(ctx, &keys.VerifyRequest{
		Signature: sig.GetSignature(),
		PublicKey: resp.GetPublicKey(),
		Message:   hash,
	})
	require.NoError(t, err)
}

func TestServerSignAndVerify(t *testing.T) {
	for _, typ := range KEY_TYPES {
		testServerSignAndVerify(t, typ)
	}
}

func testServerHash(t *testing.T, typ string) {
	hData := hashData[typ]
	data, expected := hData.data, hData.expected

	c := grpcKeysClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := c.Hash(ctx, &keys.HashRequest{Hashtype: typ, Message: []byte(data)})
	require.NoError(t, err)

	require.Equal(t, expected, resp.GetHash())
}

func TestServerHash(t *testing.T) {
	for _, typ := range HASH_TYPES {
		testServerHash(t, typ)
	}
}
