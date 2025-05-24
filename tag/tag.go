// Package tag provides an implementation of a nostr tag list, an array of
// strings with a usually single letter first "key" field, including methods to
// compare, marshal/unmarshal and access elements with their proper semantics.
package tag

import (
	"bytes"

	"golang.org/x/exp/constraints"

	"realy.lol/errorf"
	"realy.lol/log"
	"realy.lol/lol"
	"realy.lol/normalize"
	"realy.lol/text"
)

// The tag position meanings, so they are clear when reading.
const (
	Key = iota
	Value
	Relay
)

// T marker strings for e (reference) tags.
const (
	MarkerReply   = "reply"
	MarkerRoot    = "root"
	MarkerMention = "mention"
)

// BS is an abstract data type that can process strings and byte slices as byte slices.
type BS[Z []byte | string] []byte

// T is a list of strings with a literal ordering.
//
// Not a set, there can be repeating elements.
type T struct {
	field []BS[[]byte]
}

// New creates a new tag.T from a variadic parameter that can be either string or byte slice.
func New[V string | []byte](fields ...V) (t *T) {
	t = &T{field: make([]BS[[]byte], len(fields))}
	for i, field := range fields {
		t.field[i] = []byte(field)
	}
	return
}

// NewWithCap creates a new empty tag.T with a pre-allocated capacity for some number of fields.
func NewWithCap[V constraints.Integer](c V) *T { return &T{make([]BS[[]byte], 0, c)} }

// S returns a field of a tag.T as a string.
func (t *T) S(i int) (s string) {
	if t == nil {
		return
	}
	if t.Len() <= i {
		return
	}
	return string(t.field[i])
}

// B returns a field of a tag.T as a byte slice.
func (t *T) B(i int) (b []byte) {
	if t == nil {
		return
	}
	if t.Len() <= i {
		return
	}
	return t.field[i]
}

// Len returns the number of elements in a tag.T.
func (t *T) Len() int {
	if t == nil {
		return 0
	}
	return len(t.field)
}

// Less returns whether one field of a tag.T is lexicographically less than another (smaller).
// This uses bytes.Compare, which sorts strings and byte slices as though they are numbers.
func (t *T) Less(i, j int) bool {
	// Added nil checks for robustness
	if t == nil || i < 0 || j < 0 || i >= t.Len() || j >= t.Len() {
		return false // Or panic, depending on desired error handling
	}
	return bytes.Compare(t.field[i], t.field[j]) < 0
}

// Swap flips the position of two fields of a tag.T with each other.
func (t *T) Swap(i, j int) { t.field[i], t.field[j] = t.field[j], t.field[i] }

// FromBytesSlice creates a tag.T from a slice of slice of bytes.
func FromBytesSlice(fields ...[]byte) (t *T) {
	t = &T{field: make([]BS[[]byte], len(fields))}
	for i, field := range fields {
		t.field[i] = field
	}
	return
}

// Clone makes a new tag.T with the same members.
func (t *T) Clone() (c *T) {
	if t == nil {
		log.I.F("nil tag %s", lol.GetNLoc(7)) // This line is present in the `tags.go` code.
		return nil                            // Or return &T{} or panic, depending on desired behavior for nil receiver
	}
	c = &T{field: make([]BS[[]byte], 0, len(t.field))}
	for _, f := range t.field {
		l := len(f)
		b := make([]byte, l)
		copy(b, f)
		c.field = append(c.field, b)
	}
	return
}

// Append a slice of slice of bytes to a tag.T.
func (t *T) Append(b ...[]byte) (tt *T) {
	tt = t
	if t == nil {
		// we are propagating back this to tt if t was nil, else it appends
		tt = &T{}
	}
	for _, bb := range b {
		tt.field = append(tt.field, bb)
	}
	return
}

// Cap returns the capacity of a tag.T (how much elements it can hold without a re-allocation).
func (t *T) Cap() int { return cap(t.field) }

