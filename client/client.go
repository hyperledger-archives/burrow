// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package client

import (
	"fmt"
	// "strings"

	"github.com/tendermint/go-rpc/client"
	// Note [ben]: this is included to silence the logger from tendermint/go-rpc/client
	// see func init()
	tendermint_log "github.com/tendermint/log15"

	"github.com/eris-ltd/eris-cli/log"

	acc "github.com/eris-ltd/eris-db/account"
	consensus_types "github.com/eris-ltd/eris-db/consensus/types"
	core_types "github.com/eris-ltd/eris-db/core/types"
	tendermint_client "github.com/eris-ltd/eris-db/rpc/tendermint/client"
	tendermint_types "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
	"github.com/eris-ltd/eris-db/txs"
)

type NodeClient interface {
	Broadcast(transaction txs.Tx) (*txs.Receipt, error)
	DeriveWebsocketClient() (nodeWsClient NodeWebsocketClient, err error)

	Status() (ChainId []byte, ValidatorPublicKey []byte, LatestBlockHash []byte,
		LatestBlockHeight int, LatestBlockTime int64, err error)
	GetAccount(address []byte) (*acc.Account, error)
	QueryContract(callerAddress, calleeAddress, data []byte) (ret []byte, gasUsed int64, err error)
	QueryContractCode(address, code, data []byte) (ret []byte, gasUsed int64, err error)

	DumpStorage(address []byte) (storage *core_types.Storage, err error)
	GetName(name string) (owner []byte, data string, expirationBlock int, err error)
	ListValidators() (blockHeight int, bondedValidators, unbondingValidators []consensus_types.Validator, err error)
}

type NodeWebsocketClient interface {
	Subscribe(eventId string) error
	Unsubscribe(eventId string) error

	WaitForConfirmation(tx txs.Tx, chainId string, inputAddr []byte) (chan Confirmation, error)
	Close()
}

// NOTE [ben] Compiler check to ensure ErisNodeClient successfully implements
// eris-db/client.NodeClient
var _ NodeClient = (*ErisNodeClient)(nil)

// Eris-Client is a simple struct exposing the client rpc methods

type ErisNodeClient struct {
	broadcastRPC string
}

// ErisKeyClient.New returns a new eris-keys client for provided rpc location
// Eris-keys connects over http request-responses
func NewErisNodeClient(rpcString string) *ErisNodeClient {
	return &ErisNodeClient{
		broadcastRPC: rpcString,
	}
}

// Note [Ben]: This is a hack to silence Tendermint logger from tendermint/go-rpc
// it needs to be initialised before go-rpc, hence it's placement here.
func init() {
	h := tendermint_log.LvlFilterHandler(tendermint_log.LvlWarn, tendermint_log.StdoutHandler)
	tendermint_log.Root().SetHandler(h)
}

//------------------------------------------------------------------------------------
// broadcast to blockchain node
// NOTE: [ben] Eris Client first continues from tendermint rpc, but will have handshake to negotiate
// protocol version for moving towards rpc/v1

