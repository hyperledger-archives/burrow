package types

import (
	"errors"
	"strings"
)

// defined event input types
const (
	EventInputTypeInt     = "int"
	EventInputTypeUInt    = "uint"
	EventInputTypeAddress = "address"
	EventInputTypeBytes   = "bytes"
	EventInputTypeBool    = "bool"
	EventInputTypeString  = "string"
)

// IsValidEventInputType checks if the event input type is a valid one
func IsValidEventInputType(value interface{}) error {
	input, _ := value.(string)
	val := strings.ToLower(input)

	if strings.HasPrefix(val, EventInputTypeInt) ||
		strings.HasPrefix(val, EventInputTypeUInt) ||
		strings.HasPrefix(val, EventInputTypeBytes) ||
		val == EventInputTypeAddress ||
		val == EventInputTypeBool ||
		val == EventInputTypeString {
		return nil
	}

	return errors.New("invalid event input type")
}
