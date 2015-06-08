package rpc

import (
	"encoding/json"
)

// JSON-RPC 2.0 error codes.
const (
	PARSE_ERROR      = -32700
	INVALID_REQUEST  = -32600
	METHOD_NOT_FOUND = -32601
	INVALID_PARAMS   = -32602
	INTERNAL_ERROR   = -32603
)

// Request and Response objects. Id is a string. Error data not used.
type (
	RPCRequest struct {
		JSONRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
		Id      string          `json:"id"`
	}

	RPCResponse struct {
		Result  interface{} `json:"result"`
		Error   *RPCError   `json:"error"`
		Id      string      `json:"id"`
		JSONRPC string      `json:"jsonrpc"`
	}

	RPCError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
)

// Create a new response object from a result.
func NewRPCResponse(id string, res interface{}) *RPCResponse {
	if res == nil {
		res = struct{}{}
	}
	return &RPCResponse{
		Result:  res,
		Error:   nil,
		Id:      id,
		JSONRPC: "2.0",
	}
}

// Create a new error-response object from the error code and message.
func NewRPCErrorResponse(id string, code int, message string) *RPCResponse {
	return &RPCResponse{
		Result:  nil,
		Error:   &RPCError{code, message},
		Id:      id,
		JSONRPC: "2.0",
	}
}
