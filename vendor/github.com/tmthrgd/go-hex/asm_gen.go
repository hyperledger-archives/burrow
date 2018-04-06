// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.
//
// Copyright 2005-2016, Wojciech Muła. All rights reserved.
// Use of this source code is governed by a
// Simplified BSD License license that can be found in
// the LICENSE file.

// +build ignore

package main

import (
	"bytes"
	"strconv"

	"github.com/tmthrgd/asm"
)

const header = `// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.
//
// Copyright 2005-2016, Wojciech Muła. All rights reserved.
// Use of this source code is governed by a
// Simplified BSD License license that can be found in
// the LICENSE file.
//
// This file is auto-generated - do not modify

// +build amd64,!gccgo,!appengine
`

func repeat(b byte, l int) []byte {
	return bytes.Repeat([]byte{b}, l)
}

type encode struct {
	*asm.Asm

	di, si, cx asm.Register

	ret asm.Label

	alpha asm.Operand

	mask asm.Data
}

func (e *encode) vpand_sse(ops ...asm.Operand) {
	if len(ops) != 3 {
		panic("wrong number of operands")
	}

	if ops[1] == ops[2] {
		panic("invalid register choice fallback")
	}

	if ops[0] != ops[1] {
		e.Movou(ops[0], ops[1])
	}

	e.Pand(ops[0], ops[2])
}

func (e *encode) vpunpckhbw_sse(ops ...asm.Operand) {
	if len(ops) != 3 {
		panic("wrong number of operands")
	}

	if ops[1] == ops[2] {
		panic("invalid register choice fallback")
	}

	if ops[0] != ops[1] {
		e.Movou(ops[0], ops[1])
	}

	e.Punpckhbw(ops[0], ops[2])
}

func (e *encode) vpshufb_sse(ops ...asm.Operand) {
	if len(ops) != 3 {
		panic("wrong number of operands")
	}

	if ops[1] == ops[2] {
		panic("invalid register choice fallback")
	}

	if ops[0] != ops[1] {
		e.Movou(ops[0], ops[1])
	}

	e.Pshufb(ops[0], ops[2])
}

func (e *encode) Convert(out2, out1, in, tmp asm.Operand, vpand, vpunpckhbw, vpshufb func(ops ...asm.Operand)) {
	vpand(out1, in, e.mask)

	e.Psrlw(in, asm.Constant(4))
	e.Pand(in, e.mask)

	if out2 != asm.Invalid {
		vpunpckhbw(tmp, in, out1)
	}
	e.Punpcklbw(in, out1)

	vpshufb(out1, e.alpha, in)
	if out2 != asm.Invalid {
		vpshufb(out2, e.alpha, tmp)
	}
}

func (e *encode) BigLoop(l asm.Label, vpand, vpunpckhbw, vpshufb func(ops ...asm.Operand)) {
	e.Label(l)

	e.Movou(asm.X0, asm.Address(e.si, e.cx, asm.SX1, -16))

	e.Convert(asm.X2, asm.X1, asm.X0, asm.X3, vpand, vpunpckhbw, vpshufb)

	for i, r := range []asm.Operand{asm.X2, asm.X1} {
		e.Movou(asm.Address(e.di, e.cx, asm.SX2, -16*(i+1)), r)
	}

	e.Subq(e.cx, asm.Constant(16))
	e.Jz(e.ret)

	e.Cmpq(asm.Constant(16), e.cx)
	e.Jae(l)
}

