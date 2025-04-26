package conv

import "reflect"

var intType = reflect.TypeOf(0)
var uintType = reflect.TypeOf(uint(0))
var int64Type = reflect.TypeOf(int64(0))
var uint64Type = reflect.TypeOf(uint64(0))
var int32Type = reflect.TypeOf(int32(0))
var uint32Type = reflect.TypeOf(uint32(0))
var int16Type = reflect.TypeOf(int16(0))
var uint16Type = reflect.TypeOf(uint16(0))
var int8Type = reflect.TypeOf(int8(0))
var uint8Type = reflect.TypeOf(uint8(0))
var float64Type = reflect.TypeOf(float64(0))
var float32Type = reflect.TypeOf(float32(0))
var boolType = reflect.TypeOf(true)
var stringType = reflect.TypeOf("")
var bytesType = reflect.TypeOf([]byte{})

func IsNative(t reflect.Type) bool {
	switch t {
	case intType, uintType, int64Type, uint64Type, int32Type, uint32Type,
		int16Type, uint16Type, int8Type, uint8Type,
		float64Type, float32Type, boolType, stringType, bytesType:
		return true
	}
	return false
}

//MatchNative returns matched alias native type pointer or nil
func MatchNative(t reflect.Type) *reflect.Type {
	switch t.Kind() {
	case reflect.Int:
		return &intType
	case reflect.Uint:
		return &uintType
	case reflect.Int64:
		return &int64Type
	case reflect.Uint64:
		return &uint64Type
	case reflect.Int32:
		return &int32Type
	case reflect.Uint32:
		return &uint32Type
	case reflect.Int16:
		return &int16Type
	case reflect.Uint16:
		return &uint16Type
	case reflect.Int8:
		return &int8Type
	case reflect.Uint8:
		return &uint8Type
	case reflect.Float64:
		return &float64Type
	case reflect.Float32:
		return &float32Type
	case reflect.Bool:
		return &boolType
	case reflect.String:
		return &stringType
	}
	if IsBytesConvertibleTo(t) {
		return &bytesType
	}
	return nil
}

//IsByteConvertibleTo return true if []byte ConvertibleTo
func IsBytesConvertibleTo(t reflect.Type) bool {
	return t.ConvertibleTo(bytesType)
}
