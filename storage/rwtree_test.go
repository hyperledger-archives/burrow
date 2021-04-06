package storage

import (
	"fmt"
	"runtime/debug"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	dbm "github.com/tendermint/tm-db"
)

func TestSave(t *testing.T) {
	db := dbm.NewMemDB()
	rwt, err := NewRWTree(db, 100)
	require.NoError(t, err)

	var tests = []struct {
		key   []byte
		value []byte
	}{
		{[]byte("foo"), []byte("foo")},
		{[]byte("gaa"), []byte("gaa")},
		{[]byte("dam"), []byte("dam")},
	}
	for _, tt := range tests {
		rwt.Set(tt.key, tt.value)
		_, _, err = rwt.Save()
		require.NoError(t, err)
		val, err := rwt.Get(tt.key)
		require.NoError(t, err)
		assert.Equal(t, tt.value, val)
	}
}

func TestEmptyTree(t *testing.T) {
	db := dbm.NewMemDB()
	rwt, err := NewRWTree(db, 100)
	require.NoError(t, err)
	fmt.Printf("%X\n", rwt.Hash())
}

func TestRollback(t *testing.T) {
	db := dbm.NewMemDB()
	rwt, err := NewRWTree(db, 100)
	require.NoError(t, err)
	rwt.Set([]byte("Raffle"), []byte("Topper"))
	_, _, err = rwt.Save()
	require.NoError(t, err)

	foo := []byte("foo")
	gaa := []byte("gaa")
	dam := []byte("dam")
	rwt.Set(foo, gaa)
	hash1, version1, err := rwt.Save()
	require.NoError(t, err)

	// Perform some writes on top of version1
	rwt.Set(foo, gaa)
	rwt.Set(gaa, dam)
	hash2, version2, err := rwt.Save()
	require.NoError(t, err)
	err = rwt.Iterate(nil, nil, true, func(key, value []byte) error {
		fmt.Println(string(key), " => ", string(value))
		return nil
	})
	require.NoError(t, err)

	// Make a new tree
	rwt, err = NewRWTree(db, 100)
	require.NoError(t, err)
	err = rwt.Load(version1, true)
	require.NoError(t, err)
	// If you load version1 the working hash is that which you saved after version0, i.e. hash0
	require.Equal(t, hash1, rwt.Hash())

	// Run the same writes again
	rwt.Set(foo, gaa)
	rwt.Set(gaa, dam)
	hash3, version3, err := rwt.Save()
	require.NoError(t, err)
	err = rwt.Iterate(nil, nil, true, func(key, value []byte) error {
		fmt.Println(string(key), " => ", string(value))
		return nil
	})
	require.NoError(t, err)

	// Expect the same hashes
	assert.Equal(t, hash2, hash3)
	assert.Equal(t, version2, version3)
}

func TestVersionDivergence(t *testing.T) {
	// This test serves as a reminder that IAVL nodes contain the version and a new node is created for every write
	rwt1, err := NewRWTree(dbm.NewMemDB(), 100)
	require.NoError(t, err)
	rwt1.Set([]byte("Raffle"), []byte("Topper"))
	hash11, _, err := rwt1.Save()
	require.NoError(t, err)

	rwt2, err := NewRWTree(dbm.NewMemDB(), 100)
	require.NoError(t, err)
	rwt2.Set([]byte("Raffle"), []byte("Topper"))
	hash21, _, err := rwt2.Save()
	require.NoError(t, err)

	// The following 'ought' to be idempotent but isn't since it replaces the previous node with an identical one, but
	// with an incremented version number
	rwt2.Set([]byte("Raffle"), []byte("Topper"))
	hash22, _, err := rwt2.Save()
	require.NoError(t, err)

	assert.Equal(t, hash11, hash21)
	assert.NotEqual(t, hash11, hash22)
}

func TestMutableTree_Iterate(t *testing.T) {
	mut, err := NewMutableTree(dbm.NewMemDB(), 100)
	require.NoError(t, err)
	mut.Set([]byte("aa"), []byte("1"))
	mut.Set([]byte("aab"), []byte("2"))
	mut.Set([]byte("aac"), []byte("3"))
	mut.Set([]byte("aad"), []byte("4"))
	mut.Set([]byte("ab"), []byte("5"))
	_, _, err = mut.SaveVersion()
	require.NoError(t, err)
	mut.IterateRange([]byte("aab"), []byte("aad"), true, func(key []byte, value []byte) bool {
		fmt.Printf("%q -> %q\n", key, value)
		return false
	})
	fmt.Println("foo")
	mut.IterateRange([]byte("aab"), []byte("aad"), false, func(key []byte, value []byte) bool {
		fmt.Printf("%q -> %q\n", key, value)
		return false
	})
	fmt.Println("foo")
	mut.IterateRange([]byte("aad"), []byte("aab"), true, func(key []byte, value []byte) bool {
		fmt.Printf("%q -> %q\n", key, value)
		return false
	})
}

func capturePanic(f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v: %s", r, debug.Stack())
		}
	}()
	err = f()
	return
}

func TestRWTreeConcurrency(t *testing.T) {
	rwt, err := NewRWTree(dbm.NewMemDB(), 100)
	require.NoError(t, err)
	n := 100

	doneCh := make(chan struct{})
	errCh := make(chan interface{})

	// Saturate with concurrent reads and writes
	var spin func()
	spin = func() {
		for i := 0; i < n; i++ {
			val := []byte{byte(i)}
			go func() {
				err := capturePanic(func() error {
					rwt.Set(val, val)
					return nil
				})
				if err != nil {
					errCh <- err
				}
			}()
			go func() {
				err := capturePanic(func() error {
					_, err := rwt.Get(val)
					return err
				})
				if err != nil {
					errCh <- err
				}
			}()

		}
		select {
		case <-doneCh:
			return
		default:
			// Avoid starvation
			time.Sleep(time.Millisecond)
			spin()
		}
	}

	// let's
	go spin()

	// Ensure Save() is safe with concurrent read/writes
	for i := 0; i < n/10; i++ {
		err := capturePanic(func() error {
			_, _, err := rwt.Save()
			return err
		})
		if err != nil {
			break
		}
	}
	close(doneCh)

	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
}
