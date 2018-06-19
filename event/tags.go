package event

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/pubsub"
)

type Tags interface {
	pubsub.TagMap
	Map() map[string]interface{}
	Keys() []string
}

type TagMap map[string]interface{}

func (ts TagMap) Get(key string) (value string, ok bool) {
	var vint interface{}
	vint, ok = ts[key]
	if !ok {
		return
	}
	switch v := vint.(type) {
	case string:
		value = v
	case fmt.Stringer:
		value = v.String()
	default:
		value = fmt.Sprintf("%v", v)
	}
	return
}

func (ts TagMap) Len() int {
	return len(ts)
}

func (ts TagMap) Map() map[string]interface{} {
	return ts
}

func (ts TagMap) Keys() []string {
	keys := make([]string, 0, len(ts))
	for k := range ts {
		keys = append(keys, k)
	}
	return keys
}

type CombinedTags []Tags

func (ct CombinedTags) Get(key string) (value string, ok bool) {
	for _, t := range ct {
		value, ok = t.Get(key)
		if ok {
			return
		}
	}
	return
}

func (ct CombinedTags) Len() (length int) {
	for _, t := range ct {
		length += t.Len()
	}
	return length
}

func (ct CombinedTags) Map() map[string]interface{} {
	tags := make(map[string]interface{})
	for _, t := range ct {
		for _, k := range t.Keys() {
			v, ok := t.Get(k)
			if ok {
				tags[k] = v
			}
		}
	}
	return tags
}

func (ct CombinedTags) Keys() []string {
	var keys []string
	for _, t := range ct {
		keys = append(keys, t.Keys()...)
	}
	return keys
}
