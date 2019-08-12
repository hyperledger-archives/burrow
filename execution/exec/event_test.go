package exec

import (
	"testing"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmthrgd/go-hex"
)

func TestEventTagQueries(t *testing.T) {
	ev := logEvent()

	tev := ev.Tagged()

	qb := query.NewBuilder().AndEquals(event.EventTypeKey, TypeLog.String())
	qry, err := qb.Query()
	require.NoError(t, err)
	assert.True(t, qry.Matches(tev))
	require.NoError(t, qry.MatchError())

	qb = qb.AndContains(event.EventIDKey, "bar")
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.True(t, qry.Matches(tev))
	require.NoError(t, qry.MatchError())

	qb = qb.AndEquals(event.TxHashKey, hex.EncodeUpperToString(tev.Header.TxHash))
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.True(t, qry.Matches(tev))
	require.NoError(t, qry.MatchError())

	qb = qb.AndGreaterThanOrEqual(event.HeightKey, tev.Header.Height)
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.True(t, qry.Matches(tev))
	require.NoError(t, qry.MatchError())

	qb = qb.AndStrictlyLessThan(event.IndexKey, tev.Header.Index+1)
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.True(t, qry.Matches(tev))
	require.NoError(t, qry.MatchError())

	qb = qb.AndEquals(event.AddressKey, ev.Log.Address)
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.True(t, qry.Matches(tev))
	require.NoError(t, qry.MatchError())

	qb = qb.AndEquals(LogNTextKey(0), "marmot")
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.True(t, qry.Matches(tev))
	require.NoError(t, qry.MatchError())

	t.Logf("Query: %v", qry)
	t.Logf("Keys: %v", tev.Keys())
}

func BenchmarkMatching(b *testing.B) {
	b.StopTimer()
	ev := logEvent()
	qb := query.NewBuilder().AndEquals(event.EventTypeKey, TypeLog.String()).
		AndContains(event.EventIDKey, "bar").
		AndEquals(event.TxHashKey, hex.EncodeUpperToString(ev.Header.TxHash)).
		AndGreaterThanOrEqual(event.HeightKey, ev.Header.Height).
		AndStrictlyLessThan(event.IndexKey, ev.Header.Index+1).
		AndEquals(event.AddressKey, ev.Log.Address).
		AndEquals(LogNTextKey(0), "marmot")
	qry, err := qb.Query()
	require.NoError(b, err)
	tev := ev.Tagged()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		qry.Matches(tev)
	}
}

func logEvent() *Event {
	return &Event{
		Header: &Header{
			EventType: TypeLog,
			EventID:   "foo/bar",
			TxHash:    []byte{2, 3, 4},
			Height:    34,
			Index:     2,
		},
		Log: &LogEvent{
			Address: crypto.MustAddressFromHexString("DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF"),
			Topics:  []binary.Word256{binary.RightPadWord256([]byte("marmot"))},
		},
	}
}
