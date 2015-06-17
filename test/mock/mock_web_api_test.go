package mock

// Basic imports
import (
	"bytes"
	"fmt"
	// edb "github.com/eris-ltd/erisdb/erisdb"
	edb "github.com/eris-ltd/erisdb/erisdb"
	ep "github.com/eris-ltd/erisdb/erisdb/pipe"
	"github.com/eris-ltd/erisdb/rpc"
	"github.com/eris-ltd/erisdb/server"
	td "github.com/eris-ltd/erisdb/test/testdata/testdata"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/account"
	"github.com/tendermint/log15"
	"net/http"
	"testing"
	"os"
	"runtime"
	"time"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log15.Root().SetHandler(log15.LvlFilterHandler(
		log15.LvlWarn,
		log15.StreamHandler(os.Stdout, log15.TerminalFormat()),
	))
	gin.SetMode(gin.ReleaseMode)
}

type MockSuite struct {
	suite.Suite
	baseDir      string
	serveProcess *server.ServeProcess
	codec        rpc.Codec
	sUrl         string
	testData     *td.TestData
}

func (this *MockSuite) SetupSuite() {
	gin.SetMode(gin.ReleaseMode)
	// Load the supporting objects.
	testData := td.LoadTestData()
	pipe := NewMockPipe(testData)
	codec := &edb.TCodec{}
	evtSubs := edb.NewEventSubscriptions(pipe.Events())
	// The server
	restServer := edb.NewRestServer(codec, pipe, evtSubs)
	sConf := server.DefaultServerConfig()
	sConf.Bind.Port = 31473
	// Create a server process.
	proc := server.NewServeProcess(sConf, restServer)
	err := proc.Start()
	if err != nil {
		panic(err)
	}
	this.serveProcess = proc
	this.codec = edb.NewTCodec()
	this.testData = testData
	this.sUrl = "http://localhost:31473"
}

func (this *MockSuite) TearDownSuite() {
	sec := this.serveProcess.StopEventChannel()
	this.serveProcess.Stop(0)
	<-sec
	// Tests are done rapidly, this is just to give that extra milliseconds 
	// to shut down the previous server (may be excessive).
	time.Sleep(500*time.Millisecond)
}

// ********************************************* Consensus *********************************************

func (this *MockSuite) Test_A0_ConsensusState() {
	resp := this.get("/consensus")
	ret := &ep.ConsensusState{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	ret.StartTime = ""
	this.Equal(ret, this.testData.Output.ConsensusState)
}

func (this *MockSuite) Test_A1_Validators() {
	resp := this.get("/consensus/validators")
	ret := &ep.ValidatorList{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Validators)
}

// ********************************************* Network *********************************************

func (this *MockSuite) Test_B0_NetworkInfo() {
	resp := this.get("/network")
	ret := &ep.NetworkInfo{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.NetworkInfo)
}

func (this *MockSuite) Test_B1_ClientVersion() {
	resp := this.get("/network/client_version")
	ret := &ep.ClientVersion{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.ClientVersion)
}

func (this *MockSuite) Test_B2_Moniker() {
	resp := this.get("/network/moniker")
	ret := &ep.Moniker{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Moniker)
}

func (this *MockSuite) Test_B3_Listening() {
	resp := this.get("/network/listening")
	ret := &ep.Listening{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Listening)
}

func (this *MockSuite) Test_B4_Listeners() {
	resp := this.get("/network/listeners")
	ret := &ep.Listeners{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Listeners)
}

func (this *MockSuite) Test_B5_Peers() {
	resp := this.get("/network/peers")
	ret := []*ep.Peer{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Peers)
}

// ********************************************* Transactions *********************************************

func (this *MockSuite) Test_C0_TxCreate() {
	resp := this.postJson("/unsafe/txpool", this.testData.Input.TxCreate)
	ret := &ep.Receipt{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.TxCreateReceipt)
}

func (this *MockSuite) Test_C1_Tx() {
	resp := this.postJson("/unsafe/txpool", this.testData.Input.Tx)
	ret := &ep.Receipt{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.TxReceipt)
}

func (this *MockSuite) Test_C2_UnconfirmedTxs() {
	resp := this.get("/txpool")
	ret := &ep.UnconfirmedTxs{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.UnconfirmedTxs)
}

func (this *MockSuite) Test_C3_CallCode() {
	resp := this.postJson("/codecalls", this.testData.Input.CallCode)
	ret := &ep.Call{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.CallCode)
}

// ********************************************* Accounts *********************************************

func (this *MockSuite) Test_D0_Accounts() {
	resp := this.get("/accounts")
	ret := &ep.AccountList{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Accounts)
}

func (this *MockSuite) Test_D1_Account() {
	resp := this.get("/accounts/" + this.testData.Input.AccountAddress)
	ret := &account.Account{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Account)
}

func (this *MockSuite) Test_D2_Storage() {
	resp := this.get("/accounts/" + this.testData.Input.AccountAddress + "/storage")
	ret := &ep.Storage{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Storage)
}

func (this *MockSuite) Test_D3_StorageAt() {
	addr := this.testData.Input.AccountAddress
	key := this.testData.Input.StorageAddress
	resp := this.get("/accounts/" + addr + "/storage/" + key)
	ret := &ep.StorageItem{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.StorageAt)
}

// ********************************************* Blockchain *********************************************

func (this *MockSuite) Test_E0_BlockchainInfo() {
	resp := this.get("/blockchain")
	ret := &ep.BlockchainInfo{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.BlockchainInfo)
}

func (this *MockSuite) Test_E1_ChainId() {
	resp := this.get("/blockchain/chain_id")
	ret := &ep.ChainId{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.ChainId)
}

func (this *MockSuite) Test_E2_GenesisHash() {
	resp := this.get("/blockchain/genesis_hash")
	ret := &ep.GenesisHash{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.GenesisHash)
}

func (this *MockSuite) Test_E3_LatestBlockHeight() {
	resp := this.get("/blockchain/latest_block_height")
	ret := &ep.LatestBlockHeight{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.LatestBlockHeight)
}

func (this *MockSuite) Test_E4_Blocks() {
	br := this.testData.Input.BlockRange
	resp := this.get(fmt.Sprintf("/blockchain/blocks?q=height:%d..%d", br.Min, br.Max))
	ret := &ep.Blocks{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Blocks)
}

// ********************************************* Utilities *********************************************

func (this *MockSuite) get(endpoint string) *http.Response {
	resp, errG := http.Get(this.sUrl + endpoint)
	this.NoError(errG)
	this.Equal(200, resp.StatusCode)
	return resp
}

func (this *MockSuite) postJson(endpoint string, v interface{}) *http.Response {
	bts, errE := this.codec.EncodeBytes(v)
	this.NoError(errE)
	resp, errP := http.Post(this.sUrl+endpoint, "application/json", bytes.NewBuffer(bts))
	this.NoError(errP)
	this.Equal(200, resp.StatusCode)
	return resp
}

// ********************************************* Entrypoint *********************************************

func TestMockSuite(t *testing.T) {
	suite.Run(t, &MockSuite{})
}
