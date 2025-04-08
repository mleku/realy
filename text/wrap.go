package text

// AppendBytesClosure is a function type for appending data from a source to a destination and
// returning the appended-to slice.
type AppendBytesClosure func(dst, src []byte) []byte

// AppendClosure is a simple append where the caller appends to the destination and returns the
// appended-to slice.
type AppendClosure func(dst []byte) []byte

// Unquote removes the quotes around a slice of bytes.
func Unquote(b []byte) []byte { return b[1 : len(b)-1] }

// Noop simply appends the source to the destination slice and returns it.
func Noop(dst, src []byte) []byte { return append(dst, src...) }

// AppendQuote appends a source of bytes, that have been processed by an AppendBytesClosure and
// returns the appended-to slice.
func AppendQuote(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '"')
	dst = ac(dst, src)
	dst = append(dst, '"')
	return dst
}

// Quote simply quotes a provided source and attaches it to the provided destination slice.
func Quote(dst, src []byte) []byte { return AppendQuote(dst, src, Noop) }

// AppendSingleQuote appends a provided AppendBytesClosure's output from a given source of
// bytes, wrapped in single quotes ”.
func AppendSingleQuote(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '\'')
	dst = ac(dst, src)
	dst = append(dst, '\'')
	return dst
}

// AppendBackticks appends a provided AppendBytesClosure's output from a given source of
// bytes, wrapped in backticks “.
func AppendBackticks(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '`')
	dst = ac(dst, src)
	dst = append(dst, '`')
	return dst
}

// AppendBrace appends a provided AppendBytesClosure's output from a given source of
// bytes, wrapped in braces ().
func AppendBrace(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '(')
	dst = ac(dst, src)
	dst = append(dst, ')')
	return dst
}

// AppendParenthesis appends a provided AppendBytesClosure's output from a given source of
// bytes, wrapped in parentheses {}.
func AppendParenthesis(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '{')
	dst = ac(dst, src)
	dst = append(dst, '}')
	return dst
}

// AppendBracket appends a provided AppendBytesClosure's output from a given source of
// bytes, wrapped in brackets [].
func AppendBracket(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '[')
	dst = ac(dst, src)
	dst = append(dst, ']')
	return dst
}

// AppendList appends an input source bytes processed by an AppendBytesClosure and separates
// elements with the given separator byte.
func AppendList(dst []byte, src [][]byte, separator byte,
	ac AppendBytesClosure) []byte {
	last := len(src) - 1
	for i := range src {
		dst = append(dst, ac(dst, src[i])...)
		if i < last {
			dst = append(dst, separator)
		}
	}
	return dst
}
