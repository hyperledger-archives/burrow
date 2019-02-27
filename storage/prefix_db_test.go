package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tendermint/libs/db"
)

func mockDBWithStuff() dbm.DB {
	db := dbm.NewMemDB()
	// Under "key" prefix
	db.Set(bz("key"), bz("value"))
	db.Set(bz("key1"), bz("value1"))
	db.Set(bz("key2"), bz("value2"))
	db.Set(bz("key3"), bz("value3"))
	db.Set(bz("something"), bz("else"))
	db.Set(bz(""), bz(""))
	db.Set(bz("k"), bz("val"))
	db.Set(bz("ke"), bz("valu"))
	db.Set(bz("kee"), bz("valuu"))
	return db
}

func TestPrefixDBSimple(t *testing.T) {
	db := mockDBWithStuff()
	pdb := NewPrefixDB(db, "key")

	checkValue(t, pdb, bz("key"), nil)
	checkValue(t, pdb, bz(""), bz("value"))
	checkValue(t, pdb, bz("key1"), nil)
	checkValue(t, pdb, bz("1"), bz("value1"))
	checkValue(t, pdb, bz("key2"), nil)
	checkValue(t, pdb, bz("2"), bz("value2"))
	checkValue(t, pdb, bz("key3"), nil)
	checkValue(t, pdb, bz("3"), bz("value3"))
	checkValue(t, pdb, bz("something"), nil)
	checkValue(t, pdb, bz("k"), nil)
	checkValue(t, pdb, bz("ke"), nil)
	checkValue(t, pdb, bz("kee"), nil)
}

func TestPrefixDBIterator1(t *testing.T) {
	db := mockDBWithStuff()
	pdb := NewPrefixDB(db, "key")

	itr := pdb.Iterator(nil, nil)
	checkDomain(t, itr, nil, nil)
	checkItem(t, itr, bz(""), bz("value"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("1"), bz("value1"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("2"), bz("value2"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("3"), bz("value3"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBIterator2(t *testing.T) {
	db := mockDBWithStuff()
	pdb := NewPrefixDB(db, "key")

	itr := pdb.Iterator(nil, bz(""))
	checkDomain(t, itr, nil, bz(""))
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBIterator3(t *testing.T) {
	db := mockDBWithStuff()
	pdb := NewPrefixDB(db, "key")

	itr := pdb.Iterator(bz(""), nil)
	checkDomain(t, itr, bz(""), nil)
	checkItem(t, itr, bz(""), bz("value"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("1"), bz("value1"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("2"), bz("value2"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("3"), bz("value3"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBIterator4(t *testing.T) {
	db := mockDBWithStuff()
	pdb := NewPrefixDB(db, "key")

	itr := pdb.Iterator(bz(""), bz(""))
	checkDomain(t, itr, bz(""), bz(""))
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator1(t *testing.T) {
	db := mockDBWithStuff()
	pdb := NewPrefixDB(db, "key")

	itr := pdb.ReverseIterator(nil, nil)
	checkDomain(t, itr, nil, nil)
	checkItem(t, itr, bz("3"), bz("value3"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("2"), bz("value2"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("1"), bz("value1"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz(""), bz("value"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator2(t *testing.T) {
	db := mockDBWithStuff()
	pdb := NewPrefixDB(db, "key")

	itr := pdb.ReverseIterator(bz(""), nil)
	checkDomain(t, itr, bz(""), nil)
	checkItem(t, itr, bz("3"), bz("value3"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("2"), bz("value2"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("1"), bz("value1"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz(""), bz("value"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator3(t *testing.T) {
	db := mockDBWithStuff()
	pdb := NewPrefixDB(db, "key")

	itr := pdb.ReverseIterator(nil, bz(""))
	checkDomain(t, itr, nil, bz(""))
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator4(t *testing.T) {
	db := mockDBWithStuff()
	pdb := NewPrefixDB(db, "key")

	itr := pdb.ReverseIterator(bz(""), bz(""))
	checkDomain(t, itr, bz(""), bz(""))
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator5(t *testing.T) {
	db := mockDBWithStuff()
	pdb := NewPrefixDB(db, "key")

	itr := pdb.ReverseIterator(bz("1"), nil)
	checkDomain(t, itr, bz("1"), nil)
	checkItem(t, itr, bz("3"), bz("value3"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("2"), bz("value2"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("1"), bz("value1"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator6(t *testing.T) {
	db := mockDBWithStuff()
	pdb := NewPrefixDB(db, "key")

	itr := pdb.ReverseIterator(bz("2"), nil)
	checkDomain(t, itr, bz("2"), nil)
	checkItem(t, itr, bz("3"), bz("value3"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("2"), bz("value2"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator7(t *testing.T) {
	db := mockDBWithStuff()
	pdb := NewPrefixDB(db, "key")

	itr := pdb.ReverseIterator(nil, bz("2"))
	checkDomain(t, itr, nil, bz("2"))
	checkItem(t, itr, bz("1"), bz("value1"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz(""), bz("value"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func (p Prefix) BadSuffix(key []byte) []byte {
	return key[len(p):]
}

func TestBadSuffix(t *testing.T) {
	p := Prefix([]byte("foo"))
	fmt.Println(cap(p))
	key1 := p.BadSuffix([]byte("fooaaa"))
	fmt.Println(cap(p), p, string(key1))
	key2 := p.BadSuffix([]byte("foobbb"))
	fmt.Println(cap(p), p, string(key1))
	fmt.Println(cap(p), p, string(key2))

}

func checkValue(t *testing.T, db dbm.DB, key []byte, valueWanted []byte) {
	valueGot := db.Get(key)
	assert.Equal(t, valueWanted, valueGot)
}

func checkValid(t *testing.T, itr dbm.Iterator, expected bool) {
	valid := itr.Valid()
	require.Equal(t, expected, valid)
}

func checkNext(t *testing.T, itr dbm.Iterator, expected bool) {
	itr.Next()
	valid := itr.Valid()
	require.Equal(t, expected, valid)
}

func checkNextPanics(t *testing.T, itr dbm.Iterator) {
	assert.Panics(t, func() { itr.Next() }, "checkNextPanics expected panic but didn't")
}

func checkDomain(t *testing.T, itr dbm.Iterator, start, end []byte) {
	ds, de := itr.Domain()
	assert.Equal(t, start, ds, "checkDomain domain start incorrect")
	assert.Equal(t, end, de, "checkDomain domain end incorrect")
}

func checkItem(t *testing.T, itr dbm.Iterator, key []byte, value []byte) {
	k, v := itr.Key(), itr.Value()
	assert.Exactly(t, key, k)
	assert.Exactly(t, value, v)
}

func checkInvalid(t *testing.T, itr dbm.Iterator) {
	checkValid(t, itr, false)
	checkKeyPanics(t, itr)
	checkValuePanics(t, itr)
	checkNextPanics(t, itr)
}

func checkKeyPanics(t *testing.T, itr dbm.Iterator) {
	assert.Panics(t, func() { itr.Key() }, "checkKeyPanics expected panic but didn't")
}

func checkValuePanics(t *testing.T, itr dbm.Iterator) {
	assert.Panics(t, func() { itr.Key() }, "checkValuePanics expected panic but didn't")
}
