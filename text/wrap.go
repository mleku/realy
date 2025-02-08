package text

type AppendBytesClosure func(dst, src []byte) []byte

type AppendClosure func(dst []byte) []byte

func Unquote(b []byte) []byte { return b[1 : len(b)-1] }

func Noop(dst, src []byte) []byte { return append(dst, src...) }

func AppendQuote(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '"')
	dst = ac(dst, src)
	dst = append(dst, '"')
	return dst
}

func Quote(dst, src []byte) []byte { return AppendQuote(dst, src, Noop) }

func AppendSingleQuote(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '\'')
	dst = ac(dst, src)
	dst = append(dst, '\'')
	return dst
}

func AppendBackticks(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '`')
	dst = ac(dst, src)
	dst = append(dst, '`')
	return dst
}

func AppendBrace(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '(')
	dst = ac(dst, src)
	dst = append(dst, ')')
	return dst
}

func AppendParenthesis(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '{')
	dst = ac(dst, src)
	dst = append(dst, '}')
	return dst
}

func AppendBracket(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '[')
	dst = ac(dst, src)
	dst = append(dst, ']')
	return dst
}

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
