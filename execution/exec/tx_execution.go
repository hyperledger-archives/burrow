package exec

import (
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs"
)

func EventStringAccountInput(addr crypto.Address) string  { return fmt.Sprintf("Acc/%s/Input", addr) }
func EventStringAccountOutput(addr crypto.Address) string { return fmt.Sprintf("Acc/%s/Output", addr) }

func EventStringAccountCall(addr crypto.Address) string    { return fmt.Sprintf("Acc/%s/Call", addr) }
func EventStringLogEvent(addr crypto.Address) string       { return fmt.Sprintf("Log/%s", addr) }
func EventStringTxExecution(txHash []byte) string          { return fmt.Sprintf("Execution/Tx/%X", txHash) }
func EventStringGovernAccount(addr *crypto.Address) string { return fmt.Sprintf("Govern/Acc/%v", addr) }
func EventStringPayload(index uint32) string               { return fmt.Sprintf("Payload/%d", index) }

func NewTxExecution(txEnv *txs.Envelope) *TxExecution {
	return &TxExecution{
		TxHash:   txEnv.Tx.Hash(),
		TxType:   txEnv.Tx.Type(),
		Envelope: txEnv,
		Receipt:  txEnv.Tx.GenerateReceipt(),
	}
}

func DecodeTxExecution(bs []byte) (*TxExecution, error) {
	txe := new(TxExecution)
	err := cdc.UnmarshalBinary(bs, txe)
	if err != nil {
		return nil, err
	}
	return txe, nil
}

func (txe *TxExecution) Encode() ([]byte, error) {
	return cdc.MarshalBinary(txe)
}

func (*TxExecution) EventType() EventType {
	return TypeTxExecution
}

func (txe *TxExecution) Header(eventType EventType, eventID string, exception *errors.Exception) *Header {
	return &Header{
		TxType:    txe.TxType,
		TxHash:    txe.TxHash,
		EventType: eventType,
		EventID:   eventID,
		Height:    txe.Height,
		Exception: exception,
	}
}

// Emit events
func (txe *TxExecution) Input(address crypto.Address, exception *errors.Exception) {
	txe.Append(&Event{
		Header: txe.Header(TypeAccountInput, EventStringAccountInput(address), exception),
		Input: &InputEvent{
			Address: address,
		},
	})
}

func (txe *TxExecution) Output(address crypto.Address, exception *errors.Exception) {
	txe.Append(&Event{
		Header: txe.Header(TypeAccountOutput, EventStringAccountOutput(address), exception),
		Output: &OutputEvent{
			Address: address,
		},
	})
}

func (txe *TxExecution) Log(log *LogEvent) error {
	txe.Append(&Event{
		Header: txe.Header(TypeLog, EventStringLogEvent(log.Address), nil),
		Log:    log,
	})
	return nil
}

func (txe *TxExecution) Call(call *CallEvent, exception *errors.Exception) error {
	txe.Append(&Event{
		Header: txe.Header(TypeCall, EventStringAccountCall(call.CallData.Callee), exception),
		Call:   call,
	})
	return nil
}

func (txe *TxExecution) GovernAccount(governAccount *GovernAccountEvent, exception *errors.Exception) {
	txe.Append(&Event{
		Header:        txe.Header(TypeGovernAccount, EventStringGovernAccount(governAccount.AccountUpdate.Address), exception),
		GovernAccount: governAccount,
	})
}

func (txe *TxExecution) PushError(err error) {
	if txe.Exception == nil {
		// Don't forget the nil jig
		ex := errors.AsException(err)
		if ex != nil {
			txe.Exception = ex
		}
	}
}

func (txe *TxExecution) PayloadEvent(payload *PayloadEvent) {
	txe.Append(&Event{
		Header:  txe.Header(TypePayload, EventStringPayload(payload.Index), nil),
		Payload: payload,
	})
}

func (txe *TxExecution) Trace() string {
	var calls []string
	for _, ev := range txe.Events {
		if ev.Call != nil {
			ex := ""
			if ev.Header.Exception != nil {
				ex = fmt.Sprintf(" [%v]", ev.Header.Exception)
			}
			calls = append(calls, fmt.Sprintf("%v: %v -> %v: %v%s",
				ev.Call.CallType, ev.Call.CallData.Caller, ev.Call.CallData.Callee, ev.Call.Return, ex))
		}
	}
	return strings.Join(calls, "\n")
}

func (txe *TxExecution) ExceptionalCalls() []*Event {
	var exCalls []*Event
	for _, ev := range txe.Events {
		if ev.Call != nil && ev.Header.Exception != nil {
			exCalls = append(exCalls, ev)
		}
	}
	return exCalls
}

func (txe *TxExecution) CallError() *errors.CallError {
	if txe.Exception == nil {
		return nil
	}
	var nestedErrors []errors.NestedCallError
	for _, ev := range txe.Events {
		if ev.Call != nil && ev.Header.Exception != nil {
			nestedErrors = append(nestedErrors, errors.NestedCallError{
				CodedError: ev.Header.Exception,
				Caller:     ev.Call.CallData.Caller,
				Callee:     ev.Call.CallData.Callee,
				StackDepth: ev.Call.StackDepth,
			})
		}
	}
	return &errors.CallError{
		CodedError:   txe.Exception,
		NestedErrors: nestedErrors,
	}
}

// Set result
func (txe *TxExecution) Return(returnValue []byte, gasUsed uint64) {
	if txe.Result == nil {
		txe.Result = &Result{}
	}
	txe.Result.Return = returnValue
	txe.Result.GasUsed = gasUsed
}

func (txe *TxExecution) Name(entry *names.Entry) {
	if txe.Result == nil {
		txe.Result = &Result{}
	}
	txe.Result.NameEntry = entry
}

func (txe *TxExecution) Permission(permArgs *permission.PermArgs) {
	if txe.Result == nil {
		txe.Result = &Result{}
	}
	txe.Result.PermArgs = permArgs
}

func (txe *TxExecution) Append(tail ...*Event) {
	for i, ev := range tail {
		if ev != nil && ev.Header != nil {
			ev.Header.Index = uint64(len(txe.Events) + i)
			ev.Header.Height = txe.Height
		}
	}
	txe.Events = append(txe.Events, tail...)
}

// Tags
type TaggedTxExecution struct {
	query.Tagged
	*TxExecution
}

func (txe *TxExecution) Tagged() *TaggedTxExecution {
	return &TaggedTxExecution{
		Tagged: query.MergeTags(
			query.TagMap{
				event.EventIDKey:   EventStringTxExecution(txe.TxHash),
				event.EventTypeKey: txe.EventType()},
			query.MustReflectTags(txe),
			txe.Envelope.Tagged(),
		),
		TxExecution: txe,
	}
}

func (txe *TxExecution) TaggedEvents() TaggedEvents {
	tevs := make(TaggedEvents, len(txe.Events))
	for i, ev := range txe.Events {
		tevs[i] = ev.Tagged()
	}
	return tevs
}

func QueryForTxExecution(txHash []byte) query.Queryable {
	return query.NewBuilder().AndEquals(event.EventIDKey, EventStringTxExecution(txHash))
}
