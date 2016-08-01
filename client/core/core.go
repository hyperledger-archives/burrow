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

package core

package core

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eris-ltd/tendermint/account"
	ptypes "github.com/eris-ltd/tendermint/permission/types"
	rtypes "github.com/eris-ltd/tendermint/rpc/core/types"
	cclient "github.com/eris-ltd/tendermint/rpc/core_client"
	"github.com/eris-ltd/txs"
)

var (
	MaxCommitWaitTimeSeconds = 20
)

//------------------------------------------------------------------------------------
// core functions with string args.
// validates strings and forms transaction

func Send(nodeAddr, signAddr, pubkey, addr, toAddr, amtS, nonceS string) (*txs.SendTx, error) {
	pub, amt, nonce, err := checkCommon(nodeAddr, signAddr, pubkey, addr, amtS, nonceS)
	if err != nil {
		return nil, err
	}

	if toAddr == "" {
		return nil, fmt.Errorf("destination address must be given with --to flag")
	}

	toAddrBytes, err := hex.DecodeString(toAddr)
	if err != nil {
		return nil, fmt.Errorf("toAddr is bad hex: %v", err)
	}

	tx := txs.NewSendTx()
	tx.AddInputWithNonce(pub, amt, int(nonce))
	tx.AddOutput(toAddrBytes, amt)

	return tx, nil
}

func Call(nodeAddr, signAddr, pubkey, addr, toAddr, amtS, nonceS, gasS, feeS, data string) (*txs.CallTx, error) {
	pub, amt, nonce, err := checkCommon(nodeAddr, signAddr, pubkey, addr, amtS, nonceS)
	if err != nil {
		return nil, err
	}

	toAddrBytes, err := hex.DecodeString(toAddr)
	if err != nil {
		return nil, fmt.Errorf("toAddr is bad hex: %v", err)
	}

	fee, err := strconv.ParseInt(feeS, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("fee is misformatted: %v", err)
	}

	gas, err := strconv.ParseInt(gasS, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("gas is misformatted: %v", err)
	}

	dataBytes, err := hex.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("data is bad hex: %v", err)
	}

	tx := types.NewCallTxWithNonce(pub, toAddrBytes, dataBytes, amt, gas, fee, int(nonce))
	return tx, nil
}

func Name(nodeAddr, signAddr, pubkey, addr, amtS, nonceS, feeS, name, data string) (*txs.NameTx, error) {
	pub, amt, nonce, err := checkCommon(nodeAddr, signAddr, pubkey, addr, amtS, nonceS)
	if err != nil {
		return nil, err
	}

	fee, err := strconv.ParseInt(feeS, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("fee is misformatted: %v", err)
	}

	tx := types.NewNameTxWithNonce(pub, name, data, amt, fee, int(nonce))
	return tx, nil
}

type PermFunc struct {
	Name string
	Args string
}

var PermsFuncs = []PermFunc{
	{"set_base", "address, permission flag, value"},
	{"unset_base", "address, permission flag"},
	{"set_global", "permission flag, value"},
	{"add_role", "address, role"},
	{"rm_role", "address, role"},
}

