package transacting

// Basic imports
import (
	"bytes"
	"fmt"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/stretchr/testify/suite"
	// "github.com/tendermint/tendermint/types"
	edb "github.com/eris-ltd/eris-db/erisdb"
	ess "github.com/eris-ltd/eris-db/erisdb/erisdbss"
	// ep "github.com/eris-ltd/eris-db/erisdb/pipe"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/gin-gonic/gin"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/log15"
	"github.com/eris-ltd/eris-db/rpc"
	"github.com/eris-ltd/eris-db/server"
	td "github.com/eris-ltd/eris-db/test/testdata/testdata"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"testing"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log15.Root().SetHandler(log15.LvlFilterHandler(
		log15.LvlInfo,
		log15.StreamHandler(os.Stdout, log15.TerminalFormat()),
	))
	gin.SetMode(gin.ReleaseMode)
}

const (
	TX_URL        = "http://localhost:31405/server"
	CONTRACT_CODE = "60606040525b33600060006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908302179055505b609480603e6000396000f30060606040523615600d57600d565b60685b6000600060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050805033600060006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908302179055505b90565b604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f3"
)

func getCreateInput(privKey [64]byte) *edb.TransactParam {
	tp := &edb.TransactParam{}
	tp.PrivKey = privKey[:]
	tp.Address = nil
	tp.Data = []byte(CONTRACT_CODE)
	tp.GasLimit = 100000
	tp.Fee = 0
	return tp
}

type TxSuite struct {
	suite.Suite
	baseDir      string
	serveProcess *server.ServeProcess
	codec        rpc.Codec
	sUrl         string
	testData     *td.TestData
}

func (this *TxSuite) SetupSuite() {
	baseDir := path.Join(os.TempDir(), "/.edbservers")
	ss := ess.NewServerServer(baseDir)
	cfg := server.DefaultServerConfig()
	cfg.Bind.Port = uint16(31405)
	proc := server.NewServeProcess(cfg, ss)
	err := proc.Start()
	if err != nil {
		panic(err)
	}
	this.serveProcess = proc
	testData := td.LoadTestData()
	this.codec = edb.NewTCodec()

	requestData := &ess.RequestData{testData.ChainData.PrivValidator, testData.ChainData.Genesis, 30}
	rBts, _ := this.codec.EncodeBytes(requestData)
	resp, _ := http.Post(TX_URL, "application/json", bytes.NewBuffer(rBts))
	if resp.StatusCode != 200 {
		bts, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("ERROR GETTING SS ADDRESS: " + string(bts))
		fmt.Printf("%v\n", resp)
		panic(fmt.Errorf(string(bts)))
	}
	rd := &ess.ResponseData{}
	err2 := this.codec.Decode(rd, resp.Body)
	if err2 != nil {
		panic(err2)
	}
	fmt.Println("Received Port: " + rd.Port)
	this.sUrl = "http://localhost:" + rd.Port
	fmt.Println("URL: " + this.sUrl)
	this.testData = testData
}

func (this *TxSuite) TearDownSuite() {
	sec := this.serveProcess.StopEventChannel()
	this.serveProcess.Stop(0)
	<-sec
}

// ********************************************* Tests *********************************************

// TODO less duplication.
func (this *TxSuite) Test_A0_Tx_Create() {
	input := getCreateInput([64]byte(this.testData.ChainData.PrivValidator.PrivKey))
	resp := this.postJson("/unsafe/txpool?hold=true", input)
	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("%s\n", string(bts))
	}
	//ret := &types.EventMsgCall{}
	// errD := this.codec.Decode(ret, resp.Body)
	//this.NoError(errD)
	//json, _ := this.codec.EncodeBytes(ret)
	//fmt.Printf("%s\n", string(json))
}

// ********************************************* Utilities *********************************************

func (this *TxSuite) get(endpoint string) *http.Response {
	resp, errG := http.Get(this.sUrl + endpoint)
	this.NoError(errG)
	this.Equal(200, resp.StatusCode)
	return resp
}

func (this *TxSuite) postJson(endpoint string, v interface{}) *http.Response {
	bts, errE := this.codec.EncodeBytes(v)
	this.NoError(errE)
	resp, errP := http.Post(this.sUrl+endpoint, "application/json", bytes.NewBuffer(bts))
	this.NoError(errP)
	this.Equal(200, resp.StatusCode)
	return resp
}

// ********************************************* Entrypoint *********************************************

func TestQuerySuite(t *testing.T) {
	suite.Run(t, &TxSuite{})
}
