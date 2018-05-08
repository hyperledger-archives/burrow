// Copyright 2016 Tom Thorogood. All rights reserved.
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

#include "textflag.h"

DATA encodeMask<>+0x00(SB)/8, $0x0f0f0f0f0f0f0f0f
DATA encodeMask<>+0x08(SB)/8, $0x0f0f0f0f0f0f0f0f
GLOBL encodeMask<>(SB),RODATA,$16

TEXT ·encodeASM(SB),NOSPLIT,$0
	MOVQ dst+0(FP), DI
	MOVQ src+8(FP), SI
	MOVQ len+16(FP), BX
	MOVQ alpha+24(FP), DX
	MOVOU (DX), X15
	CMPQ BX, $16
	JB tail
	CMPB runtime·support_avx(SB), $1
	JNE bigloop_sse
bigloop_avx:
	MOVOU -16(SI)(BX*1), X0
	VPAND encodeMask<>(SB), X0, X1
	PSRLW $4, X0
	PAND encodeMask<>(SB), X0
	// VPUNPCKHBW X1, X0, X3
	BYTE $0xc5; BYTE $0xf9; BYTE $0x68; BYTE $0xd9
	PUNPCKLBW X1, X0
	VPSHUFB X0, X15, X1
	VPSHUFB X3, X15, X2
	MOVOU X2, -16(DI)(BX*2)
	MOVOU X1, -32(DI)(BX*2)
	SUBQ $16, BX
	JZ ret
	CMPQ BX, $16
	JAE bigloop_avx
tail:
	CMPQ BX, $2
	JB tail_in_1
	JE tail_in_2
	CMPQ BX, $4
	JB tail_in_3
	JE tail_in_4
	CMPQ BX, $6
	JB tail_in_5
	JE tail_in_6
	CMPQ BX, $8
	JB tail_in_7
tail_in_8:
	MOVQ (SI), X0
	JMP tail_conv
tail_in_7:
	PINSRB $6, 6(SI), X0
tail_in_6:
	PINSRB $5, 5(SI), X0
tail_in_5:
	PINSRB $4, 4(SI), X0
tail_in_4:
	PINSRD $0, (SI), X0
	JMP tail_conv
tail_in_3:
	PINSRB $2, 2(SI), X0
tail_in_2:
	PINSRB $1, 1(SI), X0
tail_in_1:
	PINSRB $0, (SI), X0
tail_conv:
	MOVOU X0, X1
	PAND encodeMask<>(SB), X1
	PSRLW $4, X0
	PAND encodeMask<>(SB), X0
	PUNPCKLBW X1, X0
	MOVOU X15, X1
	PSHUFB X0, X1
	CMPQ BX, $2
	JB tail_out_1
	JE tail_out_2
	CMPQ BX, $4
	JB tail_out_3
	JE tail_out_4
	CMPQ BX, $6
	JB tail_out_5
	JE tail_out_6
	CMPQ BX, $8
	JB tail_out_7
tail_out_8:
	MOVOU X1, (DI)
	SUBQ $8, BX
	JZ ret
	ADDQ $8, SI
	ADDQ $16, DI
	JMP tail
tail_out_7:
	// PEXTRW $6, X1, 12(DI)
	BYTE $0x66; BYTE $0x0f; BYTE $0x3a; BYTE $0x15; BYTE $0x4f; BYTE $0x0c; BYTE $0x06
tail_out_6:
	// PEXTRW $5, X1, 10(DI)
	BYTE $0x66; BYTE $0x0f; BYTE $0x3a; BYTE $0x15; BYTE $0x4f; BYTE $0x0a; BYTE $0x05
tail_out_5:
	// PEXTRW $4, X1, 8(DI)
	BYTE $0x66; BYTE $0x0f; BYTE $0x3a; BYTE $0x15; BYTE $0x4f; BYTE $0x08; BYTE $0x04
tail_out_4:
	MOVQ X1, (DI)
	RET
tail_out_3:
	// PEXTRW $2, X1, 4(DI)
	BYTE $0x66; BYTE $0x0f; BYTE $0x3a; BYTE $0x15; BYTE $0x4f; BYTE $0x04; BYTE $0x02
tail_out_2:
	// PEXTRW $1, X1, 2(DI)
	BYTE $0x66; BYTE $0x0f; BYTE $0x3a; BYTE $0x15; BYTE $0x4f; BYTE $0x02; BYTE $0x01
tail_out_1:
	// PEXTRW $0, X1, (DI)
	BYTE $0x66; BYTE $0x0f; BYTE $0x3a; BYTE $0x15; BYTE $0x0f; BYTE $0x00
ret:
	RET
bigloop_sse:
	MOVOU -16(SI)(BX*1), X0
	MOVOU X0, X1
	PAND encodeMask<>(SB), X1
	PSRLW $4, X0
	PAND encodeMask<>(SB), X0
	MOVOU X0, X3
	PUNPCKHBW X1, X3
	PUNPCKLBW X1, X0
	MOVOU X15, X1
	PSHUFB X0, X1
	MOVOU X15, X2
	PSHUFB X3, X2
	MOVOU X2, -16(DI)(BX*2)
	MOVOU X1, -32(DI)(BX*2)
	SUBQ $16, BX
	JZ ret
	CMPQ BX, $16
	JAE bigloop_sse
	JMP tail
