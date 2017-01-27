package config

import (
	"fmt"
	"os"

	"regexp"

	"github.com/eapache/channels"
	"github.com/eris-ltd/eris-db/common/math/integral"
	"github.com/eris-ltd/eris-db/logging/config/types"
	"github.com/eris-ltd/eris-db/logging/loggers"
	kitlog "github.com/go-kit/kit/log"
)

func BuildFilterPredicate(filterConfig *types.FilterConfig) (func([]interface{}) bool, error) {
	includePredicate, err := BuildKeyValuesPredicate(filterConfig.Include, true)
	if err != nil {
		return nil, err
	}
	excludePredicate, err := BuildKeyValuesPredicate(filterConfig.Exclude, false)
	if err != nil {
		return nil, err
	}
	return func(keyvals []interface{}) bool {
		// When do we want to exclude a log line
		return !includePredicate(keyvals) || excludePredicate(keyvals)
	}, nil
}

func BuildKeyValuesPredicate(kvpConfigs []*types.KeyValuePredicateConfig,
	matchAll bool) (func([]interface{}) bool, error) {
	length := len(kvpConfigs)
	keyRegexes := make([]*regexp.Regexp, length)
	valueRegexes := make([]*regexp.Regexp, length)

	// Compile all KV regexes
	for i, kvpConfig := range kvpConfigs {
		// Store a nil regex to indicate no key match should occur
		if kvpConfig.KeyRegex != "" {
			keyRegex, err := regexp.Compile(kvpConfig.KeyRegex)
			if err != nil {
				return nil, err
			}
			keyRegexes[i] = keyRegex
		}
		// Store a nil regex to indicate no value match should occur
		if kvpConfig.ValueRegex != "" {
			valueRegex, err := regexp.Compile(kvpConfig.ValueRegex)
			if err != nil {
				return nil, err
			}
			valueRegexes[i] = valueRegex
		}
	}

	return func(keyvals []interface{}) bool {
		return matchLogLine(keyvals, keyRegexes, valueRegexes, matchAll)
	}, nil
}

// matchLogLine tries to match a log line by trying to match each key value pair with each pair of key value regexes
// if matchAll is true then matchLogLine returns true iff every key value regexes finds a match or the line or regexes
// are empty
func matchLogLine(keyvals []interface{}, keyRegexes, valueRegexes []*regexp.Regexp, matchAll bool) bool {
	all := matchAll
	// We should be passed an aligned list of keyRegexes and valueRegexes, but since we can't error here we'll guard
	// against a failure of the caller to pass valid arguments
	length := integral.MinInt(len(keyRegexes), len(valueRegexes))
	for i := 0; i < length; i++ {
		matched := findMatchInLogLine(keyvals, keyRegexes[i], valueRegexes[i])
		if matchAll {
			all = all && matched
		} else if matched {
			return true
		}
	}
	return all
}

func findMatchInLogLine(keyvals []interface{}, keyRegex, valueRegex *regexp.Regexp) bool {
	for i := 0; i < 2*(len(keyvals)/2); i += 2 {
		key := convertToString(keyvals[i])
		val := convertToString(keyvals[i+1])
		if key == nil && val == nil {
			continue
		}
		// At least one of key or val could be converted from string, only match on either if the conversion worked
		// Try to match on both key and value, falling back to a positive match if either or both or not supplied
		if match(keyRegex, key) && match(valueRegex, val) {
			return true
		}
	}
	return false
}

func match(regex *regexp.Regexp, text *string) bool {
	// Always match on a nil regex (indicating no constraint on text),
	// and otherwise never match on nil text (indicating a non-string convertible type)
	return regex == nil || (text != nil && regex.MatchString(*text))
}

func convertToString(value interface{}) *string {
	// We have the option of returning nil here to indicate a conversion was
	// not possible/does not make sense. Although we are not opting to use it
	// currently
	switch v := value.(type) {
	case string:
		return &v
	default:
		s := fmt.Sprintf("%v", v)
		return &s
	}
}
