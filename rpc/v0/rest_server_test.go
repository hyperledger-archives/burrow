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

package v0

// Basic imports
import (
	"bytes"
	"encoding/hex"
	"net/http"
	"runtime"
	"testing"

	account "github.com/hyperledger/burrow/account"
	consensus_types "github.com/hyperledger/burrow/consensus/types"
	core_types "github.com/hyperledger/burrow/core/types"
	event "github.com/hyperledger/burrow/event"
	rpc "github.com/hyperledger/burrow/rpc"
	server "github.com/hyperledger/burrow/server"
	"github.com/hyperledger/burrow/txs"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/burrow/logging/lifecycle"
	"github.com/hyperledger/burrow/rpc/v0/shared"
	"github.com/stretchr/testify/suite"
)

var logger, _ = lifecycle.NewStdErrLogger()

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	gin.SetMode(gin.ReleaseMode)
}

type MockSuite struct {
	suite.Suite
	serveProcess *server.ServeProcess
	codec        rpc.Codec
	sUrl         string
	testData     *TestData
}

func (mockSuite *MockSuite) SetupSuite() {
	gin.SetMode(gin.ReleaseMode)
	// Load the supporting objects.
	testData := LoadTestData()
	pipe := NewMockPipe(testData)
	codec := &TCodec{}
	evtSubs := event.NewEventSubscriptions(pipe.Events())
	// The server
	restServer := NewRestServer(codec, pipe, evtSubs)
	sConf := server.DefaultServerConfig()
	sConf.Bind.Port = 31402
	// Create a server process.
	proc, _ := server.NewServeProcess(sConf, logger, restServer)
	err := proc.Start()
	if err != nil {
		panic(err)
	}
	mockSuite.serveProcess = proc
	mockSuite.codec = NewTCodec()
	mockSuite.testData = testData
	mockSuite.sUrl = "http://localhost:31402"
}

func (mockSuite *MockSuite) TearDownSuite() {
	sec := mockSuite.serveProcess.StopEventChannel()
	mockSuite.serveProcess.Stop(0)
	<-sec
}

// ********************************************* Accounts *********************************************

func (mockSuite *MockSuite) TestGetAccounts() {
	resp := mockSuite.get("/accounts")
	ret := &core_types.AccountList{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetAccounts.Output, ret)
}

func (mockSuite *MockSuite) TestGetAccount() {
	addr := hex.EncodeToString(mockSuite.testData.GetAccount.Input.Address)
	resp := mockSuite.get("/accounts/" + addr)
	ret := &account.Account{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetAccount.Output, ret)
}

func (mockSuite *MockSuite) TestGetStorage() {
	addr := hex.EncodeToString(mockSuite.testData.GetStorage.Input.Address)
	resp := mockSuite.get("/accounts/" + addr + "/storage")
	ret := &core_types.Storage{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetStorage.Output, ret)
}

func (mockSuite *MockSuite) TestGetStorageAt() {
	addr := hex.EncodeToString(mockSuite.testData.GetStorageAt.Input.Address)
	key := hex.EncodeToString(mockSuite.testData.GetStorageAt.Input.Key)
	resp := mockSuite.get("/accounts/" + addr + "/storage/" + key)
	ret := &core_types.StorageItem{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetStorageAt.Output, ret)
}

// ********************************************* Blockchain *********************************************

func (mockSuite *MockSuite) TestGetBlockchainInfo() {
	resp := mockSuite.get("/blockchain")
	ret := &core_types.BlockchainInfo{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetBlockchainInfo.Output, ret)
}

func (mockSuite *MockSuite) TestGetChainId() {
	resp := mockSuite.get("/blockchain/chain_id")
	ret := &core_types.ChainId{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetChainId.Output, ret)
}

func (mockSuite *MockSuite) TestGetGenesisHash() {
	resp := mockSuite.get("/blockchain/genesis_hash")
	ret := &core_types.GenesisHash{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetGenesisHash.Output, ret)
}

func (mockSuite *MockSuite) TestLatestBlockHeight() {
	resp := mockSuite.get("/blockchain/latest_block_height")
	ret := &core_types.LatestBlockHeight{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetLatestBlockHeight.Output, ret)
}

func (mockSuite *MockSuite) TestBlocks() {
	resp := mockSuite.get("/blockchain/blocks")
	ret := &core_types.Blocks{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetBlocks.Output, ret)
}

// ********************************************* Consensus *********************************************

