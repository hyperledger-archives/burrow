package test

import (
	"bytes"
	"hash/fnv"
	"path"
	"strconv"
	"testing"

	acm "github.com/eris-ltd/eris-db/account"
	"github.com/eris-ltd/eris-db/core"
	core_types "github.com/eris-ltd/eris-db/core/types"
	"github.com/eris-ltd/eris-db/logging/lifecycle"
	edbcli "github.com/eris-ltd/eris-db/rpc/tendermint/client"
	rpc_core "github.com/eris-ltd/eris-db/rpc/tendermint/core"
	rpc_types "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
	"github.com/eris-ltd/eris-db/server"
	"github.com/eris-ltd/eris-db/test/fixtures"
	"github.com/eris-ltd/eris-db/txs"
	"github.com/eris-ltd/eris-db/word256"

	state_types "github.com/eris-ltd/eris-db/manager/eris-mint/state/types"
	"github.com/spf13/viper"
	"github.com/tendermint/go-crypto"
	rpcclient "github.com/tendermint/go-rpc/client"
	"github.com/tendermint/tendermint/types"
)

// global variables for use across all tests
var (
	config            = server.DefaultServerConfig()
	rootWorkDir       string
	mempoolCount      = 0
	chainID           string
	websocketAddr     string
	genesisDoc        *state_types.GenesisDoc
	websocketEndpoint string
	user              = makeUsers(5) // make keys
	jsonRpcClient     rpcclient.Client
	httpClient        rpcclient.Client
	clients           map[string]rpcclient.Client
	testCore          *core.Core
)

// initialize config and create new node
func initGlobalVariables(ffs *fixtures.FileFixtures) error {
	testConfigFile := ffs.AddFile("config.toml", defaultConfig)
	rootWorkDir = ffs.AddDir("rootWorkDir")
	rootDataDir := ffs.AddDir("rootDataDir")
	genesisFile := ffs.AddFile("rootWorkDir/genesis.json", defaultGenesis)
	genesisDoc = state_types.GenesisDocFromJSON([]byte(defaultGenesis))

	if ffs.Error != nil {
		return ffs.Error
	}

	testConfig := viper.New()
	testConfig.SetConfigFile(testConfigFile)
	err := testConfig.ReadInConfig()

	if err != nil {
		return err
	}

	chainID = testConfig.GetString("chain.assert_chain_id")
	rpcAddr := config.Tendermint.RpcLocalAddress
	websocketAddr = rpcAddr
	websocketEndpoint = "/websocket"

	consensusConfig, err := core.LoadModuleConfig(testConfig, rootWorkDir,
		rootDataDir, genesisFile, chainID, "consensus")
	if err != nil {
		return err
	}

	managerConfig, err := core.LoadModuleConfig(testConfig, rootWorkDir,
		rootDataDir, genesisFile, chainID, "manager")
	if err != nil {
		return err
	}

	// Set up priv_validator.json before we start tendermint (otherwise it will
	// create its own one.
	saveNewPriv()
	logger := lifecycle.NewStdErrLogger()
	// To spill tendermint logs on the floor:
	// lifecycle.CaptureTendermintLog15Output(loggers.NewNoopInfoTraceLogger())
	lifecycle.CaptureTendermintLog15Output(logger)
	lifecycle.CaptureStdlibLogOutput(logger)

	testCore, err = core.NewCore("testCore", consensusConfig, managerConfig,
		logger)
	if err != nil {
		return err
	}

	jsonRpcClient = rpcclient.NewClientJSONRPC(rpcAddr)
	httpClient = rpcclient.NewClientURI(rpcAddr)

	clients = map[string]rpcclient.Client{
		"JSONRPC": jsonRpcClient,
		"HTTP":    httpClient,
	}
	return nil
}

// deterministic account generation, synced with genesis file in config/tendermint_test/config.go
func makeUsers(n int) []*acm.PrivAccount {
	accounts := []*acm.PrivAccount{}
	for i := 0; i < n; i++ {
		secret := ("mysecret" + strconv.Itoa(i))
		user := acm.GenPrivAccountFromSecret(secret)
		accounts = append(accounts, user)
	}
	return accounts
}

func newNode(ready chan error,
	tmServer chan *rpc_core.TendermintWebsocketServer) {
	// Run the 'tendermint' rpc server
	server, err := testCore.NewGatewayTendermint(config)
	ready <- err
	tmServer <- server
}

func saveNewPriv() {
	// Save new priv_validator file.
	priv := &types.PrivValidator{
		Address: user[0].Address,
		PubKey:  crypto.PubKeyEd25519(user[0].PubKey.(crypto.PubKeyEd25519)),
		PrivKey: crypto.PrivKeyEd25519(user[0].PrivKey.(crypto.PrivKeyEd25519)),
	}
	priv.SetFile(path.Join(rootWorkDir, "priv_validator.json"))
	priv.Save()
}

//-------------------------------------------------------------------------------
// some default transaction functions

