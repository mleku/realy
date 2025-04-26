package bintly

import "unsafe"

//PutFloat64s copy []float64 into []byte
func PutFloat64s(bs []byte, vs []float64) {
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

//GetFloat64s copy []byte  into []float64
func GetFloat64s(bs []byte, vs []float64) {
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

//PutFloat32s copy []float32 into []byte
func PutFloat32s(bs []byte, vs []float32) {
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

//GetFloat32s copy []byte  into []float32
func GetFloat32s(bs []byte, vs []float32) {
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