// Clear sets the length of the tag.T to zero so new elements can be appended.
func (t *T) Clear() { t.field = t.field[:0] }

// Slice cuts out a given start and end (exclusive) segment of the tag.T. This
// function must be called after using the Len function to ensure the `end`
// parameter does not exceed the bounds of the array.
func (t *T) Slice(start, end int) *T {
	return &T{t.field[start:end]}
}

// ToSliceOfBytes renders a tag.T as a slice of slice of bytes.
func (t *T) ToSliceOfBytes() (b [][]byte) {
	if t == nil {
		return [][]byte{}
	}
	b = make([][]byte, t.Len())
	for i := range t.field {
		b[i] = t.B(i)
	}
	return
}

// ToStringSlice converts a tag.T to a slice of strings.
func (t *T) ToStringSlice() (b []string) {
	b = make([]string, 0, len(t.field))
	for i := range t.field {
		b = append(b, string(t.field[i]))
	}
	return
}

// StartsWith checks a tag has the same initial set of elements.
//
// The last element is treated specially in that it is considered to match if
// the candidate has the same initial substring as its corresponding element.
func (t *T) StartsWith(prefix *T) bool {
	// log.I.S("StartsWith", prefix)
	prefixLen := len(prefix.field)

	if prefixLen > len(t.field) {
		return false
	}
	// check initial elements for equality
	for i := 0; i < prefixLen-1; i++ {
		if !bytes.Equal(prefix.field[i], t.field[i]) {
			return false
		}
	}
	// check last element just for a prefix
	return bytes.HasPrefix(t.field[prefixLen-1], prefix.field[prefixLen-1])
}

// Key returns the first element of the tags.
func (t *T) Key() []byte {
	if t == nil {
		return nil
	}
	if len(t.field) > Key {
		return t.field[Key]
	}
	return nil
}

// KeyString returns the first element of the tags as a string.
func (t *T) KeyString() string {
	if t == nil {
		return ""
	}
	// Get the first element.
	keyElement := t.field[Key]
	// Ensure the key element has at least two bytes to perform the slice [1:].
	// If it has 0 or 1 byte, slicing from index 1 will cause a panic or unexpected behavior.
	// A common pattern for filter keys is like "#e", "#p", so they should be at least 2 chars.
	if len(keyElement) >= 2 {
		return string(keyElement[1:])
	}
	// If the key element is too short, return an empty slice or the original key,
	// depending on desired behavior. Returning nil or an empty slice seems safer
	// than panicking. The comment implies removing '#', so if it's not present
	// or the string is too short, an empty or original string could be returned.
	// Returning nil in this context is consistent with other nil returns in this package.
	return ""
}

// FilterKey returns the first element of a filter tag (the key) with the # removed
func (t *T) FilterKey() []byte {
	if t == nil {
		return nil
	}
	// Get the first element.
	keyElement := t.field[Key]
	// Ensure the key element has at least two bytes to perform the slice [1:].
	// If it has 0 or 1 byte, slicing from index 1 will cause a panic or unexpected behavior.
	// A common pattern for filter keys is like "#e", "#p", so they should be at least 2 chars.
	if len(keyElement) >= 2 {
		return keyElement[1:]
	}
	// If the key element is too short, return an empty slice or the original key,
	// depending on desired behavior. Returning nil or an empty slice seems safer
	// than panicking. The comment implies removing '#', so if it's not present
	// or the string is too short, an empty or original string could be returned.
	// Returning nil in this context is consistent with other nil returns in this package.
	return nil
}

// Value returns the second element of the tag.
func (t *T) Value() []byte {
	if t == nil {
		return nil
	}
	if len(t.field) > Value {
		return t.field[Value]
	}
	return nil
}

var etag, ptag = []byte("e"), []byte("p")

