package test

import (
	//"fmt"
	"github.com/androlo/blockchain_rpc/server"
	edb "github.com/androlo/blockchain_rpc/erisdb"
	ep "github.com/androlo/blockchain_rpc/erisdb/pipe"
	tc "github.com/androlo/blockchain_rpc/test/client"
	"github.com/androlo/blockchain_rpc/test/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

const(
	MAX_CONNS = 10
)

// Tests the rpc server and service using a mock pipe.
type MockPipeSuite struct {
	suite.Suite
	mockData *mock.MockData
	client *tc.TestClient
	sProc *server.ServeProcess
}

func (this *MockPipeSuite) SetupSuite() {
	mockData := mock.NewDefaultMockData()
	mockPipe := mock.NewMockPipe(mockData)

	this.mockData = mockData

	edbwss := edb.NewErisDbWsService(&edb.TCodec{}, mockPipe)
	wsServer := server.NewWebSocketServer(MAX_CONNS, edbwss)
	proc, errServe := server.Serve(nil, wsServer)
	if errServe != nil {
		panic(errServe.Error())
	}
	// TODO
	proc.Start()
	this.sProc = proc
	this.client = tc.NewTestClient("ws://localhost:1337/socketrpc", mockData)
	errC := this.client.Start()
	if errC != nil {
		panic(errC.Error())
	}
}

func (this *MockPipeSuite) TearDownSuite() {
	errStop := this.sProc.Stop(time.Millisecond*100)
	if errStop != nil {
		panic(errStop.Error())
	}
}

func (this *MockPipeSuite) TestClientAccountList() {
	result, err := this.client.AccountList(&ep.AccountsParam{})
	this.NoError(err)
	this.Equal(result, this.mockData.Accounts, "Accounts not the same")
}

// Test: Account
func (this *MockPipeSuite) TestClientAccount() {
	result, err := this.client.Account([]byte{0})
	this.NoError(err)
	this.Equal(result, this.mockData.Account, "Account not the same")
}

// Test: StorageAt
func (this *MockPipeSuite) TestClientStorage() {
	result, err := this.client.Storage([]byte{0})
	this.NoError(err)
	this.Equal(result, this.mockData.Storage, "Storage not the same")
}

// Test: Storage
func (this *MockPipeSuite) TestClientStorageAt() {
	result, err := this.client.StorageAt([]byte{0}, []byte{0})
	this.NoError(err)
	this.Equal(result, this.mockData.StorageAt, "StorageItem not the same")
}


func TestMockPipeSuite(t *testing.T) {
	suite.Run(t, new(MockPipeSuite))
}