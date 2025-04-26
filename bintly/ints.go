package bintly

import (
	"unsafe"
)

const (
	n = 2048
)

//PutInts copy []int into []byte
func PutInts(bs []byte, vs []int) {
	bsLen := len(vs) * sizeIntInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / sizeIntInBytes
		copy(bs[bsOffset:bsOffset+n], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / sizeIntInBytes
	copy(bs[bsOffset:bsOffset+vsLimit], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit])
}

//Ints copy []byte  into []int
func Ints(bs []byte) []int {
	vs := make([]int, len(bs)/sizeIntInBytes)
	GetInts(bs, vs)
	return vs
}

//GetInts copy []byte  into []int
func GetInts(bs []byte, vs []int) {
	bsLen := len(vs) * sizeIntInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / sizeIntInBytes
		copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n], bs[bsOffset:bsOffset+n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / sizeIntInBytes
	copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit], bs[bsOffset:bsOffset+vsLimit])
}

//PutUints copy []uint into []byte
func PutUints(bs []byte, vs []uint) {
	bsLen := len(vs) * sizeIntInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / sizeIntInBytes
		copy(bs[bsOffset:bsOffset+n], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / sizeIntInBytes
	copy(bs[bsOffset:bsOffset+vsLimit], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit])
}

//GetUints copy []byte  into []uint
func GetUints(bs []byte, vs []uint) {
	bsLen := len(vs) * sizeIntInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / sizeIntInBytes
		copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n], bs[bsOffset:bsOffset+n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / sizeIntInBytes
	copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit], bs[bsOffset:bsOffset+vsLimit])
}

//PutUint64s copy []uint64 into []byte
func PutUint64s(bs []byte, vs []uint64) {
	bsLen := len(vs) * size64bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size64bitsInBytes
		copy(bs[bsOffset:bsOffset+n], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size64bitsInBytes
	copy(bs[bsOffset:bsOffset+vsLimit], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit])
}

//GetUint64s copy []byte  into []uint64
func GetUint64s(bs []byte, vs []uint64) {
	bsLen := len(vs) * size64bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size64bitsInBytes
		copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n], bs[bsOffset:bsOffset+n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size64bitsInBytes
	copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit], bs[bsOffset:bsOffset+vsLimit])
}

//Uint64s copy []byte  into []uint64
func Uint64s(bs []byte) []uint64 {
	vs := make([]uint64, len(bs)/size64bitsInBytes)
	GetUint64s(bs, vs)
	return vs
}

//PutUint32s copy []uint32 into []byte
func PutUint32s(bs []byte, vs []uint32) {
	bsLen := len(vs) * size32bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size32bitsInBytes
		copy(bs[bsOffset:bsOffset+n], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size32bitsInBytes
	copy(bs[bsOffset:bsOffset+vsLimit], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit])
}

//GetUint32s copy []byte  into []uint32
func GetUint32s(bs []byte, vs []uint32) {
	bsLen := len(vs) * size32bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size32bitsInBytes
		copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n], bs[bsOffset:bsOffset+n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size32bitsInBytes
	copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit], bs[bsOffset:bsOffset+vsLimit])
}

//Uint32s copy []byte  into []uint32
func Uint32s(bs []byte) []uint32 {
	vs := make([]uint32, len(bs)/size32bitsInBytes)
	GetUint32s(bs, vs)
	return vs
}

//PutUint16s copy []uint16 into []byte
func PutUint16s(bs []byte, vs []uint16) {
	bsLen := len(vs) * size16bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size16bitsInBytes
		copy(bs[bsOffset:bsOffset+n], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size16bitsInBytes
	copy(bs[bsOffset:bsOffset+vsLimit], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit])
}

//GetUint16s copy []byte  into []uint16
func GetUint16s(bs []byte, vs []uint16) {
	bsLen := len(vs) * size16bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size16bitsInBytes
		copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n], bs[bsOffset:bsOffset+n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size16bitsInBytes
	copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit], bs[bsOffset:bsOffset+vsLimit])
}

//Uint16s copy []byte  into []uint16
func Uint16s(bs []byte) []uint16 {
	vs := make([]uint16, len(bs)/size16bitsInBytes)
	GetUint16s(bs, vs)
	return vs
}

//PutInt64s copy []int64 into []byte
func PutInt64s(bs []byte, vs []int64) {
	bsLen := len(vs) * size64bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size64bitsInBytes
		copy(bs[bsOffset:bsOffset+n], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size64bitsInBytes
	copy(bs[bsOffset:bsOffset+vsLimit], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit])
}

