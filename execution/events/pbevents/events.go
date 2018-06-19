package pbevents

import (
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/txs/payload"
)

// this mostly contains tedious mapping between protobuf and our domain objects, but it may be worth
// the pain to avoid complexity and edge cases using gogoproto or other techniques.

func GetEventDataCall(edt *events.EventDataCall) *EventDataCall {
	return &EventDataCall{
		Origin:     edt.Origin.Bytes(),
		CallData:   GetCallData(edt.CallData),
		StackDepth: edt.StackDepth,
		Return:     edt.Return,
		Exception:  edt.Exception,
	}
}

func GetCallData(cd *events.CallData) *CallData {
	return &CallData{
		Caller: cd.Caller.Bytes(),
		Callee: cd.Callee.Bytes(),
		Data:   cd.Data,
		Gas:    cd.Gas,
	}
}

func GetEvent(event *events.Event) *ExecutionEvent {
	return &ExecutionEvent{
		Header:    GetEventHeader(event.Header),
		EventData: GetEventData(event),
	}
}

func GetEventHeader(header *events.Header) *EventHeader {
	return &EventHeader{
		TxType:    header.TxType.String(),
		TxHash:    header.TxHash,
		EventType: header.EventType.String(),
		EventID:   header.EventID,
		Height:    header.Height,
		Index:     header.Index,
	}
}

func GetEventData(ev *events.Event) isExecutionEvent_EventData {
	if ev.Call != nil {
		return &ExecutionEvent_EventDataCall{
			EventDataCall: &EventDataCall{
				CallData:   GetCallData(ev.Call.CallData),
				Origin:     ev.Call.Origin.Bytes(),
				StackDepth: ev.Call.StackDepth,
				Return:     ev.Call.Return,
				Exception:  ev.Call.Exception,
			},
		}
	}

	if ev.Log != nil {
		return &ExecutionEvent_EventDataLog{
			EventDataLog: &EventDataLog{
				Address: ev.Log.Address.Bytes(),
				Data:    ev.Log.Data,
				Topics:  GetTopic(ev.Log.Topics),
			},
		}
	}

	if ev.Tx != nil {
		return &ExecutionEvent_EventDataTx{
			EventDataTx: &EventDataTx{
				Return:    ev.Tx.Return,
				Exception: ev.Tx.Exception,
			},
		}
	}

	return nil
}

func GetTopic(topics []binary.Word256) [][]byte {
	topicBytes := make([][]byte, len(topics))
	for i, t := range topics {
		topicBytes[i] = t.Bytes()
	}
	return topicBytes
}

func (ee *ExecutionEvent) Event() *events.Event {
	return &events.Event{
		Header: ee.GetHeader().Header(),
		Tx:     ee.GetEventDataTx().Tx(),
		Log:    ee.GetEventDataLog().Log(ee.Header.Height),
		Call:   ee.GetEventDataCall().Call(ee.Header.TxHash),
	}
}

func (h *EventHeader) Header() *events.Header {
	return &events.Header{
		TxType:    payload.TxTypeFromString(h.TxType),
		TxHash:    h.TxHash,
		EventType: events.EventTypeFromString(h.EventType),
		EventID:   h.EventID,
		Height:    h.Height,
		Index:     h.Index,
	}
}

func (tx *EventDataTx) Tx() *events.EventDataTx {
	if tx == nil {
		return nil
	}
	return &events.EventDataTx{
		Return:    tx.Return,
		Exception: tx.Exception,
	}
}

func (log *EventDataLog) Log(height uint64) *events.EventDataLog {
	if log == nil {
		return nil
	}
	topicWords := make([]binary.Word256, len(log.Topics))
	for i, bs := range log.Topics {
		topicWords[i] = binary.LeftPadWord256(bs)
	}
	return &events.EventDataLog{
		Height:  height,
		Topics:  topicWords,
		Address: crypto.MustAddressFromBytes(log.Address),
		Data:    log.Data,
	}
}

func (call *EventDataCall) Call(txHash []byte) *events.EventDataCall {
	if call == nil {
		return nil
	}
	return &events.EventDataCall{
		Return:     call.Return,
		CallData:   call.CallData.CallData(),
		Origin:     crypto.MustAddressFromBytes(call.Origin),
		StackDepth: call.StackDepth,
		Exception:  call.Exception,
	}
}

func (cd *CallData) CallData() *events.CallData {
	return &events.CallData{
		Caller: crypto.MustAddressFromBytes(cd.Caller),
		Callee: crypto.MustAddressFromBytes(cd.Callee),
		Value:  cd.Value,
		Gas:    cd.Gas,
		Data:   cd.Data,
	}
}
