package event

import (
	"fmt"

	exe_events "github.com/hyperledger/burrow/execution/events"
	evm_events "github.com/hyperledger/burrow/execution/evm/events"
	"github.com/tendermint/go-wire/data"
	tm_types "github.com/tendermint/tendermint/types"
)

// Oh for a real sum type

// AnyEventData provides a single type for our multiplexed event categories of EVM events and Tendermint events
type AnyEventData struct {
	TMEventData     *tm_types.TMEventData `json:"tm_event_data,omitempty"`
	BurrowEventData *EventData            `json:"burrow_event_data,omitempty"`
	Err             *string               `json:"error,omitempty"`
}

type EventData struct {
	EventDataInner `json:"unwrap"`
}

type EventDataInner interface {
}

func (ed EventData) Unwrap() EventDataInner {
	return ed.EventDataInner
}

func (ed EventData) MarshalJSON() ([]byte, error) {
	return mapper.ToJSON(ed.EventDataInner)
}

func (ed *EventData) UnmarshalJSON(data []byte) (err error) {
	parsed, err := mapper.FromJSON(data)
	if err == nil && parsed != nil {
		ed.EventDataInner = parsed.(EventDataInner)
	}
	return err
}

var mapper = data.NewMapper(EventData{}).
	RegisterImplementation(exe_events.EventDataTx{}, "event_data_tx", biota()).
	RegisterImplementation(evm_events.EventDataCall{}, "event_data_call", biota()).
	RegisterImplementation(evm_events.EventDataLog{}, "event_data_log", biota())

// Get whichever element of the AnyEventData sum type that is not nil
func (aed AnyEventData) Get() interface{} {
	if aed.TMEventData != nil {
		return aed.TMEventData.Unwrap()
	}
	if aed.BurrowEventData != nil {
		return aed.BurrowEventData.Unwrap()
	}
	if aed.Err != nil {
		return *aed.Err
	}
	return nil
}

// If this AnyEventData wraps an EventDataNewBlock then return a pointer to that value, else return nil
func (aed AnyEventData) EventDataNewBlock() *tm_types.EventDataNewBlock {
	if aed.TMEventData != nil {
		eventData, _ := aed.TMEventData.Unwrap().(tm_types.EventDataNewBlock)
		return &eventData
	}
	return nil
}

// If this AnyEventData wraps an EventDataLog then return a pointer to that value, else return nil
func (aed AnyEventData) EventDataLog() *evm_events.EventDataLog {
	if aed.BurrowEventData != nil {
		eventData, _ := aed.BurrowEventData.Unwrap().(evm_events.EventDataLog)
		return &eventData
	}
	return nil
}

// If this AnyEventData wraps an EventDataCall then return a pointer to that value, else return nil
func (aed AnyEventData) EventDataCall() *evm_events.EventDataCall {
	if aed.BurrowEventData != nil {
		eventData, _ := aed.BurrowEventData.Unwrap().(evm_events.EventDataCall)
		return &eventData
	}
	return nil
}

// If this AnyEventData wraps an EventDataTx then return a pointer to that value, else return nil
func (aed AnyEventData) EventDataTx() *exe_events.EventDataTx {
	if aed.BurrowEventData != nil {
		eventData, _ := aed.BurrowEventData.Unwrap().(exe_events.EventDataTx)
		return &eventData
	}
	return nil
}

func (aed AnyEventData) Error() string {
	if aed.Err == nil {
		return ""
	}
	return *aed.Err
}

// Map any supported event data element to our AnyEventData sum type
func MapToAnyEventData(eventData interface{}) AnyEventData {
	switch ed := eventData.(type) {
	case AnyEventData:
		return ed

	case tm_types.TMEventData:
		return AnyEventData{TMEventData: &ed}

	case EventData:
		return AnyEventData{BurrowEventData: &ed}

	case EventDataInner:
		return AnyEventData{BurrowEventData: &EventData{
			EventDataInner: ed,
		}}

	default:
		errStr := fmt.Sprintf("could not map event data of type %T to AnyEventData", eventData)
		return AnyEventData{Err: &errStr}
	}
}

// Type byte helper
var nextByte byte = 1

func biota() (b byte) {
	b = nextByte
	nextByte++
	return
}
