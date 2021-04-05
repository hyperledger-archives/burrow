package exec

import (
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/errors"
)

type Events []*Event

func (evs *Events) Append(tail ...*Event) {
	for i, ev := range tail {
		if ev != nil && ev.Header != nil {
			ev.Header.Index = uint64(len(*evs) + i)
		}
	}
	*evs = append(*evs, tail...)
}

func (evs *Events) Call(call *CallEvent, exception *errors.Exception) error {
	evs.Append(&Event{
		Header: &Header{
			EventType: TypeCall,
			EventID:   EventStringAccountCall(call.CallData.Callee),
			Exception: exception,
		},
		Call: call,
	})
	return nil
}

func (evs *Events) Log(log *LogEvent) error {
	evs.Append(&Event{
		Header: &Header{
			EventType: TypeLog,
			EventID:   EventStringLogEvent(log.Address),
		},
		Log: log,
	})
	return nil
}

func (evs *Events) Print(print *PrintEvent) error {
	evs.Append(&Event{
		Header: &Header{
			EventType: TypePrint,
			EventID:   EventStringLogEvent(print.Address),
		},
		Print: print,
	})
	return nil
}

func (evs Events) CallTrace() string {
	var calls []string
	for _, ev := range evs {
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

func (evs Events) ExceptionalCalls() []*Event {
	var exCalls []*Event
	for _, ev := range evs {
		if ev.Call != nil && ev.Header.Exception != nil {
			exCalls = append(exCalls, ev)
		}
	}
	return exCalls
}

func (evs Events) NestedCallErrors() []errors.NestedCallError {
	var nestedErrors []errors.NestedCallError
	for _, ev := range evs {
		if ev.Call != nil && ev.Header.Exception != nil {
			nestedErrors = append(nestedErrors, errors.NestedCallError{
				CodedError: ev.Header.Exception,
				Caller:     ev.Call.CallData.Caller,
				Callee:     ev.Call.CallData.Callee,
				StackDepth: ev.Call.StackDepth,
			})
		}
	}
	return nestedErrors
}

func (evs Events) Filter(qry query.Query) Events {
	var filtered Events
	for _, tev := range evs {
		if qry.Matches(tev) {
			filtered = append(filtered, tev)
		}
	}
	return filtered
}
