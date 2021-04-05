package engine

import (
	"github.com/hyperledger/burrow/acm"
)

type Dispatcher interface {
	// If this Dispatcher is capable of dispatching this account (e.g. if it has the correct bytecode) then return a
	// Callable that wraps the function, otherwise return nil
	Dispatch(acc *acm.Account) Callable
}

type DispatcherFunc func(acc *acm.Account) Callable

func (d DispatcherFunc) Dispatch(acc *acm.Account) Callable {
	return d(acc)
}

// An ExternalDispatcher is able to Dispatch accounts to external engines as well as Dispatch to itself
type ExternalDispatcher interface {
	Dispatcher
	SetExternals(externals Dispatcher)
}

// An ExternalDispatcher is able to Dispatch accounts to external engines as well as Dispatch to itself
type Externals struct {
	// Provide any foreign dispatchers to allow calls between VMs
	externals Dispatcher
}

var _ ExternalDispatcher = (*Externals)(nil)

func (ed *Externals) Dispatch(acc *acm.Account) Callable {
	// Try external calls then fallback to EVM
	if ed.externals == nil {
		return nil
	}
	return ed.externals.Dispatch(acc)
}

func (ed *Externals) SetExternals(externals Dispatcher) {
	ed.externals = externals
}

type Dispatchers []Dispatcher

func NewDispatchers(dispatchers ...Dispatcher) Dispatchers {
	out := dispatchers[:0]
	// Flatten dispatchers and omit nil dispatchers (allows optional dispatchers in chain)
	for i, d := range dispatchers {
		ds, ok := d.(Dispatchers)
		if ok {
			// Add tail to nested dispatchers if one exists
			if len(dispatchers) > i {
				ds = append(ds, dispatchers[i+1:]...)
			}
			return append(out, NewDispatchers(ds...)...)
		} else if d != nil {
			out = append(out, d)
		}
	}
	return out
}

// Connect ExternalDispatchers eds to each other so that the underlying engines can mutually call contracts hosted by
// other dispatchers
func Connect(eds ...ExternalDispatcher) {
	for i, ed := range eds {
		// Collect external dispatchers excluding this one (to avoid infinite dispatcher loops!)
		others := make([]Dispatcher, 0, len(eds)-1)
		for offset := 1; offset < len(eds); offset++ {
			idx := (i + offset) % len(eds)
			others = append(others, eds[idx])
		}
		ed.SetExternals(NewDispatchers(others...))
	}
}

func (ds Dispatchers) Dispatch(acc *acm.Account) Callable {
	for _, d := range ds {
		callable := d.Dispatch(acc)
		if callable != nil {
			return callable
		}
	}
	return nil
}

type ExternalsStorage struct {
}
