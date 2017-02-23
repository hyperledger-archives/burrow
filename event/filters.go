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

package event

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
)

// TODO add generic filters for various different kinds of matching.

// Used to filter.
// Op can be any of the following:
// The usual relative operators: <, >, <=, >=, ==, != (where applicable)
// A range parameter (see: https://help.github.com/articles/search-syntax/)
type FilterData struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

// Filters based on fields.
type Filter interface {
	Match(v interface{}) bool
}

// A filter that can be configured with in-data.
type ConfigurableFilter interface {
	Filter
	Configure(*FilterData) error
}

// Filter made up of many filters.
type CompositeFilter struct {
	filters []Filter
}

func (this *CompositeFilter) SetData(filters []Filter) {
	this.filters = filters
}

func (this *CompositeFilter) Match(v interface{}) bool {
	for _, f := range this.filters {
		if !f.Match(v) {
			return false
		}
	}
	return true
}

// Rubberstamps everything.
type MatchAllFilter struct{}

func (this *MatchAllFilter) Match(v interface{}) bool { return true }

// Used to generate filters based on filter data.
// Keeping separate pools for "edge cases" (Composite and MatchAll)
type FilterFactory struct {
	filterPools         map[string]*sync.Pool
	compositeFilterPool *sync.Pool
	matchAllFilterPool  *sync.Pool
}

func NewFilterFactory() *FilterFactory {
	aff := &FilterFactory{}
	// Match all.
	aff.matchAllFilterPool = &sync.Pool{
		New: func() interface{} {
			return &MatchAllFilter{}
		},
	}
	// Composite.
	aff.compositeFilterPool = &sync.Pool{
		New: func() interface{} {
			return &CompositeFilter{}
		},
	}
	// Regular.
	aff.filterPools = make(map[string]*sync.Pool)

	return aff
}

func (this *FilterFactory) RegisterFilterPool(fieldName string, pool *sync.Pool) {
	this.filterPools[strings.ToLower(fieldName)] = pool
}

// Creates a new filter given the input data array. If the array is zero length or nil, an empty
// filter will be returned that returns true on all matches. If the array is of size 1, a regular
// filter is returned, otherwise a CompositeFieldFilter is returned, which is a special filter that
// contains a number of other filters. It implements AccountFieldFilter, and will match an account
// only if all the sub-filters matches.
func (this *FilterFactory) NewFilter(fdArr []*FilterData) (Filter, error) {

	if fdArr == nil || len(fdArr) == 0 {
		return &MatchAllFilter{}, nil
	}
	if len(fdArr) == 1 {
		return this.newSingleFilter(fdArr[0])
	}
	filters := []Filter{}
	for _, fd := range fdArr {
		f, err := this.newSingleFilter(fd)
		if err != nil {
			return nil, err
		}
		filters = append(filters, f)
	}
	cf := this.compositeFilterPool.Get().(*CompositeFilter)
	cf.filters = filters
	return cf, nil
}

func (this *FilterFactory) newSingleFilter(fd *FilterData) (ConfigurableFilter, error) {
	fp, ok := this.filterPools[strings.ToLower(fd.Field)]
	if !ok {
		return nil, fmt.Errorf("Field is not supported: " + fd.Field)
	}
	f := fp.Get().(ConfigurableFilter)
	err := f.Configure(fd)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Some standard value parsing functions.

func ParseNumberValue(value string) (int64, error) {
	var val int64
	// Check for wildcards.
	if value == "min" {
		val = math.MinInt64
	} else if value == "max" {
		val = math.MaxInt64
	} else {
		tv, err := strconv.ParseInt(value, 10, 64)

		if err != nil {
			return 0, fmt.Errorf("Wrong value type.")
		}
		val = tv
	}
	return val, nil
}

// Some standard filtering functions.

func GetRangeFilter(op, fName string) (func(a, b int64) bool, error) {
	if op == "==" {
		return func(a, b int64) bool {
			return a == b
		}, nil
	} else if op == "!=" {
		return func(a, b int64) bool {
			return a != b
		}, nil
	} else if op == "<=" {
		return func(a, b int64) bool {
			return a <= b
		}, nil
	} else if op == ">=" {
		return func(a, b int64) bool {
			return a >= b
		}, nil
	} else if op == "<" {
		return func(a, b int64) bool {
			return a < b
		}, nil
	} else if op == ">" {
		return func(a, b int64) bool {
			return a > b
		}, nil
	} else {
		return nil, fmt.Errorf("Op: " + op + " is not supported for '" + fName + "' filtering")
	}
}

func GetStringFilter(op, fName string) (func(s0, s1 string) bool, error) {
	if op == "==" {
		return func(s0, s1 string) bool {
			return strings.EqualFold(s0, s1)
		}, nil
	} else if op == "!=" {
		return func(s0, s1 string) bool {
			return !strings.EqualFold(s0, s1)
		}, nil
	} else {
		return nil, fmt.Errorf("Op: " + op + " is not supported for '" + fName + "' filtering.")
	}
}
