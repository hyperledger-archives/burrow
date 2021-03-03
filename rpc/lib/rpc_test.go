package lib

import (
	"bytes"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/hyperledger/burrow/rpc/lib/jsonrpc"

	"github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/process"

	"github.com/hyperledger/burrow/rpc/lib/server"
	"github.com/hyperledger/burrow/rpc/lib/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cmn "github.com/tendermint/tendermint/libs/bytes"
)

// Client and Server should work over tcp or unix sockets
const (
	host        = "0.0.0.0:47768"
	httpAddress = "http://" + host
	tcpAddress  = "tcp://" + host
)

type ResultEcho struct {
	Value string `json:"value"`
}

type ResultEchoInt struct {
	Value int `json:"value"`
}

type ResultEchoBytes struct {
	Value []byte `json:"value"`
}

type ResultEchoDataBytes struct {
	Value cmn.HexBytes `json:"value"`
}

// Define some routes
var Routes = map[string]*server.RPCFunc{
	"echo":            server.NewRPCFunc(EchoResult, "arg"),
	"echo_bytes":      server.NewRPCFunc(EchoBytesResult, "arg"),
	"echo_data_bytes": server.NewRPCFunc(EchoDataBytesResult, "arg"),
	"echo_int":        server.NewRPCFunc(EchoIntResult, "arg"),
}

func EchoResult(v string) (*ResultEcho, error) {
	return &ResultEcho{v}, nil
}

func EchoIntResult(v int) (*ResultEchoInt, error) {
	return &ResultEchoInt{v}, nil
}

func EchoBytesResult(v []byte) (*ResultEchoBytes, error) {
	return &ResultEchoBytes{v}, nil
}

