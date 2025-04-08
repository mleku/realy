// Package tag provides an implementation of a nostr tag list, an array of
// strings with a usually single letter first "key" field, including methods to
// compare, marshal/unmarshal and access elements with their proper semantics.
package tag

import (
	"bytes"

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
func NewWithCap(c int) *T { return &T{make([]BS[[]byte], 0, c)} }

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
	var cursor int
	for len(t.field[i]) < cursor-1 && len(t.field[j]) < cursor-1 {
		if bytes.Compare(t.field[i], t.field[j]) < 0 {
			return true
		}
		cursor++
	}
	return false
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
	if t == nil {
		// we are propagating back this to tt if t was nil, else it appends
		// t = &T{make([]ToSliceOfBytes[B], 0, len(t.field))}
		t = &T{}
	}
	for _, bb := range b {
		t.field = append(t.field, bb)
	}
	return t
}

// Cap returns the capacity of a tag.T (how much elements it can hold without a re-allocation).
func (t *T) Cap() int { return cap(t.field) }

// Clear sets the length of the tag.T to zero so new elements can be appended.
func (t *T) Clear() { t.field = t.field[:0] }

// Slice cuts out a given start and end (exclusive) segment of the tag.T.
func (t *T) Slice(start, end int) *T { return &T{t.field[start:end]} }

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
	if t.Len() > Key {
		return t.field[Key]
	}
	return nil
}

// FilterKey returns the first element of a filter tag (the key) with the # removed
func (t *T) FilterKey() []byte {
	if t == nil {
		return nil
	}
	if len(t.field) > Key {
		return t.field[Key][1:]
	}
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
	if (bytes.Equal(t.Key(), etag) ||
		bytes.Equal(t.Key(), ptag)) &&
		len(t.field) >= Relay {

		return normalize.URL([]byte(t.field[Relay]))
	}
	return
}

// Marshal encodes a tag.T as standard minified JSON array of strings.
func (t *T) Marshal(dst []byte) (b []byte) {
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
	// t.Field = []ToSliceOfBytes[B]{}
	for i := 0; i < len(b); i++ {
		if !openedBracket && b[i] == '[' {
			openedBracket = true
		} else if !inQuotes {
			if b[i] == '"' {
				inQuotes, quoteStart = true, i+1
			} else if b[i] == ']' {
				return b[i+1:], err
			}
		} else if b[i] == '\\' && i < len(b)-1 {
			i++
		} else if b[i] == '"' {
			inQuotes = false
			t.field = append(t.field, text.NostrUnescape(b[quoteStart:i]))
		}
	}
	if !openedBracket || inQuotes {
		return nil, errorf.E("tag: failed to parse tag")
	}
	log.I.S(t.field)
	return
}

// Contains returns true if the provided element is found in the tag slice.
func (t *T) Contains(s []byte) (b bool) {
	for i := range t.field {
		if bytes.Equal(t.field[i], s) {
			return true
		}
	}
	return false
}

// Equal checks that the provided tag list matches.
func (t *T) Equal(ta *T) bool {
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
