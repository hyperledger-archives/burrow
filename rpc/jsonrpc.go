// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"encoding/json"
	"fmt"
)

// JSON-RPC 2.0 error codes.
const (
	INVALID_REQUEST  = -32600
	METHOD_NOT_FOUND = -32601
	INVALID_PARAMS   = -32602
	INTERNAL_ERROR   = -32603
	PARSE_ERROR      = -32700
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

func (err RPCError) Error() string {
	return fmt.Sprintf("Error %v: %s", err.Code, err.Message)
}

// Create a new RPC request. This is the generic struct that is passed to RPC
// methods
func NewRPCRequest(id string, method string, params json.RawMessage) *RPCRequest {
	return &RPCRequest{
		JSONRPC: "2.0",
		Id:      id,
		Method:  method,
		Params:  params,
	}
}

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
