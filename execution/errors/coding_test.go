package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodings(t *testing.T) {
	a := Code.CodeOutOfBounds
	b := Code.CodeOutOfBounds
	ap := a
	bp := b
	require.True(t, ap.Equal(bp))
	require.True(t, ap == bp)
	js := Code.JSON()
	fmt.Println(js)
}