func Permissions(nodeAddr, signAddr, pubkey, addrS, nonceS, permFunc string, argsS []string) (*txs.PermissionsTx, error) {
	pub, _, nonce, err := checkCommon(nodeAddr, signAddr, pubkey, addrS, "0", nonceS)
	if err != nil {
		return nil, err
	}
	var args ptypes.PermArgs
	switch permFunc {
	case "set_base":
		addr, pF, err := decodeAddressPermFlag(argsS[0], argsS[1])
		if err != nil {
			return nil, err
		}
		if len(argsS) != 3 {
			return nil, fmt.Errorf("set_base also takes a value (true or false)")
		}
		var value bool
		if argsS[2] == "true" {
			value = true
		} else if argsS[2] == "false" {
			value = false
		} else {
			return nil, fmt.Errorf("Unknown value %s", argsS[2])
		}
		args = &ptypes.SetBaseArgs{addr, pF, value}
	case "unset_base":
		addr, pF, err := decodeAddressPermFlag(argsS[0], argsS[1])
		if err != nil {
			return nil, err
		}
		args = &ptypes.UnsetBaseArgs{addr, pF}
	case "set_global":
		pF, err := ptypes.PermStringToFlag(argsS[0])
		if err != nil {
			return nil, err
		}
		var value bool
		if argsS[1] == "true" {
			value = true
		} else if argsS[1] == "false" {
			value = false
		} else {
			return nil, fmt.Errorf("Unknown value %s", argsS[1])
		}
		args = &ptypes.SetGlobalArgs{pF, value}
	case "add_role":
		addr, err := hex.DecodeString(argsS[0])
		if err != nil {
			return nil, err
		}
		args = &ptypes.AddRoleArgs{addr, argsS[1]}
	case "rm_role":
		addr, err := hex.DecodeString(argsS[0])
		if err != nil {
			return nil, err
		}
		args = &ptypes.RmRoleArgs{addr, argsS[1]}
	default:
		return nil, fmt.Errorf("Invalid permission function for use in PermissionsTx: %s", permFunc)
	}
	// args := snativeArgs(
	tx := types.NewPermissionsTxWithNonce(pub, args, int(nonce))
	return tx, nil
}

func decodeAddressPermFlag(addrS, permFlagS string) (addr []byte, pFlag ptypes.PermFlag, err error) {
	if addr, err = hex.DecodeString(addrS); err != nil {
		return
	}
	if pFlag, err = ptypes.PermStringToFlag(permFlagS); err != nil {
		return
	}
	return
}

type NameGetter struct {
	client cclient.Client
}

func (n NameGetter) GetNameRegEntry(name string) *txs.NameRegEntry {
	entry, err := n.client.GetName(name)
	if err != nil {
		panic(err)
	}
	return entry.Entry
}

/*
func coreNewAccount(nodeAddr, pubkey, chainID string) (*types.NewAccountTx, error) {
	pub, _, _, err := checkCommon(nodeAddr, pubkey, "", "0", "0")
	if err != nil {
		return nil, err
	}

	client := cclient.NewClient(nodeAddr, "HTTP")
	return types.NewNewAccountTx(NameGetter{client}, pub, chainID)
}
*/

func Bond(nodeAddr, signAddr, pubkey, unbondAddr, amtS, nonceS string) (*txs.BondTx, error) {
	pub, amt, nonce, err := checkCommon(nodeAddr, signAddr, pubkey, "", amtS, nonceS)
	if err != nil {
		return nil, err
	}
	var pubKey account.PubKeyEd25519
	var unbondAddrBytes []byte

	if unbondAddr == "" {
		pkb, _ := hex.DecodeString(pubkey)
		copy(pubKey[:], pkb)
		unbondAddrBytes = pubKey.Address()
	} else {
		unbondAddrBytes, err = hex.DecodeString(unbondAddr)
		if err != nil {
			return nil, fmt.Errorf("unbondAddr is bad hex: %v", err)
		}

	}

	tx, err := types.NewBondTx(pub)
	if err != nil {
		return nil, err
	}
	tx.AddInputWithNonce(pub, amt, int(nonce))
	tx.AddOutput(unbondAddrBytes, amt)

	return tx, nil
}

func Unbond(addrS, heightS string) (*txs.UnbondTx, error) {
	if addrS == "" {
		return nil, fmt.Errorf("Validator address must be given with --addr flag")
	}

	addrBytes, err := hex.DecodeString(addrS)
	if err != nil {
		return nil, fmt.Errorf("addr is bad hex: %v", err)
	}

	height, err := strconv.ParseInt(heightS, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("height is misformatted: %v", err)
	}

	return &types.UnbondTx{
		Address: addrBytes,
		Height:  int(height),
	}, nil
}

