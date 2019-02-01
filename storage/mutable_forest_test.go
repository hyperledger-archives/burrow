package storage

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestMutableForest_Genesis(t *testing.T) {
	rwf, err := NewMutableForest(dbm.NewMemDB(), 100)
	require.NoError(t, err)
	prefix := bz("fooos")
	tree, err := rwf.Writer(prefix)
	require.NoError(t, err)
	key1 := bz("bar")
	val1 := bz("nog")
	tree.Set(key1, val1)

	_, _, err = rwf.Save()
	require.NoError(t, err)
	var dump string
	rwf.Iterate(nil, nil, true, func(prefix []byte, tree KVCallbackIterableReader) error {
		dump = tree.(*RWTree).Dump()
		return nil
	})
	assert.Contains(t, dump, "bar -> nog")
	reader, err := rwf.Reader(prefix)
	require.NoError(t, err)
	require.Equal(t, val1, reader.Get(key1))

}

func TestMutableForest_Save(t *testing.T) {
	forest, err := NewMutableForest(dbm.NewMemDB(), 100)
	require.NoError(t, err)
	prefix1 := bz("fooos")
	tree, err := forest.Writer(prefix1)
	require.NoError(t, err)
	key1 := bz("bar")
	val1 := bz("nog")
	tree.Set(key1, val1)

	hash1, version1, err := forest.Save()
	require.NoError(t, err)

	assertDump(t, forest, `
        .
        └── fooos
            └── bar -> nog
        `)

	prefix2 := bz("prefixo")
	key2 := bz("hogs")
	val2 := bz("they are dogs")
	tree, err = forest.Writer(prefix2)
	require.NoError(t, err)
	tree.Set(key2, val2)

	hash2, version2, err := forest.Save()
	require.NoError(t, err)
	require.NotEqual(t, hash1, hash2)
	require.Equal(t, version1+1, version2, "versions should increment")

	assertDump(t, forest, `
        .
        ├── fooos
        │   └── bar -> nog
        └── prefixo
            └── hogs -> they are dogs
        `)
}

func TestMutableForest_Load(t *testing.T) {
	db := dbm.NewMemDB()
	forest, err := NewMutableForest(db, 100)
	require.NoError(t, err)
	prefix1 := bz("prefixes can be long if you want")
	tree, err := forest.Writer(prefix1)
	require.NoError(t, err)
	key1 := bz("El Nubble")
	val1 := bz("Diplodicus")
	tree.Set(key1, val1)

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
	tree, err := forest.Writer(bz("age"))
	tree.Get(bz("foo"))
	setForest(t, forest, "age", "Lindsay", "34")
	setForest(t, forest, "age", "Cora", "1")
	_, _, err = forest.Save()
	require.NoError(t, err)
	assertDump(t, forest, `
        .
        ├── age
        │   ├── Cora -> 1
        │   └── Lindsay -> 34
        ├── balances
        │   ├── Caitlin -> 2344
        │   ├── Cora -> 654456
        │   ├── Edward -> 34
        │   └── Lindsay -> 654
        └── names
            ├── Caitlin -> female
            ├── Cora -> female
            ├── Edward -> male
            └── Lindsay -> unisex
        `)
	fmt.Println(forest.Dump())
}

func setForest(t *testing.T, forest *MutableForest, prefix, key, value string) {
	tree, err := forest.Writer([]byte(prefix))
	require.NoError(t, err)
	tree.Set([]byte(key), []byte(value))

}

func assertDump(t *testing.T, forest interface{ Dump() string }, expectedDump string) {
	actual := forest.Dump()
	expectedDump = trimMargin(expectedDump)
	if expectedDump != actual {
		t.Errorf("forest.Dump() did not match expected dump:\n%s\nDo you want this assertion instead:\n\n"+
			"assertDump(t, forest,`\n%s`)\n\n", expectedDump, actual)
	}
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
