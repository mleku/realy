package bintly

import "math/bits"

const (
	sizeIntInBytes    = bits.UintSize / 8
	size64bitsInBytes = 8
	size32bitsInBytes = 4
	size16bitsInBytes = 2
	size8bitsInBytes  = 1
	//represents nil size
	NilSize = -1
)
