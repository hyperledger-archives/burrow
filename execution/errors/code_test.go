package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodings(t *testing.T) {
	a := Codes.CodeOutOfBounds
	b := Codes.CodeOutOfBounds
	ap := a
	bp := b
	require.True(t, ap.Equal(bp))
	require.True(t, ap == bp)
	js := Codes.JSON()
	fmt.Println(js)
}
