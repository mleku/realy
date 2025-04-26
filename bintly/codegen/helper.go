package codegen

import (
	"strings"
)

func firstLetterToUppercase(text string) string {
	return strings.ToUpper(string(text[0:1])) + string(text[1:])
}

