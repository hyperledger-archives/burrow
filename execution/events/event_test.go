package events

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

func TestHeader_Key(t *testing.T) {
	h := &Header{
		EventID: "Foos",
		Height:  2345345232,
		Index:   34,
	}
	key := h.Key()
	keyString := hex.EncodeUpperToString(key)
	assert.Equal(t, "000000008BCB20D00000000000000022", keyString)
	assert.Len(t, keyString, 32, "should be 16 bytes")
	assert.Equal(t, h.Height, key.Height())
	assert.Equal(t, h.Index, key.Index())
}

func TestKey_IsSuccessorOf(t *testing.T) {
	assert.True(t, NewKey(1, 0).IsSuccessorOf(NewKey(0, 1)))
	assert.True(t, NewKey(100, 24).IsSuccessorOf(NewKey(100, 23)))
	assert.False(t, NewKey(100, 23).IsSuccessorOf(NewKey(100, 25)))
	assert.False(t, NewKey(1, 1).IsSuccessorOf(NewKey(0, 25)))
	assert.True(t, NewKey(3, 0).IsSuccessorOf(NewKey(2, 0)))
}

func TestEventTagQueries(t *testing.T) {
	addressHex := "DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF"
	address, err := crypto.AddressFromHexString(addressHex)
	require.NoError(t, err)
	ev := &Event{
		Header: &Header{
			EventType: TypeLog,
			EventID:   "foo/bar",
			TxHash:    []byte{2, 3, 4},
			Height:    34,
			Index:     2,
		},
		Log: &EventDataLog{
			Address: address,
			Topics:  []binary.Word256{binary.RightPadWord256([]byte("marmot"))},
		},
	}

	qb := query.NewBuilder().AndEquals(event.EventTypeKey, TypeLog.String())
	qry, err := qb.Query()
	require.NoError(t, err)
	assert.True(t, qry.Matches(ev))

	qb = qb.AndContains(event.EventIDKey, "bar")
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.True(t, qry.Matches(ev))

	qb = qb.AndEquals(event.TxHashKey, hex.EncodeUpperToString(ev.Header.TxHash))
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.True(t, qry.Matches(ev))

	qb = qb.AndGreaterThanOrEqual(event.HeightKey, ev.Header.Height)
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.True(t, qry.Matches(ev))

	qb = qb.AndStrictlyLessThan(event.IndexKey, ev.Header.Index+1)
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.True(t, qry.Matches(ev))

	qb = qb.AndEquals(event.AddressKey, addressHex)
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.True(t, qry.Matches(ev))

	qb = qb.AndEquals(event.LogNTextKey(0), "marmot")
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.True(t, qry.Matches(ev))

	t.Logf("Query: %v", qry)
	t.Logf("Keys: %v", ev.Keys())
}
