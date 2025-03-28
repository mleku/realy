// Copyright (c) 2020. Temple3x (temple3x@gmail.com)
// Copyright (c) 2017 Zach Bjornson
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xhex

import "github.com/templexxx/cpu"

func init() {
	if cpu.X86.HasAVX2 {
		encode = func(dst, src []byte) {
			n := len(src)
			if n == 0 {
				return
			}
			encodeAVX2(&dst[0], &src[0], n)
			done := n >> 4 << 4
			if done == n {
				return
			}

			// Deal with unaligned part.
			dst = dst[done*2:]
			src = src[done:]
			encodeBase(dst, src)
		}

		decode = func(dst, src []byte) error {
			n := len(src)
			if n == 0 {
				return nil
			}
			decodeAVX2(&dst[0], &src[0], n)
			done := n >> 5 << 5
			if done == n {
				return nil
			}
			// Deal with unaligned part.
			dst = dst[done/2:]
			src = src[done:]
			return decodeBase(dst, src)
		}
	}
}

// This mask will help to replace the byte in the higher position with byte in lower (each two bytes),
// and leave the lower part 0.
// e.g.
// Before Packed Shuffle Bytes:
// [82 0 253 0 252 0 7 0 33 0 130 0 101 0 79 0 22 0 63 0 95 0 15 0 154 0 98 0 29 0 114 0]
// After Packed Shuffle Bytes with this mask:
// [0 82 0 253 0 252 0 7 0 33 0 130 0 101 0 79 0 22 0 63 0 95 0 15 0 154 0 98 0 29 0 114]
//
// If negative integer make you uncomfortable, you could use:
// []uint8{129, 0, 129, 2, 129, 4, 129, 6, 129, 8, 129, 10, 129, 12, 129, 14,
//	129, 0, 129, 2, 129, 4, 129, 6, 129, 8, 129, 10, 129, 12, 129, 14}
// They have same effect indeed.
var replaceHighMask = []int8{-1, 0, -1, 2, -1, 4, -1, 6, -1, 8, -1, 10, -1, 12, -1, 14,
	-1, 0, -1, 2, -1, 4, -1, 6, -1, 8, -1, 10, -1, 12, -1, 14}

// This two masks are used for separating high and low nibbles and extend into 16-bit elements.
var decodeMask1 = []int8{0, -1, 2, -1, 4, -1, 6, -1, 8, -1, 10, -1, 12, -1, 14, -1}
var decodeMask2 = []int8{1, -1, 3, -1, 5, -1, 7, -1, 9, -1, 11, -1, 13, -1, 15, -1}

// encodeAVX2 encodes bytes multiple of 16(src) with AVX2 instructions.
// After lots of attempts, the algorithm described in https://github.com/zbjornson/fast-hex is the finally answer.
// There are ways to achieve same goal, fast-hex is the one of fastest,
// and the algorithm is easy to understand.
//
// More details about others:
// I have roughly read a Go version SIMD hex encoding: https://github.com/tmthrgd/go-hex,
// it should be the best choice, but the assembly codes are generated but not handwritten,
// it's awful to review it, and I found AVX-SSE transition penalty in the codes.
//go:noescape
func encodeAVX2(dst, src *byte, n int)

// decodeAVX2 decodes bytes multiple of 32(src) with AVX2 instructions.
// The main idea is still from https://github.com/zbjornson/fast-hex, but with some modification:
// In fast-hex, decode 64bytes in src every loop,
// in decodeAVX2, decode 32bytes in src every loop.
//go:noesacpe
func decodeAVX2(dst, src *byte, n int)
