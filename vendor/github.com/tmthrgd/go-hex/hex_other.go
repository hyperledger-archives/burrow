// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !amd64 gccgo appengine

package hex

// RawEncode encodes src into EncodedLen(len(src))
// bytes of dst.  As a convenience, it returns the number
// of bytes written to dst, but this value is always EncodedLen(len(src)).
// RawEncode implements hexadecimal encoding for a given alphabet.
func RawEncode(dst, src, alpha []byte) int {
	if len(alpha) != 16 {
		panic("invalid alphabet")
	}

	for i, v := range src {
		dst[i*2] = alpha[v>>4]
		dst[i*2+1] = alpha[v&0x0f]
	}

	return len(src) * 2
}

// Decode decodes src into DecodedLen(len(src)) bytes, returning the actual
// number of bytes written to dst.
//
// If Decode encounters invalid input, it returns an error describing the failure.
func Decode(dst, src []byte) (int, error) {
	if len(src)%2 == 1 {
		return 0, errLength
	}

	for i := 0; i < len(src)/2; i++ {
		a, ok := fromHexChar(src[i*2])
		if !ok {
			return 0, InvalidByteError(src[i*2])
		}

		b, ok := fromHexChar(src[i*2+1])
		if !ok {
			return 0, InvalidByteError(src[i*2+1])
		}

		dst[i] = (a << 4) | b
	}

	return len(src) / 2, nil
}

// fromHexChar converts a hex character into its value and a success flag.
func fromHexChar(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}

	return 0, false
}
