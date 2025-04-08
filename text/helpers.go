package text

import (
	"bytes"
	"io"

	"github.com/templexxx/xhex"

	"realy.lol/hex"
)

// JSONKey generates the JSON format for an object key and terminates with the semicolon.
func JSONKey(dst, k []byte) (b []byte) {
	dst = append(dst, '"')
	dst = append(dst, k...)
	dst = append(dst, '"', ':')
	b = dst
	return
}

// UnmarshalHex takes a byte string that should contain a quoted hexadecimal encoded value,
// decodes it in-place using a SIMD hex codec and returns the decoded truncated bytes (the other
// half will be as it was but no allocation is required).
func UnmarshalHex(b []byte) (h []byte, rem []byte, err error) {
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
func UnmarshalQuoted(b []byte) (content, rem []byte, err error) {
	if len(b) == 0 {
		err = io.EOF
		return
	}
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
			if !escaping {
				escaping = true
				contentLen++
				rem = rem[1:]
			} else {
				escaping = false
				contentLen++
				rem = rem[1:]
			}
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

func MarshalHexArray(dst []byte, ha [][]byte) (b []byte) {
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

// UnmarshalHexArray unpacks a JSON array containing strings with hexadecimal, and checks all
// values have the specified byte size.
func UnmarshalHexArray(b []byte, size int) (t [][]byte, rem []byte, err error) {
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
				var h []byte
				if h, rem, err = UnmarshalHex(rem); chk.E(err) {
					return
				}
				if len(h) != size {
					err = errorf.E("invalid hex array size, got %d expect %d",
						2*len(h), 2*size)
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
func UnmarshalStringArray(b []byte) (t [][]byte, rem []byte, err error) {
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
				var h []byte
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

func True() []byte  { return []byte("true") }
func False() []byte { return []byte("false") }

func MarshalBool(src []byte, truth bool) []byte {
	if truth {
		return append(src, True()...)
	}
	return append(src, False()...)
}

func UnmarshalBool(src []byte) (rem []byte, truth bool, err error) {
	rem = src
	t, f := True(), False()
	for i := range rem {
		if rem[i] == t[0] {
			if len(rem) < i+len(t) {
				err = io.EOF
				return
			}
			if bytes.Equal(t, rem[i:i+len(t)]) {
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
			if bytes.Equal(f, rem[i:i+len(f)]) {
				rem = rem[i+len(f):]
				return
			}
		}
	}
	// if a truth value is not found in the string it will run to the end
	err = io.EOF
	return
}

func Comma(b []byte) (rem []byte, err error) {
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
