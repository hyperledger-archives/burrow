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

package client

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"path"
	"strconv"
	"testing"

	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/evm"
	genesis "github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging/lifecycle"
	ptypes "github.com/hyperledger/burrow/permission"
	rpc_types "github.com/hyperledger/burrow/rpc/tm/types"
	"github.com/hyperledger/burrow/server"
	"github.com/hyperledger/burrow/test/fixtures"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/word"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/rpc/lib/client"
	"github.com/tendermint/tendermint/types"
)

const chainID = "RPC_Test_Chain"

// global variables for use across all tests
var (
	serverConfig      *server.ServerConfig
	rootWorkDir       string
	mempoolCount      = 0
	websocketAddr     string
	genesisDoc        *genesis.GenesisDoc
	websocketEndpoint string
	users             = makeUsers(5) // make keys
	jsonRpcClient     RPCClient
	httpClient        RPCClient
	clients           map[string]RPCClient
	testCore          *core.Core
)

// We use this to wrap tests
func TestWrapper(runner func() int) int {
	fmt.Println("Running with integration TestWrapper (rpc/tendermint/test/shared_test.go)...")
	ffs := fixtures.NewFileFixtures("burrow")

	defer func() {
		// Tendermint likes to try and save to priv_validator.json after its been
		// asked to shutdown so we pause to try and avoid collision
		time.Sleep(time.Second)
		ffs.RemoveAll()
	}()

	evm.SetDebug(true)
	err := initGlobalVariables(ffs)

	if err != nil {
		panic(err)
	}

	tmServer, err := testCore.NewGatewayTendermint(serverConfig)
	defer func() {
		// Shutdown -- make sure we don't hit a race on ffs.RemoveAll
		tmServer.Shutdown()
		testCore.Stop()
	}()

	if err != nil {
		panic(err)
	}

	return runner()
}

// initialize config and create new node
func initGlobalVariables(ffs *fixtures.FileFixtures) error {
	configBytes, err := config.GetConfigurationFileBytes(chainID,
		"test_single_node",
		"",
		"burrow",
		true,
		"46657",
		"burrow serve")
	if err != nil {
		return err
	}

	genesisBytes, err := genesisFileBytesFromUsers(chainID, users)
	if err != nil {
		return err
	}

	testConfigFile := ffs.AddFile("config.toml", string(configBytes))
	rootWorkDir = ffs.AddDir("rootWorkDir")
	if ffs.Error != nil {
		return ffs.Error
	}

	genesisDoc, err = genesis.GenesisDocFromJSON(genesisBytes)
	if err != nil {
		return err
	}

	testConfig := viper.New()
	testConfig.SetConfigFile(testConfigFile)
	err = testConfig.ReadInConfig()

	if err != nil {
		return err
	}

	sconf, err := core.LoadServerConfig(chainID, testConfig)
	if err != nil {
		return err
	}
	serverConfig = sconf

	rpcAddr := serverConfig.Tendermint.RpcLocalAddress
	websocketAddr = rpcAddr
	websocketEndpoint = "/websocket"

	// Set up priv_validator.json before we start tendermint (otherwise it will
	// create its own one.
	saveNewPriv()
	logger, _ := lifecycle.NewStdErrLogger()
	// To spill tendermint logs on the floor:
	// lifecycle.CaptureTendermintLog15Output(loggers.NewNoopInfoTraceLogger())
	lifecycle.CaptureStdlibLogOutput(logger)

	testCore, err = core.NewCore("testCore", logger)
	if err != nil {
		return err
	}

	jsonRpcClient = rpcclient.NewJSONRPCClient(rpcAddr)
	httpClient = rpcclient.NewURIClient(rpcAddr)

	clients = map[string]RPCClient{
		"JSONRPC": jsonRpcClient,
		"HTTP":    httpClient,
	}
	return nil
}

// Deterministic account generation helper. Pass number of accounts to make
func makeUsers(n int) []*acm.ConcretePrivateAccount {
	accounts := []*acm.ConcretePrivateAccount{}
	for i := 0; i < n; i++ {
		secret := "mysecret" + strconv.Itoa(i)
		user := acm.GenPrivAccountFromSecret(secret)
		accounts = append(accounts, user)
	}
	return accounts
}

func genesisFileBytesFromUsers(chainName string, accounts []*acm.ConcretePrivateAccount) ([]byte, error) {
	if len(accounts) < 1 {
		return nil, errors.New("Please pass in at least 1 account to be the validator")
	}
	genesisValidators := make([]*genesis.GenesisValidator, 1)
	genesisAccounts := make([]*genesis.GenesisAccount, len(accounts))
	genesisValidators[0] = genesisValidatorFromPrivAccount(accounts[0])

	for i, acc := range accounts {
		genesisAccounts[i] = genesisAccountFromPrivAccount(acc)
	}

	return genesis.GenerateGenesisFileBytes(chainName, genesisAccounts, genesisValidators)
}

func genesisValidatorFromPrivAccount(account *acm.ConcretePrivateAccount) *genesis.GenesisValidator {
	return &genesis.GenesisValidator{
		Amount: 1000000,
		Name:   fmt.Sprintf("full-account_%s", account.Address),
		PubKey: account.PubKey,
		UnbondTo: []genesis.BasicAccount{
			{
				Address: account.Address,
				Amount:  100,
			},
		},
	}
}

func genesisAccountFromPrivAccount(account *acm.ConcretePrivateAccount) *genesis.GenesisAccount {
	return genesis.NewGenesisAccount(account.Address, 100000,
		fmt.Sprintf("account_%s", account.Address), &ptypes.DefaultAccountPermissions)
}

