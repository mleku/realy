package bintly

import "unsafe"

//PutInt copy int into []byte
func PutInt(dest []byte, v int) {
	copy(dest, (*(*[sizeIntInBytes]byte)(unsafe.Pointer(&v)))[:])
}

//Int copy []byte  to int
func Int(b []byte) int {
	v := 0
	copy((*(*[sizeIntInBytes]byte)(unsafe.Pointer(&v)))[:], b)
	return v
}

//GetUint copy []byte  to uint
func GetUint(b []byte, v *uint) {
	copy((*(*[sizeIntInBytes]byte)(unsafe.Pointer(v)))[:], b)
}

//PutUint copy uint into []byte
func PutUint(dest []byte, v uint) {
	copy(dest, (*(*[sizeIntInBytes]byte)(unsafe.Pointer(&v)))[:])
}

//Uint copy []byte  to uint
func Uint(b []byte) uint {
	v := uint(0)
	copy((*(*[sizeIntInBytes]byte)(unsafe.Pointer(&v)))[:], b)
	return v
}

//GetInt copy []byte  to int
func GetInt(b []byte, v *int) {
	copy((*(*[sizeIntInBytes]byte)(unsafe.Pointer(v)))[:], b)
}

//PutUint64 copy uint64 into []byte
func PutUint64(dest []byte, v uint64) {
	copy(dest, (*(*[size64bitsInBytes]byte)(unsafe.Pointer(&v)))[:])
}

//Uint64 copy []byte  to uint64
func Uint64(b []byte) uint64 {
	v := uint64(0)
	copy((*(*[size64bitsInBytes]byte)(unsafe.Pointer(&v)))[:], b)
	return v
}

//GetUint64 copy []byte  to *uint64
func GetUint64(b []byte, v *uint64) {
	copy((*(*[size64bitsInBytes]byte)(unsafe.Pointer(v)))[:], b)
}

//PutUint32 copy uint32 into []byte
func PutUint32(dest []byte, v uint32) {
	copy(dest, (*(*[size32bitsInBytes]byte)(unsafe.Pointer(&v)))[:])
}

//Uint32 copy []byte to uint32
func Uint32(b []byte) uint32 {
	v := uint32(0)
	copy((*(*[size32bitsInBytes]byte)(unsafe.Pointer(&v)))[:], b)
	return v
}

//GetUint32 copy []byte to *uint32
func GetUint32(b []byte, v *uint32) {
	copy((*(*[size32bitsInBytes]byte)(unsafe.Pointer(v)))[:], b)
}

//PutUint16 copy uint16 into []byte
func PutUint16(dest []byte, v uint16) {
	copy(dest, (*(*[2]byte)(unsafe.Pointer(&v)))[:])
}

//Uint16 copy []byte  to uint16
func Uint16(b []byte) uint16 {
	v := uint16(0)
	copy((*(*[2]byte)(unsafe.Pointer(&v)))[:], b)
	return v
}

//GetUint16 copy []byte  to *uint16
func GetUint16(b []byte, v *uint16) {
	copy((*(*[2]byte)(unsafe.Pointer(v)))[:], b)
}

//PutInt64 copy int64 into []byte
func PutInt64(dest []byte, v int64) {
	copy(dest, (*(*[size64bitsInBytes]byte)(unsafe.Pointer(&v)))[:])
}

//Int64 copy []byte  to int64
func Int64(b []byte) int64 {
	v := int64(0)
	copy((*(*[size64bitsInBytes]byte)(unsafe.Pointer(&v)))[:], b)
	return v
}

//GetInt64 copy []byte  to int64
func GetInt64(b []byte, v *int64) {
	copy((*(*[size64bitsInBytes]byte)(unsafe.Pointer(v)))[:], b)
}

//PutInt32 copy int32 into []byte
func PutInt32(dest []byte, v int32) {
	copy(dest, (*(*[size32bitsInBytes]byte)(unsafe.Pointer(&v)))[:])
}

//Int32 copy []byte to int32
func Int32(b []byte) int32 {
	v := int32(0)
	copy((*(*[size32bitsInBytes]byte)(unsafe.Pointer(&v)))[:], b)
	return v
}

//GetInt32 copy []byte to *int32
func GetInt32(b []byte, v *int32) {
	copy((*(*[size32bitsInBytes]byte)(unsafe.Pointer(v)))[:], b)
}

//PutInt16 copy int16 into []byte
func PutInt16(dest []byte, v int16) {
	copy(dest, (*(*[2]byte)(unsafe.Pointer(&v)))[:])
}

//Int16 copy []byte  to int16
func Int16(b []byte) int16 {
	v := int16(0)
	copy((*(*[2]byte)(unsafe.Pointer(&v)))[:], b)
	return v
}

//GetInt16 copy []byte  to *int16
func GetInt16(b []byte, v *int16) {
	copy((*(*[2]byte)(unsafe.Pointer(v)))[:], b)
}