//GetInt64s copy []byte  into []int64
func GetInt64s(bs []byte, vs []int64) {
	bsLen := len(vs) * size64bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size64bitsInBytes
		copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n], bs[bsOffset:bsOffset+n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size64bitsInBytes
	copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit], bs[bsOffset:bsOffset+vsLimit])
}

//Int64s copy []byte  into []int64
func Int64s(bs []byte) []int64 {
	vs := make([]int64, len(bs)/size64bitsInBytes)
	GetInt64s(bs, vs)
	return vs
}

//PutInt32s copy []int32 into []byte
func PutInt32s(bs []byte, vs []int32) {
	bsLen := len(vs) * size32bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size32bitsInBytes
		copy(bs[bsOffset:bsOffset+n], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size32bitsInBytes
	copy(bs[bsOffset:bsOffset+vsLimit], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit])
}

//GetInt32s copy []byte  into []int32
func GetInt32s(bs []byte, vs []int32) {
	bsLen := len(vs) * size32bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size32bitsInBytes
		copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n], bs[bsOffset:bsOffset+n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size32bitsInBytes
	copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit], bs[bsOffset:bsOffset+vsLimit])
}

//Int32s copy []byte  into []int32
func Int32s(bs []byte) []int32 {
	vs := make([]int32, len(bs)/size32bitsInBytes)
	GetInt32s(bs, vs)
	return vs
}

//PutInt16s copy []int16 into []byte
func PutInt16s(bs []byte, vs []int16) {
	bsLen := len(vs) * size16bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size16bitsInBytes
		copy(bs[bsOffset:bsOffset+n], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size16bitsInBytes
	copy(bs[bsOffset:bsOffset+vsLimit], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit])
}

//GetInt16s copy []byte  into []int16
func GetInt16s(bs []byte, vs []int16) {
	bsLen := len(vs) * size16bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size16bitsInBytes
		copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n], bs[bsOffset:bsOffset+n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size16bitsInBytes
	copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit], bs[bsOffset:bsOffset+vsLimit])
}

//Int16s copy []byte  into []int16
func Int16s(bs []byte) []int16 {
	vs := make([]int16, len(bs)/size16bitsInBytes)
	GetInt16s(bs, vs)
	return vs
}

//PutInt8s copy []int8 into []byte
func PutInt8s(bs []byte, vs []int8) {
	bsLen := len(vs) * size8bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size8bitsInBytes
		copy(bs[bsOffset:bsOffset+n], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size8bitsInBytes
	copy(bs[bsOffset:bsOffset+vsLimit], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit])
}

//GetInt8s copy []byte  into []int8
func GetInt8s(bs []byte, vs []int8) {
	bsLen := len(vs) * size8bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size8bitsInBytes
		copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n], bs[bsOffset:bsOffset+n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size8bitsInBytes
	copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit], bs[bsOffset:bsOffset+vsLimit])
}

//Int8s copy []byte  into []int8
func Int8s(bs []byte) []int8 {
	vs := make([]int8, len(bs)/size8bitsInBytes)
	GetInt8s(bs, vs)
	return vs
}

//PutUint8s copy []uint8 into []byte
func PutUint8s(bs []byte, vs []uint8) {
	bsLen := len(vs) * size8bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size8bitsInBytes
		copy(bs[bsOffset:bsOffset+n], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size8bitsInBytes
	copy(bs[bsOffset:bsOffset+vsLimit], (*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit])
}

//GetUint8s copy []byte  into []uint8
func GetUint8s(bs []byte, vs []uint8) {
	bsLen := len(vs) * size8bitsInBytes
	chunks := bsLen / n
	bsOffset := 0
	for i := 0; i < chunks; i++ {
		index := bsOffset / size8bitsInBytes
		copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:n], bs[bsOffset:bsOffset+n])
		bsOffset += n
	}
	vsLimit := bsLen % n
	if vsLimit == 0 {
		return
	}
	index := bsOffset / size8bitsInBytes
	copy((*(*[n]byte)(unsafe.Pointer(&vs[index])))[:vsLimit], bs[bsOffset:bsOffset+vsLimit])
}

//Uint8s copy []byte  into []uint8
func Uint8s(bs []byte) []uint8 {
	vs := make([]uint8, len(bs)/size8bitsInBytes)
	GetUint8s(bs, vs)
	return vs
}
