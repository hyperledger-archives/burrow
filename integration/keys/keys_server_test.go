package keys

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/crypto/sha3"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type hashInfo struct {
	data     string
	expected string
}

func TestKeysServer(t *testing.T) {
	testDir, cleanup := integration.EnterTestDirectory()
	defer cleanup()
	failedCh := make(chan error)
	server := keys.StandAloneServer(testDir, false)
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	address := listener.Addr().String()
	go func() {
		failedCh <- server.Serve(listener)
	}()
	hashData := map[string]hashInfo{
		"sha256":    {"hi", "8F434346648F6B96DF89DDA901C5176B10A6D83961DD3C1AC88B59B2DC327AA4"},
		"ripemd160": {"hi", "242485AB6BFD3502BCB3442EA2E211687B8E4D89"},
	}
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	require.NoError(t, err)
	cli := keys.NewKeysClient(conn)

	t.Run("Group", func(t *testing.T) {
		for _, typ := range []string{"ed25519", "secp256k1"} {
			t.Run("KeygenAndPub", func(t *testing.T) {
				t.Parallel()
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				genresp, err := cli.GenerateKey(ctx, &keys.GenRequest{CurveType: typ})
				require.NoError(t, err)

				addr := genresp.Address
				resp, err := cli.PublicKey(ctx, &keys.PubRequest{Address: addr})
				require.NoError(t, err)

				addrB, err := crypto.AddressFromHexString(addr)
				require.NoError(t, err)

				curveType, err := crypto.CurveTypeFromString(typ)
				require.NoError(t, err)

				publicKey, err := crypto.PublicKeyFromBytes(resp.GetPublicKey(), curveType)
				require.NoError(t, err)
				assert.Equal(t, addrB, publicKey.GetAddress())
			})

			t.Run("SignAndVerify", func(t *testing.T) {
				t.Parallel()
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				genresp, err := cli.GenerateKey(ctx, &keys.GenRequest{CurveType: typ})
				require.NoError(t, err)

				addr := genresp.Address
				resp, err := cli.PublicKey(ctx, &keys.PubRequest{Address: addr})
				require.NoError(t, err)

				msg := []byte("the hash of something!")
				hash := sha3.Sha3(msg)

				sig, err := cli.Sign(ctx, &keys.SignRequest{Address: addr, Message: hash})
				require.NoError(t, err)

				_, err = cli.Verify(ctx, &keys.VerifyRequest{
					Signature: sig.GetSignature(),
					PublicKey: resp.GetPublicKey(),
					Message:   msg,
				})
				require.NoError(t, err)
			})

		}
		for _, typ := range []string{"sha256", "ripemd160"} {
			t.Run("Hash", func(t *testing.T) {
				t.Parallel()
				hData := hashData[typ]
				data, expected := hData.data, hData.expected

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				resp, err := cli.Hash(ctx, &keys.HashRequest{Hashtype: typ, Message: []byte(data)})
				require.NoError(t, err)

				require.Equal(t, expected, resp.GetHash())
			})
		}
	})
	select {
	case err := <-failedCh:
		require.NoError(t, err)
	default:
	}
}
