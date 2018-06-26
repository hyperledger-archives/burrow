package errors

import (
	"encoding/json"
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/abci/types"
)

func TestErrorCode_MarshalJSON(t *testing.T) {
	ec := NewErrorCode(ErrorCodeDataStackOverflow)
	bs, err := json.Marshal(ec)
	require.NoError(t, err)

	ecOut := new(ErrorCode)
	err = json.Unmarshal(bs, ecOut)
	require.NoError(t, err)

	assert.Equal(t, ec, ecOut)
}

func TestException_MarshalJSON(t *testing.T) {
	ex := NewCodedError(ErrorCodeExecutionReverted, "Oh noes we reverted")
	ex.BS = []byte{2, 3, 4, 5}
	bs, err := json.Marshal(ex)
	require.NoError(t, err)
	fmt.Println(string(bs))
	exOut := new(Exception)
	err = json.Unmarshal(bs, exOut)
	require.NoError(t, err)

	bb := types.RequestBeginBlock{
		Hash: []byte{2, 3, 4},
	}
	bs, err = json.Marshal(bb)
	require.NoError(t, err)
	fmt.Println(string(bs))
}
