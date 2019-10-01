package errors

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorCode_MarshalJSON(t *testing.T) {
	ec := NewException(Codes.DataStackOverflow, "arrgh")
	bs, err := json.Marshal(ec)
	require.NoError(t, err)

	ecOut := new(Exception)
	err = json.Unmarshal(bs, ecOut)
	require.NoError(t, err)

	assert.Equal(t, ec, ecOut)
}

func TestCode_String(t *testing.T) {
	err := Codes.CodeOutOfBounds
	fmt.Println(err.Error())
}
