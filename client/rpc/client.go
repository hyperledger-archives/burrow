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

	ptypes "github.com/hyperledger/burrow/permission/types"

	"github.com/hyperledger/burrow/client"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/txs"
)

//------------------------------------------------------------------------------------
// core functions with string args.
// validates strings and forms transaction

func Send(nodeClient client.NodeClient, keyClient keys.KeyClient, pubkey, addr, toAddr, amtS, nonceS string) (*txs.SendTx, error) {
	pub, amt, nonce, err := checkCommon(nodeClient, keyClient, pubkey, addr, amtS, nonceS)
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

func Call(nodeClient client.NodeClient, keyClient keys.KeyClient, pubkey, addr, toAddr, amtS, nonceS, gasS, feeS, data string) (*txs.CallTx, error) {
	pub, amt, nonce, err := checkCommon(nodeClient, keyClient, pubkey, addr, amtS, nonceS)
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

	tx := txs.NewCallTxWithNonce(pub, toAddrBytes, dataBytes, amt, gas, fee, int(nonce))
	return tx, nil
}

func Name(nodeClient client.NodeClient, keyClient keys.KeyClient, pubkey, addr, amtS, nonceS, feeS, name, data string) (*txs.NameTx, error) {
	pub, amt, nonce, err := checkCommon(nodeClient, keyClient, pubkey, addr, amtS, nonceS)
	if err != nil {
		return nil, err
	}

	fee, err := strconv.ParseInt(feeS, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("fee is misformatted: %v", err)
	}

	tx := txs.NewNameTxWithNonce(pub, name, data, amt, fee, int(nonce))
	return tx, nil
}

func Permissions(nodeClient client.NodeClient, keyClient keys.KeyClient, pubkey, addrS, nonceS, permFunc string, argsS []string) (*txs.PermissionsTx, error) {
	pub, _, nonce, err := checkCommon(nodeClient, keyClient, pubkey, addrS, "0", nonceS)
	if err != nil {
		return nil, err
	}
	var args ptypes.PermArgs
	switch permFunc {
	case "setBase":
		addr, pF, err := decodeAddressPermFlag(argsS[0], argsS[1])
		if err != nil {
			return nil, err
		}
		if len(argsS) != 3 {
			return nil, fmt.Errorf("setBase also takes a value (true or false)")
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
	case "unsetBase":
		addr, pF, err := decodeAddressPermFlag(argsS[0], argsS[1])
		if err != nil {
			return nil, err
		}
		args = &ptypes.UnsetBaseArgs{addr, pF}
	case "setGlobal":
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
	case "addRole":
		addr, err := hex.DecodeString(argsS[0])
		if err != nil {
			return nil, err
		}
		args = &ptypes.AddRoleArgs{addr, argsS[1]}
	case "removeRole":
		addr, err := hex.DecodeString(argsS[0])
		if err != nil {
			return nil, err
		}
		args = &ptypes.RmRoleArgs{addr, argsS[1]}
	default:
		return nil, fmt.Errorf("Invalid permission function for use in PermissionsTx: %s", permFunc)
	}
	// args := snativeArgs(
	tx := txs.NewPermissionsTxWithNonce(pub, args, int(nonce))
	return tx, nil
}

func Bond(nodeClient client.NodeClient, keyClient keys.KeyClient, pubkey, unbondAddr, amtS, nonceS string) (*txs.BondTx, error) {
	return nil, fmt.Errorf("Bond Transaction formation to be implemented on 0.12.0")
	// pub, amt, nonce, err := checkCommon(nodeAddr, signAddr, pubkey, "", amtS, nonceS)
	// if err != nil {
	// 	return nil, err
	// }
	// var pubKey crypto.PubKeyEd25519
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
	// tx.AddInputWithNonce(pub, amt, int(nonce))
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
	Address   []byte // only for new contracts
	Return    []byte
	Exception string

	//TODO: make Broadcast() errors more responsive so we
	// can differentiate mempool errors from other
}

// Preserve
func SignAndBroadcast(chainID string, nodeClient client.NodeClient, keyClient keys.KeyClient, tx txs.Tx, sign,
	broadcast, wait bool) (txResult *TxResult, err error) {
	var inputAddr []byte
	if sign {
		inputAddr, tx, err = signTx(keyClient, chainID, tx)
		if err != nil {
			return nil, err
		}
	}

	if broadcast {
		if wait {
			wsClient, err := nodeClient.DeriveWebsocketClient()
			if err != nil {
				return nil, err
			}
			var confirmationChannel chan client.Confirmation
			confirmationChannel, err = wsClient.WaitForConfirmation(tx, chainID, inputAddr)
			if err != nil {
				return nil, err
			} else {
				defer func() {
					if err != nil {
						// if broadcast threw an error, just return
						return
					}
					confirmation := <-confirmationChannel
					if confirmation.Error != nil {
						err = fmt.Errorf("Encountered error waiting for event: %s", confirmation.Error)
						return
					}
					if confirmation.Exception != nil {
						err = fmt.Errorf("Encountered Exception from chain: %s", confirmation.Exception)
						return
					}
					txResult.BlockHash = confirmation.BlockHash
					txResult.Exception = ""
					eventDataTx, ok := confirmation.Event.(*txs.EventDataTx)
					if !ok {
						err = fmt.Errorf("Received wrong event type.")
						return
					}
					txResult.Return = eventDataTx.Return
				}()
			}
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
			if len(tx_.Address) == 0 {
				txResult.Address = txs.NewContractAddress(tx_.Input.Address, tx_.Input.Sequence)
			}
		}
	}
	return
}
