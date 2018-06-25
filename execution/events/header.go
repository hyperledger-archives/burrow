package events

import (
	"fmt"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/txs/payload"
)

type Header struct {
	TxType    payload.Type
	TxHash    binary.HexBytes
	EventType Type
	EventID   string
	Height    uint64
	Index     uint64
}

var headerTagKeys = []string{
	event.TxTypeKey,
	event.TxHashKey,
	event.EventTypeKey,
	event.EventIDKey,
	event.HeightKey,
	event.IndexKey,
}

// Implements Tags for events
func (h *Header) Get(key string) (value string, ok bool) {
	var v interface{}
	switch key {
	case event.TxTypeKey:
		v = h.TxType
	case event.TxHashKey:
		v = h.TxHash
	case event.EventTypeKey:
		v = h.EventType
	case event.EventIDKey:
		v = h.EventID
	case event.HeightKey:
		v = h.Height
	case event.IndexKey:
		v = h.Index
	default:
		return "", false
	}
	return query.StringFromValue(v), true
}

func (h *Header) Len() int {
	return len(headerTagKeys)
}

func (h *Header) Map() map[string]interface{} {
	tags := make(map[string]interface{})
	for _, key := range headerTagKeys {
		tags[key], _ = h.Get(key)
	}
	return tags
}

func (h *Header) Keys() []string {
	return headerTagKeys
}

// Returns a lexicographically sortable key encoding the height and index of this event
func (h *Header) Key() Key {
	return NewKey(h.Height, h.Index)
}

func (h *Header) String() string {
	return fmt.Sprintf("Header{Tx{%v}: %v; Event{%v}: %v; Height: %v; Index: %v}",
		h.TxType, h.TxHash, h.EventType, h.EventID, h.Height, h.Index)
}
