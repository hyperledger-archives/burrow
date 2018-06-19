package events

import (
	"fmt"

	"reflect"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/txs"
)

var cdc = txs.NewAminoCodec()

var eventMessageTag = event.TagMap{event.MessageTypeKey: reflect.TypeOf(&Event{}).String()}

type Provider interface {
	GetEvents(startBlock, finalBlock uint64, queryable query.Queryable) (<-chan *Event, error)
}

type Event struct {
	Header *Header
	Call   *EventDataCall `json:",omitempty"`
	Log    *EventDataLog  `json:",omitempty"`
	Tx     *EventDataTx   `json:",omitempty"`
	tags   event.Tags
}

func DecodeEvent(bs []byte) (*Event, error) {
	ev := new(Event)
	err := cdc.UnmarshalBinary(bs, ev)
	if err != nil {
		return nil, err
	}
	return ev, nil
}

func (ev *Event) Encode() ([]byte, error) {
	return cdc.MarshalBinary(ev)
}

func (ev *Event) Tags() event.Tags {
	if ev.tags == nil {
		ev.tags = event.CombinedTags{
			ev.Header,
			eventMessageTag,
			ev.Tx,
			ev.Call,
			ev.Log,
		}
	}
	return ev.tags
}

func (ev *Event) Get(key string) (value string, ok bool) {
	return ev.Tags().Get(key)
}

func (ev *Event) Keys() []string {
	return ev.Tags().Keys()
}

func (ev *Event) Len() int {
	return ev.Tags().Len()
}

func (ev *Event) Map() map[string]interface{} {
	return ev.Tags().Map()
}

// event.Cache will provide an index through this methods of Indexable
func (ev *Event) ProvideIndex(index uint64) {
	ev.Header.Index = index
}

func (ev *Event) String() string {
	return fmt.Sprintf("%v", ev.Header)
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