func encodeASM(a *asm.Asm) {
	mask := a.Data("encodeMask", repeat(0x0f, 16))

	a.NewFunction("encodeASM")
	a.NoSplit()

	dst := a.Argument("dst", 8)
	src := a.Argument("src", 8)
	length := a.Argument("len", 8)
	alpha := a.Argument("alpha", 8)

	a.Start()

	bigloop_avx := a.NewLabel("bigloop_avx")
	bigloop_sse := a.NewLabel("bigloop_sse")
	tail := a.NewLabel("tail")
	ret := a.NewLabel("ret")

	e := &encode{
		a,

		asm.DI, asm.SI, asm.BX,

		ret,

		asm.X15,

		mask,
	}

	a.Movq(e.di, dst)
	a.Movq(e.si, src)
	a.Movq(e.cx, length)
	a.Movq(asm.DX, alpha)

	a.Movou(e.alpha, asm.Address(asm.DX))

	a.Cmpq(asm.Constant(16), e.cx)
	a.Jb(tail)

	a.Cmpb(asm.Constant(1), asm.Data("runtime·support_avx"))
	a.Jne(bigloop_sse)

	e.BigLoop(bigloop_avx, a.Vpand, a.Vpunpckhbw, a.Vpshufb)

	a.Label(tail)

	tailIn := [9]asm.Label{tail.Suffix("in")}
	tailConv := tail.Suffix("conv")
	tailOut := [9]asm.Label{tail.Suffix("out")}

	for i := 1; i <= 8; i++ {
		tailIn[i] = tailIn[0].Suffix(strconv.Itoa(i))
		tailOut[i] = tailOut[0].Suffix(strconv.Itoa(i))
	}

	for i := 2; i < 8; i += 2 {
		a.Cmpq(asm.Constant(i), e.cx)
		a.Jb(tailIn[i-1])
		a.Je(tailIn[i])
	}

	a.Cmpq(asm.Constant(8), e.cx)
	a.Jb(tailIn[7])

	a.Label(tailIn[8])
	a.Movq(asm.X0, asm.Address(e.si))
	a.Jmp(tailConv)

	for i := 7; i >= 5; i-- {
		a.Label(tailIn[i])
		a.Pinsrb(asm.X0, asm.Address(e.si, i-1), asm.Constant(i-1))
	}

	a.Label(tailIn[4])
	a.Pinsrd(asm.X0, asm.Address(e.si), asm.Constant(0))
	a.Jmp(tailConv)

	for i := 3; i >= 1; i-- {
		a.Label(tailIn[i])
		a.Pinsrb(asm.X0, asm.Address(e.si, i-1), asm.Constant(i-1))
	}

	a.Label(tailConv)

	e.Convert(asm.Invalid, asm.X1, asm.X0, asm.X3, e.vpand_sse, e.vpunpckhbw_sse, e.vpshufb_sse)

	for i := 2; i < 8; i += 2 {
		a.Cmpq(asm.Constant(i), e.cx)
		a.Jb(tailOut[i-1])
		a.Je(tailOut[i])
	}

	a.Cmpq(asm.Constant(8), e.cx)
	a.Jb(tailOut[7])

	a.Label(tailOut[8])
	a.Movou(asm.Address(e.di), asm.X1)

	a.Subq(e.cx, asm.Constant(8))
	a.Jz(ret)

	a.Addq(e.si, asm.Constant(8))
	a.Addq(e.di, asm.Constant(16))

	a.Jmp(tail)

	for i := 7; i >= 5; i-- {
		a.Label(tailOut[i])
		a.Pextrw(asm.Address(e.di, (i-1)*2), asm.X1, asm.Constant(i-1))
	}

	a.Label(tailOut[4])
	a.Movq(asm.Address(e.di), asm.X1)
	a.Ret()

	for i := 3; i >= 1; i-- {
		a.Label(tailOut[i])
		a.Pextrw(asm.Address(e.di, (i-1)*2), asm.X1, asm.Constant(i-1))
	}

	a.Label(ret)
	a.Ret()

	e.BigLoop(bigloop_sse, e.vpand_sse, e.vpunpckhbw_sse, e.vpshufb_sse)
	a.Jmp(tail)
}

type decode struct {
	*asm.Asm

	di, si, cx, invldMskOut, invldMskIn asm.Register

	ret, invalid asm.Label

	base, toLower, high, low, valid, sign asm.Data

	valid0, valid2 asm.Operand
}

func (d *decode) vpxor_sse(ops ...asm.Operand) {
	if len(ops) != 3 {
		panic("wrong number of operands")
	}

	if ops[1] == ops[2] {
		panic("invalid register choice fallback")
	}

	if ops[0] != ops[1] {
		d.Movou(ops[0], ops[1])
	}

	d.Pxor(ops[0], ops[2])
}

