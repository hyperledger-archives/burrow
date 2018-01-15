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
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

const OBJECTS = 100

type FilterableObject struct {
	Integer int
	String  string
}

// Filter for integer value.
// Ops: All
type IntegerFilter struct {
	op    string
	value uint64
	match func(uint64, uint64) bool
}

func (this *IntegerFilter) Configure(fd *FilterData) error {
	val, err := ParseNumberValue(fd.Value)
	if err != nil {
		return err
	}
	match, err2 := GetRangeFilter(fd.Op, "integer")
	if err2 != nil {
		return err2
	}
	this.match = match
	this.op = fd.Op
	this.value = val
	return nil
}

func (this *IntegerFilter) Match(v interface{}) bool {
	fo, ok := v.(FilterableObject)
	if !ok {
		return false
	}
	return this.match(uint64(fo.Integer), this.value)
}

// Filter for integer value.
// Ops: All
type StringFilter struct {
	op    string
	value string
	match func(string, string) bool
}

func (this *StringFilter) Configure(fd *FilterData) error {
	match, err := GetStringFilter(fd.Op, "string")
	if err != nil {
		return err
	}
	this.match = match
	this.op = fd.Op
	this.value = fd.Value
	return nil
}

func (this *StringFilter) Match(v interface{}) bool {
	fo, ok := v.(FilterableObject)
	if !ok {
		return false
	}
	return this.match(fo.String, this.value)
}

// Test suite
type FilterSuite struct {
	suite.Suite
	objects       []FilterableObject
	filterFactory *FilterFactory
}

func (fs *FilterSuite) SetupSuite() {
	objects := make([]FilterableObject, OBJECTS)

	for i := 0; i < 100; i++ {
		objects[i] = FilterableObject{i, fmt.Sprintf("string%d", i)}
	}

	ff := NewFilterFactory()

	ff.RegisterFilterPool("integer", &sync.Pool{
		New: func() interface{} {
			return &IntegerFilter{}
		},
	})

	ff.RegisterFilterPool("string", &sync.Pool{
		New: func() interface{} {
			return &StringFilter{}
		},
	})

	fs.objects = objects
	fs.filterFactory = ff
}

func (fs *FilterSuite) TearDownSuite() {

}

// ********************************************* Tests *********************************************

func (fs *FilterSuite) Test_FilterIntegersEquals() {
	fd := &FilterData{"integer", "==", "5"}
	filter, err := fs.filterFactory.NewFilter([]*FilterData{fd})
	fs.NoError(err)
	arr := []FilterableObject{}
	for _, o := range fs.objects {
		if filter.Match(o) {
			arr = append(arr, o)
			break
		}
	}
	fs.Equal(arr, fs.objects[5:6])
}

func (fs *FilterSuite) Test_FilterIntegersLT() {
	fd := &FilterData{"integer", "<", "5"}
	filter, err := fs.filterFactory.NewFilter([]*FilterData{fd})
	fs.NoError(err)
	arr := []FilterableObject{}
	for _, o := range fs.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	fs.Equal(arr, fs.objects[:5])
}

func (fs *FilterSuite) Test_FilterIntegersLTEQ() {
	fd := &FilterData{"integer", "<=", "10"}
	filter, err := fs.filterFactory.NewFilter([]*FilterData{fd})
	fs.NoError(err)
	arr := []FilterableObject{}
	for _, o := range fs.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	fs.Equal(arr, fs.objects[:11])
}

func (fs *FilterSuite) Test_FilterIntegersGT() {
	fd := &FilterData{"integer", ">", "50"}
	filter, err := fs.filterFactory.NewFilter([]*FilterData{fd})
	fs.NoError(err)
	arr := []FilterableObject{}
	for _, o := range fs.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	fs.Equal(arr, fs.objects[51:])
}

func (fs *FilterSuite) Test_FilterIntegersRange() {
	fd0 := &FilterData{"integer", ">", "5"}
	fd1 := &FilterData{"integer", "<", "38"}
	filter, err := fs.filterFactory.NewFilter([]*FilterData{fd0, fd1})
	fs.NoError(err)
	arr := []FilterableObject{}
	for _, o := range fs.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	fs.Equal(arr, fs.objects[6:38])
}

func (fs *FilterSuite) Test_FilterIntegersGTEQ() {
	fd := &FilterData{"integer", ">=", "77"}
	filter, err := fs.filterFactory.NewFilter([]*FilterData{fd})
	fs.NoError(err)
	arr := []FilterableObject{}
	for _, o := range fs.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	fs.Equal(arr, fs.objects[77:])
}

func (fs *FilterSuite) Test_FilterIntegersNEQ() {
	fd := &FilterData{"integer", "!=", "50"}
	filter, err := fs.filterFactory.NewFilter([]*FilterData{fd})
	fs.NoError(err)
	arr := []FilterableObject{}
	for _, o := range fs.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	res := make([]FilterableObject, OBJECTS)
	copy(res, fs.objects)
	res = append(res[:50], res[51:]...)
	fs.Equal(arr, res)
}

func (fs *FilterSuite) Test_FilterStringEquals() {
	fd := &FilterData{"string", "==", "string7"}
	filter, err := fs.filterFactory.NewFilter([]*FilterData{fd})
	fs.NoError(err)
	arr := []FilterableObject{}
	for _, o := range fs.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	fs.Equal(arr, fs.objects[7:8])
}

func (fs *FilterSuite) Test_FilterStringNEQ() {
	fd := &FilterData{"string", "!=", "string50"}
	filter, err := fs.filterFactory.NewFilter([]*FilterData{fd})
	fs.NoError(err)
	arr := []FilterableObject{}

	for _, o := range fs.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	res := make([]FilterableObject, OBJECTS)
	copy(res, fs.objects)
	res = append(res[:50], res[51:]...)
	fs.Equal(arr, res)
}

// ********************************************* Entrypoint *********************************************

func TestFilterSuite(t *testing.T) {
	suite.Run(t, &FilterSuite{})
}
