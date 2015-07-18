package web_api

// Basic imports
import (
	"bytes"
	"encoding/hex"
	"fmt"
	// edb "github.com/eris-ltd/erisdb/erisdb"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/gin-gonic/gin"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/stretchr/testify/suite"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/account"
	edb "github.com/eris-ltd/eris-db/erisdb"
	ess "github.com/eris-ltd/eris-db/erisdb/erisdbss"
	ep "github.com/eris-ltd/eris-db/erisdb/pipe"
	"github.com/eris-ltd/eris-db/rpc"
	"github.com/eris-ltd/eris-db/server"
	td "github.com/eris-ltd/eris-db/test/testdata/testdata"
	"io/ioutil"
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
	this.Equal(ret, this.testData.GetConsensusState.Output)
}

func (this *WebApiSuite) Test_A1_Validators() {
	resp := this.get("/consensus/validators")
	ret := &ep.ValidatorList{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetValidators.Output)
}

// ********************************************* Network *********************************************

func (this *WebApiSuite) Test_B0_NetworkInfo() {
	resp := this.get("/network")
	ret := &ep.NetworkInfo{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetNetworkInfo.Output)
}

func (this *WebApiSuite) Test_B1_ClientVersion() {
	resp := this.get("/network/client_version")
	ret := &ep.ClientVersion{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetClientVersion.Output)
}

func (this *WebApiSuite) Test_B2_Moniker() {
	resp := this.get("/network/moniker")
	ret := &ep.Moniker{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetMoniker.Output)
}

func (this *WebApiSuite) Test_B3_Listening() {
	resp := this.get("/network/listening")
	ret := &ep.Listening{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.IsListening.Output)
}

func (this *WebApiSuite) Test_B4_Listeners() {
	resp := this.get("/network/listeners")
	ret := &ep.Listeners{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetListeners.Output)
}

func (this *WebApiSuite) Test_B5_Peers() {
	resp := this.get("/network/peers")
	ret := []*ep.Peer{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetPeers.Output)
}

// ********************************************* Transactions *********************************************

func (this *WebApiSuite) Test_C0_TxCreate() {
	resp := this.postJson("/unsafe/txpool", this.testData.TransactCreate.Input)
	ret := &ep.Receipt{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.TransactCreate.Output)
}

func (this *WebApiSuite) Test_C1_Tx() {
	resp := this.postJson("/unsafe/txpool", this.testData.Transact.Input)
	ret := &ep.Receipt{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.Transact.Output)
}

func (this *WebApiSuite) Test_C2_UnconfirmedTxs() {
	resp := this.get("/txpool")
	ret := &ep.UnconfirmedTxs{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetUnconfirmedTxs.Output)
}

func (this *WebApiSuite) Test_C3_CallCode() {
	resp := this.postJson("/codecalls", this.testData.CallCode.Input)
	ret := &ep.Call{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.CallCode.Output)
}

// ********************************************* Accounts *********************************************

func (this *WebApiSuite) Test_D0_GetAccounts() {
	resp := this.get("/accounts")
	ret := &ep.AccountList{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetAccounts.Output)
}

func (this *WebApiSuite) Test_D1_GetAccount() {
	addr := hex.EncodeToString(this.testData.GetAccount.Input.Address)
	resp := this.get("/accounts/" + addr)
	ret := &account.Account{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetAccount.Output)
}

func (this *WebApiSuite) Test_D2_GetStorage() {
	addr := hex.EncodeToString(this.testData.GetStorage.Input.Address)
	resp := this.get("/accounts/" + addr + "/storage")
	ret := &ep.Storage{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetStorage.Output)
}

func (this *WebApiSuite) Test_D3_GetStorageAt() {
	addr := hex.EncodeToString(this.testData.GetStorageAt.Input.Address)
	key := hex.EncodeToString(this.testData.GetStorageAt.Input.Key)
	resp := this.get("/accounts/" + addr + "/storage/" + key)
	ret := &ep.StorageItem{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetStorageAt.Output)
}

// ********************************************* Blockchain *********************************************

func (this *WebApiSuite) Test_E0_GetBlockchainInfo() {
	resp := this.get("/blockchain")
	ret := &ep.BlockchainInfo{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetBlockchainInfo.Output)
}

func (this *WebApiSuite) Test_E1_GetChainId() {
	resp := this.get("/blockchain/chain_id")
	ret := &ep.ChainId{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetChainId.Output)
}

func (this *WebApiSuite) Test_E2_GetGenesisHash() {
	resp := this.get("/blockchain/genesis_hash")
	ret := &ep.GenesisHash{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetGenesisHash.Output)
}

func (this *WebApiSuite) Test_E3_GetLatestBlockHeight() {
	resp := this.get("/blockchain/latest_block_height")
	ret := &ep.LatestBlockHeight{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetLatestBlockHeight.Output)
}

func (this *WebApiSuite) Test_E4_GetBlocks() {
	resp := this.get("/blockchain/blocks")
	ret := &ep.Blocks{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(ret, this.testData.GetBlocks.Output)
}

// ********************************************* Utilities *********************************************

func (this *WebApiSuite) get(endpoint string) *http.Response {
	resp, errG := http.Get(this.sUrl + endpoint)
	this.NoError(errG)
	this.Equal(200, resp.StatusCode)
	return resp
}

func (this *WebApiSuite) postJson(endpoint string, v interface{}) *http.Response {
	bts, errE := this.codec.EncodeBytes(v)
	this.NoError(errE)
	resp, errP := http.Post(this.sUrl+endpoint, "application/json", bytes.NewBuffer(bts))
	if resp.StatusCode != 200 {
		bts, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("ERROR: " + string(bts))
	}
	this.NoError(errP)
	this.Equal(200, resp.StatusCode)
	return resp
}

// ********************************************* Entrypoint *********************************************

func TestWebApiSuite(t *testing.T) {
	suite.Run(t, &WebApiSuite{})
}
