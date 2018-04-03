// +build integration

// Space above here matters
// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"strconv"
	"testing"

	"os"

	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/consensus/tendermint/validator"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/config"
	"github.com/hyperledger/burrow/logging/lifecycle"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/rpc"
	tm_client "github.com/hyperledger/burrow/rpc/tm/client"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/require"
	tm_config "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/rpc/lib/client"
)

const (
	chainName         = "RPC_Test_Chain"
	rpcAddr           = "0.0.0.0:46657"
	websocketAddr     = rpcAddr
	websocketEndpoint = "/websocket"
	testDir           = "./test_scratch/tm_test"
)

// Enable logger output during tests
var debugLogging = true

// global variables for use across all tests
var (
	privateAccounts = makePrivateAccounts(5) // make keys
	jsonRpcClient   = rpcclient.NewJSONRPCClient(rpcAddr)
	httpClient      = rpcclient.NewURIClient(rpcAddr)
	clients         = map[string]tm_client.RPCClient{
		"JSONRPC": jsonRpcClient,
		"HTTP":    httpClient,
	}
	// Initialised in initGlobalVariables
	genesisDoc = new(genesis.GenesisDoc)
)

// We use this to wrap tests
func TestWrapper(runner func() int) int {
	fmt.Println("Running with integration TestWrapper (rpc/tm/integration/shared.go)...")

	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.Chdir(testDir)

	tmConf := tm_config.DefaultConfig()
	logger := logging.NewNoopLogger()
	if debugLogging {
		var err error
		// Change config as needed
		logger, err = lifecycle.NewLoggerFromLoggingConfig(&config.LoggingConfig{
			RootSink: config.Sink().
				SetTransform(config.FilterTransform(config.IncludeWhenAnyMatches,
					//"","",
					"method", "GetAccount",
					"method", "BroadcastTx",
					"tag", "sequence",
					"tag", "Commit",
					"tag", "CheckTx",
					"tag", "DeliverTx",
				)).
				//AddSinks(config.Sink().SetTransform(config.FilterTransform(config.ExcludeWhenAnyMatches, "run_call", "false")).
				AddSinks(config.Sink().SetTransform(config.PruneTransform("log_channel", "trace", "scope", "returns", "run_id", "args")).
					AddSinks(config.Sink().SetTransform(config.SortTransform("tx_hash", "time", "message", "method")).
						SetOutput(config.StdoutOutput()))),
		})
		if err != nil {
			panic(err)
		}
	}

	privValidator := validator.NewPrivValidatorMemory(privateAccounts[0], privateAccounts[0])
	genesisDoc = testGenesisDoc()
	kernel, err := core.NewKernel(context.Background(), privValidator, genesisDoc, tmConf, rpc.DefaultRPCConfig(),
		nil, logger)
	if err != nil {
		panic(err)
	}
	// Sometimes better to not shutdown as logging errors on shutdown may obscure real issue
	defer func() {
		//kernel.Shutdown(context.Background())
	}()

	err = kernel.Boot()
	if err != nil {
		panic(err)
	}

	return runner()
}

func testGenesisDoc() *genesis.GenesisDoc {
	accounts := make(map[string]acm.Account, len(privateAccounts))
	for i, pa := range privateAccounts {
		account := acm.FromAddressable(pa)
		account.AddToBalance(1 << 32)
		account.SetPermissions(permission.AllAccountPermissions.Clone())
		accounts[fmt.Sprintf("user_%v", i)] = account
	}
	genesisTime, err := time.Parse("02-01-2006", "27-10-2017")
	if err != nil {
		panic("could not parse test genesis time")
	}
	return genesis.MakeGenesisDocFromAccounts(chainName, nil, genesisTime, accounts,
		map[string]acm.Validator{
			"genesis_validator": acm.AsValidator(accounts["user_0"]),
		})
}

