package pipe

import (
	"fmt"
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

// Parses a range param.
func RangeToNumbers() (int64, int64, error){
	return 1, 1, nil
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

func (this *FilterFactory) RegisterFilterPool(fieldName string, pool *sync.Pool){
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