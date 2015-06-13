package test

// Basic imports
import (
	"bytes"
	"fmt"
	// edb "github.com/eris-ltd/erisdb/erisdb"
	ess "github.com/eris-ltd/erisdb/erisdb/erisdbss"
	edb "github.com/eris-ltd/erisdb/erisdb"
	ep "github.com/eris-ltd/erisdb/erisdb/pipe"
	"github.com/eris-ltd/erisdb/rpc"
	"github.com/eris-ltd/erisdb/server"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	// "io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
	"time"
)

const SS_URL = "http://localhost:1337/server"

//
type WebApiSuite struct {
	suite.Suite
	baseDir      string
	serveProcess *server.ServeProcess
	codec        rpc.Codec
	sUrl         string
	testData     *TestData
}

func (this *WebApiSuite) SetupSuite() {
	gin.SetMode(gin.ReleaseMode)
	baseDir := path.Join(os.TempDir(), "/.edbservers")
	ss := ess.NewServerServer(baseDir)
	proc := server.NewServeProcess(nil, ss)
	_ = proc.Start()
	this.serveProcess = proc
	time.Sleep(1*time.Second)
	this.testData = LoadTestData()
	this.codec = edb.NewTCodec()
	rBts, _ := this.codec.EncodeBytes(this.testData.requestData)
	resp, _ := http.Post(SS_URL, "application/json", bytes.NewBuffer(rBts))
	rd := &ess.ResponseData{}
	_ = this.codec.Decode(rd, resp.Body)
	fmt.Println("Received URL: " + rd.URL)
	this.sUrl = rd.URL
	time.Sleep(1*time.Second)
}

func (this *WebApiSuite) TearDownSuite() {
	sec := this.serveProcess.StopEventChannel()
	this.serveProcess.Stop(time.Millisecond)
	<-sec
	os.RemoveAll(this.baseDir)
}

// ********************************************* Consensus *********************************************

func (this *WebApiSuite) Test_A0_ConsensusState() {
	resp, errG := http.Get(this.sUrl + "/consensus")
	this.NoError(errG)
	ret := &ep.ConsensusState{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	fmt.Printf("Consensus state: %v\n", ret)
}

func (this *WebApiSuite) Test_A1_Validators() {
	resp, errG := http.Get(this.sUrl + "/consensus/validators")
	this.NoError(errG)
	ret := &ep.ValidatorList{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	fmt.Printf("Validators: %v\n", ret)
}

// ********************************************* Network *********************************************

func (this *WebApiSuite) Test_B0_NetworkInfo() {
	resp, errG := http.Get(this.sUrl + "/network")
	this.NoError(errG)
	ni := &ep.NetworkInfo{}
	errD := this.codec.Decode(ni, resp.Body)
	this.NoError(errD)
	fmt.Printf("NetworkInfo: %v\n", ni)
}

func (this *WebApiSuite) Test_B1_Moniker() {
	resp, errG := http.Get(this.sUrl + "/network/moniker")
	this.NoError(errG)
	ret := &ep.Moniker{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	fmt.Printf("Moniker: %v\n", ret)
}

func (this *WebApiSuite) Test_B2_Listening() {
	resp, errG := http.Get(this.sUrl + "/network/listening")
	this.NoError(errG)
	ret := &ep.Listening{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	fmt.Printf("Listening: %v\n", ret)
}

func (this *WebApiSuite) Test_B3_Listeners() {
	resp, errG := http.Get(this.sUrl + "/network/listeners")
	this.NoError(errG)
	ret := &ep.Listeners{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	fmt.Printf("Listeners: %v\n", ret)
}

func (this *WebApiSuite) Test_B4_Peers() {
	resp, errG := http.Get(this.sUrl + "/network/peers")
	this.NoError(errG)
	ret := &ep.Peers{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	fmt.Printf("Listeners: %v\n", ret)
}

// ********************************************* Transactions *********************************************

func (this *WebApiSuite) Test_C0_TransactContractCreate() {
	resp, errG := http.Get(this.sUrl + "/unsafe/txpool")
	this.NoError(errG)
	ni := &ep.NetworkInfo{}
	errD := this.codec.Decode(ni, resp.Body)
	this.NoError(errD)
	fmt.Printf("NetworkInfo: %v\n", ni)
}

// ********************************************* Teardown *********************************************

func TestWebApiSuite(t *testing.T) {
	suite.Run(t, &WebApiSuite{})
}