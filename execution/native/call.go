package native

import (
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
)

// Call provides a standard wrapper for implementing Callable.Call with appropriate error handling and event firing.
func Call(state engine.State, params engine.CallParams, execute func(engine.State, engine.CallParams) ([]byte, error)) ([]byte, error) {
	maybe := new(errors.Maybe)
	if params.CallType == exec.CallTypeCall || params.CallType == exec.CallTypeCode {
		// NOTE: DelegateCall does not transfer the value to the callee.

		// DelegateCall is executed by the DELEGATECALL opcode, introduced as off Ethereum Homestead.
		// The intent of delegate call is to run the code of the callee in the storage context of the caller;
		// while preserving the original caller to the previous callee.
		// Different to the normal CALL or CALLCODE, the value does not need to be transferred to the callee.
		maybe.PushError(Transfer(state.CallFrame, params.Caller, params.Callee, params.Value))
	}

	output := maybe.Bytes(execute(state, params))
	// fire the post call event (including exception if applicable) and make sure we return the accumulated call error
	maybe.PushError(FireCallEvent(state.CallFrame, maybe.Error(), state.EventSink, output, params))
	return output, maybe.Error()
}

func FireCallEvent(callFrame *engine.CallFrame, callErr error, eventSink exec.EventSink, output []byte,
	params engine.CallParams) error {
	// fire the post call event (including exception if applicable)
	return eventSink.Call(&exec.CallEvent{
		CallType: params.CallType,
		CallData: &exec.CallData{
			Caller: params.Caller,
			Callee: params.Callee,
			Data:   params.Input,
			Value:  params.Value,
			Gas:    *params.Gas,
		},
		Origin:     params.Origin,
		StackDepth: callFrame.CallStackDepth(),
		Return:     output,
	}, errors.AsException(callErr))
}