func EchoDataBytesResult(v cmn.HexBytes) (*ResultEchoDataBytes, error) {
	return &ResultEchoDataBytes{v}, nil
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

// launch unix and tcp servers
func setup() {
	logger, err := logconfig.New().NewLogger()
	if err != nil {
		panic(err)
	}

	tcpLogger := logger.With("socket", "tcp")
	mux := http.NewServeMux()
	server.RegisterRPCFuncs(mux, Routes, tcpLogger)
	go func() {
		l, err := process.ListenerFromAddress(tcpAddress)
		if err != nil {
			panic(err)
		}
		_, err = server.StartHTTPServer(l, mux, tcpLogger)
		if err != nil {
			panic(err)
		}
	}()

	// wait for servers to start
	time.Sleep(time.Second * 2)
}

func echoViaHTTP(cl jsonrpc.HTTPClient, val string) (string, error) {
	params := map[string]interface{}{
		"arg": val,
	}
	result := new(ResultEcho)
	if err := cl.Call("echo", params, result); err != nil {
		return "", err
	}
	return result.Value, nil
}

func echoIntViaHTTP(cl jsonrpc.HTTPClient, val int) (int, error) {
	params := map[string]interface{}{
		"arg": val,
	}
	result := new(ResultEchoInt)
	if err := cl.Call("echo_int", params, result); err != nil {
		return 0, err
	}
	return result.Value, nil
}

func echoBytesViaHTTP(cl jsonrpc.HTTPClient, bytes []byte) ([]byte, error) {
	params := map[string]interface{}{
		"arg": bytes,
	}
	result := new(ResultEchoBytes)
	if err := cl.Call("echo_bytes", params, result); err != nil {
		return []byte{}, err
	}
	return result.Value, nil
}

func echoDataBytesViaHTTP(cl jsonrpc.HTTPClient, bytes cmn.HexBytes) (cmn.HexBytes, error) {
	params := map[string]interface{}{
		"arg": bytes,
	}
	result := new(ResultEchoDataBytes)
	if err := cl.Call("echo_data_bytes", params, result); err != nil {
		return []byte{}, err
	}
	return result.Value, nil
}

func testWithHTTPClient(t *testing.T, cl jsonrpc.HTTPClient) {
	val := "acbd"
	got, err := echoViaHTTP(cl, val)
	require.NoError(t, err)
	assert.Equal(t, val, got)

	val2 := randBytes(t)
	got2, err := echoBytesViaHTTP(cl, val2)
	require.Nil(t, err)
	assert.Equal(t, val2, got2)

	val3 := cmn.HexBytes(randBytes(t))
	got3, err := echoDataBytesViaHTTP(cl, val3)
	require.Nil(t, err)
	assert.Equal(t, val3, got3)

	val4 := rand.Intn(10000)
	got4, err := echoIntViaHTTP(cl, val4)
	require.Nil(t, err)
	assert.Equal(t, val4, got4)
}

func TestServersAndClientsBasic(t *testing.T) {
	serverAddrs := [...]string{httpAddress}
	for _, addr := range serverAddrs {
		cl1 := jsonrpc.NewURIClient(addr)
		fmt.Printf("=== testing server on %s using %v client", addr, cl1)
		testWithHTTPClient(t, cl1)

		cl2 := jsonrpc.NewClient(addr)
		fmt.Printf("=== testing server on %s using %v client", addr, cl2)
		testWithHTTPClient(t, cl2)
	}
}

func TestHexStringArg(t *testing.T) {
	cl := jsonrpc.NewURIClient(httpAddress)
	// should NOT be handled as hex
	val := "0xabc"
	got, err := echoViaHTTP(cl, val)
	require.Nil(t, err)
	assert.Equal(t, got, val)
}

func TestQuotedStringArg(t *testing.T) {
	cl := jsonrpc.NewURIClient(httpAddress)
	// should NOT be unquoted
	val := "\"abc\""
	got, err := echoViaHTTP(cl, val)
	require.Nil(t, err)
	assert.Equal(t, got, val)
}

func TestUnmarshalError(t *testing.T) {
	respString := `{"id":"jsonrpc-client","jsonrpc":"2.0","error":{"message":"Method status not supported.","code":-32000,"data":{"stack":"Error: Method status not supported.\n    at GethApiDouble.handleRequest (/home/silas/code/go/src/github.com/hyperledger/burrow/tests/vent/eth/node_modules/truffle/build/webpack:/node_modules/ganache-core/lib/subproviders/geth_api_double.js:70:1)\n    at next (/home/silas/code/go/src/github.com/hyperledger/burrow/tests/vent/eth/node_modules/truffle/build/webpack:/node_modules/web3-provider-engine/index.js:136:1)\n    at GethDefaults.handleRequest (/home/silas/code/go/src/github.com/hyperledger/burrow/tests/vent/eth/node_modules/truffle/build/webpack:/node_modules/ganache-core/lib/subproviders/gethdefaults.js:15:1)\n    at next (/home/silas/code/go/src/github.com/hyperledger/burrow/tests/vent/eth/node_modules/truffle/build/webpack:/node_modules/web3-provider-engine/index.js:136:1)\n    at SubscriptionSubprovider.FilterSubprovider.handleRequest (/home/silas/code/go/src/github.com/hyperledger/burrow/tests/vent/eth/node_modules/truffle/build/webpack:/node_modules/web3-provider-engine/subproviders/filters.js:89:1)\n    at SubscriptionSubprovider.handleRequest (/home/silas/code/go/src/github.com/hyperledger/burrow/tests/vent/eth/node_modules/truffle/build/webpack:/node_modules/web3-provider-engine/subproviders/subscriptions.js:137:1)\n    at next (/home/silas/code/go/src/github.com/hyperledger/burrow/tests/vent/eth/node_modules/truffle/build/webpack:/node_modules/web3-provider-engine/index.js:136:1)\n    at DelayedBlockFilter.handleRequest (/home/silas/code/go/src/github.com/hyperledger/burrow/tests/vent/eth/node_modules/truffle/build/webpack:/node_modules/ganache-core/lib/subproviders/delayedblockfilter.js:32:1)\n    at next (/home/silas/code/go/src/github.com/hyperledger/burrow/tests/vent/eth/node_modules/truffle/build/webpack:/node_modules/web3-provider-engine/index.js:136:1)\n    at RequestFunnel.handleRequest (/home/silas/code/go/src/github.com/hyperledger/burrow/tests/vent/eth/node_modules/truffle/build/webpack:/node_modules/ganache-core/lib/subproviders/requestfunnel.js:32:1)\n    at next (/home/silas/code/go/src/github.com/hyperledger/burrow/tests/vent/eth/node_modules/truffle/build/webpack:/node_modules/web3-provider-engine/index.js:136:1)\n    at Web3ProviderEngine._handleAsync (/home/silas/code/go/src/github.com/hyperledger/burrow/tests/vent/eth/node_modules/truffle/build/webpack:/node_modules/web3-provider-engine/index.js:123:1)\n    at Timeout._onTimeout (/home/silas/code/go/src/github.com/hyperledger/burrow/tests/vent/eth/node_modules/truffle/build/webpack:/node_modules/web3-provider-engine/index.js:107:1)\n    at listOnTimeout (node:internal/timers:557:17)\n    at processTimers (node:internal/timers:500:7)","name":"Error"}}}`

	response := &types.RPCResponse{}
	err := json.Unmarshal([]byte(respString), response)
	require.NoError(t, err)
}

func randBytes(t *testing.T) []byte {
	n := rand.Intn(10) + 2
	buf := make([]byte, n)
	_, err := crand.Read(buf)
	require.Nil(t, err)
	return bytes.Replace(buf, []byte("="), []byte{100}, -1)
}
