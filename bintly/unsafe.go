package bintly

import "unsafe"

func unsafeGetBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

func unsafeGetString(bs []byte) string {
	if len(bs) == 0 {
		return ""
	}
	return *(*string)(unsafe.Pointer(&bs))
}