// TODO: re-enable these when implemented
//func (mockSuite *MockSuite) TestGetConsensusState() {
//	resp := mockSuite.get("/consensus")
//	ret := &core_types.ConsensusState{}
//	errD := mockSuite.codec.Decode(ret, resp.Body)
//	mockSuite.NoError(errD)
//	ret.StartTime = ""
//	mockSuite.Equal(mockSuite.testData.GetConsensusState.Output, ret)
//}
//
//func (mockSuite *MockSuite) TestGetValidators() {
//	resp := mockSuite.get("/consensus/validators")
//	ret := &core_types.ValidatorList{}
//	errD := mockSuite.codec.Decode(ret, resp.Body)
//	mockSuite.NoError(errD)
//	mockSuite.Equal(mockSuite.testData.GetValidators.Output, ret)
//}

// ********************************************* NameReg *********************************************

func (mockSuite *MockSuite) TestGetNameRegEntry() {
	resp := mockSuite.get("/namereg/" + mockSuite.testData.GetNameRegEntry.Input.Name)
	ret := &core_types.NameRegEntry{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetNameRegEntry.Output, ret)
}

func (mockSuite *MockSuite) TestGetNameRegEntries() {
	resp := mockSuite.get("/namereg")
	ret := &core_types.ResultListNames{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetNameRegEntries.Output, ret)
}

// ********************************************* Network *********************************************

func (mockSuite *MockSuite) TestGetNetworkInfo() {
	resp := mockSuite.get("/network")
	ret := &shared.NetworkInfo{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetNetworkInfo.Output, ret)
}

func (mockSuite *MockSuite) TestGetClientVersion() {
	resp := mockSuite.get("/network/client_version")
	ret := &core_types.ClientVersion{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetClientVersion.Output, ret)
}

func (mockSuite *MockSuite) TestGetMoniker() {
	resp := mockSuite.get("/network/moniker")
	ret := &core_types.Moniker{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetMoniker.Output, ret)
}

func (mockSuite *MockSuite) TestIsListening() {
	resp := mockSuite.get("/network/listening")
	ret := &core_types.Listening{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.IsListening.Output, ret)
}

func (mockSuite *MockSuite) TestGetListeners() {
	resp := mockSuite.get("/network/listeners")
	ret := &core_types.Listeners{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetListeners.Output, ret)
}

func (mockSuite *MockSuite) TestGetPeers() {
	resp := mockSuite.get("/network/peers")
	ret := []*consensus_types.Peer{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetPeers.Output, ret)
}

/*
func (mockSuite *MockSuite) TestGetPeer() {
	addr := mockSuite.testData.GetPeer.Input.Address
	resp := mockSuite.get("/network/peer/" + addr)
	ret := []*core_types.Peer{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetPeers.Output)
}
*/

// ********************************************* Transactions *********************************************

func (mockSuite *MockSuite) TestTransactCreate() {
	resp := mockSuite.postJson("/unsafe/txpool", mockSuite.testData.TransactCreate.Input)
	ret := &txs.Receipt{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.TransactCreate.Output, ret)
}

func (mockSuite *MockSuite) TestTransact() {
	resp := mockSuite.postJson("/unsafe/txpool", mockSuite.testData.Transact.Input)
	ret := &txs.Receipt{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.Transact.Output, ret)
}

func (mockSuite *MockSuite) TestTransactNameReg() {
	resp := mockSuite.postJson("/unsafe/namereg/txpool", mockSuite.testData.TransactNameReg.Input)
	ret := &txs.Receipt{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.TransactNameReg.Output, ret)
}

func (mockSuite *MockSuite) TestGetUnconfirmedTxs() {
	resp := mockSuite.get("/txpool")
	ret := &txs.UnconfirmedTxs{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetUnconfirmedTxs.Output, ret)
}

func (mockSuite *MockSuite) TestCallCode() {
	resp := mockSuite.postJson("/codecalls", mockSuite.testData.CallCode.Input)
	ret := &core_types.Call{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.CallCode.Output, ret)
}

func (mockSuite *MockSuite) TestCall() {
	resp := mockSuite.postJson("/calls", mockSuite.testData.Call.Input)
	ret := &core_types.Call{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.CallCode.Output, ret)
}

// ********************************************* Utilities *********************************************

func (mockSuite *MockSuite) get(endpoint string) *http.Response {
	resp, errG := http.Get(mockSuite.sUrl + endpoint)
	mockSuite.NoError(errG)
	mockSuite.Equal(200, resp.StatusCode)
	return resp
}

func (mockSuite *MockSuite) postJson(endpoint string, v interface{}) *http.Response {
	bts, errE := mockSuite.codec.EncodeBytes(v)
	mockSuite.NoError(errE)
	resp, errP := http.Post(mockSuite.sUrl+endpoint, "application/json", bytes.NewBuffer(bts))
	mockSuite.NoError(errP)
	mockSuite.Equal(200, resp.StatusCode)
	return resp
}

// ********************************************* Entrypoint *********************************************

func TestMockSuite(t *testing.T) {
	suite.Run(t, &MockSuite{})
}
