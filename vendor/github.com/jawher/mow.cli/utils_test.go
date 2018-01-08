package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJoinStrings(t *testing.T) {
	cases := []struct {
		input    []string
		expected string
	}{
		{nil, ""},
		{[]string{""}, ""},
		{[]string{" "}, ""},
		{[]string{"\t"}, ""},
		{[]string{"", " ", "\t"}, ""},
		{[]string{"a"}, "a"},
		{[]string{"a", "b c"}, "a b c"},
		{[]string{"", "a", " ", "b", "\t"}, "a b"},
	}

	for _, cas := range cases {
		t.Logf("Testing %#v", cas.input)
		actual := joinStrings(cas.input...)

		require.Equal(t, cas.expected, actual)
	}
}