func saveNewPriv() {
	// Save new priv_validator file.
	priv := &types.PrivValidator{
		Address: users[0].Address.Bytes(),
		PubKey:  users[0].PubKey,
		PrivKey: users[0].PrivKey,
	}
	priv.SetFile(path.Join(rootWorkDir, "priv_validator.json"))
	priv.Save()
}

//-------------------------------------------------------------------------------
// some default transaction functions

func makeDefaultSendTx(t *testing.T, client RPCClient, addr acm.Address,
	amt int64) *txs.SendTx {
	nonce := getNonce(t, client, users[0].Address)
	tx := txs.NewSendTx()
	tx.AddInputWithNonce(users[0].PubKey, amt, nonce+1)
	tx.AddOutput(addr, amt)
	return tx
}

func makeDefaultSendTxSigned(t *testing.T, client RPCClient, addr acm.Address,
	amt int64) *txs.SendTx {
	tx := makeDefaultSendTx(t, client, addr, amt)
	tx.SignInput(chainID, 0, users[0])
	return tx
}

func makeDefaultCallTx(t *testing.T, client RPCClient, addr acm.Address, code []byte, amt, gasLim,
	fee int64) *txs.CallTx {
	nonce := getNonce(t, client, users[0].Address)
	tx := txs.NewCallTxWithNonce(users[0].PubKey, addr, code, amt, gasLim, fee,
		nonce+1)
	tx.Sign(chainID, users[0])
	return tx
}

func makeDefaultNameTx(t *testing.T, client RPCClient, name, value string, amt,
	fee int64) *txs.NameTx {
	nonce := getNonce(t, client, users[0].Address)
	tx := txs.NewNameTxWithNonce(users[0].PubKey, name, value, amt, fee, nonce+1)
	tx.Sign(chainID, users[0])
	return tx
}

//-------------------------------------------------------------------------------
// rpc call wrappers (fail on err)

// get an account's nonce
func getNonce(t *testing.T, client RPCClient, addr acm.Address) int64 {
	ac, err := GetAccount(client, addr)
	if err != nil {
		t.Fatal(err)
	}
	if ac == nil {
		return 0
	}
	return ac.Sequence
}

// get the account
func getAccount(t *testing.T, client RPCClient, addr acm.Address) *acm.ConcreteAccount {
	ac, err := GetAccount(client, addr)
	if err != nil {
		t.Fatal(err)
	}
	return ac
}

// sign transaction
func signTx(t *testing.T, client RPCClient, tx txs.Tx,
	privAcc *acm.ConcretePrivateAccount) txs.Tx {
	signedTx, err := SignTx(client, tx, []*acm.ConcretePrivateAccount{privAcc})
	if err != nil {
		t.Fatal(err)
	}
	return signedTx
}

// broadcast transaction
func broadcastTx(t *testing.T, client RPCClient, tx txs.Tx) txs.Receipt {
	rec, err := BroadcastTx(client, tx)
	if err != nil {
		t.Fatal(err)
	}
	mempoolCount += 1
	return rec
}

// dump all storage for an account. currently unused
func dumpStorage(t *testing.T, addr acm.Address) *rpc_types.ResultDumpStorage {
	client := clients["HTTP"]
	resp, err := DumpStorage(client, addr)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func getStorage(t *testing.T, client RPCClient, addr acm.Address, key []byte) []byte {
	resp, err := GetStorage(client, addr, key)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func callCode(t *testing.T, client RPCClient, fromAddress, code, data,
	expected []byte) {
	resp, err := CallCode(client, fromAddress, code, data)
	if err != nil {
		t.Fatal(err)
	}
	ret := resp.Return
	// NOTE: we don't flip memory when it comes out of RETURN (?!)
	if bytes.Compare(ret, word.LeftPadWord256(expected).Bytes()) != 0 {
		t.Fatalf("Conflicting return value. Got %x, expected %x", ret, expected)
	}
}

func callContract(t *testing.T, client RPCClient, fromAddress, toAddress acm.Address,
	data, expected []byte) {
	resp, err := Call(client, fromAddress, toAddress, data)
	if err != nil {
		t.Fatal(err)
	}
	ret := resp.Return
	// NOTE: we don't flip memory when it comes out of RETURN (?!)
	if bytes.Compare(ret, word.LeftPadWord256(expected).Bytes()) != 0 {
		t.Fatalf("Conflicting return value. Got %x, expected %x", ret, expected)
	}
}

// get the namereg entry
func getNameRegEntry(t *testing.T, client RPCClient, name string) *execution.NameRegEntry {
	entry, err := GetName(client, name)
	if err != nil {
		t.Fatal(err)
	}
	return entry
}

// Returns a positive int64 hash of text (consumers want int64 instead of uint64)
func hashString(text string) int64 {
	hasher := fnv.New64()
	hasher.Write([]byte(text))
	value := int64(hasher.Sum64())
	// Flip the sign if we wrapped
	if value < 0 {
		return -value
	}
	return value
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
		word.RightPadWord256(contractCode).Bytes()...)
	// store it in memory
	code = append(code, []byte{0x60, 0x0, 0x52}...)
	// return whats in memory
	//code = append(code, []byte{0x60, byte(32 - lenCode), 0x60, byte(lenCode), 0xf3}...)
	code = append(code, []byte{0x60, byte(lenCode), 0x60, 0x0, 0xf3}...)
	// return init code, contract code, expected return
	return code, contractCode, word.LeftPadBytes([]byte{0xb}, 32)
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
	contractCode = append(contractCode, addr...)
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
	return code, contractCode, word.LeftPadBytes([]byte{0xb}, 32)
}
