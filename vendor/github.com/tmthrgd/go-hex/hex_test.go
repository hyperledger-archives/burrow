// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

package hex

import (
	ref "encoding/hex"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

func testEncode(t *testing.T, enc, ref func([]byte) string, scale float64, maxsize int) {
	if err := quick.CheckEqual(ref, enc, &quick.Config{
		Values: func(args []reflect.Value, rand *rand.Rand) {
			off := rand.Intn(32)

			data := make([]byte, 1+rand.Intn(maxsize)+off)
			rand.Read(data[off:])
			args[0] = reflect.ValueOf(data[off:])
		},

		MaxCountScale: scale,
	}); err != nil {
		t.Error(err)
	}
}

func TestEncodeShort(t *testing.T) {
	testEncode(t, EncodeToString, ref.EncodeToString, 1000, 15)
}

func TestEncode(t *testing.T) {
	testEncode(t, EncodeToString, ref.EncodeToString, 1.5, 1024*1024)
}

func TestEncodeUpper(t *testing.T) {
	testEncode(t, EncodeUpperToString, func(src []byte) string {
		return strings.ToUpper(ref.EncodeToString(src))
	}, 0.375, 1024*1024)
}

func testDecode(t *testing.T, enc func([]byte) string, scale float64, maxsize int) {
	if err := quick.CheckEqual(func(s string) (string, error) {
		return s, nil
	}, func(s string) (string, error) {
		b, err := DecodeString(s)
		return enc(b), err
	}, &quick.Config{
		Values: func(args []reflect.Value, rand *rand.Rand) {
			off := rand.Intn(32)

			src := make([]byte, 1+rand.Intn(maxsize)+off)
			rand.Read(src)
			data := enc(src)
			args[0] = reflect.ValueOf(data[2*off:])
		},

		MaxCountScale: scale,
	}); err != nil {
		t.Error(err)
	}
}

func TestDecodeShort(t *testing.T) {
	testDecode(t, EncodeToString, 1000, 7)
}

func TestDecode(t *testing.T) {
	testDecode(t, EncodeToString, 2, 1024*1024)
}

func TestDecodeUpper(t *testing.T) {
	testDecode(t, EncodeUpperToString, 2, 1024*1024)
}

func TestDecodeInvalid(t *testing.T) {
	src := make([]byte, 19)
	rand.Read(src)

	dst := make([]byte, EncodedLen(len(src)))
	Encode(dst, src)

	tmp := make([]byte, len(src))

	for pos := 0; pos < len(dst); pos++ {
		old := dst[pos]

		for c := rune(0); c < rune(0x100); c++ {
			dst[pos] = byte(c)

			_, err := Decode(tmp, dst)
			if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f') {
				if err != nil {
					t.Errorf("unexpected error for %d:%#U: %v", pos, c, err)
				}
			} else if err, ok := err.(InvalidByteError); ok {
				if byte(err) != byte(c) {
					t.Errorf("expected error for %d:%#U, got %v", pos, c, err)
				}
			} else {
				t.Errorf("expected error for %d:%#U, got %v", pos, c, err)
			}
		}

		dst[pos] = old
	}
}

func catchPanic(fn func()) (e interface{}) {
	defer func() { e = recover() }()
	fn()
	return nil
}

func TestDstTooShort(t *testing.T) {
	dst := make([]byte, 10)
	src := make([]byte, 100)

	for i := range src {
		src[i] = '0'
	}

	if catchPanic(func() {
		Encode(dst, src)
	}) == nil {
		t.Fatal("did not catch encode into small dst buffer")
	}

	if catchPanic(func() {
		if _, err := Decode(dst, src); err != nil {
			t.Fatal(err)
		}
	}) == nil {
		t.Fatal("did not catch decode into small dst buffer")
	}
}

func TestInvalidAlphabetSize(t *testing.T) {
	if err := catchPanic(func() {
		RawEncode(nil, nil, nil)
	}); err != "invalid alphabet" {
		t.Fatal("did not catch invalid alphabet size")
	}
}

type size struct {
	name string
	l    int
}

var encodeSize = []size{
	{"15", 15},
	{"32", 32},
	{"128", 128},
	{"1K", 1 * 1024},
	{"16K", 16 * 1024},
	{"128K", 128 * 1024},
	{"1M", 1024 * 1024},
	{"16M", 16 * 1024 * 1024},
	{"128M", 128 * 1024 * 1024},
}

func BenchmarkEncode(b *testing.B) {
	for _, size := range encodeSize {
		b.Run(size.name, func(b *testing.B) {
			src := make([]byte, size.l)
			rand.Read(src)

			dst := make([]byte, EncodedLen(size.l))

			b.SetBytes(int64(size.l))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				Encode(dst, src)
			}
		})
	}
}

func BenchmarkRefEncode(b *testing.B) {
	for _, size := range encodeSize {
		b.Run(size.name, func(b *testing.B) {
			src := make([]byte, size.l)
			rand.Read(src)

			dst := make([]byte, ref.EncodedLen(size.l))

			b.SetBytes(int64(size.l))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				ref.Encode(dst, src)
			}
		})
	}
}

var decodeSize = []size{
	{"14", 14},
	{"32", 32},
	{"128", 128},
	{"1K", 1 * 1024},
	{"16K", 16 * 1024},
	{"128K", 128 * 1024},
	{"1M", 1024 * 1024},
	{"16M", 16 * 1024 * 1024},
	{"128M", 128 * 1024 * 1024},
}

func BenchmarkDecode(b *testing.B) {
	for _, size := range decodeSize {
		b.Run(size.name, func(b *testing.B) {
			m := DecodedLen(size.l)

			src := make([]byte, size.l)
			rand.Read(src[:m])
			Encode(src, src[:m])

			dst := make([]byte, m)

			b.SetBytes(int64(size.l))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				Decode(dst, src)
			}
		})
	}
}

func benchmarkRefDecode(b *testing.B, l int) {
	for _, size := range decodeSize {
		b.Run(size.name, func(b *testing.B) {
			m := ref.DecodedLen(size.l)

			src := make([]byte, size.l)
			rand.Read(src[:m])
			Encode(src, src[:m])

			dst := make([]byte, m)

			b.SetBytes(int64(size.l))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				ref.Decode(dst, src)
			}
		})
	}
}
