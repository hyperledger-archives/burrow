// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package structure

import . "github.com/hyperledger/burrow/util/slice"

const (
	// Log time (time.Time)
	TimeKey = "time"
	// Call site for log invocation (go-stack.Call)
	CallerKey = "caller"
	// Trace for log call
	TraceKey = "trace"
	// Level name (string)
	LevelKey = "level"
	// Channel name in a vector channel logging context
	ChannelKey = "log_channel"
	// Log message (string)
	MessageKey = "message"
	// Captured logging source (like tendermint_log15, stdlib_log)
	CapturedLoggingSourceKey = "captured_logging_source"
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

// Drops all key value pairs where the key is in keys
func RemoveKeys(keyvals []interface{}, keys ...interface{}) []interface{} {
	keyvalsWithoutKeys := make([]interface{}, 0, len(keyvals))
NEXT_KEYVAL:
	for i := 0; i < 2*(len(keyvals)/2); i += 2 {
		for _, key := range keys {
			if keyvals[i] == key {
				continue NEXT_KEYVAL
			}
		}
		keyvalsWithoutKeys = append(keyvalsWithoutKeys, keyvals[i], keyvals[i+1])
	}
	return keyvalsWithoutKeys
}

// Stateful index that tracks the location of a possible vector value
type vectorValueindex struct {
	// Location of the value belonging to a key in output slice
	valueIndex int
	// Whether or not the value is currently a vector
	vector bool
}

// 'Vectorises' values associated with repeated string keys member by collapsing many values into a single vector value.
// The result is a copy of keyvals where the first occurrence of each matching key and its first value are replaced by
// that key and all of its values in a single slice.
func Vectorise(keyvals []interface{}, vectorKeys ...string) []interface{} {
	// We rely on working against a single backing array, so we use a capacity that is the maximum possible size of the
	// slice after vectorising (in the case there are no duplicate keys and this is a no-op)
	outputKeyvals := make([]interface{}, 0, len(keyvals))
	// Track the location and vector status of the values in the output
	valueIndices := make(map[string]*vectorValueindex, len(vectorKeys))
	elided := 0
	for i := 0; i < 2*(len(keyvals)/2); i += 2 {
		key := keyvals[i]
		val := keyvals[i+1]

		// Only attempt to vectorise string keys
		if k, ok := key.(string); ok {
			if valueIndices[k] == nil {
				// Record that this key has been seen once
				valueIndices[k] = &vectorValueindex{
					valueIndex: i + 1 - elided,
				}
				// Copy the key-value to output with the single value
				outputKeyvals = append(outputKeyvals, key, val)
			} else {
				// We have seen this key before
				vi := valueIndices[k]
				if !vi.vector {
					// This must be the only second occurrence of the key so now vectorise the value
					outputKeyvals[vi.valueIndex] = []interface{}{outputKeyvals[vi.valueIndex]}
					vi.vector = true
				}
				// Grow the vector value
				outputKeyvals[vi.valueIndex] = append(outputKeyvals[vi.valueIndex].([]interface{}), val)
				// We are now running two more elements behind the input keyvals because we have absorbed this key-value pair
				elided += 2
			}
		} else {
			// Just copy the key-value to the output for non-string keys
			outputKeyvals = append(outputKeyvals, key, val)
		}
	}
	return outputKeyvals
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

// Maps key values pairs with a function (key, value) -> (new key, new value)
func MapKeyValues(keyvals []interface{}, fn func(interface{}, interface{}) (interface{}, interface{})) []interface{} {
	mappedKeyvals := make([]interface{}, len(keyvals))
	for i := 0; i < 2*(len(keyvals)/2); i += 2 {
		key := keyvals[i]
		val := keyvals[i+1]
		mappedKeyvals[i], mappedKeyvals[i+1] = fn(key, val)
	}
	return mappedKeyvals
}
