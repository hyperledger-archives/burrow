// +build integration,ethereum

package ethclient

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/solidity"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/web3"
	"github.com/hyperledger/burrow/tests/web3/web3test"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

var client = web3test.GetChainRPCClient()

func TestEthAccounts(t *testing.T) {
	accounts, err := EthAccounts(client)
	require.NoError(t, err)
	fmt.Println(accounts)
}

func TestEthSendTransaction(t *testing.T) {
	pk := web3test.GetPrivateKey(t)

	d := new(web3.HexDecoder)
	param := &EthSendTransactionParam{
		From: web3.HexEncoder.Address(pk.GetAddress()),
		Gas:  web3.HexEncoder.Uint64(999999),
		Data: web3.HexEncoder.BytesTrim(solidity.Bytecode_EventEmitter),
	}
	txHash, err := EthSendTransaction(client, param)
	require.NoError(t, err)
	require.NotEmpty(t, txHash)

	tx, err := EthGetTransactionByHash(client, txHash)
	require.NoError(t, err)
	assert.Greater(t, d.Uint64(tx.BlockNumber), uint64(0))

	receipt, err := EthGetTransactionReceipt(client, txHash)
	require.NoError(t, err)
	assert.Equal(t, txHash, receipt.TransactionHash)

	require.NoError(t, d.Err())
}

func TestNonExistentTransaction(t *testing.T) {
	txHash := "0x990258f47aba0cf913c14cc101ddf5b589c04765429d5709f643c891442bfcf7"
	receipt, err := EthGetTransactionReceipt(client, txHash)
	require.NoError(t, err)
	require.Equal(t, "", receipt.TransactionHash)
	require.Equal(t, "", receipt.BlockNumber)
	require.Equal(t, "", receipt.BlockHash)
	tx, err := EthGetTransactionByHash(client, txHash)
	require.NoError(t, err)
	require.Equal(t, "", tx.Hash)
	require.Equal(t, "", tx.BlockNumber)
	require.Equal(t, "", tx.BlockHash)
}

func TestEthGetLogs(t *testing.T) {
	// TODO: make this test generate its own fixutres
	filter := &Filter{
		BlockRange: rpcevents.AbsoluteRange(1, 34340),
		Addresses: []crypto.Address{
			crypto.MustAddressFromHexString("a1e378f122fec6aa8c841397042e21bc19368768"),
			crypto.MustAddressFromHexString("f73aaa468496a87675d27638878a1600b0db3c71"),
		},
	}
	result, err := EthGetLogs(client, filter)
	require.NoError(t, err)
	bs, err := json.Marshal(result)
	require.NoError(t, err)
	fmt.Printf("%s\n", string(bs))
}

func TestNetVersion(t *testing.T) {
	chainID, err := NetVersion(client)
	require.NoError(t, err)
	require.NotEmpty(t, chainID)
}

func TestWeb3ClientVersion(t *testing.T) {
	version, err := Web3ClientVersion(client)
	require.NoError(t, err)
	require.NotEmpty(t, version)
}

func TestEthSyncing(t *testing.T) {
	result, err := EthSyncing(client)
	require.NoError(t, err)
	fmt.Printf("%#v\n", result)
}

func TestEthBlockNumber(t *testing.T) {
	height, err := EthBlockNumber(client)
	require.NoError(t, err)
	require.Greater(t, height, uint64(0))
}