func makeDefaultSendTx(t *testing.T, client rpcclient.Client, addr []byte,
	amt int64) *txs.SendTx {
	nonce := getNonce(t, client, user[0].Address)
	tx := txs.NewSendTx()
	tx.AddInputWithNonce(user[0].PubKey, amt, nonce+1)
	tx.AddOutput(addr, amt)
	return tx
}

func makeDefaultSendTxSigned(t *testing.T, client rpcclient.Client, addr []byte,
	amt int64) *txs.SendTx {
	tx := makeDefaultSendTx(t, client, addr, amt)
	tx.SignInput(chainID, 0, user[0])
	return tx
}

func makeDefaultCallTx(t *testing.T, client rpcclient.Client, addr, code []byte, amt, gasLim,
	fee int64) *txs.CallTx {
	nonce := getNonce(t, client, user[0].Address)
	tx := txs.NewCallTxWithNonce(user[0].PubKey, addr, code, amt, gasLim, fee,
		nonce+1)
	tx.Sign(chainID, user[0])
	return tx
}

func makeDefaultNameTx(t *testing.T, client rpcclient.Client, name, value string, amt,
	fee int64) *txs.NameTx {
	nonce := getNonce(t, client, user[0].Address)
	tx := txs.NewNameTxWithNonce(user[0].PubKey, name, value, amt, fee, nonce+1)
	tx.Sign(chainID, user[0])
	return tx
}

//-------------------------------------------------------------------------------
// rpc call wrappers (fail on err)

// get an account's nonce
func getNonce(t *testing.T, client rpcclient.Client, addr []byte) int {
	ac, err := edbcli.GetAccount(client, addr)
	if err != nil {
		t.Fatal(err)
	}
	if ac == nil {
		return 0
	}
	return ac.Sequence
}

// get the account
func getAccount(t *testing.T, client rpcclient.Client, addr []byte) *acm.Account {
	ac, err := edbcli.GetAccount(client, addr)
	if err != nil {
		t.Fatal(err)
	}
	return ac
}

// sign transaction
func signTx(t *testing.T, client rpcclient.Client, tx txs.Tx,
	privAcc *acm.PrivAccount) txs.Tx {
	signedTx, err := edbcli.SignTx(client, tx, []*acm.PrivAccount{privAcc})
	if err != nil {
		t.Fatal(err)
	}
	return signedTx
}

// broadcast transaction
func broadcastTx(t *testing.T, client rpcclient.Client, tx txs.Tx) txs.Receipt {
	rec, err := edbcli.BroadcastTx(client, tx)
	if err != nil {
		t.Fatal(err)
	}
	mempoolCount += 1
	return rec
}

// dump all storage for an account. currently unused
func dumpStorage(t *testing.T, addr []byte) *rpc_types.ResultDumpStorage {
	client := clients["HTTP"]
	resp, err := edbcli.DumpStorage(client, addr)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func getStorage(t *testing.T, client rpcclient.Client, addr, key []byte) []byte {
	resp, err := edbcli.GetStorage(client, addr, key)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func callCode(t *testing.T, client rpcclient.Client, fromAddress, code, data,
	expected []byte) {
	resp, err := edbcli.CallCode(client, fromAddress, code, data)
	if err != nil {
		t.Fatal(err)
	}
	ret := resp.Return
	// NOTE: we don't flip memory when it comes out of RETURN (?!)
	if bytes.Compare(ret, word256.LeftPadWord256(expected).Bytes()) != 0 {
		t.Fatalf("Conflicting return value. Got %x, expected %x", ret, expected)
	}
}

func callContract(t *testing.T, client rpcclient.Client, fromAddress, toAddress,
	data, expected []byte) {
	resp, err := edbcli.Call(client, fromAddress, toAddress, data)
	if err != nil {
		t.Fatal(err)
	}
	ret := resp.Return
	// NOTE: we don't flip memory when it comes out of RETURN (?!)
	if bytes.Compare(ret, word256.LeftPadWord256(expected).Bytes()) != 0 {
		t.Fatalf("Conflicting return value. Got %x, expected %x", ret, expected)
	}
}

// get the namereg entry
func getNameRegEntry(t *testing.T, client rpcclient.Client, name string) *core_types.NameRegEntry {
	entry, err := edbcli.GetName(client, name)
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
		word256.RightPadWord256(contractCode).Bytes()...)
	// store it in memory
	code = append(code, []byte{0x60, 0x0, 0x52}...)
	// return whats in memory
	//code = append(code, []byte{0x60, byte(32 - lenCode), 0x60, byte(lenCode), 0xf3}...)
	code = append(code, []byte{0x60, byte(lenCode), 0x60, 0x0, 0xf3}...)
	// return init code, contract code, expected return
	return code, contractCode, word256.LeftPadBytes([]byte{0xb}, 32)
}

// simple call contract calls another contract
func simpleCallContract(addr []byte) ([]byte, []byte, []byte) {
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
	return code, contractCode, word256.LeftPadBytes([]byte{0xb}, 32)
}
