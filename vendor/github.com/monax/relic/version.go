package relic

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// Base of minor, major, and patch version numbers
	numberBase = 10
	// Number of bits to represent version numbers
	uintBits = 8
)

type Version struct {
	major uint8
	minor uint8
	patch uint8
}

func (v Version) Major() uint8 {
	return v.major
}

func (v Version) Minor() uint8 {
	return v.minor
}

func (v Version) Patch() uint8 {
	return v.patch
}

func (v Version) String() string {
	return fmt.Sprintf("%v.%v.%v", v.major, v.minor, v.patch)
}

func AsVersion(versionLike interface{}) (Version, error) {
	switch v := versionLike.(type) {
	case Version:
		return v, nil
	case string:
		return ParseVersion(v)
	default:
		return Version{}, fmt.Errorf("unsupported type for version: %t, must be Version or string", v)
	}
}

func ParseVersion(versionString string) (Version, error) {
	parts := strings.Split(versionString, ".")
	if len(parts) != 3 {
		return Version{},
			fmt.Errorf("version string must have three '.' separated parts but '%s' does not", versionString)
	}
	maj, err := strconv.ParseUint(parts[0], numberBase, uintBits)
	if err != nil {
		return Version{}, err
	}
	min, err := strconv.ParseUint(parts[1], numberBase, uintBits)
	if err != nil {
		return Version{}, err
	}
	pat, err := strconv.ParseUint(parts[2], numberBase, uintBits)
	if err != nil {
		return Version{}, err
	}
	return Version{
		major: uint8(maj),
		minor: uint8(min),
		patch: uint8(pat),
	}, err
}
