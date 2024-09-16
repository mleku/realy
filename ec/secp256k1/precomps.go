package secp256k1

import (
	_ "embed"
)

//go:embed rawbytepoints.bin
var bytepoints []byte
var BytePointTable [32][256]JacobianPoint

func init() {
	var cursor int
	for i := range BytePointTable {
		for j := range BytePointTable[i] {
			BytePointTable[i][j].X.SetByteSlice(bytepoints[cursor:])
			cursor += 32
			BytePointTable[i][j].Y.SetByteSlice(bytepoints[cursor:])
			cursor += 32
			BytePointTable[i][j].Z.SetInt(1)
		}
	}
}
