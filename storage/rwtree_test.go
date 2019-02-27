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
	_, _, err := rwt.Save()
	require.NoError(t, err)
	assert.Equal(t, dam, rwt.Get(foo))
}

func TestEmptyTree(t *testing.T) {
	db := dbm.NewMemDB()
	rwt := NewRWTree(db, 100)
	fmt.Printf("%X\n", rwt.Hash())
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
	rwt.Iterate(nil, nil, true, func(key, value []byte) error {
		fmt.Println(string(key), " => ", string(value))
		return nil
	})
	require.NoError(t, err)

	// Make a new tree
	rwt = NewRWTree(db, 100)
	err = rwt.Load(version1, true)
	require.NoError(t, err)
	// If you load version1 the working hash is that which you saved after version0, i.e. hash0
	require.Equal(t, hash1, rwt.Hash())

	// Run the same writes again
	rwt.Set(foo, gaa)
	rwt.Set(gaa, dam)
	hash3, version3, err := rwt.Save()
	require.NoError(t, err)
	rwt.Iterate(nil, nil, true, func(key, value []byte) error {
		fmt.Println(string(key), " => ", string(value))
		return nil
	})

	// Expect the same hashes
	assert.Equal(t, hash2, hash3)
	assert.Equal(t, version2, version3)
}

func TestVersionDivergence(t *testing.T) {
	// This test serves as a reminder that IAVL nodes contain the version and a new node is created for every write
	rwt1 := NewRWTree(dbm.NewMemDB(), 100)
	rwt1.Set(bz("Raffle"), bz("Topper"))
	hash11, _, err := rwt1.Save()
	require.NoError(t, err)

	rwt2 := NewRWTree(dbm.NewMemDB(), 100)
	rwt2.Set(bz("Raffle"), bz("Topper"))
	hash21, _, err := rwt2.Save()
	require.NoError(t, err)

	// The following 'ought' to be idempotent but isn't since it replaces the previous node with an identical one, but
	// with an incremented version number
	rwt2.Set(bz("Raffle"), bz("Topper"))
	hash22, _, err := rwt2.Save()
	require.NoError(t, err)

	assert.Equal(t, hash11, hash21)
	assert.NotEqual(t, hash11, hash22)
}

func TestMutableTree_Iterate(t *testing.T) {
	mut := NewMutableTree(dbm.NewMemDB(), 100)
	mut.Set(bz("aa"), bz("1"))
	mut.Set(bz("aab"), bz("2"))
	mut.Set(bz("aac"), bz("3"))
	mut.Set(bz("aad"), bz("4"))
	mut.Set(bz("ab"), bz("5"))
	_, _, err := mut.SaveVersion()
	require.NoError(t, err)
	mut.IterateRange(bz("aab"), bz("aad"), true, func(key []byte, value []byte) bool {
		fmt.Printf("%q -> %q\n", key, value)
		return false
	})
	fmt.Println("foo")
	mut.IterateRange(bz("aab"), bz("aad"), false, func(key []byte, value []byte) bool {
		fmt.Printf("%q -> %q\n", key, value)
		return false
	})
	fmt.Println("foo")
	mut.IterateRange(bz("aad"), bz("aab"), true, func(key []byte, value []byte) bool {
		fmt.Printf("%q -> %q\n", key, value)
		return false
	})
}
