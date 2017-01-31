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
