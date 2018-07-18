package exec

import (
	"reflect"

	"fmt"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
)

var eventMessageTag = query.TagMap{event.MessageTypeKey: reflect.TypeOf(&Event{}).String()}

type EventType uint32

// Execution event types
const (
	TypeCall           = EventType(0x00)
	TypeLog            = EventType(0x01)
	TypeAccountInput   = EventType(0x02)
	TypeAccountOutput  = EventType(0x03)
	TypeTxExecution    = EventType(0x04)
	TypeBlockExecution = EventType(0x05)
	TypeGovernAccount  = EventType(0x06)
)

var nameFromType = map[EventType]string{
	TypeCall:           "CallEvent",
	TypeLog:            "LogEvent",
	TypeAccountInput:   "AccountInputEvent",
	TypeAccountOutput:  "AccountOutputEvent",
	TypeTxExecution:    "TxExecutionEvent",
	TypeBlockExecution: "BlockExecutionEvent",
	TypeGovernAccount:  "GovernAccountEvent",
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
	return "UnknownTx"
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

// Tags
type TaggedEvent struct {
	query.Tagged
	*Event
}

type TaggedEvents []*TaggedEvent

func (ev *Event) Tagged() *TaggedEvent {
	return &TaggedEvent{
		Tagged: query.MergeTags(
			query.MustReflectTags(ev.Header),
			eventMessageTag,
			query.MustReflectTags(ev.Input),
			query.MustReflectTags(ev.Output),
			query.MustReflectTags(ev.Call),
			ev.Log,
		),
		Event: ev,
	}
}

func (tevs TaggedEvents) Filter(qry query.Query) TaggedEvents {
	var filtered TaggedEvents
	for _, tev := range tevs {
		if qry.Matches(tev) {
			filtered = append(filtered, tev)
		}
	}
	return filtered
}
