package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiIterator(t *testing.T) {
	t.Log("Testing forward iterator...")
	ci1 := iteratorOver(kvPairs("a", "dogs"))
	ci2 := iteratorOver(kvPairs("b", "frogs", "x", "mogs"))
	ci3 := iteratorOver(kvPairs("d", "bar", "h", "flobs"))
	ci4 := iteratorOver(kvPairs("c", "zfoo", "A", "nibble", "\xFF", "HIGH"))
	mi := NewMultiIterator(false, ci4, ci2, ci3, ci1)
	start, end := mi.Domain()
	assert.Equal(t, "A", string(start))
	assert.Equal(t, "\xFF", string(end))
	assertIteratorSorted(t, mi, false)

	t.Log("Testing reverse iterator...")
	ci1 = iteratorOver(kvPairs("a", "dogs"), true)
	ci2 = iteratorOver(kvPairs("b", "frogs", "x", "mogs"), true)
	ci3 = iteratorOver(kvPairs("d", "bar", "h", "flobs"), true)
	ci4 = iteratorOver(kvPairs("c", "zfoo", "A", "nibble", "", ""), true)
	mi = NewMultiIterator(true, ci4, ci2, ci3, ci1)
	start, end = mi.Domain()
	assert.Equal(t, "x", string(start))
	assert.Equal(t, "", string(end))
	assertIteratorSorted(t, mi, true)
}

func TestDuplicateKeys(t *testing.T) {
	t.Log("Testing iterators with duplicate keys...")
	ci1 := iteratorOver(kvPairs("a", "dogs"))
	ci2 := iteratorOver(kvPairs("a", "frogs", "x", "mogs"))
	ci3 := iteratorOver(kvPairs("a", "bar", "h", "flobs"))
	ci4 := iteratorOver(kvPairs("a", "zfoo", "A", "nibble", "\xFF", "HIGH"))
	mi := NewMultiIterator(false, ci1, ci2, ci3, ci4)
	var as []string
	for mi.Valid() {
		if string(mi.Key()) == "a" {
			as = append(as, string(mi.Value()))
		}
		mi.Next()
	}
	assert.Equal(t, []string{"dogs", "frogs", "bar", "zfoo"}, as,
		"duplicate keys should appear in iterator order")
}
