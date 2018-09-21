package storage

import "bytes"

type Trie struct {
	Root *TrieNode
}

type TrieNode struct {
	key      []byte
	value    []byte
	children [][]*TrieNode
	mask     [4]uint64
}

func NewTrie() *Trie {
	return &Trie{
		Root: &TrieNode{},
	}
}

func (tn *TrieNode) Insert(key, value []byte) {
	bytes.Compare(tn.key, key)
}