func (d *decode) vpcmpgtb_sse(ops ...asm.Operand) {
	if len(ops) != 3 {
		panic("wrong number of operands")
	}

	if ops[1] == ops[2] {
		panic("invalid register choice fallback")
	}

	if ops[0] != ops[1] {
		d.Movou(ops[0], ops[1])
	}

	d.Pcmpgtb(ops[0], ops[2])
}

func (d *decode) vpshufb_sse(ops ...asm.Operand) {
	if len(ops) != 3 {
		panic("wrong number of operands")
	}

	if ops[1] == ops[2] {
		panic("invalid register choice fallback")
	}

	if ops[0] != ops[1] {
		d.Movou(ops[0], ops[1])
	}

	d.Pshufb(ops[0], ops[2])
}

func (d *decode) Convert(io, tmp4, tmp3, tmp2, tmp1 asm.Operand, vpxor, vpcmpgtb, vpshufb func(ops ...asm.Operand)) {
	vpxor(tmp1, io, d.sign)

	d.Por(io, d.toLower)

	vpxor(tmp2, io, d.sign)

	vpcmpgtb(tmp3, d.valid0, tmp1)
	d.Pcmpgtb(tmp1, d.valid.Offset(16))
	vpcmpgtb(tmp4, d.valid2, tmp2)
	d.Pcmpgtb(tmp2, d.valid.Offset(48))

	d.Pand(tmp1, tmp4)
	d.Por(tmp3, tmp2)
	d.Por(tmp3, tmp1)

	d.Pmovmskb(d.invldMskOut, tmp3)

	d.Testw(d.invldMskIn, d.invldMskOut)
	d.Jnz(d.invalid)

	d.Psubb(io, d.base.Offset(0))

	d.Pandn(tmp4, d.base.Offset(16))
	d.Psubb(io, tmp4)

	vpshufb(tmp3, io, d.low)
	d.Pshufb(io, d.high)

	d.Psllw(io, asm.Constant(4))
	d.Por(io, tmp3)
}

func (d *decode) BigLoop(l asm.Label, vpxor, vpcmpgtb, vpshufb func(ops ...asm.Operand)) {
	d.Label(l)

	d.Movou(asm.X0, asm.Address(d.si))

	d.Convert(asm.X0, asm.X4, asm.X3, asm.X2, asm.X1, vpxor, vpcmpgtb, vpshufb)

	d.Movq(asm.Address(d.di), asm.X0)

	d.Subq(d.cx, asm.Constant(16))
	d.Jz(d.ret)

	d.Addq(d.si, asm.Constant(16))
	d.Addq(d.di, asm.Constant(8))

	d.Cmpq(asm.Constant(16), d.cx)
	d.Jae(l)
}

