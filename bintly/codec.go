package bintly

const (
	codecEOF = uint8(iota)
	codecAlloc  //alloc upto 2Gb
	codecMAlloc //medium size alloc upto 64k
	codecInts
	codecUints
	codecInt64s
	codecUint64s
	codecInt32s
	codecUint32s
	codecInt16s
	codecUint16s
	codecInt8s
	codecUint8s
	codecFloat64s
	codecFloat32s
)
