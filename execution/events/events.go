package events

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/errors"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/tmthrgd/go-hex"
)

func EventStringAccountInput(addr crypto.Address) string  { return fmt.Sprintf("Acc/%s/Input", addr) }
func EventStringAccountOutput(addr crypto.Address) string { return fmt.Sprintf("Acc/%s/Output", addr) }
func EventStringNameReg(name string) string               { return fmt.Sprintf("NameReg/%s", name) }
func EventStringPermissions(perm ptypes.PermFlag) string  { return fmt.Sprintf("Permissions/%v", perm) }
func EventStringBond() string                             { return "Bond" }
func EventStringUnbond() string                           { return "Unbond" }
func EventStringRebond() string                           { return "Rebond" }

// All txs fire EventDataTx, but only CallTx might have Return or Exception
type EventDataTx struct {
	Tx        *txs.Tx
	Return    []byte
	Exception *errors.Exception
}

// For re-use
var sendTxQuery = event.NewQueryBuilder().
	AndEquals(event.MessageTypeKey, reflect.TypeOf(&EventDataTx{}).String()).
	AndEquals(event.TxTypeKey, payload.TypeSend.String())

var callTxQuery = event.NewQueryBuilder().
	AndEquals(event.MessageTypeKey, reflect.TypeOf(&EventDataTx{}).String()).
	AndEquals(event.TxTypeKey, payload.TypeCall.String())

// Publish/Subscribe
func SubscribeAccountOutputSendTx(ctx context.Context, subscribable event.Subscribable, subscriber string,
	address crypto.Address, txHash []byte, ch chan<- *payload.SendTx) error {

	query := sendTxQuery.And(event.QueryForEventID(EventStringAccountOutput(address))).
		AndEquals(event.TxHashKey, hex.EncodeUpperToString(txHash))

	return event.SubscribeCallback(ctx, subscribable, subscriber, query, func(message interface{}) bool {
		if edt, ok := message.(*EventDataTx); ok {
			if sendTx, ok := edt.Tx.Payload.(*payload.SendTx); ok {
				ch <- sendTx
			}
		}
		return true
	})
}

func PublishAccountOutput(publisher event.Publisher, address crypto.Address, tx *txs.Tx, ret []byte,
	exception *errors.Exception) error {

	return event.PublishWithEventID(publisher, EventStringAccountOutput(address),
		&EventDataTx{
			Tx:        tx,
			Return:    ret,
			Exception: exception,
		},
		map[string]interface{}{
			"address":       address,
			event.TxTypeKey: tx.Type().String(),
			event.TxHashKey: hex.EncodeUpperToString(tx.Hash()),
		})
}

func PublishAccountInput(publisher event.Publisher, address crypto.Address, tx *txs.Tx, ret []byte,
	exception *errors.Exception) error {

	return event.PublishWithEventID(publisher, EventStringAccountInput(address),
		&EventDataTx{
			Tx:        tx,
			Return:    ret,
			Exception: exception,
		},
		map[string]interface{}{
			"address":       address,
			event.TxTypeKey: tx.Type().String(),
			event.TxHashKey: hex.EncodeUpperToString(tx.Hash()),
		})
}

func PublishNameReg(publisher event.Publisher, tx *txs.Tx) error {
	nameTx, ok := tx.Payload.(*payload.NameTx)
	if !ok {
		return fmt.Errorf("Tx payload must be NameTx to PublishNameReg")
	}
	return event.PublishWithEventID(publisher, EventStringNameReg(nameTx.Name), &EventDataTx{Tx: tx},
		map[string]interface{}{
			"name":          nameTx.Name,
			event.TxTypeKey: tx.Type().String(),
			event.TxHashKey: hex.EncodeUpperToString(tx.Hash()),
		})
}

func PublishPermissions(publisher event.Publisher, perm ptypes.PermFlag, tx *txs.Tx) error {
	_, ok := tx.Payload.(*payload.PermissionsTx)
	if !ok {
		return fmt.Errorf("Tx payload must be PermissionsTx to PublishPermissions")
	}
	return event.PublishWithEventID(publisher, EventStringPermissions(perm), &EventDataTx{Tx: tx},
		map[string]interface{}{
			"name":          perm.String(),
			event.TxTypeKey: tx.Type().String(),
			event.TxHashKey: hex.EncodeUpperToString(tx.Hash()),
		})
}
