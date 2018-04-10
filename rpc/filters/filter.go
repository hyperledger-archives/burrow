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

package filters

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// TODO add generic filters for various different kinds of matching.

// Filters based on fields.
type Filter interface {
	Match(v interface{}) bool
}

// A filter that can be configured with in-data.
type ConfigurableFilter interface {
	Filter
	Configure(*FilterData) error
}

// Used to filter.
// Op can be any of the following:
// The usual relative operators: <, >, <=, >=, ==, != (where applicable)
// A range parameter (see: https://help.github.com/articles/search-syntax/)
type FilterData struct {
	Field string
	Op    string
	Value string
}

// Filter made up of many filters.
type CompositeFilter struct {
	filters []Filter
}

func (cf *CompositeFilter) SetData(filters []Filter) {
	cf.filters = filters
}

func (cf *CompositeFilter) Match(v interface{}) bool {
	for _, f := range cf.filters {
		if !f.Match(v) {
			return false
		}
	}
	return true
}

// Rubberstamps everything.
type MatchAllFilter struct{}

func (maf *MatchAllFilter) Match(v interface{}) bool { return true }

// Some standard value parsing functions.

func ParseNumberValue(value string) (uint64, error) {
	var val uint64
	// Check for wildcards.
	if value == "min" {
		val = 0
	} else if value == "max" {
		val = math.MaxUint64
	} else {
		tv, err := strconv.ParseUint(value, 10, 64)

		if err != nil {
			return 0, fmt.Errorf("Wrong value type.")
		}
		val = tv
	}
	return val, nil
}

// Some standard filtering functions.

func GetRangeFilter(op, fName string) (func(a, b uint64) bool, error) {
	if op == "==" {
		return func(a, b uint64) bool {
			return a == b
		}, nil
	} else if op == "!=" {
		return func(a, b uint64) bool {
			return a != b
		}, nil
	} else if op == "<=" {
		return func(a, b uint64) bool {
			return a <= b
		}, nil
	} else if op == ">=" {
		return func(a, b uint64) bool {
			return a >= b
		}, nil
	} else if op == "<" {
		return func(a, b uint64) bool {
			return a < b
		}, nil
	} else if op == ">" {
		return func(a, b uint64) bool {
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
