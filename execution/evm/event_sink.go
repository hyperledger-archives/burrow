package evm

import (
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
)

type EventSink interface {
	Call(call *exec.CallEvent, exception *errors.Exception) error
	Log(log *exec.LogEvent) error
}

type noopEventSink struct {
}

func NewNoopEventSink() *noopEventSink {
	return &noopEventSink{}
}

func (es *noopEventSink) Call(call *exec.CallEvent, exception *errors.Exception) error {
	return nil
}

func (es *noopEventSink) Log(log *exec.LogEvent) error {
	return nil
}

type logFreeEventSink struct {
	EventSink
}

func NewLogFreeEventSink(eventSink EventSink) *logFreeEventSink {
	return &logFreeEventSink{
		EventSink: eventSink,
	}
}

func (esc *logFreeEventSink) Log(log *exec.LogEvent) error {
	return errors.ErrorCodef(errors.ErrorCodeIllegalWrite,
		"Log emitted from contract %v, but current call should be log-free", log.Address)
}
