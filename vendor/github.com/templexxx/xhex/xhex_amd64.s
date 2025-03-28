#include "textflag.h"

#define dst R8
#define src R9
#define len R11
#define mask Y0
#define rmask Y1
#define tbl Y2

// HEX_TBL doubles normal hex table for AVX2 register (expand 16Bytes to 32Bytes)
DATA HEX_TBL<>+0x00(SB)/32, $"0123456789abcdef0123456789abcdef"
GLOBL HEX_TBL<>(SB), RODATA, $32

// func encodeAVX2(dst, src *byte, n int)
TEXT ·encodeAVX2(SB), NOSPLIT, $0
	MOVQ  d+0(FP), dst
	MOVQ  s+8(FP), src
	MOVQ  n+16(FP), len
	SHRQ  $4, len       // n / 16.
	TESTQ len, len
	JZ    ret           // If 0, return.

	// Preparation.
	// Make a 32Bytes mask filled with $0x0f.
	MOVQ         $0x0f, DX
	MOVQ         DX, X0
	VPBROADCASTB X0, mask
	MOVQ         ·replaceHighMask(SB), AX
	VMOVDQU      (AX), rmask
	VMOVDQU      HEX_TBL<>(SB), tbl

loop16b:
	VMOVDQU   (src), X3     // Load 16bytes source.
	VPMOVZXBW X3, Y3        // Zero extend 16bytes to 32bytes.
	VPSRLW    $4, Y3, Y4    // >> 4.
	VPSHUFB   rmask, Y3, Y3
	VPOR      Y4, Y3, Y4
	VPAND     mask, Y4, Y4  // Clean high bits.
	VPSHUFB   Y4, tbl, Y4
	VMOVDQU   Y4, (dst)

	ADDQ $16, src
	ADDQ $32, dst
	SUBQ $1, len
	JNE  loop16b
	VZEROUPPER

ret:
	RET

#define mask15 X0
#define maska X1
#define maskb X2
#define mask9 X10

// func decodeAVX2(dst, src *byte, n int)
TEXT ·decodeAVX2(SB), NOSPLIT, $0
	MOVQ  d+0(FP), dst
	MOVQ  s+8(FP), src
	MOVQ  n+16(FP), len
	SHRQ  $5, len       // n / 32.
	TESTQ len, len
	JZ    ret           // If 0, return.

	MOVQ         $0x0f, DX
	MOVQ         DX, X0
	VPBROADCASTW X0, mask15
	MOVQ         $9, DX
	MOVQ         DX, X10
	VPBROADCASTW X10, mask9

	MOVQ    ·decodeMask1(SB), AX
	VMOVDQU (AX), maska
	MOVQ    ·decodeMask2(SB), BX
	VMOVDQU (BX), maskb

loop32b:
	VMOVDQU (src), X3
	VMOVDQU 16(src), X4

	VPSHUFB maska, X3, X5
	VPSHUFB maskb, X3, X6
	VPSHUFB maska, X4, X7
	VPSHUFB maskb, X4, X8

	VPAND      mask15, X5, X11
	VPSRAW     $6, X5, X12
	VPMADDUBSW mask9, X12, X12
	VPADDW     X11, X12, X5

	VPAND      mask15, X6, X11
	VPSRAW     $6, X6, X12
	VPMADDUBSW mask9, X12, X12
	VPADDW     X11, X12, X6

	VPAND      mask15, X7, X11
	VPSRAW     $6, X7, X12
	VPMADDUBSW mask9, X12, X12
	VPADDW     X11, X12, X7

	VPAND      mask15, X8, X11
	VPSRAW     $6, X8, X12
	VPMADDUBSW mask9, X12, X12
	VPADDW     X11, X12, X8

	VPSLLW $4, X5, X5
	VPSLLW $4, X7, X7
	VPOR   X5, X6, X5
	VPOR   X7, X8, X7

	VPACKUSWB X5, X7, X9
	VPSHUFD   $78, X9, X9

	VMOVDQU X9, (dst)

	ADDQ $32, src
	ADDQ $16, dst
	SUBQ $1, len
	JNE  loop32b
	VZEROUPPER

ret:
	RET