func Rebond(addrS, heightS string) (*txs.RebondTx, error) {
	if addrS == "" {
		return nil, fmt.Errorf("Validator address must be given with --addr flag")
	}

	addrBytes, err := hex.DecodeString(addrS)
	if err != nil {
		return nil, fmt.Errorf("addr is bad hex: %v", err)
	}

	height, err := strconv.ParseInt(heightS, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("height is misformatted: %v", err)
	}

	return &types.RebondTx{
		Address: addrBytes,
		Height:  int(height),
	}, nil
}

//------------------------------------------------------------------------------------
// sign and broadcast

func Pub(addr, rpcAddr string) (pubBytes []byte, err error) {
	args := map[string]string{
		"addr": addr,
	}
	pubS, err := RequestResponse(rpcAddr, "pub", args)
	if err != nil {
		return
	}
	return hex.DecodeString(pubS)
}

func Sign(signBytes, signAddr, signRPC string) (sig [64]byte, err error) {
	args := map[string]string{
		"msg":  signBytes,
		"hash": signBytes, // backwards compatibility
		"addr": signAddr,
	}
	sigS, err := RequestResponse(signRPC, "sign", args)
	if err != nil {
		return
	}
	sigBytes, err := hex.DecodeString(sigS)
	if err != nil {
		return
	}
	copy(sig[:], sigBytes)
	return
}

func Broadcast(tx types.Tx, broadcastRPC string) (*txs.Receipt, error) {
	client := cclient.NewClient(broadcastRPC, "JSONRPC")
	rec, err := client.BroadcastTx(tx)
	if err != nil {
		return nil, err
	}
	return &rec.Receipt, nil
}

//------------------------------------------------------------------------------------
// utils for talking to the key server

type HTTPResponse struct {
	Response string
	Error    string
}

func RequestResponse(addr, method string, args map[string]string) (string, error) {
	b, err := json.Marshal(args)
	if err != nil {
		return "", err
	}
	endpoint := fmt.Sprintf("%s/%s", addr, method)
	logger.Debugf("Sending request body (%s): %s\n", endpoint, string(b))
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	res, errS, err := requestResponse(req)
	if err != nil {
		return "", fmt.Errorf("Error calling eris-keys at %s: %s", endpoint, err.Error())
	}
	if errS != "" {
		return "", fmt.Errorf("Error (string) calling eris-keys at %s: %s", endpoint, errS)
	}
	return res, nil
}

func requestResponse(req *http.Request) (string, string, error) {
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode >= 400 {
		return "", "", fmt.Errorf(resp.Status)
	}
	return unpackResponse(resp)
}

func unpackResponse(resp *http.Response) (string, string, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	r := new(HTTPResponse)
	if err := json.Unmarshal(b, r); err != nil {
		return "", "", err
	}
	return r.Response, r.Error, nil
}

//------------------------------------------------------------------------------------
// sign and broadcast convenience

// tx has either one input or we default to the first one (ie for send/bond)
// TODO: better support for multisig and bonding
func signTx(signAddr, chainID string, tx_ txs.Tx) ([]byte, txs.Tx, error) {
	signBytes := fmt.Sprintf("%X", account.SignBytes(chainID, tx_))
	var inputAddr []byte
	var sigED account.SignatureEd25519
	switch tx := tx_.(type) {
	case *types.SendTx:
		inputAddr = tx.Inputs[0].Address
		defer func(s *account.SignatureEd25519) { tx.Inputs[0].Signature = *s }(&sigED)
	case *types.NameTx:
		inputAddr = tx.Input.Address
		defer func(s *account.SignatureEd25519) { tx.Input.Signature = *s }(&sigED)
	case *types.CallTx:
		inputAddr = tx.Input.Address
		defer func(s *account.SignatureEd25519) { tx.Input.Signature = *s }(&sigED)
	case *types.PermissionsTx:
		inputAddr = tx.Input.Address
		defer func(s *account.SignatureEd25519) { tx.Input.Signature = *s }(&sigED)
	case *types.BondTx:
		inputAddr = tx.Inputs[0].Address
		defer func(s *account.SignatureEd25519) {
			tx.Signature = *s
			tx.Inputs[0].Signature = *s
		}(&sigED)
	case *types.UnbondTx:
		inputAddr = tx.Address
		defer func(s *account.SignatureEd25519) { tx.Signature = *s }(&sigED)
	case *types.RebondTx:
		inputAddr = tx.Address
		defer func(s *account.SignatureEd25519) { tx.Signature = *s }(&sigED)
	}
	addrHex := fmt.Sprintf("%X", inputAddr)
	sig, err := Sign(signBytes, addrHex, signAddr)
	if err != nil {
		return nil, nil, err
	}
	sigED = account.SignatureEd25519(sig)
	logger.Debugf("SIG: %X\n", sig)
	return inputAddr, tx_, nil
}

