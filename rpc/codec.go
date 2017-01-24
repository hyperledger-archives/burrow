package rpc

import (
	"io"
)

// Used for rpc request and response data.
type Codec interface {
	EncodeBytes(interface{}) ([]byte, error)
	Encode(interface{}, io.Writer) error
	DecodeBytes(interface{}, []byte) error
	DecodeBytesPtr(interface{}, []byte) error
	Decode(interface{}, io.Reader) error
}
