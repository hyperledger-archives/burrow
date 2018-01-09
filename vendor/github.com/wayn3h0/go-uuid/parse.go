package uuid

import (
	"encoding/hex"
	"errors"
	"fmt"
)

// Parse parses the UUID string.
func Parse(str string) (UUID, error) {
	length := len(str)
	buffer := make([]byte, 16)
	charIndexes := []int{}
	switch length {
	case 36:
		if str[8] != '-' || str[13] != '-' || str[18] != '-' || str[23] != '-' {
			return Nil, fmt.Errorf("uuid: format of UUID string \"%s\" is invalid, it should be xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (8-4-4-4-12)", str)
		}
		charIndexes = []int{0, 2, 4, 6, 9, 11, 14, 16, 19, 21, 24, 26, 28, 30, 32, 34}
	case 32:
		charIndexes = []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30}
	default:
		return Nil, fmt.Errorf("uuid: length of UUID string \"%s\" is invalid, it should be 36 (standard) or 32 (without dash)", str)
	}
	for i, v := range charIndexes {
		if c, e := hex.DecodeString(str[v : v+2]); e != nil {
			return Nil, fmt.Errorf("uuid: UUID string \"%s\" is invalid: %s", str, e.Error())
		} else {
			buffer[i] = c[0]
		}
	}

	uuid := UUID{}
	copy(uuid[:], buffer)

	if !uuid.Equal(Nil) {
		if uuid.Layout() == LayoutInvalid {
			return Nil, errors.New("uuid: layout is invalid")
		}

		if uuid.Version() == VersionUnknown {
			return Nil, errors.New("uuid: version is unknown")
		}
	}

	return uuid, nil
}
