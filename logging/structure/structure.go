package structure

import (
	"reflect"

	. "github.com/eris-ltd/eris-db/util/slice"
)

const (
	// Log time (time.Time)
	TimeKey = "time"
	// Call site for log invocation (go-stack.Call)
	CallerKey = "caller"
	// Level name (string)
	LevelKey = "level"
	// Channel name in a multiple channel logging context
	ChannelKey = "channel"
	// Log message (string)
	MessageKey = "message"
	// Top-level component (choose one) name
	ComponentKey = "component"
	// Vector-valued scope
	ScopeKey = "scope"
	// Globally unique identifier persisting while a single instance (root process)
	// of this program/service is running
	RunId = "run_id"
)

// Pull the specified values from a structured log line into a map.
// Assumes keys are single-valued.
// Returns a map of the key-values from the requested keys and
// the unmatched remainder keyvals as context as a slice of key-values.
func ValuesAndContext(keyvals []interface{},
	keys ...interface{}) (map[interface{}]interface{}, []interface{}) {
	vals := make(map[interface{}]interface{}, len(keys))
	context := make([]interface{}, len(keyvals))
	copy(context, keyvals)
	deletions := 0
	// We can't really do better than a linear scan of both lists here. N is small
	// so screw the asymptotics.
	// Guard against odd-length list
	for i := 0; i < 2*(len(keyvals)/2); i += 2 {
		for k := 0; k < len(keys); k++ {
			if keyvals[i] == keys[k] {
				// Pull the matching key-value pair into vals to return
				vals[keys[k]] = keyvals[i+1]
				// Delete the key once it's found
				keys = DeleteAt(keys, k)
				// And remove the key-value pair from context
				context = Delete(context, i-deletions, 2)
				// Keep a track of how much we've shrunk the context to offset next
				// deletion
				deletions += 2
				break
			}
		}
	}
	return vals, context
}

type vector []interface{}

func newVector(vals... interface{}) vector {
	return vals
}

func (v vector) Slice() []interface{} {
	return v
}

// Returns the unique keys in keyvals and a map of values where values of
// repeated keys are merged into a slice of those values in the order which they
// appeared
func KeyValuesVector(keyvals []interface{}) ([]interface{}, map[interface{}]interface{}) {
	keys := []interface{}{}
	vals := make(map[interface{}]interface{}, len(keyvals)/2)
	for i := 0; i < 2*(len(keyvals)/2); i += 2 {
		key := keyvals[i]
		val := keyvals[i+1]
		switch oldVal := vals[key].(type) {
		case nil:
			keys = append(keys, key)
			vals[key] = val
		case vector:
			// if this is, in fact, only the second time we have seen key and the
			// first time it had a value of []interface{} then here we are doing the
			// wrong thing by appending val. We will end up with
			// Slice(..., val) rather than Slice(Slice(...), val)
			vals[key] = vector(append(oldVal, val))
		default:
			vals[key] = newVector(oldVal, val)
		}
	}
	return keys, vals
}

// Return a single value corresponding to key in keyvals
func Value(keyvals []interface{}, key interface{}) interface{} {
	for i := 0; i < 2*(len(keyvals)/2); i += 2 {
		if keyvals[i] == key {
			return keyvals[i+1]
		}
	}
	return nil
}

// Obtain a canonical key from a value. Useful for structured logging where the
// type of value alone may be sufficient to determine its key. Providing this
// function centralises any convention over type names
func KeyFromValue(val interface{}) string {
	switch val.(type) {
	case string:
		return "text"
	default:
		return reflect.TypeOf(val).Name()
	}
}
