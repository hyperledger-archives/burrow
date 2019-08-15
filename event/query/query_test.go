package query

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatches(t *testing.T) {
	var (
		txDate = "2017-01-01"
		txTime = "2018-05-03T14:45:00Z"
	)

	now := time.Now().Format(TimeLayout)

	testCases := []struct {
		s       string
		tags    map[string]interface{}
		err     bool
		matches bool
	}{
		{"Height CONTAINS '2'", map[string]interface{}{"Height": uint64(12)}, false, true},
		{"Height CONTAINS '2'", map[string]interface{}{"Height": uint64(11)}, false, false},
		{"foo > 10", map[string]interface{}{"foo": 11}, false, true},
		{"foo >= 10", map[string]interface{}{"foo": uint64(11)}, false, true},
		{"foo >= 10", map[string]interface{}{"foo": uint32(11)}, false, true},
		{"foo >= 10", map[string]interface{}{"foo": uint(11)}, false, true},
		{fmt.Sprintf("(foo >= 10 OR foo CONTAINS 'frogs') AND badger < TIME %s", now),
			map[string]interface{}{"foo": "Ilikefrogs", "badger": time.Unix(343, 0)}, false, true},
		{fmt.Sprintf("foo >= 10 OR foo CONTAINS 'frogs' AND badger < TIME %s", now),
			map[string]interface{}{"foo": "Ilikefrogs", "badger": time.Unix(343, 0)}, false, true},
		{fmt.Sprintf("foo CONTAINS 'frosgs' OR  (foo >= 10 AND badger < TIME %s)", now),
			map[string]interface{}{"foo": "Ilikefrogs", "badger": time.Unix(343, 0)}, false, false},
		{fmt.Sprintf("foo CONTAINS 'mute' AND foo >= 10 OR badger < TIME %s", now),
			map[string]interface{}{"foo": "Ilikefrogs", "badger": time.Unix(343, 0)}, false, true},
		{fmt.Sprintf("foo CONTAINS 'mute' AND (foo >= 10 OR badger < TIME %s)", now),
			map[string]interface{}{"foo": "Ilikefrogs", "badger": time.Unix(343, 0)}, false, false},
		{"tm.events.type='NewBlock'", map[string]interface{}{"tm.events.type": "NewBlock"}, false, true},
		{"tx.gas > 7", map[string]interface{}{"tx.gas": "8"}, false, true},
		{"tx.gas > 7 AND tx.gas < 9", map[string]interface{}{"tx.gas": "8"}, false, true},
		{"body.weight >= 3.5", map[string]interface{}{"body.weight": "3.5"}, false, true},
		{"account.balance < 1000.0", map[string]interface{}{"account.balance": "900"}, false, true},
		{"apples.kg <= 4", map[string]interface{}{"apples.kg": "4.0"}, false, true},
		{"body.weight >= 4.5", map[string]interface{}{"body.weight": fmt.Sprintf("%v", float32(4.5))}, false, true},
		{"oranges.kg < 4 AND watermellons.kg > 10", map[string]interface{}{"oranges.kg": "3", "watermellons.kg": "12"}, false, true},
		{"peaches.kg < 4", map[string]interface{}{"peaches.kg": "5"}, false, false},

		{"tx.date > DATE 2017-01-01", map[string]interface{}{"tx.date": time.Now().Format(DateLayout)}, false, true},
		{"tx.date = DATE 2017-01-01", map[string]interface{}{"tx.date": txDate}, false, true},
		{"tx.date = DATE 2018-01-01", map[string]interface{}{"tx.date": txDate}, false, false},

		{"tx.time >= TIME 2013-05-03T14:45:00Z", map[string]interface{}{"tx.time": time.Now().Format(TimeLayout)}, false, true},
		{"tx.time = TIME 2013-05-03T14:45:00Z", map[string]interface{}{"tx.time": txTime}, false, false},

		{"abci.owner.name CONTAINS 'Igor'", map[string]interface{}{"abci.owner.name": "Igor,Ivan"}, false, true},
		{"abci.owner.name CONTAINS 'Igor'", map[string]interface{}{"abci.owner.name": "Pavel,Ivan"}, false, false},
	}

	for _, tc := range testCases {
		q, err := New(tc.s)
		require.NoError(t, err)
		if !tc.err {
			require.Nil(t, err)
		}

		q.ExplainTo(func(format string, args ...interface{}) {
			fmt.Printf(format, args...)
		})

		matches := q.Matches(TagMap(tc.tags))
		err = q.MatchError()
		require.NoError(t, err)
		if tc.matches {
			assert.True(t, matches, "Query '%s' should match %v", tc.s, tc.tags)
		} else {
			assert.False(t, matches, "Query '%s' should not match %v", tc.s, tc.tags)
		}
	}
}

func TestMustParse(t *testing.T) {
	assert.Panics(t, func() { MustParse("=") })
	assert.NotPanics(t, func() { MustParse("tm.events.type='NewBlock'") })
}