type TxResult struct {
	BlockHash []byte // all txs get in a block
	Hash      []byte // all txs get a hash

	// only CallTx
	Address   []byte // only for new contracts
	Return    []byte
	Exception string

	//TODO: make Broadcast() errors more responsive so we
	// can differentiate mempool errors from other
}

func SignAndBroadcast(chainID, nodeAddr, signAddr string, tx types.Tx, sign, broadcast, wait bool) (txResult *TxResult, err error) {
	var inputAddr []byte
	if sign {
		inputAddr, tx, err = signTx(signAddr, chainID, tx)
		if err != nil {
			return nil, err
		}
	}

	if broadcast {
		if wait {
			var ch chan Msg
			ch, err = subscribeAndWait(tx, chainID, nodeAddr, inputAddr)
			if err != nil {
				return nil, err
			} else {
				defer func() {
					if err != nil {
						// if broadcast threw an error, just return
						return
					}
					logger.Debugln("Waiting for tx to be committed ...")
					msg := <-ch
					if msg.Error != nil {
						logger.Infof("Encountered error waiting for event: %v\n", msg.Error)
						err = msg.Error
					} else {
						txResult.BlockHash = msg.BlockHash
						txResult.Return = msg.Value
						txResult.Exception = msg.Exception
					}
				}()
			}
		}
		var receipt *rtypes.Receipt
		receipt, err = Broadcast(tx, nodeAddr)
		if err != nil {
			return nil, err
		}
		txResult = &TxResult{
			Hash: receipt.TxHash,
		}
		if tx_, ok := tx.(*types.CallTx); ok {
			if len(tx_.Address) == 0 {
				txResult.Address = types.NewContractAddress(tx_.Input.Address, tx_.Input.Sequence)
			}
		}
	}
	return
}

//------------------------------------------------------------------------------------
// wait for events

type Msg struct {
	BlockHash []byte
	Value     []byte
	Exception string
	Error     error
}

