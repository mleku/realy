package envelopes

func Identify(b by) (t st, rem by, err er) {
	var openBrackets, openQuotes, afterQuotes bo
	var label by
	rem = b
	for ; len(rem) > 0; rem = rem[1:] {
		if !openBrackets && rem[0] == '[' {
			openBrackets = true
		} else if openBrackets {
			if !openQuotes && rem[0] == '"' {
				openQuotes = true
			} else if afterQuotes {
				// return the remainder after the comma
				if rem[0] == ',' {
					rem = rem[1:]
					return
				}
			} else if openQuotes {
				for i := range rem {
					if rem[i] == '"' {
						label = rem[:i]
						rem = rem[i:]
						t = st(label)
						afterQuotes = true
						break
					}
				}
			}
		}
	}
	return
}
