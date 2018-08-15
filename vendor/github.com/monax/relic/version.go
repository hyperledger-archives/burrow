package relic

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	VersionDateSeparator = " - "
	// Base of minor, major, and patch version numbers
	NumberBase = 10
	// Number of bits to represent version numbers
	UintBits = 8
)

var ZeroVersion = Version{}

type Version struct {
	Major uint8
	Minor uint8
	Patch uint8
	Date  time.Time
}

func (v Version) String() string {
	if v == ZeroVersion {
		return "Unreleased"
	}
	return v.Semver()
}

func (v Version) Semver() string {
	return fmt.Sprintf("%v.%v.%v", v.Major, v.Minor, v.Patch)
}

func (v Version) Ref() string {
	if v == ZeroVersion {
		return "HEAD"
	}
	return fmt.Sprintf("v%s", v.Semver())
}

func (v Version) Dated() bool {
	return v.Date != time.Time{}
}

func (v Version) FormatDate() string {
	return v.Date.Format(DefaultDateLayout)
}

func ParseDatedVersion(versionString string) (Version, error) {
	parts := strings.Split(versionString, VersionDateSeparator)
	switch len(parts) {
	case 1:
		return ParseVersion(versionString)
	case 2:
		v, err := ParseVersion(parts[0])
		if err != nil {
			return Version{}, err
		}
		v.Date, err = AsDate(parts[1])
		if err != nil {
			return Version{}, err
		}
		return v, nil
	default:
		return Version{}, fmt.Errorf("could interpret %v as date version, should be be of the form '2.3.4 - 2018-08-14'",
			versionString)
	}
}

func ParseVersion(versionString string) (Version, error) {
	parts := strings.Split(versionString, ".")
	if len(parts) != 3 {
		return Version{},
			fmt.Errorf("version string must have three '.' separated parts but '%s' does not", versionString)
	}
	maj, err := strconv.ParseUint(parts[0], NumberBase, UintBits)
	if err != nil {
		return Version{}, err
	}
	min, err := strconv.ParseUint(parts[1], NumberBase, UintBits)
	if err != nil {
		return Version{}, err
	}
	pat, err := strconv.ParseUint(parts[2], NumberBase, UintBits)
	if err != nil {
		return Version{}, err
	}
	return Version{
		Major: uint8(maj),
		Minor: uint8(min),
		Patch: uint8(pat),
	}, err
}
