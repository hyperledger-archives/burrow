package storage

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"

	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

func TestMutableForest_Genesis(t *testing.T) {
	rwf, err := NewMutableForest(dbm.NewMemDB(), 100)
	require.NoError(t, err)
	prefix := []byte("fooos")
	key1 := []byte("bar")
	val1 := []byte("nog")
	err = rwf.Write(prefix, func(tree *RWTree) error {
		tree.Set(key1, val1)
		return nil
	})
	require.NoError(t, err)

	_, _, err = rwf.Save()
	require.NoError(t, err)
	var dump string
	err = rwf.Iterate(nil, nil, true, func(prefix []byte, tree KVCallbackIterableReader) error {
		dump = tree.(*RWTree).Dump()
		return nil
	})
	require.NoError(t, err)
	assert.Contains(t, dump, "\"bar\" -> \"nog\"")
	reader, err := rwf.Reader(prefix)
	require.NoError(t, err)
	val2, err := reader.Get(key1)
	require.NoError(t, err)
	require.Equal(t, val1, val2)

}

func TestMutableForest_Save(t *testing.T) {
	forest, err := NewMutableForest(dbm.NewMemDB(), 100)
	require.NoError(t, err)
	prefix1 := []byte("fooos")
	key1 := []byte("bar")
	val1 := []byte("nog")
	err = forest.Write(prefix1, func(tree *RWTree) error {
		tree.Set(key1, val1)
		return nil
	})
	require.NoError(t, err)

	hash1, version1, err := forest.Save()
	require.NoError(t, err)
	assertDump(t, forest, `
        	            	.
        	            	├── "Commits"
        	            	│   └── "fooos" -> "\b\x01\x12 ym.\xb8fw\xdcIK\xe8QQ\xb6\x8a\x1fT\x15\xff\x80\xd5\xd91\xf6YKf\x12wx\x16l\xf5"
        	            	└── "fooos"
        	            	    └── "bar" -> "nog"
        	            	`)

	prefix2 := []byte("prefixo")
	key2 := []byte("hogs")
	val2 := []byte("they are dogs")
	err = forest.Write(prefix2, func(tree *RWTree) error {
		tree.Set(key2, val2)
		return nil
	})
	require.NoError(t, err)

	hash2, version2, err := forest.Save()
	require.NoError(t, err)
	require.NotEqual(t, hash1, hash2)
	require.Equal(t, version1+1, version2, "versions should increment")

	assertDump(t, forest, `
        	            	.
        	            	├── "Commits"
        	            	│   ├── "fooos" -> "\b\x01\x12 ym.\xb8fw\xdcIK\xe8QQ\xb6\x8a\x1fT\x15\xff\x80\xd5\xd91\xf6YKf\x12wx\x16l\xf5"
        	            	│   └── "prefixo" -> "\b\x01\x12 E\xb2\xa4{аA\xddf\xcc\x02ȭ\xfa\xd1\xceZ\xa0nP\xe0\xd3\\X\x9c\x16M\xc1\x88t\x15\x8c"
        	            	├── "fooos"
        	            	│   └── "bar" -> "nog"
        	            	└── "prefixo"
        	            	    └── "hogs" -> "they are dogs"
        	            	`)
}

func TestMutableForest_Load(t *testing.T) {
	db := dbm.NewMemDB()
	forest, err := NewMutableForest(db, 100)
	require.NoError(t, err)
	prefix1 := []byte("prefixes can be long if you want")
	err = forest.Write(prefix1, func(tree *RWTree) error {
		key1 := []byte("El Nubble")
		val1 := []byte("Diplodicus")
		tree.Set(key1, val1)
		return nil
	})
	require.NoError(t, err)

	hash, version, err := forest.Save()
	require.NoError(t, err)

	dump := forest.Dump()

	forest, err = NewMutableForest(db, 100)
	require.NoError(t, err)
	err = forest.Load(version)
	require.NoError(t, err)

	require.Equal(t, hash, forest.Hash())
	require.Equal(t, version, forest.Version())
	require.Equal(t, dump, forest.Dump())
}

