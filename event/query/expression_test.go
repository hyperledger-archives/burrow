package query

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpression(t *testing.T) {
	t.Run("Basic OR AND", func(t *testing.T) {
		qry, err := New("(something = 'awful' OR something = 'nice') AND another_thing = 'OKAY'")
		require.NoError(t, err)
		out := qry.parser.String()
		require.Equal(t, "something, 'awful', =, something, 'nice', =, OR, another_thing, 'OKAY', =, AND", out)

		getter := func(key string) (interface{}, bool) {
			switch key {
			case "something":
				return "awful", true

			case "another_thing":
				return "OKAY", true

			default:
				return "", false
			}
		}

		matches, err := qry.parser.Evaluate(getter)
		require.NoError(t, err)
		require.True(t, matches)
	})
}
