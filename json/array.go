package json

// An Array is an ordered list of values.
type Array struct{ V []I }

func (a *Array) Marshal(dst B) (b B) {
	b = dst
	b = append(b, '[')
	last := len(a.V) - 1
	for i, v := range a.V {
		b = v.Marshal(b)
		if i != last {
			b = append(b, ',')
		}
	}
	b = append(b, ']')
	return
}

func (a *Array) Unmarshal(dst B) (rem B, err E) {
	rem = dst
	var openBracket bool
	var element N
	for ; len(rem) > 0; rem = rem[1:] {
		if !openBracket && rem[0] == '[' {
			openBracket = true
			continue
		}
		if openBracket {
			if rem[0] == ',' {
				continue
			} else if rem[0] == ']' {
				rem = rem[1:]
				return
			}
			// element marshallers already know to skip until the known sign of the beginning of
			// their content, eg quotes, numerical value, etc.
			if rem, err = a.V[element].Unmarshal(rem); chk.E(err) {
				return
			}
			if rem[0] == ']' {
				rem = rem[1:]
				// done
				return
			}
		}
	}
	return
}