func decodeASM(a *asm.Asm) {
	base := a.Data("decodeBase", bytes.Join([][]byte{
		repeat('0', 16),
		repeat('a'-'0'-10, 16),
	}, nil))
	toLower := a.Data("decodeToLower", repeat(0x20, 16))
	high := a.Data64("decodeHigh", []uint64{0x0e0c0a0806040200, ^uint64(0)})
	low := a.Data64("decodeLow", []uint64{0x0f0d0b0907050301, ^uint64(0)})
	valid := a.Data("decodeValid", bytes.Join([][]byte{
		repeat('0'^0x80, 16),
		repeat('9'^0x80, 16),
		repeat('a'^0x80, 16),
		repeat('f'^0x80, 16),
	}, nil))
	sign := a.Data("decodeToSigned", repeat(0x80, 16))

	a.NewFunction("decodeASM")
	a.NoSplit()

	dst := a.Argument("dst", 8)
	src := a.Argument("src", 8)
	length := a.Argument("len", 8)
	n := a.Argument("n", 8)
	ok := a.Argument("ok", 4)

	a.Start()

	bigloop_avx := a.NewLabel("bigloop_avx")
	bigloop_sse := a.NewLabel("bigloop_sse")
	tail := a.NewLabel("tail")
	ret := a.NewLabel("ret")
	invalid := a.NewLabel("invalid")

	d := &decode{
		a,

		asm.DI, asm.SI, asm.BX, asm.AX, asm.DX,

		ret, invalid,

		base, toLower, high, low, valid, sign,

		asm.X14, asm.X15,
	}

	a.Movq(d.di, dst)
	a.Movq(d.si, src)
	a.Movq(d.cx, length)

	a.Movq(asm.R15, d.si)

	a.Movou(d.valid0, d.valid.Offset(0))
	a.Movou(d.valid2, d.valid.Offset(32))

	a.Movw(d.invldMskIn, asm.Constant(0xffff))

	a.Cmpq(asm.Constant(16), d.cx)
	a.Jb(tail)

	a.Cmpb(asm.Constant(1), asm.Data("runtime·support_avx"))
	a.Jne(bigloop_sse)

	d.BigLoop(bigloop_avx, a.Vpxor, a.Vpcmpgtb, a.Vpshufb)

	a.Label(tail)

	tailIn := [15]asm.Label{tail.Suffix("in")}
	tailConv := tail.Suffix("conv")
	tailOut := [15]asm.Label{tail.Suffix("out")}

	for i := 2; i <= 14; i += 2 {
		tailIn[i] = tailIn[0].Suffix(strconv.Itoa(i))
		tailOut[i] = tailOut[0].Suffix(strconv.Itoa(i))
	}

	d.Movq(asm.CX, asm.Constant(16))
	d.Subq(asm.CX, d.cx)
	a.Shrw(d.invldMskIn, asm.CX)

	for i := 4; i <= 12; i += 4 {
		a.Cmpq(asm.Constant(i), d.cx)
		a.Jb(tailIn[i-2])
		a.Je(tailIn[i])
	}

	for i := 14; i >= 10; i -= 2 {
		a.Label(tailIn[i])
		a.Pinsrw(asm.X0, asm.Address(d.si, i-2), asm.Constant(i/2-1))
	}

	a.Label(tailIn[8])
	a.Pinsrq(asm.X0, asm.Address(d.si), asm.Constant(0))
	a.Jmp(tailConv)

	for i := 6; i >= 2; i -= 2 {
		a.Label(tailIn[i])
		a.Pinsrw(asm.X0, asm.Address(d.si, i-2), asm.Constant(i/2-1))
	}

	a.Label(tailConv)

	d.Convert(asm.X0, asm.X4, asm.X3, asm.X2, asm.X1, d.vpxor_sse, d.vpcmpgtb_sse, d.vpshufb_sse)

	for i := 4; i <= 12; i += 4 {
		a.Cmpq(asm.Constant(i), d.cx)
		a.Jb(tailOut[i-2])
		a.Je(tailOut[i])
	}

	for i := 14; i >= 10; i -= 2 {
		a.Label(tailOut[i])
		a.Pextrb(asm.Address(d.di, i/2-1), asm.X0, asm.Constant(i/2-1))
	}

	a.Label(tailOut[8])
	a.Movl(asm.Address(d.di), asm.X0)
	a.Jmp(ret)

	for i := 6; i >= 2; i -= 2 {
		a.Label(tailOut[i])
		a.Pextrb(asm.Address(d.di, i/2-1), asm.X0, asm.Constant(i/2-1))
	}

	a.Label(ret)
	a.Movb(ok, asm.Constant(1))
	a.Ret()

	a.Label(invalid)

	a.Bsfw(d.invldMskOut, d.invldMskOut)

	a.Subq(d.si, asm.R15)
	a.Addq(d.invldMskOut, d.si)

	a.Movq(n, d.invldMskOut)
	a.Movb(ok, asm.Constant(0))

	a.Ret()

	d.BigLoop(bigloop_sse, d.vpxor_sse, d.vpcmpgtb_sse, d.vpshufb_sse)
	a.Jmp(tail)
}

func main() {
	if err := asm.Do("hex_encode_amd64.s", header, encodeASM); err != nil {
		panic(err)
	}

	if err := asm.Do("hex_decode_amd64.s", header, decodeASM); err != nil {
		panic(err)
	}
}
