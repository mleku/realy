package text

type AppendBytesClosure func(dst, src by) by

type AppendClosure func(dst by) by

func Unquote(b by) by { return b[1 : len(b)-1] }

func Noop(dst, src by) by { return append(dst, src...) }

func AppendQuote(dst, src by, ac AppendBytesClosure) by {
	dst = append(dst, '"')
	dst = ac(dst, src)
	dst = append(dst, '"')
	return dst
}

func Quote(dst, src by) by { return AppendQuote(dst, src, Noop) }

func AppendSingleQuote(dst, src by, ac AppendBytesClosure) by {
	dst = append(dst, '\'')
	dst = ac(dst, src)
	dst = append(dst, '\'')
	return dst
}

func AppendBackticks(dst, src by, ac AppendBytesClosure) by {
	dst = append(dst, '`')
	dst = ac(dst, src)
	dst = append(dst, '`')
	return dst
}

func AppendBrace(dst, src by, ac AppendBytesClosure) by {
	dst = append(dst, '(')
	dst = ac(dst, src)
	dst = append(dst, ')')
	return dst
}

func AppendParenthesis(dst, src by, ac AppendBytesClosure) by {
	dst = append(dst, '{')
	dst = ac(dst, src)
	dst = append(dst, '}')
	return dst
}

func AppendBracket(dst, src by, ac AppendBytesClosure) by {
	dst = append(dst, '[')
	dst = ac(dst, src)
	dst = append(dst, ']')
	return dst
}

func AppendList(dst by, src []by, separator byte,
	ac AppendBytesClosure) by {
	last := len(src) - 1
	for i := range src {
		dst = append(dst, ac(dst, src[i])...)
		if i < last {
			dst = append(dst, separator)
		}
	}
	return dst
}
