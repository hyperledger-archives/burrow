package structure

import . "github.com/eris-ltd/eris-db/util/slice"

const (
	// Key for go time.Time object
	TimeKey = "time"
	// Key for call site for log invocation
	CallerKey = "caller"
	// Key for String name for level
	LevelKey   = "level"
	ChannelKey = "channel"
	// String message key
	MessageKey = "message"
	// Key for module or function or struct that is the subject of the logging
	ComponentKey = "component"
)

// Pull the specified values from a structured log line into a map.
// Assumes keys are single-valued. And returns the rest as context.
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

// Return a single value corresponding to key in keyvals
func Value(keyvals []interface{}, key interface{}) interface{} {
	for i := 0; i < 2*(len(keyvals)/2); i += 2 {
		if keyvals[i] == key {
			return keyvals[i+1]
		}
	}
	return nil
}