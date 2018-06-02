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

package rpc

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/hyperledger/burrow/crypto"
	ptypes "github.com/hyperledger/burrow/permission"

	"github.com/hyperledger/burrow/client"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/permission/snatives"
	"github.com/hyperledger/burrow/txs"
)

//------------------------------------------------------------------------------------
// core functions with string args.
// validates strings and forms transaction

func Send(nodeClient client.NodeClient, keyClient keys.KeyClient, pubkey, addr, toAddr, amtS, sequenceS string) (*txs.SendTx, error) {
	pub, amt, sequence, err := checkCommon(nodeClient, keyClient, pubkey, addr, amtS, sequenceS)
	if err != nil {
		return nil, err
	}

	if toAddr == "" {
		return nil, fmt.Errorf("destination address must be given with --to flag")
	}

	toAddress, err := addressFromHexString(toAddr)
	if err != nil {
		return nil, err
	}

	tx := txs.NewSendTx()
	tx.AddInputWithSequence(pub, amt, sequence)
	tx.AddOutput(toAddress, amt)

	return tx, nil
}

func Call(nodeClient client.NodeClient, keyClient keys.KeyClient, pubkey, addr, toAddr, amtS, sequenceS, gasS, feeS, data string) (*txs.CallTx, error) {
	pub, amt, sequence, err := checkCommon(nodeClient, keyClient, pubkey, addr, amtS, sequenceS)
	if err != nil {
		return nil, err
	}

	var toAddress *crypto.Address

	if toAddr != "" {
		address, err := addressFromHexString(toAddr)
		if err != nil {
			return nil, fmt.Errorf("toAddr is bad hex: %v", err)
		}
		toAddress = &address
	}

	fee, err := strconv.ParseUint(feeS, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("fee is misformatted: %v", err)
	}

	gas, err := strconv.ParseUint(gasS, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("gas is misformatted: %v", err)
	}

	dataBytes, err := hex.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("data is bad hex: %v", err)
	}

	tx := txs.NewCallTxWithSequence(pub, toAddress, dataBytes, amt, gas, fee, sequence)
	return tx, nil
}

func Name(nodeClient client.NodeClient, keyClient keys.KeyClient, pubkey, addr, amtS, sequenceS, feeS, name, data string) (*txs.NameTx, error) {
	pub, amt, sequence, err := checkCommon(nodeClient, keyClient, pubkey, addr, amtS, sequenceS)
	if err != nil {
		return nil, err
	}

	fee, err := strconv.ParseUint(feeS, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("fee is misformatted: %v", err)
	}

	tx := txs.NewNameTxWithSequence(pub, name, data, amt, fee, sequence)
	return tx, nil
}

func Permissions(nodeClient client.NodeClient, keyClient keys.KeyClient, pubkey, addrS, sequenceS string,
	action, target, permissionFlag, role, value string) (*txs.PermissionsTx, error) {

	pub, _, sequence, err := checkCommon(nodeClient, keyClient, pubkey, addrS, "0", sequenceS)
	if err != nil {
		return nil, err
	}
	permFlag, err := ptypes.PermStringToFlag(action)
	if err != nil {
		return nil, fmt.Errorf("could not convert action '%s' to PermFlag: %v", action, err)
	}
	permArgs := snatives.PermArgs{
		PermFlag: permFlag,
	}

	// Try and set each PermArg field for which a string has been provided we'll validate afterwards
	if target != "" {
		address, err := crypto.AddressFromHexString(target)
		if err != nil {
			return nil, err
		}
		permArgs.Address = &address
	}

	if value != "" {
		valueBool := value == "true"
		permArgs.Value = &valueBool
		if !valueBool && value != "false" {
			return nil, fmt.Errorf("did not recognise value %s as boolean, use 'true' or 'false'", value)
		}
		permArgs.Value = &valueBool
	}

	if permissionFlag != "" {
		permission, err := ptypes.PermStringToFlag(permissionFlag)
		if err != nil {
			return nil, err
		}
		permArgs.Permission = &permission
	}

	if role != "" {
		permArgs.Role = &role
	}

	err = permArgs.EnsureValid()
	if err != nil {
		return nil, err
	}

	tx := txs.NewPermissionsTxWithSequence(pub, permArgs, sequence)
	return tx, nil
}

func Bond(nodeClient client.NodeClient, keyClient keys.KeyClient, pubkey, unbondAddr, amtS, sequenceS string) (*txs.BondTx, error) {
	return nil, fmt.Errorf("Bond Transaction formation to be implemented on 0.12.0")
	// pub, amt, sequence, err := checkCommon(nodeAddr, signAddr, pubkey, "", amtS, sequenceS)
	// if err != nil {
	// 	return nil, err
	// }
	// var pubKey acm.PublicKeyEd25519
	// var unbondAddrBytes []byte

	// if unbondAddr == "" {
	// 	pkb, _ := hex.DecodeString(pubkey)
	// 	copy(pubKey[:], pkb)
	// 	unbondAddrBytes = pubKey.Address()
	// } else {
	// 	unbondAddrBytes, err = hex.DecodeString(unbondAddr)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("unbondAddr is bad hex: %v", err)
	// 	}

	// }

	// tx, err := types.NewBondTx(pub)
	// if err != nil {
	// 	return nil, err
	// }
	// tx.AddInputWithSequence(pub, amt, int(sequence))
	// tx.AddOutput(unbondAddrBytes, amt)

	// return tx, nil
}

