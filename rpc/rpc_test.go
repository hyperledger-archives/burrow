package rpc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// ...
func TestNewJsonRpcResponse(t *testing.T) {
	id := "testId"
	data := "a string"
	resp := &RPCResponse{
		Result:  data,
		Error:   nil,
		Id:      id,
		JSONRPC: "2.0",
	}
	respGen := NewRPCResponse(id, data)
	assert.Equal(t, respGen, resp)
}

// ...
func TestNewJsonRpcErrorResponse(t *testing.T) {
	id := "testId"
	code := 100
	message := "the error"
	resp := &RPCResponse{
		Result:  nil,
		Error:   &RPCError{code, message},
		Id:      id,
		JSONRPC: "2.0",
	}
	respGen := NewRPCErrorResponse(id, code, message)
	assert.Equal(t, respGen, resp)
}