// Deterministic account generation helper. Pass number of accounts to make
func makePrivateAccounts(n int) []acm.SigningAccount {
	accounts := make([]acm.SigningAccount, n)
	for i := 0; i < n; i++ {
		accounts[i] = acm.GeneratePrivateAccountFromSecret("mysecret" + strconv.Itoa(i))
	}
	return accounts
}

//-------------------------------------------------------------------------------
// some default transaction functions

func makeDefaultSendTx(t *testing.T, client tm_client.RPCClient, addr acm.Address, amt uint64) *txs.SendTx {
	sequence := getSequence(t, client, privateAccounts[0].Address())
	tx := txs.NewSendTx()
	tx.AddInputWithSequence(privateAccounts[0].PublicKey(), amt, sequence+1)
	tx.AddOutput(addr, amt)
	return tx
}

func makeDefaultSendTxSigned(t *testing.T, client tm_client.RPCClient, addr acm.Address, amt uint64) *txs.SendTx {
	tx := makeDefaultSendTx(t, client, addr, amt)
	tx.SignInput(genesisDoc.ChainID(), 0, privateAccounts[0])
	return tx
}

func makeDefaultCallTx(t *testing.T, client tm_client.RPCClient, addr *acm.Address, code []byte, amt, gasLim,
	fee uint64) *txs.CallTx {
	sequence := getSequence(t, client, privateAccounts[0].Address())
	tx := txs.NewCallTxWithSequence(privateAccounts[0].PublicKey(), addr, code, amt, gasLim, fee,
		sequence+1)
	tx.Sign(genesisDoc.ChainID(), privateAccounts[0])
	return tx
}

func makeDefaultNameTx(t *testing.T, client tm_client.RPCClient, name, value string, amt, fee uint64) *txs.NameTx {
	sequence := getSequence(t, client, privateAccounts[0].Address())
	tx := txs.NewNameTxWithSequence(privateAccounts[0].PublicKey(), name, value, amt, fee, sequence+1)
	tx.Sign(genesisDoc.ChainID(), privateAccounts[0])
	return tx
}

//-------------------------------------------------------------------------------
// rpc call wrappers (fail on err)

// get an account's sequence number
func getSequence(t *testing.T, client tm_client.RPCClient, addr acm.Address) uint64 {
	acc, err := tm_client.GetAccount(client, addr)
	if err != nil {
		t.Fatal(err)
	}
	if acc == nil {
		return 0
	}
	return acc.Sequence()
}

// get the account
func getAccount(t *testing.T, client tm_client.RPCClient, addr acm.Address) acm.Account {
	ac, err := tm_client.GetAccount(client, addr)
	if err != nil {
		t.Fatal(err)
	}
	return ac
}

// sign transaction
func signTx(t *testing.T, client tm_client.RPCClient, tx txs.Tx,
	privAcc *acm.ConcretePrivateAccount) txs.Tx {
	signedTx, err := tm_client.SignTx(client, tx, []*acm.ConcretePrivateAccount{privAcc})
	if err != nil {
		t.Fatal(err)
	}
	return signedTx
}

// broadcast transaction
func broadcastTx(t *testing.T, client tm_client.RPCClient, tx txs.Tx) *txs.Receipt {
	rec, err := tm_client.BroadcastTx(client, tx)
	require.NoError(t, err)
	return rec
}

