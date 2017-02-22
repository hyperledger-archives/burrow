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

	event "github.com/eris-ltd/eris-db/event"
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
	value int64
	match func(int64, int64) bool
}

func (this *IntegerFilter) Configure(fd *event.FilterData) error {
	val, err := event.ParseNumberValue(fd.Value)
	if err != nil {
		return err
	}
	match, err2 := event.GetRangeFilter(fd.Op, "integer")
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
	return this.match(int64(fo.Integer), this.value)
}

// Filter for integer value.
// Ops: All
type StringFilter struct {
	op    string
	value string
	match func(string, string) bool
}

func (this *StringFilter) Configure(fd *event.FilterData) error {
	match, err := event.GetStringFilter(fd.Op, "string")
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
	filterFactory *event.FilterFactory
}

func (this *FilterSuite) SetupSuite() {
	objects := make([]FilterableObject, OBJECTS, OBJECTS)

	for i := 0; i < 100; i++ {
		objects[i] = FilterableObject{i, fmt.Sprintf("string%d", i)}
	}

	ff := event.NewFilterFactory()

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

	this.objects = objects
	this.filterFactory = ff
}

func (this *FilterSuite) TearDownSuite() {

}

// ********************************************* Tests *********************************************

func (this *FilterSuite) Test_FilterIntegersEquals() {
	fd := &event.FilterData{"integer", "==", "5"}
	filter, err := this.filterFactory.NewFilter([]*event.FilterData{fd})
	this.NoError(err)
	arr := []FilterableObject{}
	for _, o := range this.objects {
		if filter.Match(o) {
			arr = append(arr, o)
			break
		}
	}
	this.Equal(arr, this.objects[5:6])
}

func (this *FilterSuite) Test_FilterIntegersLT() {
	fd := &event.FilterData{"integer", "<", "5"}
	filter, err := this.filterFactory.NewFilter([]*event.FilterData{fd})
	this.NoError(err)
	arr := []FilterableObject{}
	for _, o := range this.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	this.Equal(arr, this.objects[:5])
}

func (this *FilterSuite) Test_FilterIntegersLTEQ() {
	fd := &event.FilterData{"integer", "<=", "10"}
	filter, err := this.filterFactory.NewFilter([]*event.FilterData{fd})
	this.NoError(err)
	arr := []FilterableObject{}
	for _, o := range this.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	this.Equal(arr, this.objects[:11])
}

func (this *FilterSuite) Test_FilterIntegersGT() {
	fd := &event.FilterData{"integer", ">", "50"}
	filter, err := this.filterFactory.NewFilter([]*event.FilterData{fd})
	this.NoError(err)
	arr := []FilterableObject{}
	for _, o := range this.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	this.Equal(arr, this.objects[51:])
}

func (this *FilterSuite) Test_FilterIntegersRange() {
	fd0 := &event.FilterData{"integer", ">", "5"}
	fd1 := &event.FilterData{"integer", "<", "38"}
	filter, err := this.filterFactory.NewFilter([]*event.FilterData{fd0, fd1})
	this.NoError(err)
	arr := []FilterableObject{}
	for _, o := range this.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	this.Equal(arr, this.objects[6:38])
}

func (this *FilterSuite) Test_FilterIntegersGTEQ() {
	fd := &event.FilterData{"integer", ">=", "77"}
	filter, err := this.filterFactory.NewFilter([]*event.FilterData{fd})
	this.NoError(err)
	arr := []FilterableObject{}
	for _, o := range this.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	this.Equal(arr, this.objects[77:])
}

func (this *FilterSuite) Test_FilterIntegersNEQ() {
	fd := &event.FilterData{"integer", "!=", "50"}
	filter, err := this.filterFactory.NewFilter([]*event.FilterData{fd})
	this.NoError(err)
	arr := []FilterableObject{}
	for _, o := range this.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	res := make([]FilterableObject, OBJECTS)
	copy(res, this.objects)
	res = append(res[:50], res[51:]...)
	this.Equal(arr, res)
}

func (this *FilterSuite) Test_FilterStringEquals() {
	fd := &event.FilterData{"string", "==", "string7"}
	filter, err := this.filterFactory.NewFilter([]*event.FilterData{fd})
	this.NoError(err)
	arr := []FilterableObject{}
	for _, o := range this.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	this.Equal(arr, this.objects[7:8])
}

func (this *FilterSuite) Test_FilterStringNEQ() {
	fd := &event.FilterData{"string", "!=", "string50"}
	filter, err := this.filterFactory.NewFilter([]*event.FilterData{fd})
	this.NoError(err)
	arr := []FilterableObject{}

	for _, o := range this.objects {
		if filter.Match(o) {
			arr = append(arr, o)
		}
	}
	res := make([]FilterableObject, OBJECTS)
	copy(res, this.objects)
	res = append(res[:50], res[51:]...)
	this.Equal(arr, res)
}

// ********************************************* Entrypoint *********************************************

func TestFilterSuite(t *testing.T) {
	suite.Run(t, &FilterSuite{})
}
