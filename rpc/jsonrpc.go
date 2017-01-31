// Copyright 2015-2017 Monax Industries Limited.
// This file is part of the Monax platform (Monax)

// Monax is free software: you can use, redistribute it and/or modify
// it only under the terms of the GNU General Public License, version
// 3, as published by the Free Software Foundation.

// Monax is distributed WITHOUT ANY WARRANTY pursuant to
// the terms of the Gnu General Public Licence, version 3, including
// (but not limited to) Clause 15 thereof. See the text of the
// GNU General Public License, version 3 for full terms.

// You should have received a copy of the GNU General Public License,
// version 3, with Monax.  If not, see <http://www.gnu.org/licenses/>.

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
// Refer to JSON-RPC specification http://www.jsonrpc.org/specification
type (
	RPCRequest struct {
		JSONRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
		Id      string          `json:"id"`
	}

	// RPCResponse MUST follow the JSON-RPC specification for Response object
	// reference: http://www.jsonrpc.org/specification#response_object
	RPCResponse interface {
		AssertIsRPCResponse() bool
	}

	// RPCResultResponse MUST NOT contain the error member if no error occurred
	RPCResultResponse struct {
		Result  interface{} `json:"result"`
		Id      string      `json:"id"`
		JSONRPC string      `json:"jsonrpc"`
	}

	// RPCErrorResponse MUST NOT contain the result member if an error occured
	RPCErrorResponse struct {
		Error   *RPCError `json:"error"`
		Id      string    `json:"id"`
		JSONRPC string    `json:"jsonrpc"`
	}

	// RPCError MUST be included in the Response object if an error occured
	RPCError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		// Note: Data is currently unused, and the data member may be omitted
		// Data  interface{} `json:"data"`
	}
)

// NewRPCResponse creates a new response object from a result
func NewRPCResponse(id string, res interface{}) RPCResponse {
	return RPCResponse(&RPCResultResponse{
		Result:  res,
		Id:      id,
		JSONRPC: "2.0",
	})
}

// NewRPCErrorResponse creates a new error-response object from the error code and message
func NewRPCErrorResponse(id string, code int, message string) RPCResponse {
	return RPCResponse(&RPCErrorResponse{
		Error:   &RPCError{code, message},
		Id:      id,
		JSONRPC: "2.0",
	})
}

// AssertIsRPCResponse implements a marker method for RPCResultResponse
// to implement the interface RPCResponse
func (rpcResultResponse *RPCResultResponse) AssertIsRPCResponse() bool {
	return true
}

// AssertIsRPCResponse implements a marker method for RPCErrorResponse
// to implement the interface RPCResponse
func (rpcErrorResponse *RPCErrorResponse) AssertIsRPCResponse() bool {
	return true
}
