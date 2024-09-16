package text

type AppendBytesClosure func(dst, src B) B

type AppendClosure func(dst B) B

func Unquote(b B) B { return b[1 : len(b)-1] }

func Noop(dst, src B) B { return append(dst, src...) }

func AppendQuote(dst, src B, ac AppendBytesClosure) B {
	dst = append(dst, '"')
	dst = ac(dst, src)
	dst = append(dst, '"')
	return dst
}

func Quote(dst, src B) B { return AppendQuote(dst, src, Noop) }

func AppendSingleQuote(dst, src B, ac AppendBytesClosure) B {
	dst = append(dst, '\'')
	dst = ac(dst, src)
	dst = append(dst, '\'')
	return dst
}

func AppendBackticks(dst, src B, ac AppendBytesClosure) B {
	dst = append(dst, '`')
	dst = ac(dst, src)
	dst = append(dst, '`')
	return dst
}

func AppendBrace(dst, src B, ac AppendBytesClosure) B {
	dst = append(dst, '(')
	dst = ac(dst, src)
	dst = append(dst, ')')
	return dst
}

func AppendParenthesis(dst, src B, ac AppendBytesClosure) B {
	dst = append(dst, '{')
	dst = ac(dst, src)
	dst = append(dst, '}')
	return dst
}

func AppendBracket(dst, src B, ac AppendBytesClosure) B {
	dst = append(dst, '[')
	dst = ac(dst, src)
	dst = append(dst, ']')
	return dst
}

func AppendList(dst B, src []B, separator byte,
	ac AppendBytesClosure) B {
	last := len(src) - 1
	for i := range src {
		dst = append(dst, ac(dst, src[i])...)
		if i < last {
			dst = append(dst, separator)
		}
	}
	return dst
}
