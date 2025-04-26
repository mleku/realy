package codegen

import "github.com/viant/toolbox"

//isBaseType checks if typeName is primitive types
func isBaseType(typeName string) bool {
	switch typeName {
	case "string", "bool", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "time.Time",
		"[]string", "[]bool", "[]int", "[]int8", "[]int16", "[]int32", "[]int64", "[]uint", "[]uint8", "[]uint16", "[]uint32", "[]uint64", "[]float32", "[]float64":
		return true
	}
	return false
}

//isStruct return true if type has at least 1 field
func isStruct(aType *toolbox.TypeInfo) bool {
	return len(aType.Fields()) > 0
}
