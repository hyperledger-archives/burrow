//
// See https://eth.wiki/fundamentals/rlp
//
package rlp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/bits"
	"reflect"

	binary2 "github.com/hyperledger/burrow/binary"
)

type magicOffset uint8

const (
	ShortLength              = 55
	StringOffset magicOffset = 0x80 // 128 - if string length is less than or equal to 55 [inclusive]
	SliceOffset  magicOffset = 0xC0 // 192 - if slice length is less than or equal to 55 [inclusive]
	SmallByte                = 0x7f // 247 - value less than or equal is itself [inclusive
)

type Code uint32

const (
	ErrUnknown Code = iota
	ErrNoInput
	ErrInvalid
)

var bigIntType = reflect.TypeOf(&big.Int{})

func (c Code) Error() string {
	switch c {
	case ErrNoInput:
		return "no input"
	case ErrInvalid:
		return "input not valid RLP encoding"
	default:
		return "unknown error"
	}
}

func Encode(input interface{}) ([]byte, error) {
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return encode(val)
}

func Decode(src []byte, dst interface{}) error {
	fields, err := decode(src)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(dst)
	typ := reflect.TypeOf(dst)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Slice:
		switch typ.Elem().Kind() {
		case reflect.Uint8:
			out, ok := dst.([]byte)
			if !ok {
				return fmt.Errorf("cannot decode into type %s", val.Type())
			}
			found := bytes.Join(fields, []byte(""))
			if len(out) < len(found) {
				return fmt.Errorf("cannot decode %d bytes into slice of size %d", len(found), len(out))
			}
			for i, b := range found {
				out[i] = b
			}
		default:
			for i := 0; i < val.Len(); i++ {
				elem := val.Index(i)
				err = decodeField(elem, fields[i])
				if err != nil {
					return err
				}
			}
		}
	case reflect.Struct:
		if val.NumField() != len(fields) {
			return fmt.Errorf("wrong number of fields; have %d, want %d", len(fields), val.NumField())
		}
		for i := 0; i < val.NumField(); i++ {
			err := decodeField(val.Field(i), fields[i])
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("cannot decode into unsupported type %v", reflect.TypeOf(dst))
	}
	return nil
}

func encodeUint8(input uint8) ([]byte, error) {
	if input == 0 {
		// yes this makes no sense, but it does seem to be what everyone else does, apparently 'no leading zeroes'.
		// It means we cannot store []byte{0} because that is indistinguishable from byte{}
		return []byte{uint8(StringOffset)}, nil
	} else if input <= SmallByte {
		return []byte{input}, nil
	} else if input >= uint8(StringOffset) {
		return []byte{0x81, input}, nil
	}
	return []byte{uint8(StringOffset)}, nil
}

func encodeUint64(i uint64) ([]byte, error) {
	size := bits.Len64(i)/8 + 1
	if size == 1 {
		return encodeUint8(uint8(i))
	}
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i))
	return encodeString(b[8-size:])
}

func encodeBigInt(b *big.Int) ([]byte, error) {
	if b.Sign() == -1 {
		return nil, fmt.Errorf("cannot RLP encode negative number")
	}
	if b.IsUint64() {
		return encodeUint64(b.Uint64())
	}
	bs := b.Bytes()
	length := encodeLength(len(bs), StringOffset)
	return append(length, bs...), nil
}

func encodeLength(n int, offset magicOffset) []byte {
	// > if a string is 0-55 bytes long, the RLP encoding consists of a single byte with value 0x80 plus
	// > the length of the string followed by the string.
	if n <= ShortLength {
		return []uint8{uint8(offset) + uint8(n)}
	}

	i := uint64(n)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	byteLengthOfLength := bits.Len64(i)/8 + 1
	// > If a string is more than 55 bytes long, the RLP encoding consists of a single byte with value 0xb7
	// > plus the length in bytes of the length of the string in binary form, followed by the length of the string,
	// > followed by the string
	return append([]byte{uint8(offset) + ShortLength + uint8(byteLengthOfLength)}, b[8-byteLengthOfLength:]...)
}

func encodeString(input []byte) ([]byte, error) {
	if len(input) == 1 && input[0] <= SmallByte {
		return encodeUint8(input[0])
	} else {
		return append(encodeLength(len(input), StringOffset), input...), nil
	}
}

func encodeList(val reflect.Value) ([]byte, error) {
	if val.Len() == 0 {
		return []byte{uint8(SliceOffset)}, nil
	}

	out := make([][]byte, 0)
	for i := 0; i < val.Len(); i++ {
		data, err := encode(val.Index(i))
		if err != nil {
			return nil, err
		}
		out = append(out, data)
	}

	sum := bytes.Join(out, []byte{})
	return append(encodeLength(len(sum), SliceOffset), sum...), nil
}

