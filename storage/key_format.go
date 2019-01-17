package storage

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	DelimiterSegmentLength = -1
	VariadicSegmentLength  = 0
)

type ByteSlicable interface {
	Bytes() []byte
}

func NewMustKeyFormat(prefix string, layout ...int) *MustKeyFormat {
	kf, err := NewKeyFormat(prefix, layout...)
	if err != nil {
		panic(err)
	}
	return &MustKeyFormat{
		KeyFormat: *kf,
	}
}

// Provides a fixed-width lexicographically sortable []byte key format
type KeyFormat struct {
	prefix Prefix
	layout []int
	length int
}

// Create a []byte key format based on a single byte prefix and fixed width key segments each of whose length is
// specified by by the corresponding element of layout. A final segment length of 0 can be used to indicate a variadic
// final element that may be of arbitrary length.
//
// For example, to store keys that could index some objects by a version number and their SHA256 hash using the form:
// 'c<version uint64><hash [32]byte>' then you would define the KeyFormat with:
//
//  var keyFormat = NewKeyFormat('c', 8, 32)
//
// Then you can create a key with:
//
//  func ObjectKey(version uint64, objectBytes []byte) []byte {
//  	hasher := sha256.New()
//  	hasher.Sum(nil)
//  	return keyFormat.Key(version, hasher.Sum(nil))
//  }}
func NewKeyFormat(prefix string, layout ...int) (*KeyFormat, error) {
	kf := &KeyFormat{
		prefix: Prefix(prefix),
		layout: layout,
	}
	err := kf.init()
	if err != nil {
		return nil, err
	}
	return kf, nil
}

// Format the byte segments into the key format - will panic if the segment lengths do not match the layout.
func (kf *KeyFormat) KeyBytes(segments ...[]byte) ([]byte, error) {
	key := make([]byte, kf.length)
	n := copy(key, kf.prefix)
	var offset int
	for i, l := range kf.layout {
		si := i + offset
		if len(segments) <= si {
			break
		}
		s := segments[si]
		switch l {
		case VariadicSegmentLength:
			// Must be a final variadic element
			key = append(key, s...)
			n += len(s)
		case DelimiterSegmentLength:
			// ignore
			offset--
		default:
			if len(s) != l {
				return nil, fmt.Errorf("the segment '0x%X' provided to KeyFormat.KeyBytes() does not have required "+
					"%d bytes required by layout for segment %d", s, l, i)
			}
			n += l
			// Big endian so pad on left if not given the full width for this segment
			copy(key[n-len(s):n], s)
		}
	}
	return key[:n], nil
}

// Format the args passed into the key format - will panic if the arguments passed do not match the length
// of the segment to which they correspond. When called with no arguments returns the raw prefix (useful as a start
// element of the entire keys space when sorted lexicographically).
func (kf *KeyFormat) Key(args ...interface{}) ([]byte, error) {
	if len(args) > len(kf.layout) {
		return nil, fmt.Errorf("KeyFormat.Key() is provided with %d args but format only has %d segments",
			len(args), len(kf.layout))
	}
	segments := make([][]byte, len(args))
	for i, a := range args {
		segments[i] = format(a)
	}
	return kf.KeyBytes(segments...)
}

// Reads out the bytes associated with each segment of the key format from key.
func (kf *KeyFormat) ScanBytes(key []byte) [][]byte {
	segments := make([][]byte, len(kf.layout))
	n := kf.prefix.Length()
	for i, l := range kf.layout {
		if l == 0 {
			// Must be final variadic segment
			segments[i] = key[n:]
			return segments
		}
		n += l
		if n > len(key) {
			return segments[:i]
		}
		segments[i] = key[n-l : n]
	}
	return segments
}

// Extracts the segments into the values pointed to by each of args. Each arg must be a pointer to int64, uint64, or
// []byte, and the width of the args must match layout.
func (kf *KeyFormat) Scan(key []byte, args ...interface{}) error {
	segments := kf.ScanBytes(key)
	if len(args) > len(segments) {
		return fmt.Errorf("KeyFormat.Scan() is provided with %d args but format only has %d segments in key %X",
			len(args), len(segments), key)
	}
	for i, a := range args {
		scan(a, segments[i])
	}
	return nil
}

// Return the Key as a prefix - may just be the literal prefix, or an entire key
func (kf *KeyFormat) Prefix() Prefix {
	return kf.prefix
}

// Like Prefix but removes the prefix string
func (kf *KeyFormat) KeyNoPrefix(args ...interface{}) (Prefix, error) {
	key, err := kf.Key(args...)
	if err != nil {
		return nil, err
	}
	return key[len(kf.prefix):], nil
}

