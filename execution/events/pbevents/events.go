package pbevents

import (
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/events"
)

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

func GetEvent(event interface{}) *ExecutionEvent {
	switch ev := event.(type) {
	case *events.Event:
		return &ExecutionEvent{
			Header:    GetEventHeader(ev.Header),
			EventData: GetEventData(ev),
		}
	default:
		return nil
	}
}

func GetEventHeader(header *events.Header) *EventHeader {
	return &EventHeader{
		TxHash: header.TxHash,
		Height: header.Height,
		Index:  header.Index,
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
