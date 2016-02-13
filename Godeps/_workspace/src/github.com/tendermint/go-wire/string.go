package wire

import (
	"io"

	. "github.com/tendermint/go-common"
)

// String

func WriteString(s string, w io.Writer, n *int, err *error) {
	WriteVarint(len(s), w, n, err)
	WriteTo([]byte(s), w, n, err)
}

func ReadString(r io.Reader, lmt int, n *int, err *error) string {
	length := ReadVarint(r, n, err)
	if *err != nil {
		return ""
	}
	if length < 0 {
		*err = ErrBinaryReadSizeUnderflow
		return ""
	}
	if lmt != 0 && lmt < MaxInt(length, *n+length) {
		*err = ErrBinaryReadSizeOverflow
		return ""
	}

	buf := make([]byte, length)
	ReadFull(buf, r, n, err)
	return string(buf)
}
