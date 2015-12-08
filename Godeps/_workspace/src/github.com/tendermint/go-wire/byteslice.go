package wire

import (
	"io"

	. "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/go-common"
)

func WriteByteSlice(bz []byte, w io.Writer, n *int, err *error) {
	WriteVarint(len(bz), w, n, err)
	WriteTo(bz, w, n, err)
}

func ReadByteSlice(r io.Reader, lmt int, n *int, err *error) []byte {
	length := ReadVarint(r, n, err)
	if *err != nil {
		return nil
	}
	if length < 0 {
		*err = ErrBinaryReadSizeUnderflow
		return nil
	}
	if lmt != 0 && lmt < MaxInt(length, *n+length) {
		*err = ErrBinaryReadSizeOverflow
		return nil
	}

	buf := make([]byte, length)
	ReadFull(buf, r, n, err)
	return buf
}

//-----------------------------------------------------------------------------

func WriteByteSlices(bzz [][]byte, w io.Writer, n *int, err *error) {
	WriteVarint(len(bzz), w, n, err)
	for _, bz := range bzz {
		WriteByteSlice(bz, w, n, err)
		if *err != nil {
			return
		}
	}
}

func ReadByteSlices(r io.Reader, lmt int, n *int, err *error) [][]byte {
	length := ReadVarint(r, n, err)
	if *err != nil {
		return nil
	}
	if length < 0 {
		*err = ErrBinaryReadSizeUnderflow
		return nil
	}
	if lmt != 0 && lmt < MaxInt(length, *n+length) {
		*err = ErrBinaryReadSizeOverflow
		return nil
	}

	bzz := make([][]byte, length)
	for i := 0; i < length; i++ {
		bz := ReadByteSlice(r, lmt, n, err)
		if *err != nil {
			return nil
		}
		bzz[i] = bz
	}
	return bzz
}