func (erisNodeClient *ErisNodeClient) Broadcast(tx txs.Tx) (*txs.Receipt, error) {
	client := rpcclient.NewClientURI(erisNodeClient.broadcastRPC)
	receipt, err := tendermint_client.BroadcastTx(client, tx)
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

func (erisNodeClient *ErisNodeClient) DeriveWebsocketClient() (nodeWsClient NodeWebsocketClient, err error) {
	var wsAddr string
	// TODO: clean up this inherited mess on dealing with the address prefixes.
	nodeAddr := erisNodeClient.broadcastRPC
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
	log.WithFields(log.Fields{
		"websocket address": wsAddr,
		"endpoint":          "/websocket",
	}).Debug("Subscribing to websocket address")
	wsClient := rpcclient.NewWSClient(wsAddr, "/websocket")
	if _, err = wsClient.Start(); err != nil {
		return nil, err
	}
	derivedErisNodeWebsocketClient := &ErisNodeWebsocketClient{
		tendermintWebsocket: wsClient,
	}
	return derivedErisNodeWebsocketClient, nil
}

//------------------------------------------------------------------------------------
// RPC requests other than transaction related

// Status returns the ChainId (GenesisHash), validator's PublicKey, latest block hash
// the block height and the latest block time.
func (erisNodeClient *ErisNodeClient) Status() (GenesisHash []byte, ValidatorPublicKey []byte, LatestBlockHash []byte, LatestBlockHeight int, LatestBlockTime int64, err error) {
	client := rpcclient.NewClientJSONRPC(erisNodeClient.broadcastRPC)
	res, err := tendermint_client.Status(client)
	if err != nil {
		err = fmt.Errorf("Error connecting to node (%s) to get status: %s",
			erisNodeClient.broadcastRPC, err.Error())
		return nil, nil, nil, int(0), int64(0), err
	}

	// unwrap return results
	GenesisHash = res.GenesisHash
	ValidatorPublicKey = res.PubKey.Bytes()
	LatestBlockHash = res.LatestBlockHash
	LatestBlockHeight = res.LatestBlockHeight
	LatestBlockTime = res.LatestBlockTime
	return
}

func (erisNodeClient *ErisNodeClient) ChainId() (ChainName, ChainId string, GenesisHash []byte, err error) {
	client := rpcclient.NewClientJSONRPC(erisNodeClient.broadcastRPC)
	chainIdResult, err := tendermint_client.ChainId(client)
	if err != nil {
		err = fmt.Errorf("Error connecting to node (%s) to get chain id: %s",
			erisNodeClient.broadcastRPC, err.Error())
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
func (erisNodeClient *ErisNodeClient) QueryContract(callerAddress, calleeAddress, data []byte) (ret []byte, gasUsed int64, err error) {
	client := rpcclient.NewClientJSONRPC(erisNodeClient.broadcastRPC)
	callResult, err := tendermint_client.Call(client, callerAddress, calleeAddress, data)
	if err != nil {
		err = fmt.Errorf("Error connnecting to node (%s) to query contract at (%X) with data (%X)",
			erisNodeClient.broadcastRPC, calleeAddress, data, err.Error())
		return nil, int64(0), err
	}
	return callResult.Return, callResult.GasUsed, nil
}

// QueryContractCode executes the contract code at address with the given data but with provided code
func (erisNodeClient *ErisNodeClient) QueryContractCode(address, code, data []byte) (ret []byte, gasUsed int64, err error) {
	client := rpcclient.NewClientJSONRPC(erisNodeClient.broadcastRPC)
	// TODO: [ben] Call and CallCode have an inconsistent signature; it makes sense for both to only
	// have a single address that is the contract to query.
	callResult, err := tendermint_client.CallCode(client, address, code, data)
	if err != nil {
		err = fmt.Errorf("Error connnecting to node (%s) to query contract code at (%X) with data (%X) and code (%X)",
			erisNodeClient.broadcastRPC, address, data, code, err.Error())
		return nil, int64(0), err
	}
	return callResult.Return, callResult.GasUsed, nil
}

// GetAccount returns a copy of the account
func (erisNodeClient *ErisNodeClient) GetAccount(address []byte) (*acc.Account, error) {
	client := rpcclient.NewClientJSONRPC(erisNodeClient.broadcastRPC)
	account, err := tendermint_client.GetAccount(client, address)
	if err != nil {
		err = fmt.Errorf("Error connecting to node (%s) to fetch account (%X): %s",
			erisNodeClient.broadcastRPC, address, err.Error())
		return nil, err
	}
	if account == nil {
		err = fmt.Errorf("Unknown account %X at node (%s)", address, erisNodeClient.broadcastRPC)
		return nil, err
	}

	return account.Copy(), nil
}

// DumpStorage returns the full storage for an account.
func (erisNodeClient *ErisNodeClient) DumpStorage(address []byte) (storage *core_types.Storage, err error) {
	client := rpcclient.NewClientJSONRPC(erisNodeClient.broadcastRPC)
	resultStorage, err := tendermint_client.DumpStorage(client, address)
	if err != nil {
		err = fmt.Errorf("Error connecting to node (%s) to get storage for account (%X): %s",
			erisNodeClient.broadcastRPC, address, err.Error())
		return nil, err
	}
	// UnwrapResultDumpStorage is an inefficient full deep copy,
	// to transform the type to /core/types.Storage
	// TODO: removing go-wire and go-rpc allows us to collapse these types
	storage = tendermint_types.UnwrapResultDumpStorage(resultStorage)
	return
}

//--------------------------------------------------------------------------------------------
// Name registry

func (erisNodeClient *ErisNodeClient) GetName(name string) (owner []byte, data string, expirationBlock int, err error) {
	client := rpcclient.NewClientJSONRPC(erisNodeClient.broadcastRPC)
	entryResult, err := tendermint_client.GetName(client, name)
	if err != nil {
		err = fmt.Errorf("Error connecting to node (%s) to get name registrar entry for name (%s)",
			erisNodeClient.broadcastRPC, name)
		return nil, "", 0, err
	}
	// unwrap return results
	owner = entryResult.Owner
	data = entryResult.Data
	expirationBlock = entryResult.Expires
	return
}

//--------------------------------------------------------------------------------------------

func (erisNodeClient *ErisNodeClient) ListValidators() (blockHeight int,
	bondedValidators []consensus_types.Validator, unbondingValidators []consensus_types.Validator, err error) {
	client := rpcclient.NewClientJSONRPC(erisNodeClient.broadcastRPC)
	validatorsResult, err := tendermint_client.ListValidators(client)
	if err != nil {
		err = fmt.Errorf("Error connecting to node (%s) to get validators",
			erisNodeClient.broadcastRPC)
		return 0, nil, nil, err
	}
	// unwrap return results
	blockHeight = validatorsResult.BlockHeight
	bondedValidators = validatorsResult.BondedValidators
	unbondingValidators = validatorsResult.UnbondingValidators
	return
}
