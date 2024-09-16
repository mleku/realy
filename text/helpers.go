package text

import (
	"io"

	"github.com/templexxx/xhex"
	"realy.lol/hex"
	"realy.lol/ints"
	"realy.lol/kind"
	"realy.lol/kinds"
)

// JSONKey generates the JSON format for an object key and terminates with the
// semicolon.
func JSONKey(dst, k B) (b B) {
	dst = append(dst, '"')
	dst = append(dst, k...)
	dst = append(dst, '"', ':')
	b = dst
	return
}

// UnmarshalHex takes a byte string that should contain a quoted hexadecimal
// encoded value, decodes it in-place using a SIMD hex codec and returns the
// decoded truncated bytes (the other half will be as it was but no allocation
// is required).
func UnmarshalHex(b B) (h B, rem B, err error) {
	rem = b[:]
	var inQuote bool
	var start int
	for i := 0; i < len(b); i++ {
		if !inQuote {
			if b[i] == '"' {
				inQuote = true
				start = i + 1
			}
		} else if b[i] == '"' {
			h = b[start:i]
			rem = b[i+1:]
			break
		}
	}
	if !inQuote {
		err = io.EOF
		return
	}
	l := len(h)
	if l%2 != 0 {
		err = errorf.E("invalid length for hex: %d, %0x", len(h), h)
		return
	}
	if err = xhex.Decode(h, h); chk.E(err) {
		return
	}
	h = h[:l/2]
	return
}

// UnmarshalQuoted performs an in-place unquoting of NIP-01 quoted byte string.
func UnmarshalQuoted(b B) (content, rem B, err error) {
	rem = b[:]
	for ; len(rem) >= 0; rem = rem[1:] {
		// advance to open quotes
		if rem[0] == '"' {
			rem = rem[1:]
			content = rem
			break
		}
	}
	if len(rem) == 0 {
		err = io.EOF
		return
	}
	var escaping bool
	var contentLen int
	for len(rem) > 0 {
		if rem[0] == '\\' {
			escaping = true
			contentLen++
			rem = rem[1:]
		} else if rem[0] == '"' {
			if !escaping {
				rem = rem[1:]
				content = content[:contentLen]
				content = NostrUnescape(content)
				return
			}
			contentLen++
			rem = rem[1:]
			escaping = false
		} else {
			escaping = false
			switch rem[0] {
			// none of these characters are allowed inside a JSON string:
			//
			// backspace, tab, newline, form feed or carriage return.
			case '\b', '\t', '\n', '\f', '\r':
				err = errorf.E("invalid character '%s' in quoted string",
					NostrEscape(nil, rem[:1]))
				return
			}
			contentLen++
			rem = rem[1:]
		}
	}
	return
}

func MarshalHexArray(dst B, ha []B) (b B) {
	dst = append(dst, '[')
	for i := range ha {
		dst = AppendQuote(dst, ha[i], hex.EncAppend)
		if i != len(ha)-1 {
			dst = append(dst, ',')
		}
	}
	dst = append(dst, ']')
	b = dst
	return
}

// UnmarshalHexArray unpacks a JSON array containing strings with hexadecimal,
// and checks all values have the specified byte size..
func UnmarshalHexArray(b B, size int) (t []B, rem B, err error) {
	rem = b
	var openBracket bool
	for ; len(rem) > 0; rem = rem[1:] {
		if rem[0] == '[' {
			openBracket = true
		} else if openBracket {
			if rem[0] == ',' {
				continue
			} else if rem[0] == ']' {
				rem = rem[1:]
				return
			} else if rem[0] == '"' {
				var h B
				if h, rem, err = UnmarshalHex(rem); chk.E(err) {
					return
				}
				if len(h) != size {
					err = errorf.E("invalid hex array size, got %d expect %d",
						len(h), size)
					return
				}
				t = append(t, h)
				if rem[0] == ']' {
					rem = rem[1:]
					// done
					return
				}
			}
		}
	}
	return
}

// UnmarshalStringArray unpacks a JSON array containing strings.
func UnmarshalStringArray(b B) (t []B, rem B, err error) {
	rem = b
	var openBracket bool
	for ; len(rem) > 0; rem = rem[1:] {
		if rem[0] == '[' {
			openBracket = true
		} else if openBracket {
			if rem[0] == ',' {
				continue
			} else if rem[0] == ']' {
				rem = rem[1:]
				return
			} else if rem[0] == '"' {
				var h B
				if h, rem, err = UnmarshalQuoted(rem); chk.E(err) {
					return
				}
				t = append(t, h)
				if rem[0] == ']' {
					rem = rem[1:]
					// done
					return
				}
			}
		}
	}
	return
}

func MarshalKindsArray(dst B, ka *kinds.T) (b B) {
	dst = append(dst, '[')
	for i := range ka.K {
		dst, _ = ka.K[i].MarshalJSON(dst)
		if i != len(ka.K)-1 {
			dst = append(dst, ',')
		}
	}
	dst = append(dst, ']')
	b = dst
	return
}

func UnmarshalKindsArray(b B) (k *kinds.T, rem B, err error) {
	rem = b
	k = &kinds.T{}
	var openedBracket bool
	for ; len(rem) > 0; rem = rem[1:] {
		if !openedBracket && rem[0] == '[' {
			openedBracket = true
			continue
		} else if openedBracket {
			if rem[0] == ']' {
				// done
				return
			} else if rem[0] == ',' {
				continue
			}
			kk := ints.New(0)
			if rem, err = kk.UnmarshalJSON(rem); chk.E(err) {
				return
			}
			k.K = append(k.K, kind.New(kk.Uint16()))
			if rem[0] == ']' {
				rem = rem[1:]
				return
			}
		}
	}
	if !openedBracket {
		log.I.F("\n%v\n%s", k, rem)
		return nil, nil, errorf.E("kinds: failed to unmarshal\n%s\n%s\n%s", k,
			b, rem)
	}
	return
}

func True() B  { return B("true") }
func False() B { return B("false") }

func MarshalBool(src B, truth bool) B {
	if truth {
		return append(src, True()...)
	}
	return append(src, False()...)
}

func UnmarshalBool(src B) (rem B, truth bool, err error) {
	rem = src
	t, f := True(), False()
	for i := range rem {
		if rem[i] == t[0] {
			if len(rem) < i+len(t) {
				err = io.EOF
				return
			}
			if equals(t, rem[i:i+len(t)]) {
				truth = true
				rem = rem[i+len(t):]
				return
			}
		}
		if rem[i] == f[0] {
			if len(rem) < i+len(f) {
				err = io.EOF
				return
			}
			if equals(f, rem[i:i+len(f)]) {
				rem = rem[i+len(f):]
				return
			}
		}
	}
	// if a truth value is not found in the string it will run to the end
	err = io.EOF
	return
}

func Comma(b B) (rem B, err error) {
	rem = b
	for i := range rem {
		if rem[i] == ',' {
			rem = rem[i:]
			return
		}
	}
	err = io.EOF
	return
}
