package relic

import (
	"fmt"
	"time"
)

var DefaultDateLayout = "2006-01-02"

func AsString(stringLike interface{}) (string, error) {
	switch s := stringLike.(type) {
	case string:
		return s, nil
	case fmt.Stringer:
		return s.String(), nil
	default:
		return "", fmt.Errorf("unsupported type for string: %t, must string or fmt.Stringer", s)
	}
}

func AsDate(dateLike interface{}) (time.Time, error) {
	switch d := dateLike.(type) {
	case time.Time:
		return d, nil
	default:
		s, err := AsString(d)
		if err != nil {
			return time.Time{}, err
		}
		return time.Parse(DefaultDateLayout, s)
	}
}
func AsVersion(versionLike interface{}) (Version, error) {
	switch v := versionLike.(type) {
	case Version:
		return v, nil
	default:
		s, err := AsString(v)
		if err != nil {
			return Version{}, err
		}
		if s == "" {
			// for unreleased notes
			return Version{}, nil
		}
		return ParseDatedVersion(s)
	}
}