func subscribeAndWait(tx types.Tx, chainID, nodeAddr string, inputAddr []byte) (chan Msg, error) {
	// subscribe to event and wait for tx to be committed
	wsAddr := strings.TrimPrefix(nodeAddr, "http://")
	wsAddr = "ws://" + wsAddr + "websocket"
	logger.Debugf("Websocket Address %s\n", wsAddr)
	wsClient := cclient.NewWSClient(wsAddr)
	wsClient.Start()
	eid := types.EventStringAccInput(inputAddr)
	if err := wsClient.Subscribe(eid); err != nil {
		return nil, fmt.Errorf("Error subscribing to AccInput event: %v", err)
	}
	if err := wsClient.Subscribe(types.EventStringNewBlock()); err != nil {
		return nil, fmt.Errorf("Error subscribing to NewBlock event: %v", err)
	}

	resultChan := make(chan Msg, 1)

	var latestBlockHash []byte

	// Read message
	go func() {
		for {
			result := <-wsClient.EventsCh
			// if its a block, remember the block hash
			blockData, ok := result.Data.(types.EventDataNewBlock)
			if ok {
				logger.Infoln(blockData.Block)
				latestBlockHash = blockData.Block.Hash()
				continue
			}

			// we don't accept events unless they came after a new block (ie. in)
			if latestBlockHash == nil {
				continue
			}

			if result.Event != eid {
				logger.Debugf("received unsolicited event! Got %s, expected %s\n", result.Event, eid)
				continue
			}

			data, ok := result.Data.(types.EventDataTx)
			if !ok {
				resultChan <- Msg{Error: fmt.Errorf("response error: expected result.Data to be *types.EventDataTx")}
				return
			}

			if !bytes.Equal(types.TxID(chainID, data.Tx), types.TxID(chainID, tx)) {
				logger.Debugf("Received event for same input from another transaction: %X\n", types.TxID(chainID, data.Tx))
				continue
			}

			if data.Exception != "" {
				resultChan <- Msg{BlockHash: latestBlockHash, Value: data.Return, Exception: data.Exception}
				return
			}

			// GOOD!
			resultChan <- Msg{BlockHash: latestBlockHash, Value: data.Return}
			return
		}
	}()

	// txs should take no more than 10 seconds
	timeoutTicker := time.Tick(time.Duration(MaxCommitWaitTimeSeconds) * time.Second)

	go func() {
		<-timeoutTicker
		resultChan <- Msg{Error: fmt.Errorf("timed out waiting for event")}
		return
	}()
	return resultChan, nil
}

//------------------------------------------------------------------------------------
// convenience function

func checkCommon(nodeAddr, signAddr, pubkey, addr, amtS, nonceS string) (pub account.PubKey, amt int64, nonce int64, err error) {
	if amtS == "" {
		err = fmt.Errorf("input must specify an amount with the --amt flag")
		return
	}

	var pubKeyBytes []byte
	if pubkey == "" && addr == "" {
		err = fmt.Errorf("at least one of --pubkey or --addr must be given")
		return
	} else if pubkey != "" {
		if addr != "" {
			// NOTE: if --addr given byt MINTX_PUBKEY is set, the pubkey still wins
			// TODO: fix this
			logger.Infoln("you have specified both a pubkey and an address. the pubkey takes precedent")
		}
		pubKeyBytes, err = hex.DecodeString(pubkey)
		if err != nil {
			err = fmt.Errorf("pubkey is bad hex: %v", err)
			return
		}
	} else {
		// grab the pubkey from eris-keys
		pubKeyBytes, err = Pub(addr, signAddr)
		if err != nil {
			err = fmt.Errorf("failed to fetch pubkey for address (%s): %v", addr, err)
			return
		}

	}

	if len(pubKeyBytes) == 0 {
		err = fmt.Errorf("Error resolving public key")
		return
	}

	amt, err = strconv.ParseInt(amtS, 10, 64)
	if err != nil {
		err = fmt.Errorf("amt is misformatted: %v", err)
	}

	var pubArray [32]byte
	copy(pubArray[:], pubKeyBytes)
	pub = account.PubKeyEd25519(pubArray)
	addrBytes := pub.Address()

	if nonceS == "" {
		if nodeAddr == "" {
			err = fmt.Errorf("input must specify a nonce with the --nonce flag or use --node-addr (or MINTX_NODE_ADDR) to fetch the nonce from a node")
			return
		}

		// fetch nonce from node
		client := cclient.NewClient(nodeAddr, "HTTP")
		ac, err2 := client.GetAccount(addrBytes)
		if err2 != nil {
			err = fmt.Errorf("Error connecting to node (%s) to fetch nonce: %s", nodeAddr, err2.Error())
			return
		}
		if ac == nil || ac.Account == nil {
			err = fmt.Errorf("unknown account %X", addrBytes)
			return
		}
		nonce = int64(ac.Account.Sequence) + 1
	} else {
		nonce, err = strconv.ParseInt(nonceS, 10, 64)
		if err != nil {
			err = fmt.Errorf("nonce is misformatted: %v", err)
			return
		}
	}

	return
}
