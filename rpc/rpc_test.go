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
	"testing"

	"github.com/stretchr/testify/assert"
)

// ...
func TestNewJsonRpcResponse(t *testing.T) {
	id := "testId"
	data := "a string"
	resp := RPCResponse(&RPCResultResponse{
		Result:  data,
		Id:      id,
		JSONRPC: "2.0",
	})
	respGen := NewRPCResponse(id, data)
	assert.Equal(t, respGen, resp)
}

// ...
func TestNewJsonRpcErrorResponse(t *testing.T) {
	id := "testId"
	code := 100
	message := "the error"
	resp := RPCResponse(&RPCErrorResponse{
		Error:   &RPCError{code, message},
		Id:      id,
		JSONRPC: "2.0",
	})
	respGen := NewRPCErrorResponse(id, code, message)
	assert.Equal(t, respGen, resp)
}
