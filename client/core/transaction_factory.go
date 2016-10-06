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

import (
	"encoding/hex"
	"fmt"
	"strconv"

	log "github.com/eris-ltd/eris-logger"

	ptypes "github.com/eris-ltd/eris-db/permission/types"

	"github.com/eris-ltd/eris-db/account"
	"github.com/eris-ltd/eris-db/client"
	"github.com/eris-ltd/eris-db/keys"
	"github.com/eris-ltd/eris-db/txs"
)

var (
	MaxCommitWaitTimeSeconds = 20
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

func Permissions(nodeClient client.NodeClient, keyClient keys.KeyClient, pubkey, addrS, nonceS, permFunc string, argsS []string) (*txs.PermissionsTx, error) {
	pub, _, nonce, err := checkCommon(nodeClient, keyClient, pubkey, addrS, "0", nonceS)
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
func SignAndBroadcast(chainID string, nodeClient client.NodeClient, keyClient keys.KeyClient, tx txs.Tx, sign, broadcast, wait bool) (txResult *TxResult, err error) {
	var inputAddr []byte
	if sign {
		inputAddr, tx, err = signTx(keyClient, chainID, tx)
		if err != nil {
			return nil, err
		}
		log.WithFields(log.Fields{
			"transaction": string(account.SignBytes(chainID, tx)),
		}).Debug("Signed transaction")
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
					log.Debug("Waiting for transaction to be confirmed.")
					confirmation := <-confirmationChannel
					if confirmation.Error != nil {
						log.Errorf("Encountered error waiting for event: %s\n", confirmation.Error)
						err = confirmation.Error
						return
					}
					if confirmation.Exception != nil {
						log.Errorf("Encountered Exception from chain w: %s\n", confirmation.Error)
						err = confirmation.Exception
						return
					}
					txResult.BlockHash = confirmation.BlockHash
					txResult.Exception = ""
					eventDataTx, ok := confirmation.Event.(*txs.EventDataTx)
					if !ok {
						log.Errorf("Received wrong event type.")
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
