package errors

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorCode_MarshalJSON(t *testing.T) {
	ec := NewException(ErrorCodeDataStackOverflow, "arrgh")
	bs, err := json.Marshal(ec)
	require.NoError(t, err)

	ecOut := new(Exception)
	err = json.Unmarshal(bs, ecOut)
	require.NoError(t, err)

	assert.Equal(t, ec, ecOut)
}

func TestCode_String(t *testing.T) {
	err := ErrorCodeCodeOutOfBounds
	fmt.Println(err.Error())
}

func TestFirstOnly(t *testing.T) {
	err := FirstOnly()
	// This will be a wrapped nil - it should not register as first error
	var ex CodedError = (*Exception)(nil)
	err.PushError(ex)
	// This one should
	realErr := ErrorCodef(ErrorCodeInsufficientBalance, "real error")
	err.PushError(realErr)
	assert.True(t, realErr.Equal(err.Error()))
}
