package engine

import (
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/stretchr/testify/require"
)

type testDispatcher struct {
	name      string
	externals Dispatcher
}

func newDispatcher(name string) *testDispatcher {
	return &testDispatcher{
		name: name,
	}
}

func (t *testDispatcher) SetExternals(externals Dispatcher) {
	t.externals = externals
}

func (t *testDispatcher) String() string {
	if t.externals != nil {
		ds, ok := t.externals.(Dispatchers)
		if ok {
			var exts []string
			for _, d := range ds {
				td, ok := d.(*testDispatcher)
				if ok {
					exts = append(exts, td.name)
				}
			}
			if len(exts) > 0 {
				return fmt.Sprintf("%s -> %v", t.name, exts)
			}
		}
	}
	return t.name
}

func (t *testDispatcher) Dispatch(acc *acm.Account) Callable {
	return nil
}

func TestNewDispatchers(t *testing.T) {
	d1 := newDispatcher("1")
	d2 := newDispatcher("2")
	d3 := newDispatcher("3")

	dd1 := NewDispatchers(d1, d2, d3)

	require.Equal(t, "[1 2 3 1 2 3]", fmt.Sprint(NewDispatchers(dd1, dd1)))
	require.Equal(t, "[1 2 3 1 3]", fmt.Sprint(NewDispatchers(dd1, d1, d3)))
	require.Equal(t, "[3 1 1 2 3]", fmt.Sprint(NewDispatchers(d3, d1, dd1)))
}

func TestConnect(t *testing.T) {
	d1 := newDispatcher("1")
	d2 := newDispatcher("2")
	d3 := newDispatcher("3")

	// Check we don't panic
	Connect()

	// Does nothing but fine
	Connect(d1)
	require.Equal(t, "1", d1.String())

	Connect(d1, d2)
	require.Equal(t, "1 -> [2]", d1.String())
	require.Equal(t, "2 -> [1]", d2.String())
	require.Equal(t, "3", d3.String())

	Connect(d1, d2, d3)
	require.Equal(t, "1 -> [2 3]", d1.String())
	require.Equal(t, "2 -> [3 1]", d2.String())
	require.Equal(t, "3 -> [1 2]", d3.String())
}
