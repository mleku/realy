package envelopes

// Identify handles determining what kind of codec.Envelope is, by the Label,
// the first step in identifying the structure of the message. This first step
// is not sufficient because the same labels are used on several codec.Envelope
// types in the nostr specification. The rest of the context is in whether this
// is a client or a relay receiving it.
func Identify(b []byte) (t string, rem []byte, err error) {
	var openBrackets, openQuotes, afterQuotes bool
	var label []byte
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
						t = string(label)
						afterQuotes = true
						break
					}
				}
			}
		}
	}
	return
}
