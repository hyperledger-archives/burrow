package client

import (
	"encoding/json"
	"fmt"
	edb "github.com/eris-ltd/erisdb/erisdb"
	ep "github.com/eris-ltd/erisdb/erisdb/pipe"
	rpc "github.com/eris-ltd/erisdb/rpc"
	"github.com/eris-ltd/erisdb/test/mock"
	"github.com/tendermint/tendermint/binary"
	"github.com/tendermint/tendermint/account"
)

type RPCSendRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []byte `json:"params"`
	Id      string `json:"id"`
}

type RPCSendResponse struct {
	Result  json.RawMessage `json:"result"`
	Error   *rpc.RPCError   `json:"error"`
	Id      string          `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
}

// This client is used to test the server, nothing else. Use the built-in tendermint
// rpc server to RPC from a Go lib.
type TestClient struct {
	mockData *mock.MockData
	codec    rpc.Codec
	client   *WSClient
	idCount  int
	readChan chan []byte
}

// Create a new testclient.
func NewTestClient(addr string, mockData *mock.MockData) *TestClient {
	wsc := NewWSClient(addr)
	return &TestClient{
		mockData: mockData,
		client:   wsc,
		idCount:  1,
		codec:    &edb.TCodec{},
	}
}

// Start the client.
func (this *TestClient) Start() error {
	r, err := this.client.Dial()
	if err != nil {
		fmt.Printf("Response: %v\n", r)
		return err
	}
	this.readChan = this.client.Read()
	return nil
}

// Get the next request id.
func (this *TestClient) nextId() string {
	s := fmt.Sprintf("%d", this.idCount) // ++ not working?
	this.idCount++
	return s
}

// write a request to the socket.
func (this *TestClient) writeRequest(method string, params interface{}) {
	pbts := binary.JSONBytes(params)
	req := &RPCSendRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  pbts,
		Id:      this.nextId(),
	}
	msg := binary.JSONBytes(req)
	this.client.WriteMsg(msg)
}

// Read a response from the read channel.
func (this *TestClient) readResponse() (*RPCSendResponse, error) {
	// TODO probably timeout here.
	msg := <-this.readChan
	fmt.Printf("RESPONSE: %s\n", string(msg))
	resp := &RPCSendResponse{}
	err := json.Unmarshal(msg, resp)
	if err != nil {
		return nil, err
	}
	if resp.JSONRPC != "2.0" {
		return nil, fmt.Errorf("Wrong protocol version used: %s\n", resp.JSONRPC)
	}
	return resp, nil
}

// API methods

// ********************************** Accounts **********************************

func (this *TestClient) AccountList(ap *edb.AccountsParam) (interface{}, error) {
	this.writeRequest(edb.GET_ACCOUNTS, ap)
	resp, errR := this.readResponse()
	if errR != nil {
		return nil, errR
	}
	accs := &ep.AccountList{}
	err := this.codec.DecodeBytes(accs, resp.Result)
	if err != nil {
		return nil, err
	}
	return accs, nil
}

func (this *TestClient) Account(address []byte) (interface{}, error) {
	params := &edb.AddressParam{address}
	this.writeRequest(edb.GET_ACCOUNT, params)
	resp, errR := this.readResponse()
	if errR != nil {
		return nil, errR
	}
	acc := &account.Account{}
	err := this.codec.DecodeBytes(acc, resp.Result)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (this *TestClient) Storage(address []byte) (interface{}, error) {
	params := &edb.AddressParam{address}
	this.writeRequest(edb.GET_STORAGE, params)
	resp, errR := this.readResponse()
	if errR != nil {
		return nil, errR
	}
	s := &ep.Storage{}
	err := this.codec.DecodeBytes(s, resp.Result)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (this *TestClient) StorageAt(address, key []byte) (interface{}, error) {
	params := &edb.StorageAtParam{address, key}
	this.writeRequest(edb.GET_STORAGE_AT, params)
	resp, errR := this.readResponse()
	if errR != nil {
		return nil, errR
	}
	sa := &ep.StorageItem{}
	err := this.codec.DecodeBytes(sa, resp.Result)
	if err != nil {
		return nil, err
	}
	return sa, nil
}
