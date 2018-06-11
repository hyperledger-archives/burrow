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

package client

import (
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/rpc"
	tmClient "github.com/hyperledger/burrow/rpc/tm/client"
	rpcClient "github.com/hyperledger/burrow/rpc/tm/lib/client"
	"github.com/hyperledger/burrow/txs"
)

type NodeClient interface {
	Broadcast(transaction *txs.Envelope) (*txs.Receipt, error)
	DeriveWebsocketClient() (nodeWsClient NodeWebsocketClient, err error)

	Status() (ChainId []byte, ValidatorPublicKey []byte, LatestBlockHash []byte,
		LatestBlockHeight uint64, LatestBlockTime int64, err error)
	GetAccount(address crypto.Address) (acm.Account, error)
	QueryContract(callerAddress, calleeAddress crypto.Address, data []byte) (ret []byte, gasUsed uint64, err error)
	QueryContractCode(address crypto.Address, code, data []byte) (ret []byte, gasUsed uint64, err error)

	DumpStorage(address crypto.Address) (storage *rpc.ResultDumpStorage, err error)
	GetName(name string) (owner crypto.Address, data string, expirationBlock uint64, err error)
	ListValidators() (blockHeight uint64, bondedValidators, unbondingValidators []acm.Validator, err error)

	// Logging context for this NodeClient
	Logger() *logging.Logger
}

type NodeWebsocketClient interface {
	Subscribe(eventId string) error
	Unsubscribe(eventId string) error
	WaitForConfirmation(tx *txs.Envelope, inputAddr crypto.Address) (chan Confirmation, error)
	Close()
}

// NOTE [ben] Compiler check to ensure burrowNodeClient successfully implements
// burrow/client.NodeClient
var _ NodeClient = (*burrowNodeClient)(nil)

// burrow-client is a simple struct exposing the client rpc methods
type burrowNodeClient struct {
	broadcastRPC string
	logger       *logging.Logger
}

// BurrowKeyClient.New returns a new monax-keys client for provided rpc location
// Monax-keys connects over http request-responses
func NewBurrowNodeClient(rpcString string, logger *logging.Logger) *burrowNodeClient {
	return &burrowNodeClient{
		broadcastRPC: rpcString,
		logger:       logger.WithScope("BurrowNodeClient"),
	}
}

//------------------------------------------------------------------------------------
// broadcast to blockchain node