func Unbond(addrS, heightS string) (*txs.UnbondTx, error) {
	return nil, fmt.Errorf("Unbond Transaction formation to be implemented on 0.12.0")
	// if addrS == "" {
	// 	return nil, fmt.Errorf("Validator address must be given with --addr flag")
	// }

	// addrBytes, err := hex.DecodeString(addrS)
	// if err != nil {
	// 	return nil, fmt.Errorf("addr is bad hex: %v", err)
	// }

	// height, err := strconv.ParseInt(heightS, 10, 32)
	// if err != nil {
	// 	return nil, fmt.Errorf("height is misformatted: %v", err)
	// }

	// return &types.UnbondTx{
	// 	Address: addrBytes,
	// 	Height:  int(height),
	// }, nil
}

func Rebond(addrS, heightS string) (*txs.RebondTx, error) {
	return nil, fmt.Errorf("Rebond Transaction formation to be implemented on 0.12.0")
	// 	if addrS == "" {
	// 		return nil, fmt.Errorf("Validator address must be given with --addr flag")
	// 	}

	// 	addrBytes, err := hex.DecodeString(addrS)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("addr is bad hex: %v", err)
	// 	}

	// 	height, err := strconv.ParseInt(heightS, 10, 32)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("height is misformatted: %v", err)
	// 	}

	// 	return &types.RebondTx{
	// 		Address: addrBytes,
	// 		Height:  int(height),
	// 	}, nil
}

type TxResult struct {
	BlockHash []byte // all txs get in a block
	Hash      []byte // all txs get a hash

	// only CallTx
	Address   *crypto.Address // only for new contracts
	Return    []byte
	Exception string

	//TODO: make Broadcast() errors more responsive so we
	// can differentiate mempool errors from other
}

// Preserve
func SignAndBroadcast(chainID string, nodeClient client.NodeClient, keyClient keys.KeyClient, tx txs.Tx, sign,
	broadcast, wait bool) (txResult *TxResult, err error) {

	var inputAddr crypto.Address
	if sign {
		inputAddr, tx, err = signTx(keyClient, chainID, tx)
		if err != nil {
			return nil, err
		}
	}

	if broadcast {
		if wait {
			var wsClient client.NodeWebsocketClient
			wsClient, err = nodeClient.DeriveWebsocketClient()
			if err != nil {
				return nil, err
			}
			var confirmationChannel chan client.Confirmation
			confirmationChannel, err = wsClient.WaitForConfirmation(tx, chainID, inputAddr)
			if err != nil {
				return nil, err
			}
			defer func() {
				if err != nil {
					// if broadcast threw an error, just return
					return
				}
				if txResult == nil {
					err = fmt.Errorf("txResult unexpectedly not initialised in SignAndBroadcast")
					return
				}
				confirmation := <-confirmationChannel
				if confirmation.Error != nil {
					err = fmt.Errorf("encountered error waiting for event: %s", confirmation.Error)
					return
				}
				if confirmation.Exception != nil {
					err = fmt.Errorf("encountered Exception from chain: %s", confirmation.Exception)
					return
				}
				txResult.BlockHash = confirmation.BlockHash
				txResult.Exception = ""
				eventDataTx := confirmation.EventDataTx
				if eventDataTx == nil {
					err = fmt.Errorf("EventDataTx was nil")
					return
				}
				txResult.Return = eventDataTx.Return
			}()
		}

		var receipt *txs.Receipt
		receipt, err = nodeClient.Broadcast(tx)
		if err != nil {
			return nil, err
		}
		txResult = &TxResult{
			Hash: receipt.TxHash,
		}
		// NOTE: [ben] is this consistent with the Ethereum protocol?  It should seem
		// reasonable to get this returned from the chain directly.  Alternatively,
		// the benefit is that the we don't need to trust the chain node
		if tx_, ok := tx.(*txs.CallTx); ok {
			if tx_.Address == nil {
				address := crypto.NewContractAddress(tx_.Input.Address, tx_.Input.Sequence)
				txResult.Address = &address
			}
		}
	}
	return
}

func addressFromHexString(addrString string) (crypto.Address, error) {
	addrBytes, err := hex.DecodeString(addrString)
	if err != nil {
		return crypto.Address{}, err
	}
	return crypto.AddressFromBytes(addrBytes)
}