// dump all storage for an account. currently unused
func dumpStorage(t *testing.T, addr acm.Address) *rpc.ResultDumpStorage {
	client := clients["HTTP"]
	resp, err := tm_client.DumpStorage(client, addr)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func getStorage(t *testing.T, client tm_client.RPCClient, addr acm.Address, key []byte) []byte {
	resp, err := tm_client.GetStorage(client, addr, key)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func callCode(t *testing.T, client tm_client.RPCClient, fromAddress acm.Address, code, data,
	expected []byte) {
	resp, err := tm_client.CallCode(client, fromAddress, code, data)
	if err != nil {
		t.Fatal(err)
	}
	ret := resp.Return
	// NOTE: we don't flip memory when it comes out of RETURN (?!)
	if bytes.Compare(ret, binary.LeftPadWord256(expected).Bytes()) != 0 {
		t.Fatalf("Conflicting return value. Got %x, expected %x", ret, expected)
	}
}

func callContract(t *testing.T, client tm_client.RPCClient, fromAddress, toAddress acm.Address,
	data, expected []byte) {
	resp, err := tm_client.Call(client, fromAddress, toAddress, data)
	if err != nil {
		t.Fatal(err)
	}
	ret := resp.Return
	// NOTE: we don't flip memory when it comes out of RETURN (?!)
	if bytes.Compare(ret, binary.LeftPadWord256(expected).Bytes()) != 0 {
		t.Fatalf("Conflicting return value. Got %x, expected %x", ret, expected)
	}
}

// get the namereg entry
func getNameRegEntry(t *testing.T, client tm_client.RPCClient, name string) *execution.NameRegEntry {
	entry, err := tm_client.GetName(client, name)
	if err != nil {
		t.Fatal(err)
	}
	return entry
}

// Returns a positive int64 hash of text (consumers want int64 instead of uint64)
func hashString(text string) uint64 {
	hasher := fnv.New64()
	hasher.Write([]byte(text))
	return uint64(hasher.Sum64())
}

//--------------------------------------------------------------------------------
// utility verification function

// simple contract returns 5 + 6 = 0xb
func simpleContract() ([]byte, []byte, []byte) {
	// this is the code we want to run when the contract is called
	contractCode := []byte{0x60, 0x5, 0x60, 0x6, 0x1, 0x60, 0x0, 0x52, 0x60, 0x20,
		0x60, 0x0, 0xf3}
	// the is the code we need to return the contractCode when the contract is initialized
	lenCode := len(contractCode)
	// push code to the stack
	//code := append([]byte{byte(0x60 + lenCode - 1)}, RightPadWord256(contractCode).Bytes()...)
	code := append([]byte{0x7f},
		binary.RightPadWord256(contractCode).Bytes()...)
	// store it in memory
	code = append(code, []byte{0x60, 0x0, 0x52}...)
	// return whats in memory
	//code = append(code, []byte{0x60, byte(32 - lenCode), 0x60, byte(lenCode), 0xf3}...)
	code = append(code, []byte{0x60, byte(lenCode), 0x60, 0x0, 0xf3}...)
	// return init code, contract code, expected return
	return code, contractCode, binary.LeftPadBytes([]byte{0xb}, 32)
}

// simple call contract calls another contract
func simpleCallContract(addr acm.Address) ([]byte, []byte, []byte) {
	gas1, gas2 := byte(0x1), byte(0x1)
	value := byte(0x1)
	inOff, inSize := byte(0x0), byte(0x0) // no call data
	retOff, retSize := byte(0x0), byte(0x20)
	// this is the code we want to run (call a contract and return)
	contractCode := []byte{0x60, retSize, 0x60, retOff, 0x60, inSize, 0x60, inOff,
		0x60, value, 0x73}
	contractCode = append(contractCode, addr.Bytes()...)
	contractCode = append(contractCode, []byte{0x61, gas1, gas2, 0xf1, 0x60, 0x20,
		0x60, 0x0, 0xf3}...)

	// the is the code we need to return; the contractCode when the contract is initialized
	// it should copy the code from the input into memory
	lenCode := len(contractCode)
	memOff := byte(0x0)
	inOff = byte(0xc) // length of code before codeContract
	length := byte(lenCode)

	code := []byte{0x60, length, 0x60, inOff, 0x60, memOff, 0x37}
	// return whats in memory
	code = append(code, []byte{0x60, byte(lenCode), 0x60, 0x0, 0xf3}...)
	code = append(code, contractCode...)
	// return init code, contract code, expected return
	return code, contractCode, binary.LeftPadBytes([]byte{0xb}, 32)
}
