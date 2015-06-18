package web_api

// Basic imports
import (
	"bytes"
	"fmt"
	// edb "github.com/eris-ltd/erisdb/erisdb"
	edb "github.com/eris-ltd/eris-db/erisdb"
	ess "github.com/eris-ltd/eris-db/erisdb/erisdbss"
	ep "github.com/eris-ltd/eris-db/erisdb/pipe"
	"github.com/eris-ltd/eris-db/rpc"
	"github.com/eris-ltd/eris-db/server"
	td "github.com/eris-ltd/eris-db/test/testdata/testdata"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/account"
	"net/http"
	"os"
	"path"
	"testing"
)

const WAPIS_URL = "http://localhost:31404/server"

type WebApiSuite struct {
	suite.Suite
	baseDir      string
	serveProcess *server.ServeProcess
	codec        rpc.Codec
	sUrl         string
	testData     *td.TestData
}

func (this *WebApiSuite) SetupSuite() {
	gin.SetMode(gin.ReleaseMode)
	baseDir := path.Join(os.TempDir(), "/.edbservers")
	ss := ess.NewServerServer(baseDir)
	cfg := server.DefaultServerConfig()
	cfg.Bind.Port = uint16(31404)
	proc := server.NewServeProcess(cfg, ss)
	err := proc.Start()
	if err != nil {
		panic(err)
	}
	this.serveProcess = proc
	testData := td.LoadTestData()
	this.codec = edb.NewTCodec()

	requestData := &ess.RequestData{testData.ChainData.PrivValidator, testData.ChainData.Genesis, SERVER_DURATION}
	rBts, _ := this.codec.EncodeBytes(requestData)
	resp, _ := http.Post(WAPIS_URL, "application/json", bytes.NewBuffer(rBts))
	rd := &ess.ResponseData{}
	err2 := this.codec.Decode(rd, resp.Body)
	if err2 != nil {
		panic(err2)
	}
	fmt.Println("Received Port: " + rd.Port)
	this.sUrl = "http://localhost:" + rd.Port
	this.testData = testData
}

func (this *WebApiSuite) TearDownSuite() {
	sec := this.serveProcess.StopEventChannel()
	this.serveProcess.Stop(0)
	<-sec
}

// ********************************************* Consensus *********************************************

func (this *WebApiSuite) Test_A0_ConsensusState() {
	resp := this.get("/consensus")
	ret := &ep.ConsensusState{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	ret.StartTime = ""
	this.Equal(ret, this.testData.Output.ConsensusState)
}

func (this *WebApiSuite) Test_A1_Validators() {
	resp := this.get("/consensus/validators")
	ret := &ep.ValidatorList{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Validators)
}

// ********************************************* Network *********************************************

func (this *WebApiSuite) Test_B0_NetworkInfo() {
	resp := this.get("/network")
	ret := &ep.NetworkInfo{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.NetworkInfo)
}

func (this *WebApiSuite) Test_B1_ClientVersion() {
	resp := this.get("/network/client_version")
	ret := &ep.ClientVersion{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.ClientVersion)
}

func (this *WebApiSuite) Test_B2_Moniker() {
	resp := this.get("/network/moniker")
	ret := &ep.Moniker{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Moniker)
}

func (this *WebApiSuite) Test_B3_Listening() {
	resp := this.get("/network/listening")
	ret := &ep.Listening{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Listening)
}

func (this *WebApiSuite) Test_B4_Listeners() {
	resp := this.get("/network/listeners")
	ret := &ep.Listeners{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Listeners)
}

func (this *WebApiSuite) Test_B5_Peers() {
	resp := this.get("/network/peers")
	ret := []*ep.Peer{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Peers)
}

// ********************************************* Transactions *********************************************

func (this *WebApiSuite) Test_C0_TxCreate() {
	resp := this.postJson("/unsafe/txpool", this.testData.Input.TxCreate)
	ret := &ep.Receipt{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.TxCreateReceipt)
}

func (this *WebApiSuite) Test_C1_Tx() {
	resp := this.postJson("/unsafe/txpool", this.testData.Input.Tx)
	ret := &ep.Receipt{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.TxReceipt)
}

func (this *WebApiSuite) Test_C2_UnconfirmedTxs() {
	resp := this.get("/txpool")
	ret := &ep.UnconfirmedTxs{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.UnconfirmedTxs)
}

func (this *WebApiSuite) Test_C3_CallCode() {
	resp := this.postJson("/codecalls", this.testData.Input.CallCode)
	ret := &ep.Call{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.CallCode)
}

// ********************************************* Accounts *********************************************

func (this *WebApiSuite) Test_D0_Accounts() {
	resp := this.get("/accounts")
	ret := &ep.AccountList{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Accounts)
}

func (this *WebApiSuite) Test_D1_Account() {
	resp := this.get("/accounts/" + this.testData.Input.AccountAddress)
	ret := &account.Account{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Account)
}

func (this *WebApiSuite) Test_D2_Storage() {
	resp := this.get("/accounts/" + this.testData.Input.AccountAddress + "/storage")
	ret := &ep.Storage{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Storage)
}

func (this *WebApiSuite) Test_D3_StorageAt() {
	addr := this.testData.Input.AccountAddress
	key := this.testData.Input.StorageAddress
	resp := this.get("/accounts/" + addr + "/storage/" + key)
	ret := &ep.StorageItem{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.StorageAt)
}

// ********************************************* Blockchain *********************************************

func (this *WebApiSuite) Test_E0_BlockchainInfo() {
	resp := this.get("/blockchain")
	ret := &ep.BlockchainInfo{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.BlockchainInfo)
}

func (this *WebApiSuite) Test_E1_ChainId() {
	resp := this.get("/blockchain/chain_id")
	ret := &ep.ChainId{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.ChainId)
}

func (this *WebApiSuite) Test_E2_GenesisHash() {
	resp := this.get("/blockchain/genesis_hash")
	ret := &ep.GenesisHash{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.GenesisHash)
}

func (this *WebApiSuite) Test_E3_LatestBlockHeight() {
	resp := this.get("/blockchain/latest_block_height")
	ret := &ep.LatestBlockHeight{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.LatestBlockHeight)
}

func (this *WebApiSuite) Test_E4_Blocks() {
	br := this.testData.Input.BlockRange
	resp := this.get(fmt.Sprintf("/blockchain/blocks?q=height:%d..%d", br.Min, br.Max))
	ret := &ep.Blocks{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Output.Blocks)
}

// ********************************************* Utilities *********************************************

func (this *WebApiSuite) get(endpoint string) *http.Response {
	resp, errG := http.Get(this.sUrl + endpoint)
	this.NoError(errG)
	return resp
}

func (this *WebApiSuite) postJson(endpoint string, v interface{}) *http.Response {
	bts, errE := this.codec.EncodeBytes(v)
	this.NoError(errE)
	resp, errP := http.Post(this.sUrl+endpoint, "application/json", bytes.NewBuffer(bts))
	this.NoError(errP)
	this.Equal(200, resp.StatusCode)
	return resp
}

// ********************************************* Entrypoint *********************************************

func TestWebApiSuite(t *testing.T) {
	suite.Run(t, &WebApiSuite{})
}
