package filters

import (
	"fmt"
	"strings"
	"sync"
)

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

func (ff *FilterFactory) RegisterFilterPool(fieldName string, pool *sync.Pool) {
	ff.filterPools[strings.ToLower(fieldName)] = pool
}

// Creates a new filter given the input data array. If the array is zero length or nil, an empty
// filter will be returned that returns true on all matches. If the array is of size 1, a regular
// filter is returned, otherwise a CompositeFieldFilter is returned, which is a special filter that
// contains a number of other filters. It implements AccountFieldFilter, and will match an account
// only if all the sub-filters matches.
func (ff *FilterFactory) NewFilter(fdArr []*FilterData) (Filter, error) {

	if len(fdArr) == 0 {
		return &MatchAllFilter{}, nil
	}
	if len(fdArr) == 1 {
		return ff.newSingleFilter(fdArr[0])
	}
	filters := []Filter{}
	for _, fd := range fdArr {
		f, err := ff.newSingleFilter(fd)
		if err != nil {
			return nil, err
		}
		filters = append(filters, f)
	}
	cf := ff.compositeFilterPool.Get().(*CompositeFilter)
	cf.filters = filters
	return cf, nil
}

func (ff *FilterFactory) newSingleFilter(fd *FilterData) (ConfigurableFilter, error) {
	fp, ok := ff.filterPools[strings.ToLower(fd.Field)]
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
