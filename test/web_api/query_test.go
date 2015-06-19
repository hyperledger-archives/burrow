package web_api

// Basic imports
import (
	"io/ioutil"
	"bytes"
	"fmt"
	edb "github.com/eris-ltd/eris-db/erisdb"
	ess "github.com/eris-ltd/eris-db/erisdb/erisdbss"
	ep "github.com/eris-ltd/eris-db/erisdb/pipe"
	"github.com/eris-ltd/eris-db/rpc"
	"github.com/eris-ltd/eris-db/server"
	fd "github.com/eris-ltd/eris-db/test/testdata/filters"
	"github.com/stretchr/testify/suite"
	"net/http"
	"os"
	"path"
	"testing"
)

const QS_URL = "http://localhost:31403/server"

type QuerySuite struct {
	suite.Suite
	baseDir      string
	serveProcess *server.ServeProcess
	codec        rpc.Codec
	sUrl         string
	testData     *fd.TestData
}

func (this *QuerySuite) SetupSuite() {
	baseDir := path.Join(os.TempDir(), "/.edbservers")
	ss := ess.NewServerServer(baseDir)
	cfg := server.DefaultServerConfig()
	cfg.Bind.Port = uint16(31403)
	proc := server.NewServeProcess(cfg, ss)
	err := proc.Start()
	if err != nil {
		panic(err)
	}
	this.serveProcess = proc
	testData := fd.LoadTestData()
	this.codec = edb.NewTCodec()

	requestData := &ess.RequestData{testData.ChainData.PrivValidator, testData.ChainData.Genesis, SERVER_DURATION}
	rBts, _ := this.codec.EncodeBytes(requestData)
	resp, _ := http.Post(QS_URL, "application/json", bytes.NewBuffer(rBts))
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

func (this *QuerySuite) TearDownSuite() {
	sec := this.serveProcess.StopEventChannel()
	this.serveProcess.Stop(0)
	<-sec
}

// ********************************************* Tests *********************************************


// TODO make these functions into one.
func (this *QuerySuite) Test_Accounts0() {
	fd := this.testData.Input.Filters0
	resp := this.get("/accounts?" + generateQuery(fd))
	ret := &ep.AccountList{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(this.testData.Output.Accounts0, ret)
}

func (this *QuerySuite) Test_Accounts1() {
	fd := this.testData.Input.Filters1
	resp := this.get("/accounts?" + generateQuery(fd))
	ret := &ep.AccountList{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(this.testData.Output.Accounts1, ret)
}

func (this *QuerySuite) Test_Accounts2() {
	fd := this.testData.Input.Filters2
	resp := this.get("/accounts?" + generateQuery(fd))
	ret := &ep.AccountList{}
	errD := this.codec.Decode(ret, resp.Body)
	this.NoError(errD)
	this.Equal(this.testData.Output.Accounts2, ret)
}

// ********************************************* Utilities *********************************************

func (this *QuerySuite) get(endpoint string) *http.Response {
	resp, errG := http.Get(this.sUrl + endpoint)
	this.NoError(errG)
	this.Equal(200, resp.StatusCode)
	return resp
}

func (this *QuerySuite) postJson(endpoint string, v interface{}) *http.Response {
	bts, errE := this.codec.EncodeBytes(v)
	this.NoError(errE)
	resp, errP := http.Post(this.sUrl+endpoint, "application/json", bytes.NewBuffer(bts))
	this.NoError(errP)
	this.Equal(200, resp.StatusCode)
	return resp
}

func generateQuery(fda []*ep.FilterData) string {
	query := "q="
	for i := 0; i < len(fda); i++ {
		fd := fda[i]
		query += fd.Field + ":" + fd.Op + fd.Value
		if i != len(fda)-1 {
			query += "+"
		}
	}
	return query
}

// ********************************************* Entrypoint *********************************************

func TestQuerySuite(t *testing.T) {
	suite.Run(t, &QuerySuite{})
}
