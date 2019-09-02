package dump

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb/opt"
	dbm "github.com/tendermint/tm-db"
)

// The tests in this package are quite a good starting point for investigating the inadequacies of IAVL...
func TestMain(m *testing.M) {
	// For pprof
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	code := m.Run()
	os.Exit(code)
}

func BenchmarkDump(b *testing.B) {
	b.StopTimer()
	st := testLoad(b, NewMockSource(1000, 1000, 100, 10000))
	dumper := NewDumper(st, &bcm.Blockchain{})
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		err := dumper.Transmit(NullSink{}, 0, 0, All)
		require.NoError(b, err)
	}
}

func TestDump(t *testing.T) {
	mockSource := NewMockSource(50, 50, 100, 100)
	st := testLoad(t, mockSource)
	dumper := NewDumper(st, mockSource)
	sink := CollectSink{
		Rows: make([]string, 0),
	}
	err := dumper.Transmit(&sink, 0, 0, All)
	require.NoError(t, err)

	sort.Strings(sink.Rows)

	m := NewMockSource(50, 50, 100, 100)
	data := make([]string, 0)

	for {
		row, err := m.Recv()
		if err == io.EOF {
			break
		}
		bs, _ := json.Marshal(row)
		data = append(data, string(bs))
	}

	sort.Strings(data)

	require.Equal(t, sink.Rows, data)
}

// Test util

func normaliseDump(dump string) string {
	rows := strings.Split(dump, "\n")
	sort.Stable(sort.StringSlice(rows))
	return strings.Join(rows, "\n")
}

func dumpToJSONString(t *testing.T, st *state.State, blockchain Blockchain) string {
	buf := new(bytes.Buffer)
	receiver := NewDumper(st, blockchain).Source(0, 0, All)
	err := Write(buf, receiver, false, All)
	require.NoError(t, err)
	return string(buf.Bytes())
}

func loadDumpFromJSONString(t *testing.T, st *state.State, jsonDump string) {
	reader, err := NewJSONReader(bytes.NewBufferString(jsonDump))
	require.NoError(t, err)
	err = Load(reader, st)
	require.NoError(t, err)
}

func testDB(t testing.TB) dbm.DB {
	testDir, err := ioutil.TempDir("", "TestDump")
	require.NoError(t, err)
	var options *opt.Options
	db, err := dbm.NewGoLevelDBWithOpts("TestDumpDB", testDir, options)
	require.NoError(t, err)
	return db
}