// Relay returns the third element of the tag.
func (t *T) Relay() (s []byte) {
	if t == nil {
		return nil
	}
	// Check if the key is 'e' or 'p' and if there are enough fields for the Relay.
	if (bytes.Equal(t.Key(), etag) ||
		bytes.Equal(t.Key(), ptag)) &&
		len(t.field) > Relay {

		return normalize.URL([]byte(t.field[Relay]))
	}
	return
}

// Marshal encodes a tag.T as standard minified JSON array of strings.
func (t *T) Marshal(dst []byte) (b []byte) {
	if t == nil {
		// A nil tag should marshal to an empty JSON array.
		return append(dst, []byte("[]")...)
	}
	dst = append(dst, '[')
	for i, s := range t.field {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = text.AppendQuote(dst, s, text.NostrEscape)
	}
	dst = append(dst, ']')
	return dst
}

// Unmarshal decodes a standard minified JSON array of strings to a tags.T.
func (t *T) Unmarshal(b []byte) (r []byte, err error) {
	var inQuotes, openedBracket bool
	var quoteStart int
	t.field = []BS[[]byte]{} // Clear the field to ensure a fresh unmarshal

	for i := 0; i < len(b); i++ {
		if !openedBracket {
			if b[i] == '[' {
				openedBracket = true
				if i+1 == len(b) { // Handle empty array "[]" if it's the end of input
					return nil, nil // No remaining bytes, no error
				}
				continue // Move to the next character after '['
			} else {
				// If we haven't opened a bracket yet and current char isn't '[', it's an error.
				return nil, errorf.E("tag: failed to parse tag: expected opening bracket '['")
			}
		}

		// We are inside the bracket now
		if !inQuotes {
			switch b[i] {
			case '"':
				inQuotes, quoteStart = true, i+1
			case ']':
				// Found the closing bracket. Return the remaining bytes after it.
				return b[i+1:], nil // Correctly return remaining bytes and no error
			case ',':
				// Expecting a comma only if we've already parsed at least one tag.
				// This case covers a comma before the first element or multiple commas.
				if len(t.field) == 0 {
					return nil, errorf.E("tag: failed to parse tag: unexpected comma before first element")
				}
			case ' ':
				// Allow spaces outside quotes but within the array structure
				continue
			default:
				// Unexpected character outside of quotes, e.g., "invalid"
				return nil, errorf.E("tag: failed to parse tag: unexpected character '%c' outside quotes", b[i])
			}
		} else { // In quotes
			if b[i] == '\\' && i < len(b)-1 {
				i++ // Skip escaped character
			} else if b[i] == '"' {
				inQuotes = false
				t.field = append(t.field, text.NostrUnescape(b[quoteStart:i]))
			}
			// If it's not '\' or '"', just continue as it's part of the string content.
		}
	}

	// If we reach here, it means we didn't find a closing bracket or are still in quotes
	// when the input ended. This indicates an incomplete or malformed tag.
	if inQuotes {
		return nil, errorf.E("tag: failed to parse tag: unclosed quote")
	}
	if openedBracket {
		return nil, errorf.E("tag: failed to parse tag: unclosed bracket")
	}
	return nil, errorf.E("tag: failed to parse tag: unexpected end of input")
}

// Contains returns true if the provided element is found in the tag slice.
func (t *T) Contains(s []byte) (b bool) {
	if t == nil {
		return false // A nil tag list cannot contain any elements.
	}
	for i := range t.field {
		if bytes.Equal(t.field[i], s) {
			return true
		}
	}
	return false
}

// Equal checks that the provided tag list matches.
func (t *T) Equal(ta *T) bool {
	// Handle nil cases:
	// If both are nil, they are equal.
	if t == nil && ta == nil {
		return true
	}
	// If one is nil and the other is not, they are not equal.
	if t == nil || ta == nil {
		return false
	}
	if len(t.field) != len(ta.field) {
		return false
	}
	for i := range t.field {
		if !bytes.Equal(t.field[i], ta.field[i]) {
			return false
		}
	}
	return true
}