// Fixes the first args many segments as the prefix of a new KeyFormat by using the args to generate a key that becomes
// that prefix. Any remaining unassigned segments become the layout of the new KeyFormat.
func (kf *KeyFormat) Fix(args ...interface{}) (*KeyFormat, error) {
	key, err := kf.Key(args...)
	if err != nil {
		return nil, err
	}
	return NewKeyFormat(string(key), kf.layout[len(args):]...)
}

// Returns an iterator over the underlying iterable using this KeyFormat's prefix. This is to support proper iteration over the
// prefix in the presence of nil start or end which requests iteration to the inclusive edges of the domain. An optional
// argument for reverse can be passed to get reverse iteration.
func (kf *KeyFormat) Iterator(iterable KVIterable, start, end []byte, reverse ...bool) KVIterator {
	if len(reverse) > 0 && reverse[0] {
		return kf.prefix.Iterator(iterable.ReverseIterator, start, end)
	}
	return kf.prefix.Iterator(iterable.Iterator, start, end)
}

func (kf *KeyFormat) Unprefixed() *KeyFormat {
	return &KeyFormat{
		prefix: []byte{},
		layout: kf.layout[:len(kf.layout):len(kf.layout)],
		length: kf.length - len(kf.prefix),
	}
}

func (kf *KeyFormat) NumSegments() int {
	return len(kf.layout)
}

func (kf *KeyFormat) Layout() []int {
	l := make([]int, len(kf.layout))
	copy(l, kf.layout)
	return l
}

func (kf *KeyFormat) String() string {
	ls := make([]string, len(kf.layout))
	for i, l := range kf.layout {
		ls[i] = fmt.Sprintf("[%d]byte", l)
	}
	return fmt.Sprintf("KeyFormat{0x%s|%s}", kf.prefix.HexString(), strings.Join(ls, "|"))
}

func (kf *KeyFormat) init() error {
	kf.length = kf.prefix.Length()
	for i, l := range kf.layout {
		switch l {
		case VariadicSegmentLength:
			if i != len(kf.layout)-1 {
				return fmt.Errorf("KeyFormat may only have a 0 in the last place of its layout to indicate a " +
					"variadic segment")
			}
		case DelimiterSegmentLength:
			// ignore
		default:
			if l < 0 {
				panic(fmt.Errorf("KeyFormat layout must contain non-negative integers"))
			}
			kf.length += int(l)
		}
	}
	return nil
}

func scan(a interface{}, value []byte) {
	switch v := a.(type) {
	case *int64:
		// Negative values will be mapped correctly when read in as uint64 and then type converted
		*v = int64(binary.BigEndian.Uint64(value))
	case *uint64:
		*v = binary.BigEndian.Uint64(value)
	case *[]byte:
		*v = value
	case *string:
		*v = string(value)
	default:
		panic(fmt.Errorf("KeyFormat scan() does not support scanning value of type %T: %v", a, a))
	}
}

func format(a interface{}) []byte {
	switch v := a.(type) {
	case uint64:
		return formatUint64(v)
	case int64:
		return formatUint64(uint64(v))
	// Provide formatting from int,uint as a convenience to avoid casting arguments
	case uint:
		return formatUint64(uint64(v))
	case int:
		return formatUint64(uint64(v))
	case []byte:
		return v
	case ByteSlicable:
		return v.Bytes()
	case string:
		return []byte(v)
	default:
		panic(fmt.Errorf("KeyFormat format() does not support formatting value of type %T: %v", a, a))
	}
}

func formatUint64(v uint64) []byte {
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, v)
	return bs
}

// MustKeyFormat for panicking early when a KeyFormat does not parse
type MustKeyFormat struct {
	KeyFormat
}

func (kf *MustKeyFormat) KeyBytes(segments ...[]byte) []byte {
	key, err := kf.KeyFormat.KeyBytes(segments...)
	if err != nil {
		panic(err)
	}
	return key
}

func (kf *MustKeyFormat) Key(args ...interface{}) []byte {
	key, err := kf.KeyFormat.Key(args...)
	if err != nil {
		panic(err)
	}
	return key
}

func (kf *MustKeyFormat) Scan(key []byte, args ...interface{}) {
	err := kf.KeyFormat.Scan(key, args...)
	if err != nil {
		panic(err)
	}
}

func (kf *MustKeyFormat) Unprefixed() *MustKeyFormat {
	return &MustKeyFormat{*kf.KeyFormat.Unprefixed()}
}

func (kf *MustKeyFormat) Fix(args ...interface{}) *MustKeyFormat {
	fkf, err := kf.KeyFormat.Fix(args...)
	if err != nil {
		panic(err)
	}
	return &MustKeyFormat{*fkf}
}

func (kf *MustKeyFormat) KeyNoPrefix(args ...interface{}) Prefix {
	prefix, err := kf.KeyFormat.KeyNoPrefix(args...)
	if err != nil {
		panic(err)
	}
	return prefix
}
