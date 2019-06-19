package exec

import (
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/binary"

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

func NewTxExecution(txEnv *txs.Envelope) *TxExecution {
	return &TxExecution{
		TxHeader: &TxHeader{
			TxHash: txEnv.Tx.Hash(),
			TxType: txEnv.Tx.Type(),
		},
		Envelope: txEnv,
		Receipt:  txEnv.Tx.GenerateReceipt(),
	}
}

func DecodeTxExecution(bs []byte) (*TxExecution, error) {
	txe := new(TxExecution)
	err := cdc.UnmarshalBinaryBare(bs, txe)
	if err != nil {
		return nil, err
	}
	return txe, nil
}

func DecodeTxExecutionKey(bs []byte) (*TxExecutionKey, error) {
	be := new(TxExecutionKey)

	err := cdc.UnmarshalBinaryLengthPrefixed(bs, be)
	if err != nil {
		return nil, err
	}
	return be, nil
}

func (key *TxExecutionKey) Encode() ([]byte, error) {
	// At height 0 index 0, the B cdc.MarshalBinaryBase() returns a string of 0 bytes,
	// which cannot be stored in iavl. So, abuse MarshalBinaryLengthPrefixed() to
	// ensure we have > 0 bytes.
	return cdc.MarshalBinaryLengthPrefixed(key)
}

func (txe *TxExecution) StreamEvents() StreamEvents {
	var ses StreamEvents
	ses = append(ses,
		&StreamEvent{
			BeginTx: &BeginTx{
				TxHeader:  txe.TxHeader,
				Exception: txe.Exception,
				Result:    txe.Result,
			},
		},
		&StreamEvent{
			Envelope: txe.Envelope,
		},
	)
	for _, ev := range txe.Events {
		ses = append(ses, &StreamEvent{
			Event: ev,
		})
	}
	for _, txeNested := range txe.TxExecutions {
		ses = append(ses, txeNested.StreamEvents()...)
	}
	return append(ses, &StreamEvent{
		EndTx: &EndTx{
			TxHash: txe.TxHash,
		},
	})
}

func (txe *TxExecution) Encode() ([]byte, error) {
	return cdc.MarshalBinaryBare(txe)
}

func (*TxExecution) EventType() EventType {
	return TypeTxExecution
}

func (txe *TxExecution) GetTxHash() binary.HexBytes {
	if txe == nil || txe.TxHeader == nil {
		return nil
	}
	return txe.TxHeader.TxHash
}

func (txe *TxExecution) Header(eventType EventType, eventID string, exception *errors.Exception) *Header {
	return &Header{
		TxType:    txe.GetTxType(),
		TxHash:    txe.GetTxHash(),
		Height:    txe.GetHeight(),
		EventType: eventType,
		EventID:   eventID,
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

// Errors pushed to TxExecutions end up in merkle state so it is essential that they are deterministic and independent
// of the code path taken to execution (e.g. replay takes a different path to that of normal consensus reactor so stack
// traces may differ - as they may across architectures)
func (txe *TxExecution) PushError(err error) {
	if txe.Exception == nil {
		// Don't forget the nil jig
		ex := errors.AsException(err)
		if ex != nil {
			txe.Exception = ex
		}
	}
}

func (txe *TxExecution) CallTrace() string {
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
			ev.Header.Height = txe.GetHeight()
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
	var tagged query.Tagged = query.TagMap{}
	if txe != nil {
		tagged = query.MergeTags(
			query.TagMap{
				event.EventIDKey:   EventStringTxExecution(txe.TxHash),
				event.EventTypeKey: txe.EventType()},
			query.MustReflectTags(txe),
			query.MustReflectTags(txe.TxHeader),
			txe.Envelope.Tagged(),
		)
	}
	return &TaggedTxExecution{
		Tagged:      tagged,
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
	return event.QueryForEventID(EventStringTxExecution(txHash))
}
