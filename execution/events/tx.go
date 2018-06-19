package events

import (
	"context"
	"fmt"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
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

// All txs fire EventDataTx, but only CallTx might have Return or Exception
type EventDataTx struct {
	Tx        *txs.Tx
	Return    binary.HexBytes
	Exception *errors.Exception
}

var txTagKeys = []string{event.ExceptionKey}

func (tx *EventDataTx) Get(key string) (string, bool) {
	var value interface{}
	switch key {
	case event.ExceptionKey:
		value = tx.Exception
	default:
		return "", false
	}
	return query.StringFromValue(value), true
}

func (tx *EventDataTx) Len() int {
	return len(txTagKeys)
}

func (tx *EventDataTx) Map() map[string]interface{} {
	tags := make(map[string]interface{})
	for _, key := range txTagKeys {
		tags[key], _ = tx.Get(key)
	}
	return tags
}

func (tx *EventDataTx) Keys() []string {
	return txTagKeys
}

// For re-use
var sendTxQuery = query.NewBuilder().
	AndEquals(event.TxTypeKey, payload.TypeSend.String())

var callTxQuery = query.NewBuilder().
	AndEquals(event.TxTypeKey, payload.TypeCall.String())

// Publish/Subscribe
func PublishAccountInput(publisher event.Publisher, height uint64, address crypto.Address, tx *txs.Tx, ret []byte,
	exception *errors.Exception) error {

	ev := txEvent(height, TypeAccountInput, EventStringAccountInput(address), tx, ret, exception)
	return publisher.Publish(context.Background(), ev, event.CombinedTags{ev.Tags(), event.TagMap{
		event.AddressKey: address,
	}})
}

func PublishAccountOutput(publisher event.Publisher, height uint64, address crypto.Address, tx *txs.Tx, ret []byte,
	exception *errors.Exception) error {

	ev := txEvent(height, TypeAccountOutput, EventStringAccountOutput(address), tx, ret, exception)
	return publisher.Publish(context.Background(), ev, event.CombinedTags{ev.Tags(), event.TagMap{
		event.AddressKey: address,
	}})
}

func PublishNameReg(publisher event.Publisher, height uint64, tx *txs.Tx) error {
	nameTx, ok := tx.Payload.(*payload.NameTx)
	if !ok {
		return fmt.Errorf("Tx payload must be NameTx to PublishNameReg")
	}
	ev := txEvent(height, TypeAccountInput, EventStringNameReg(nameTx.Name), tx, nil, nil)
	return publisher.Publish(context.Background(), ev, event.CombinedTags{ev.Tags(), event.TagMap{
		event.NameKey: nameTx.Name,
	}})
}

func PublishPermissions(publisher event.Publisher, height uint64, tx *txs.Tx) error {
	permTx, ok := tx.Payload.(*payload.PermissionsTx)
	if !ok {
		return fmt.Errorf("Tx payload must be PermissionsTx to PublishPermissions")
	}
	ev := txEvent(height, TypeAccountInput, EventStringPermissions(permTx.PermArgs.PermFlag), tx, nil, nil)
	return publisher.Publish(context.Background(), ev, event.CombinedTags{ev.Tags(), event.TagMap{
		event.PermissionKey: permTx.PermArgs.PermFlag.String(),
	}})
}

func SubscribeAccountOutputSendTx(ctx context.Context, subscribable event.Subscribable, subscriber string,
	address crypto.Address, txHash []byte, ch chan<- *payload.SendTx) error {

	query := sendTxQuery.And(event.QueryForEventID(EventStringAccountOutput(address))).
		AndEquals(event.TxHashKey, hex.EncodeUpperToString(txHash))

	return event.SubscribeCallback(ctx, subscribable, subscriber, query, func(message interface{}) bool {
		if ev, ok := message.(*Event); ok && ev.Tx != nil {
			if sendTx, ok := ev.Tx.Tx.Payload.(*payload.SendTx); ok {
				ch <- sendTx
			}
		}
		return true
	})
}

func txEvent(height uint64, eventType Type, eventID string, tx *txs.Tx, ret []byte, exception *errors.Exception) *Event {
	return &Event{
		Header: &Header{
			TxType:    tx.Type(),
			TxHash:    tx.Hash(),
			EventType: eventType,
			EventID:   eventID,
			Height:    height,
		},
		Tx: &EventDataTx{
			Tx:        tx,
			Return:    ret,
			Exception: exception,
		},
	}
}
