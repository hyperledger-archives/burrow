package rpc_test

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/hyperledger/burrow/encoding"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/project"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/web3"
	"github.com/stretchr/testify/require"
)

func TestWeb3RPCService(t *testing.T) {
	ctx := context.Background()

	genesisAccounts := integration.MakePrivateAccounts("accounts", 2)
	genesisDoc := integration.TestGenesisDoc(genesisAccounts, 0)

	config, _ := integration.NewTestConfig(genesisDoc)
	logger := logging.NewNoopLogger()
	kern, err := integration.TestKernel(genesisAccounts[0], genesisAccounts, config, nil)
	require.NoError(t, err)
	err = kern.Boot()
	defer kern.Shutdown(ctx)

	dir, err := ioutil.TempDir(os.TempDir(), "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	store := keys.NewKeyStore(dir, true)
	for _, acc := range genesisAccounts {
		err = store.StoreKeyPlain(&keys.Key{
			CurveType:  acc.PrivateKey().CurveType,
			Address:    acc.GetAddress(),
			PublicKey:  acc.GetPublicKey(),
			PrivateKey: acc.PrivateKey(),
		})
		require.NoError(t, err)
	}

	nodeView, err := kern.GetNodeView()
	require.NoError(t, err)

	accountState := kern.State
	eventsState := kern.State
	validatorState := kern.State
	eth := rpc.NewEthService(accountState, eventsState, kern.Blockchain, validatorState, nodeView, kern.Transactor, store, kern.Logger)

	t.Run("Web3Sha3", func(t *testing.T) {
		result, err := eth.Web3Sha3(&web3.Web3Sha3Params{"0x68656c6c6f20776f726c64"}) // hello world
		require.NoError(t, err)
		// hex encoded
		require.Equal(t, "0x47173285a8d7341e5e972fc677286384f802f8ef42a5ec5f03bbfa254cb01fad", result.HashedData)
	})

	t.Run("NetListening", func(t *testing.T) {
		result, err := eth.NetListening()
		require.NoError(t, err)
		require.Equal(t, true, result.IsNetListening)
	})

	t.Run("NetPeerCount", func(t *testing.T) {
		result, err := eth.NetPeerCount()
		require.NoError(t, err)
		require.Equal(t, "0x0", result.NumConnectedPeers)
	})

	t.Run("Version+ID", func(t *testing.T) {
		t.Run("Web3ClientVersion", func(t *testing.T) {
			result, err := eth.Web3ClientVersion()
			require.NoError(t, err)
			require.Equal(t, project.FullVersion(), result.ClientVersion)
		})

		t.Run("NetVersion", func(t *testing.T) {
			result, err := eth.NetVersion()
			require.NoError(t, err)
			doc := config.GenesisDoc
			require.Equal(t, encoding.HexEncodeBytes(doc.ShortHash()), result.ChainID)
		})

		t.Run("EthProtocolVersion", func(t *testing.T) {
			result, err := eth.EthProtocolVersion()
			require.NoError(t, err)
			require.NotEmpty(t, result.ProtocolVersion)
		})

		t.Run("EthChainId", func(t *testing.T) {
			result, err := eth.EthChainId()
			require.NoError(t, err)
			require.Equal(t, "Integration_Test_Chain-281EDE", result.ChainId)
		})
	})

	t.Run("EthCreateContract", func(t *testing.T) {
		account := genesisAccounts[1]
		var txHash, contractAddress string

		// create contract on chain
		t.Run("EthSendTransaction", func(t *testing.T) {
			result, err := eth.EthSendTransaction(&web3.EthSendTransactionParams{
				Transaction: web3.Transaction{
					From: encoding.HexEncodeBytes(account.GetAddress().Bytes()),
					Gas:  encoding.HexEncodeNumber(40),
					Data: encoding.HexEncodeBytes(rpc.Bytecode_HelloWorld),
				},
			})
			require.NoError(t, err)
			txHash = result.TransactionHash
			require.NotEmpty(t, txHash)
		})

		// t.Run("EthSendRawTransaction", func(t *testing.T) {
		// 	raw := `0xf866068609184e72a0008303000094fa3caabc8eefec2b5e2895e5afbf79379e7268a7808025a06d35f407f418737eec80cba738c4301e683cfcecf19bac9a1aeb2316cac19d3ba002935ee46e3b6bd69168b0b07670699d71df5b32d5f66dbca5758bce2431c9e8`

		// 	_, err := eth.EthSendRawTransaction(&web3.EthSendRawTransactionParams{
		// 		SignedTransactionData: raw,
		// 	})
		// 	require.NoError(t, err)
		// })

		t.Run("EthGetTransactionReceipt", func(t *testing.T) {
			require.NotEmpty(t, txHash, "need tx hash to get tx receipt")
			result, err := eth.EthGetTransactionReceipt(&web3.EthGetTransactionReceiptParams{
				TransactionHash: txHash,
			})
			require.NoError(t, err)
			contractAddress = result.Receipt.ContractAddress
			require.NotEmpty(t, contractAddress)
		})

		t.Run("EthCall", func(t *testing.T) {
			require.NotEmpty(t, contractAddress, "need contract address to call")

			packed, _, err := abi.EncodeFunctionCall(string(rpc.Abi_HelloWorld), "Hello", logger)
			require.NoError(t, err)

			result, err := eth.EthCall(&web3.EthCallParams{
				Transaction: web3.Transaction{
					From: encoding.HexEncodeBytes(account.GetAddress().Bytes()),
					To:   contractAddress,
					Data: encoding.HexAddPrefix(string(packed)),
				},
			})

			value, err := encoding.HexDecodeToBytes(result.ReturnValue)
			require.NoError(t, err)
			vars, err := abi.DecodeFunctionReturn(string(rpc.Abi_HelloWorld), "Hello", value)
			require.NoError(t, err)
			require.Len(t, vars, 1)
			require.Equal(t, "Hello, World", vars[0].Value)
		})

		t.Run("EthGetCode", func(t *testing.T) {
			require.NotEmpty(t, contractAddress, "need contract address get code")
			result, err := eth.EthGetCode(&web3.EthGetCodeParams{
				Address: contractAddress,
			})
			require.NoError(t, err)
			require.Equal(t, encoding.HexEncodeBytes(rpc.DeployedBytecode_HelloWorld), strings.ToLower(result.Bytes))
		})
	})

	t.Run("EthMining", func(t *testing.T) {
		result, err := eth.EthMining()
		require.NoError(t, err)
		require.Equal(t, true, result.Mining)
	})

	t.Run("EthAccounts", func(t *testing.T) {
		result, err := eth.EthAccounts()
		require.NoError(t, err)
		require.Len(t, result.Addresses, len(genesisAccounts))
		for _, acc := range genesisAccounts {
			require.Contains(t, result.Addresses, encoding.HexEncodeBytes(acc.GetAddress().Bytes()))
		}
	})
}