func (burrowNodeClient *burrowNodeClient) Broadcast(txEnv *txs.Envelope) (*txs.Receipt, error) {
	client := rpcClient.NewURIClient(burrowNodeClient.broadcastRPC)
	receipt, err := tmClient.BroadcastTx(client, txEnv)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func (burrowNodeClient *burrowNodeClient) DeriveWebsocketClient() (nodeWsClient NodeWebsocketClient, err error) {
	var wsAddr string
	// TODO: clean up this inherited mess on dealing with the address prefixes.
	nodeAddr := burrowNodeClient.broadcastRPC
	// if strings.HasPrefix(nodeAddr, "http://") {
	// 	wsAddr = strings.TrimPrefix(nodeAddr, "http://")
	// }
	// if strings.HasPrefix(nodeAddr, "tcp://") {
	// 	wsAddr = strings.TrimPrefix(nodeAddr, "tcp://")
	// }
	// if strings.HasPrefix(nodeAddr, "unix://") {
	// 	log.WithFields(log.Fields{
	// 		"node address": nodeAddr,
	// 	}).Error("Unable to subscribe to websocket from unix socket.")
	// 	return nil, fmt.Errorf("Unable to construct websocket from unix socket: %s", nodeAddr)
	// }
	// wsAddr = "ws://" + wsAddr
	wsAddr = nodeAddr
	burrowNodeClient.logger.TraceMsg("Subscribing to websocket address",
		"websocket address", wsAddr,
		"endpoint", "/websocket",
	)
	wsClient := rpcClient.NewWSClient(wsAddr, "/websocket")
	if err = wsClient.Start(); err != nil {
		return nil, err
	}
	derivedBurrowNodeWebsocketClient := &burrowNodeWebsocketClient{
		tendermintWebsocket: wsClient,
		logger:              burrowNodeClient.logger.WithScope("BurrowNodeWebsocketClient"),
	}
	return derivedBurrowNodeWebsocketClient, nil
}

//------------------------------------------------------------------------------------
// RPC requests other than transaction related

// Status returns the ChainId (GenesisHash), validator's PublicKey, latest block hash
// the block height and the latest block time.
func (burrowNodeClient *burrowNodeClient) Status() (GenesisHash []byte, ValidatorPublicKey []byte,
	LatestBlockHash []byte, LatestBlockHeight uint64, LatestBlockTime int64, err error) {

	client := rpcClient.NewJSONRPCClient(burrowNodeClient.broadcastRPC)
	res, err := tmClient.Status(client)
	if err != nil {
		err = fmt.Errorf("error connecting to node (%s) to get status: %s",
			burrowNodeClient.broadcastRPC, err.Error())
		return
	}

	// unwrap return results
	GenesisHash = res.GenesisHash
	ValidatorPublicKey = res.PubKey.RawBytes()
	LatestBlockHash = res.LatestBlockHash
	LatestBlockHeight = res.LatestBlockHeight
	LatestBlockTime = res.LatestBlockTime
	return
}

func (burrowNodeClient *burrowNodeClient) ChainId() (ChainName, ChainId string, GenesisHash []byte, err error) {
	client := rpcClient.NewJSONRPCClient(burrowNodeClient.broadcastRPC)
	chainIdResult, err := tmClient.ChainId(client)
	if err != nil {
		err = fmt.Errorf("error connecting to node (%s) to get chain id: %s",
			burrowNodeClient.broadcastRPC, err.Error())
		return "", "", nil, err
	}
	// unwrap results
	ChainName = chainIdResult.ChainName
	ChainId = chainIdResult.ChainId
	GenesisHash = make([]byte, len(chainIdResult.GenesisHash))
	copy(GenesisHash[:], chainIdResult.GenesisHash)
	return
}

// QueryContract executes the contract code at address with the given data
// NOTE: there is no check on the caller;
func (burrowNodeClient *burrowNodeClient) QueryContract(callerAddress, calleeAddress crypto.Address,
	data []byte) (ret []byte, gasUsed uint64, err error) {

	client := rpcClient.NewJSONRPCClient(burrowNodeClient.broadcastRPC)
	callResult, err := tmClient.Call(client, callerAddress, calleeAddress, data)
	if err != nil {
		err = fmt.Errorf("error (%v) connnecting to node (%s) to query contract at (%s) with data (%X)",
			err.Error(), burrowNodeClient.broadcastRPC, calleeAddress, data)
		return
	}
	return callResult.Return, callResult.GasUsed, nil
}

// QueryContractCode executes the contract code at address with the given data but with provided code
func (burrowNodeClient *burrowNodeClient) QueryContractCode(address crypto.Address, code,
	data []byte) (ret []byte, gasUsed uint64, err error) {

	client := rpcClient.NewJSONRPCClient(burrowNodeClient.broadcastRPC)
	// TODO: [ben] Call and CallCode have an inconsistent signature; it makes sense for both to only
	// have a single address that is the contract to query.
	callResult, err := tmClient.CallCode(client, address, code, data)
	if err != nil {
		err = fmt.Errorf("error connnecting to node (%s) to query contract code at (%s) with data (%X) and code (%X): %v",
			burrowNodeClient.broadcastRPC, address, data, code, err.Error())
		return nil, uint64(0), err
	}
	return callResult.Return, callResult.GasUsed, nil
}

// GetAccount returns a copy of the account
func (burrowNodeClient *burrowNodeClient) GetAccount(address crypto.Address) (acm.Account, error) {
	client := rpcClient.NewJSONRPCClient(burrowNodeClient.broadcastRPC)
	account, err := tmClient.GetAccount(client, address)
	if err != nil {
		err = fmt.Errorf("error connecting to node (%s) to fetch account (%s): %s",
			burrowNodeClient.broadcastRPC, address, err.Error())
		return nil, err
	}
	if account == nil {
		err = fmt.Errorf("unknown account %s at node (%s)", address, burrowNodeClient.broadcastRPC)
		return nil, err
	}

	return account, nil
}

// DumpStorage returns the full storage for an acm.
func (burrowNodeClient *burrowNodeClient) DumpStorage(address crypto.Address) (*rpc.ResultDumpStorage, error) {
	client := rpcClient.NewJSONRPCClient(burrowNodeClient.broadcastRPC)
	resultStorage, err := tmClient.DumpStorage(client, address)
	if err != nil {
		return nil, fmt.Errorf("error connecting to node (%s) to get storage for account (%X): %s",
			burrowNodeClient.broadcastRPC, address, err.Error())
	}
	return resultStorage, nil
}

//--------------------------------------------------------------------------------------------
// Name registry

func (burrowNodeClient *burrowNodeClient) GetName(name string) (owner crypto.Address, data string,
	expirationBlock uint64, err error) {

	client := rpcClient.NewJSONRPCClient(burrowNodeClient.broadcastRPC)
	entryResult, err := tmClient.GetName(client, name)
	if err != nil {
		err = fmt.Errorf("error connecting to node (%s) to get name registrar entry for name (%s)",
			burrowNodeClient.broadcastRPC, name)
		return crypto.ZeroAddress, "", 0, err
	}
	// unwrap return results
	owner = entryResult.Owner
	data = entryResult.Data
	expirationBlock = entryResult.Expires
	return
}

//--------------------------------------------------------------------------------------------

func (burrowNodeClient *burrowNodeClient) ListValidators() (blockHeight uint64,
	bondedValidators, unbondingValidators []acm.Validator, err error) {

	client := rpcClient.NewJSONRPCClient(burrowNodeClient.broadcastRPC)
	validatorsResult, err := tmClient.ListValidators(client)
	if err != nil {
		err = fmt.Errorf("error connecting to node (%s) to get validators", burrowNodeClient.broadcastRPC)
		return
	}
	// unwrap return results
	blockHeight = validatorsResult.BlockHeight
	bondedValidators = make([]acm.Validator, len(validatorsResult.BondedValidators))
	for i, cv := range validatorsResult.BondedValidators {
		bondedValidators[i] = cv.Validator()
	}
	unbondingValidators = make([]acm.Validator, len(validatorsResult.UnbondingValidators))
	for i, cv := range validatorsResult.UnbondingValidators {
		unbondingValidators[i] = cv.Validator()
	}
	return
}

func (burrowNodeClient *burrowNodeClient) Logger() *logging.Logger {
	return burrowNodeClient.logger
}
