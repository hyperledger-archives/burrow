package errors

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorCode_MarshalJSON(t *testing.T) {
	ec := NewCodedError(ErrorCodeDataStackOverflow, "arrg")
	bs, err := json.Marshal(ec)
	require.NoError(t, err)

	ecOut := new(Exception)
	err = json.Unmarshal(bs, ecOut)
	require.NoError(t, err)

	assert.Equal(t, ec, ecOut)
}
