// Copyright 2015-2017 Monax Industries Limited.
// This file is part of the Monax platform (Monax)

// Monax is free software: you can use, redistribute it and/or modify
// it only under the terms of the GNU General Public License, version
// 3, as published by the Free Software Foundation.

// Monax is distributed WITHOUT ANY WARRANTY pursuant to
// the terms of the Gnu General Public Licence, version 3, including
// (but not limited to) Clause 15 thereof. See the text of the
// GNU General Public License, version 3 for full terms.

// You should have received a copy of the GNU General Public License,
// version 3, with Monax.  If not, see <http://www.gnu.org/licenses/>.

// TODO: [ben] this is not specific for word256, but it is used exclusively
// for word256 (and for word160* 20byte addresses) so consider stronger typing.

package word256

import (
	"bytes"
)

func Fingerprint(slice []byte) []byte {
	fingerprint := make([]byte, 6)
	copy(fingerprint, slice)
	return fingerprint
}

func IsZeros(slice []byte) bool {
	for _, byt := range slice {
		if byt != byte(0) {
			return false
		}
	}
	return true
}

func RightPadBytes(slice []byte, l int) []byte {
	if l < len(slice) {
		return slice
	}
	padded := make([]byte, l)
	copy(padded[0:len(slice)], slice)
	return padded
}

func LeftPadBytes(slice []byte, l int) []byte {
	if l < len(slice) {
		return slice
	}
	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)
	return padded
}

func TrimmedString(b []byte) string {
	trimSet := string([]byte{0})
	return string(bytes.TrimLeft(b, trimSet))

}
