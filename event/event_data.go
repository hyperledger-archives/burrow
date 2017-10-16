package event

import (
	"fmt"

	evm_events "github.com/hyperledger/burrow/execution/evm/events"
	tm_types "github.com/tendermint/tendermint/types"
)

// Oh for a real sum type

// AnyEventData provides a single type for our multiplexed event categories of EVM events and Tendermint events
type AnyEventData struct {
	TMEventData  *tm_types.TMEventData `json:"tm_event_data,omitempty"`
	EVMEventData evm_events.EventData  `json:"evm_event_data,omitempty"`
	Error        error                 `json:"error,omitempty"`
}

// Get whichever element of the AnyEventData sum type that is not nil
func (aed AnyEventData) Get() interface{} {
	if aed.TMEventData != nil {
		return aed.TMEventData
	}
	if aed.EVMEventData != nil {
		return aed.EVMEventData
	}
	return aed.Error
}

// If this AnyEventData wraps an EventDataNewBlock then return a pointer to that value, else return nil
func (aed AnyEventData) EventDataNewBlock() *tm_types.EventDataNewBlock {
	if aed.TMEventData != nil {
		if eventData, ok := aed.TMEventData.Unwrap().(tm_types.EventDataNewBlock); ok {
			return &eventData
		}
	}
	return nil
}

// If this AnyEventData wraps an EventDataCall then return a pointer to that value, else return nil
func (aed AnyEventData) EventDataCall() *evm_events.EventDataCall {
	if aed.EVMEventData != nil {
		if eventData, ok := aed.EVMEventData.(evm_events.EventDataCall); ok {
			return &eventData
		}
	}
	return nil
}

// If this AnyEventData wraps an EventDataTx then return a pointer to that value, else return nil
func (aed AnyEventData) EventDataTx() *evm_events.EventDataTx {
	if aed.EVMEventData != nil {
		if eventData, ok := aed.EVMEventData.(evm_events.EventDataTx); ok {
			return &eventData
		}
	}
	return nil
}

// Map any supported event data element to our AnyEventData sum type
func MapToAnyEventData(eventData interface{}) AnyEventData {
	switch ed := eventData.(type) {
	case tm_types.TMEventData:
		return AnyEventData{TMEventData: &ed}
	case evm_events.EventData:
		return AnyEventData{EVMEventData: ed}
	default:
		return AnyEventData{Error: fmt.Errorf("could not map event data of type %T to AnyEventData", eventData)}
	}
}