func TestSorted(t *testing.T) {
	forest, err := NewMutableForest(dbm.NewMemDB(), 100)
	require.NoError(t, err)
	setForest(t, forest, "names", "Edward", "male")
	setForest(t, forest, "names", "Caitlin", "female")
	setForest(t, forest, "names", "Lindsay", "unisex")
	setForest(t, forest, "names", "Cora", "female")
	_, _, err = forest.Save()
	require.NoError(t, err)
	setForest(t, forest, "balances", "Edward", "34")
	setForest(t, forest, "balances", "Caitlin", "2344")
	_, _, err = forest.Save()
	require.NoError(t, err)
	setForest(t, forest, "balances", "Lindsay", "654")
	setForest(t, forest, "balances", "Cora", "654456")
	_, _, err = forest.Save()
	require.NoError(t, err)
	err = forest.Write([]byte("age"), func(tree *RWTree) error {
		_, err = tree.Get([]byte("foo"))
		return err
	})
	require.NoError(t, err)
	setForest(t, forest, "age", "Lindsay", "34")
	setForest(t, forest, "age", "Cora", "1")
	_, _, err = forest.Save()
	require.NoError(t, err)

	assertDump(t, forest, `
        	            	.
        	            	├── "Commits"
        	            	│   ├── "age" -> "\b\x01\x12 \x1dwd_\xbaRB\xf5\xa6\xf0\n\xab\x9aWY\xf7\t\x16t웿\xb6\x89O\n\xcf&\xf7\xe6\xcd\n"
        	            	│   ├── "balances" -> "\b\x02\x12 \x9f\xab\xd3s\x18{\xbc\xe8\x98\xdai\xf5\x9f\x16\xden\xac(\xc9ԷU\x99\x17\xda'\xfa3-\x98\xd4\xc9"
        	            	│   └── "names" -> "\b\x01\x12 \xbf\xf8\xf9vt>\xbc\x06@C\xe9I\x01C\xa3\xc3O \xbc\xaf\xbf\xb3\b\xb2UHh\xe8TM\xb3\xba"
        	            	├── "age"
        	            	│   ├── "Cora" -> "1"
        	            	│   └── "Lindsay" -> "34"
        	            	├── "balances"
        	            	│   ├── "Caitlin" -> "2344"
        	            	│   ├── "Cora" -> "654456"
        	            	│   ├── "Edward" -> "34"
        	            	│   └── "Lindsay" -> "654"
        	            	└── "names"
        	            	    ├── "Caitlin" -> "female"
        	            	    ├── "Cora" -> "female"
        	            	    ├── "Edward" -> "male"
        	            	    └── "Lindsay" -> "unisex"
        	            	`)
}

func TestForestConcurrency(t *testing.T) {
	db := dbm.NewMemDB()
	cacheSize := 10
	forest, err := NewMutableForest(db, cacheSize)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	var prefixes [][]byte

	// Make sure we see some cache evictions
	for i := 0; i < cacheSize*2; i++ {
		prefixes = append(prefixes, []byte(fmt.Sprintf("prefix-%d", i)))
	}

	n := 10000
	for i := 0; i < n; i++ {
		index := i
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, uint64(index))
		g.Go(func() error {
			return forest.Write(prefixes[index%len(prefixes)], func(tree *RWTree) error {
				tree.Set(key, key)
				return nil
			})
		})
		g.Go(func() error {
			tree, err := forest.Reader(prefixes[index%2])
			if err != nil {
				return err
			}
			_, err = tree.Get(key)
			return err
		})
	}

	for i := 0; i < 501; i++ {
		_, _, err := forest.Save()
		require.NoError(t, err)
	}

	require.NoError(t, g.Wait())
	g, ctx = errgroup.WithContext(ctx)

	_, _, err = forest.Save()
	require.NoError(t, err)

	checkKeysStored := func(when string) {
		g := new(errgroup.Group)
		for i := 0; i < n; i++ {
			index := i
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, uint64(index))
			checker := func() error {
				tree, err := forest.Reader(prefixes[index%len(prefixes)])
				if err != nil {
					return err
				}
				value, err := tree.Get(key)
				if err != nil {
					return err
				}
				if !bytes.Equal(key, value) {
					return fmt.Errorf("%s: key-value %d/%d, expected key '%X' to store its own value, but got: '%X'",
						when, index, n, key, value)
				}
				return nil
			}
			g.Go(checker)
		}
		require.NoError(t, g.Wait(), "error checking keys in %s", when)
	}

	checkKeysStored("after saves, before reload")

	latestVersion := forest.Version()

	// Check persistence and start with cold caches
	forest, err = NewMutableForest(db, 100)
	require.NoError(t, err)

	err = forest.Load(latestVersion)
	require.NoError(t, err)

	checkKeysStored("after reload")

	require.NoError(t, g.Wait())
}

func setForest(t *testing.T, forest *MutableForest, prefix, key, value string) {
	err := forest.Write([]byte(prefix), func(tree *RWTree) error {
		tree.Set([]byte(key), []byte(value))
		return nil
	})
	require.NoError(t, err)

}

func assertDump(t *testing.T, forest interface{ Dump() string }, expectedDump string) {
	actual := forest.Dump()
	expectedDump = trimMargin(expectedDump)
	assert.Equal(t, expectedDump, actual,
		"forest.Dump() did not match expected dump:\n%s\nDo you want this assertion instead:\n\n"+
			"assertDump(t, forest,`\n%s`)\n\n", expectedDump, actual)
}

func trimMargin(str string) string {
	buf := new(bytes.Buffer)
	margin := 0
	for _, l := range strings.Split(str, "\n") {
		if margin == 0 {
			// Find the full-stop
			i := strings.Index(l, ".")
			if i > 0 {
				// trim this off every subsequent line
				margin = i
			} else {
				continue
			}
		}
		// Ignore blank lines
		if strings.Trim(l, " \t") != "" {
			buf.WriteString(l[margin:])
			buf.WriteString("\n")
		}
	}
	return buf.String()
}
