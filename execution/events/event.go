package events

import "github.com/hyperledger/burrow/binary"

type Event struct {
	Header *Header
	Call   *EventDataCall `json:",omitempty"`
	Log    *EventDataLog  `json:",omitempty"`
	Tx     *EventDataTx   `json:",omitempty"`
}

func (ev *Event) GetTx() *EventDataTx {
	if ev == nil {
		return nil
	}
	return ev.Tx
}

func (ev *Event) GetCall() *EventDataCall {
	if ev == nil {
		return nil
	}
	return ev.Call
}

func (ev *Event) GetLog() *EventDataLog {
	if ev == nil {
		return nil
	}
	return ev.Log
}

type Header struct {
	TxHash binary.HexBytes
	Height uint64
	Index  uint64
}
