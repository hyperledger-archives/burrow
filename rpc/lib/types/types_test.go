package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type SampleResult struct {
	Value string
}

func TestResponses(t *testing.T) {
	assert := assert.New(t)

	a := NewRPCSuccessResponse("1", &SampleResult{"hello"})
	b, _ := json.Marshal(a)
	s := `{"jsonrpc":"2.0","id":"1","result":{"Value":"hello"}}`
	assert.Equal(string(s), string(b))

	d := RPCParseError("1", errors.New("Hello world"))
	e, _ := json.Marshal(d)
	f := `{"jsonrpc":"2.0","id":"1","error":{"code":-32700,"message":"Parse Error","data":"Hello world"}}`
	assert.Equal(string(f), string(e))

	g := RPCMethodNotFoundError("2")
	h, _ := json.Marshal(g)
	i := `{"jsonrpc":"2.0","id":"2","error":{"code":-32601,"message":"Method Not Found"}}`
	assert.Equal(i, string(h))
}

func TestRPCError(t *testing.T) {
	assert.Equal(t, `Unknown Error 12 - Badness: "One worse than a code 11"`,
		fmt.Sprintf("%v", &RPCError{
			Code:    12,
			Message: "Badness",
			Data:    raw("One worse than a code 11"),
		}))

	assert.Equal(t, "Unknown Error 12 - Badness",
		fmt.Sprintf("%v", &RPCError{
			Code:    12,
			Message: "Badness",
		}))
}

func raw(v interface{}) json.RawMessage {
	bs, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		panic(err)
	}
	return bs
}
