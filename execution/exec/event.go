package exec

import (
	"fmt"
	"reflect"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
)

var eventMessageType = reflect.TypeOf(&Event{}).String()

type EventType uint32

// Execution event types
const (
	TypeUnknown EventType = iota
	TypeCall
	TypeLog
	TypeAccountInput
	TypeAccountOutput
	TypeTxExecution
	TypeBlockExecution
	TypeGovernAccount
	TypeBeginBlock
	TypeBeginTx
	TypeEnvelope
	TypeEndTx
	TypeEndBlock
	TypePrint
)

var nameFromType = map[EventType]string{
	TypeUnknown:        "UnknownEvent",
	TypeCall:           "CallEvent",
	TypeLog:            "LogEvent",
	TypeAccountInput:   "AccountInputEvent",
	TypeAccountOutput:  "AccountOutputEvent",
	TypeTxExecution:    "TxExecutionEvent",
	TypeBlockExecution: "BlockExecutionEvent",
	TypeGovernAccount:  "GovernAccountEvent",
	TypeBeginBlock:     "BeginBlockEvent",
	TypeEndBlock:       "EndBlockEvent",
}

var typeFromName = make(map[string]EventType)

func init() {
	for t, n := range nameFromType {
		typeFromName[n] = t
	}
}

func EventTypeFromString(name string) EventType {
	return typeFromName[name]
}

func (ev *Event) EventType() EventType {
	return ev.Header.EventType
}

func (typ EventType) String() string {
	name, ok := nameFromType[typ]
	if ok {
		return name
	}
	return "UnknownEventType"
}

func (typ EventType) MarshalText() ([]byte, error) {
	return []byte(typ.String()), nil
}

func (typ *EventType) UnmarshalText(data []byte) error {
	*typ = EventTypeFromString(string(data))
	return nil
}

// Event

func (ev *Event) String() string {
	return fmt.Sprintf("ExecutionEvent{%v: %s}", ev.Header.String(), ev.Body())
}

func (ev *Event) Body() string {
	if ev.Input != nil {
		return ev.Input.String()
	}
	if ev.Output != nil {
		return ev.Output.String()
	}
	if ev.Log != nil {
		return ev.Log.String()
	}
	if ev.Call != nil {
		return ev.Call.String()
	}
	return "<empty>"
}

func (ev *Event) Get(key string) (value interface{}, ok bool) {
	switch key {
	case event.MessageTypeKey:
		return eventMessageType, true
	}
	if ev == nil {
		return nil, false
	}
	v, ok := ev.Log.Get(key)
	if ok {
		return v, true
	}
	v, ok = query.GetReflect(reflect.ValueOf(ev.Header), key)
	if ok {
		return v, true
	}
	return query.GetReflect(reflect.ValueOf(ev), key)
}
