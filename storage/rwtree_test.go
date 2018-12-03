package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestSave(t *testing.T) {
	db := dbm.NewMemDB()
	rwt := NewRWTree(db, 100)
	foo := bz("foo")
	gaa := bz("gaa")
	dam := bz("dam")
	rwt.Set(foo, gaa)
	rwt.Save()
	assert.Equal(t, gaa, rwt.Get(foo))
	rwt.Set(foo, dam)
	rwt.Save()
	assert.Equal(t, dam, rwt.Get(foo))
}

func TestRollback(t *testing.T) {
	db := dbm.NewMemDB()
	rwt := NewRWTree(db, 100)
	rwt.Set(bz("Raffle"), bz("Topper"))
	_, _, err := rwt.Save()

	foo := bz("foo")
	gaa := bz("gaa")
	dam := bz("dam")
	rwt.Set(foo, gaa)
	hash1, version1, err := rwt.Save()
	require.NoError(t, err)

	// Perform some writes on top of version1
	rwt.Set(foo, gaa)
	rwt.Set(gaa, dam)
	hash2, version2, err := rwt.Save()
	rwt.IterateRange(nil, nil, true, func(key []byte, value []byte) bool {
		fmt.Println(string(key), " => ", string(value))
		return false
	})
	require.NoError(t, err)

	// Make a new tree
	rwt = NewRWTree(db, 100)
	err = rwt.Load(version1)
	require.NoError(t, err)
	// If you load version1 the working hash is that which you saved after version0, i.e. hash0
	require.Equal(t, hash1, rwt.Hash())

	// Run the same writes again
	rwt.Set(foo, gaa)
	rwt.Set(gaa, dam)
	hash3, version3, err := rwt.Save()
	require.NoError(t, err)
	rwt.IterateRange(nil, nil, true, func(key []byte, value []byte) bool {
		fmt.Println(string(key), " => ", string(value))
		return false
	})

	// Expect the same hashes
	assert.Equal(t, hash2, hash3)
	assert.Equal(t, version2, version3)
}