func encodeStruct(val reflect.Value) ([]byte, error) {
	out := make([][]byte, 0)

	for i := 0; i < val.NumField(); i++ {
		data, err := encode(val.Field(i))
		if err != nil {
			return nil, err
		}
		out = append(out, data)
	}
	sum := bytes.Join(out, []byte{})
	length := encodeLength(len(sum), SliceOffset)
	return append(length, sum...), nil
}

func encode(val reflect.Value) ([]byte, error) {
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Ptr:
		if !val.Type().AssignableTo(bigIntType) {
			return nil, fmt.Errorf("cannot encode pointer type %v", val.Type())
		}
		return encodeBigInt(val.Interface().(*big.Int))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i := val.Int()
		if i < 0 {
			return nil, fmt.Errorf("cannot rlp encode negative integer")
		}
		return encodeUint64(uint64(i))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return encodeUint64(val.Uint())
	case reflect.Bool:
		if val.Bool() {
			return []byte{0x01}, nil
		}
		return []byte{uint8(StringOffset)}, nil
	case reflect.String:
		return encodeString([]byte(val.String()))
	case reflect.Slice:
		switch val.Type().Elem().Kind() {
		case reflect.Uint8:
			i, err := encodeString(val.Bytes())
			return i, err
		default:
			return encodeList(val)
		}
	case reflect.Struct:
		return encodeStruct(val)
	default:
		return []byte{uint8(StringOffset)}, nil
	}
}

// Split into RLP fields by reading length prefixes and consuming chunks
func decode(in []byte) ([][]byte, error) {
	if len(in) == 0 {
		return nil, nil
	}

	offset, length, typ := decodeLength(in)
	end := offset + length

	if end > uint64(len(in)) {
		return nil, fmt.Errorf("read length prefix of %d but there is only %d bytes of unconsumed input",
			length, uint64(len(in))-offset)
	}

	suffix, err := decode(in[end:])
	if err != nil {
		return nil, err
	}
	switch typ {
	case reflect.String:
		return append([][]byte{in[offset:end]}, suffix...), nil
	case reflect.Slice:
		prefix, err := decode(in[offset:end])
		if err != nil {
			return nil, err
		}
		return append(prefix, suffix...), nil
	}

	return suffix, nil
}

func decodeLength(input []byte) (uint64, uint64, reflect.Kind) {
	magicByte := magicOffset(input[0])

	switch {
	case magicByte <= SmallByte:
		// small byte: sufficiently small single byte
		return 0, 1, reflect.String

	case magicByte <= StringOffset+ShortLength:
		// short string: length less than or equal to 55 bytes
		length := uint64(magicByte - StringOffset)
		return 1, length, reflect.String

	case magicByte < SliceOffset:
		// long string: length described by magic = 0xb7 + <byte length of length of string>
		byteLengthOfLength := magicByte - StringOffset - ShortLength
		length := getUint64(input[1:byteLengthOfLength])
		offset := uint64(byteLengthOfLength + 1)
		return offset, length, reflect.String

	case magicByte <= SliceOffset+ShortLength:
		// short slice: length less than or equal to 55 bytes
		length := uint64(magicByte - SliceOffset)
		return 1, length, reflect.Slice

	// Note this takes us all the way up to <= 255 so this switch is exhaustive
	default:
		// long string: length described by magic = 0xf7 + <byte length of length of string>
		byteLengthOfLength := magicByte - SliceOffset - ShortLength
		length := getUint64(input[1:byteLengthOfLength])
		offset := uint64(byteLengthOfLength + 1)
		return offset, length, reflect.Slice
	}
}

func getUint64(bs []byte) uint64 {
	bs = binary2.LeftPadBytes(bs, 8)
	return binary.BigEndian.Uint64(bs)
}

func decodeField(val reflect.Value, field []byte) error {
	typ := val.Type()

	switch val.Kind() {
	case reflect.Ptr:
		if !typ.AssignableTo(bigIntType) {
			return fmt.Errorf("cannot decode into pointer type %v", typ)
		}
		bi := new(big.Int).SetBytes(field)
		val.Set(reflect.ValueOf(bi))

	case reflect.String:
		val.SetString(string(field))
	case reflect.Uint64:
		out := make([]byte, 8)
		for j := range field {
			out[len(out)-(len(field)-j)] = field[j]
		}
		val.SetUint(binary.BigEndian.Uint64(out))
	case reflect.Slice:
		if typ.Elem().Kind() != reflect.Uint8 {
			// skip
			return nil
		}
		out := make([]byte, len(field))
		for i, b := range field {
			out[i] = b
		}
		val.SetBytes(out)
	}
	return nil
}
