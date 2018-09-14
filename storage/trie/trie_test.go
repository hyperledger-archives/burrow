package trie

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	trie := NewTrie()
	set(trie, "hello")
	set(trie, "hellp")
	set(trie, "hellpl")
	set(trie, "apple")
	set(trie, "arnold")
	set(trie, "butter")
	set(trie, "buttercup")
	fmt.Println(trie.Dump())
}

func TestDelete(t *testing.T) {
	trie := NewTrie()
	set(trie, "hello")
	set(trie, "hellp")
	set(trie, "hellpl")
	set(trie, "apple")
	set(trie, "arnold")
	set(trie, "butter")
	set(trie, "buttercup")
	del(trie, "hellpl")
	del(trie, "hellp")
	set(trie, "hellp")
	fmt.Println(trie.Dump())
}

func TestDelete2(t *testing.T) {
	trie := NewTrie()
	set(trie, "hello")
	del(trie, "hello")
	set(trie, "hello")
	set(trie, "help")
	fmt.Println(trie.Dump())
}

func TestEmpty(t *testing.T) {
	trie := NewTrie()
	set(trie, "")
	fmt.Println(trie.Dump())
}

func TestGrow(t *testing.T) {
	tb := NewBranch(2)
	branchBytes := []byte("axcdpjks")
	for i := 0; i < len(branchBytes); i++ {
		bb := branchBytes[i]
		tb.Add(uint(bb), NewBranch(int(bb)))
	}
	require.Len(t, tb.children, len(branchBytes))
	sort.Slice(branchBytes, func(i, j int) bool {
		return branchBytes[i] < branchBytes[j]
	})
	for i := 0; i < len(branchBytes); i++ {
		index := tb.children[i].index
		require.Equal(t, int(branchBytes[i]), index)
	}
}

func TestCriticalIndex(t *testing.T) {
	assert.Equal(t, 0, criticalIndex([]byte(""), []byte("")))
	assert.Equal(t, 0, criticalIndex([]byte("a"), []byte("b")))
	assert.Equal(t, 1, criticalIndex([]byte("aa"), []byte("ab")))
	assert.Equal(t, 5, criticalIndex([]byte("aabra"), []byte("aabracadabra")))
	assert.Equal(t, 5, criticalIndex([]byte("aabra"), []byte("aabra")))
}

func TestChildIndex(t *testing.T) {
	tb := NewBranch(0)
	assert.Equal(t, uint(0), tb.childIndex('x'))
	tb.bitmap.Set('a')
	assert.Equal(t, uint(1), tb.childIndex('x'))
	tb.bitmap.Set('y')
	assert.Equal(t, uint(1), tb.childIndex('x'))
	tb.bitmap.Set('w')
	assert.Equal(t, uint(2), tb.childIndex('x'))
	tb.bitmap.Set('x')
	assert.Equal(t, uint(2), tb.childIndex('x'))
}

func set(trie *Trie, key string) bool {
	return trie.Set([]byte(key), key)
}

func del(trie *Trie, key string) bool {
	return trie.Delete([]byte(key))
}

func BenchmarkArrayAccessn(b *testing.B) {
	key := []byte("alongishkindofkey")
	keycopy := make([]uint, len(key))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(key); j++ {
			keycopy[j] = uint(key[j])
		}
	}
}

func BenchmarkCharAt(b *testing.B) {
	key := []byte("alongishkindofkey")
	keycopy := make([]uint, len(key)+1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 1; j <= len(key); j++ {
			keycopy[j] = charAt(key, j)
		}
	}
}
