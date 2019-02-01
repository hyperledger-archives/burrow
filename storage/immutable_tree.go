package storage

import "github.com/tendermint/iavl"

// We wrap IAVL's tree types in order to provide iteration helpers and to harmonise other interface types with what we
// expect

type ImmutableTree struct {
	*iavl.ImmutableTree
}

func (imt *ImmutableTree) Get(key []byte) []byte {
	_, value := imt.ImmutableTree.Get(key)
	return value
}

func (imt *ImmutableTree) Iterate(start, end []byte, ascending bool, fn func(key []byte, value []byte) error) error {
	var err error
	imt.ImmutableTree.IterateRange(start, end, ascending, func(key, value []byte) bool {
		err = fn(key, value)
		if err != nil {
			// stop
			return true
		}
		return false
	})
	return err
}
