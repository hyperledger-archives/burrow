package query

import (
	"sort"
)

type Tagged interface {
	Keys() []string
	Get(key string) (value interface{}, ok bool)
	// Len returns the number of tags.
	Len() int
}

type TagMap map[string]interface{}

func MapFromTagged(tagged Tagged) map[string]interface{} {
	tags := make(map[string]interface{})
	for _, k := range tagged.Keys() {
		v, ok := tagged.Get(k)
		if ok {
			tags[k] = v
		}
	}
	return tags
}

func (ts TagMap) Get(key string) (value interface{}, ok bool) {
	var vint interface{}
	vint, ok = ts[key]
	if !ok {
		return "", false
	}
	return vint, true
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
	sort.Strings(keys)
	return keys
}

type CombinedTags struct {
	keys []string
	ks   map[string][]Tagged
}

func NewCombinedTags() *CombinedTags {
	return &CombinedTags{
		ks: make(map[string][]Tagged),
	}
}

func MergeTags(tags ...Tagged) *CombinedTags {
	ct := NewCombinedTags()
	ct.AddTags(false, tags...)
	return ct
}

func ConcatTags(tags ...Tagged) *CombinedTags {
	ct := NewCombinedTags()
	ct.AddTags(true, tags...)
	return ct
}

// Adds each of tagsList to CombinedTags - choosing whether values for the same key should
// be concatenated or whether the first should value should stick
func (ct *CombinedTags) AddTags(concat bool, tagsList ...Tagged) {
	for _, t := range tagsList {
		for _, k := range t.Keys() {
			if len(ct.ks[k]) == 0 {
				ct.keys = append(ct.keys, k)
				// Store reference to key-holder amongst Taggeds
				ct.ks[k] = append(ct.ks[k], t)
			} else if concat {
				// Store additional tag reference only if concat, otherwise first key-holder wins
				ct.ks[k] = append(ct.ks[k], t)
			}
		}
	}
}

func (ct *CombinedTags) Get(key string) (interface{}, bool) {
	ts := ct.ks[key]
	if len(ts) == 0 {
		return "", false
	}
	values := make([]interface{}, 0, len(ts))
	for _, t := range ts {
		value, ok := t.Get(key)
		if ok {
			values = append(values, value)
		}
	}
	if len(values) == 0 {
		return nil, false
	}
	if len(values) == 1 {
		return values[0], true
	}
	return values, true
}

func (ct *CombinedTags) Len() (length int) {
	return len(ct.keys)
}

func (ct *CombinedTags) Keys() []string {
	return ct.keys
}
