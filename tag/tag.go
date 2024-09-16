package tag

import (
	"bytes"

	"realy.lol/normalize"
	"realy.lol/text"
)

// The tag position meanings so they are clear when reading.
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

type BS[Z B | S] B

// T is a list of strings with a literal ordering.
//
// Not a set, there can be repeating elements.
type T struct {
	Field []BS[B]
}

func (t *T) Len() int { return len(t.Field) }

func (t *T) Less(i, j int) bool {
	var cursor N
	for len(t.Field[i]) < cursor-1 && len(t.Field[j]) < cursor-1 {
		if bytes.Compare(t.Field[i], t.Field[j]) < 0 {
			return true
		}
		cursor++
	}
	return false
}

func (t *T) Swap(i, j int) { t.Field[i], t.Field[j] = t.Field[j], t.Field[i] }

func NewWithCap(c int) *T { return &T{make([]BS[B], 0, c)} }

func New[V S | B](fields ...V) (t *T) {
	t = &T{Field: make([]BS[B], len(fields))}
	for i, field := range fields {
		t.Field[i] = B(field)
	}
	return
}

func FromBytesSlice(fields ...B) (t *T) {
	t = &T{Field: make([]BS[B], len(fields))}
	for i, field := range fields {
		t.Field[i] = field
	}
	return
}

// Clone makes a new tag.T with the same members.
func (t *T) Clone() (c *T) {
	c = &T{Field: make([]BS[B], 0, len(t.Field))}
	for _, f := range t.Field {
		l := len(f)
		b := make([]byte, l)
		copy(b, f)
		c.Field = append(c.Field, b)
	}
	return
}

func (t *T) Append(b B)              { t.Field = append(t.Field, b) }
func (t *T) Cap() int                { return cap(t.Field) }
func (t *T) Clear()                  { t.Field = t.Field[:0] }
func (t *T) Slice(start, end int) *T { return &T{t.Field[start:end]} }

func (t *T) ToByteSlice() (b []B) {
	for i := range t.Field {
		b = append(b, t.Field[i])
	}
	return
}

func (t *T) ToStringSlice() (b []S) {
	b = make([]S, 0, len(t.Field))
	for i := range t.Field {
		b = append(b, S(t.Field[i]))
	}
	return
}

// StartsWith checks a tag has the same initial set of elements.
//
// The last element is treated specially in that it is considered to match if
// the candidate has the same initial substring as its corresponding element.
func (t *T) StartsWith(prefix *T) bool {
	prefixLen := len(prefix.Field)

	if prefixLen > len(t.Field) {
		return false
	}
	// check initial elements for equality
	for i := 0; i < prefixLen-1; i++ {
		if !equals(prefix.Field[i], t.Field[i]) {
			return false
		}
	}
	// check last element just for a prefix
	return bytes.HasPrefix(t.Field[prefixLen-1], prefix.Field[prefixLen-1])
}

// Key returns the first element of the tags.
func (t *T) Key() B {
	if t == nil {
		return nil
	}
	if len(t.Field) > Key {
		return t.Field[Key]
	}
	return nil
}

// FilterKey returns the first element of a filter tag (the key) with the # removed
func (t *T) FilterKey() B {
	if t == nil {
		return nil
	}
	if len(t.Field) > Key {
		return t.Field[Key][1:]
	}
	return nil
}

// Value returns the second element of the tag.
func (t *T) Value() B {
	if t == nil {
		return nil
	}
	if len(t.Field) > Value {
		return t.Field[Value]
	}
	return nil
}

var etag, ptag = B("e"), B("p")

// Relay returns the third element of the tag.
func (t *T) Relay() (s B) {
	if t == nil {
		return nil
	}
	if (equals(t.Key(), etag) ||
		equals(t.Key(), ptag)) &&
		len(t.Field) >= Relay {

		return normalize.URL(B(t.Field[Relay]))
	}
	return
}

// MarshalJSON appends the JSON form to the passed bytes.
func (t *T) MarshalJSON(dst B) (b B, err error) {
	dst = append(dst, '[')
	for i, s := range t.Field {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = text.AppendQuote(dst, s, text.NostrEscape)
	}
	dst = append(dst, ']')
	return dst, err
}

// UnmarshalJSON decodes the provided JSON tag list (array of strings), and
// returns any remainder after the close bracket has been encountered.
func (t *T) UnmarshalJSON(b B) (r B, err error) {
	var inQuotes, openedBracket bool
	var quoteStart int
	// t.Field = []BS[B]{}
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
			t.Field = append(t.Field, text.NostrUnescape(b[quoteStart:i]))
		}
	}
	if !openedBracket || inQuotes {
		log.I.F("\n%v\n%s", t, r)
		return nil, errorf.E("tag: failed to parse tag")
	}
	log.I.S(t.Field)
	return
}

// func (t *T) String() string {
// 	b, _ := t.MarshalJSON(nil)
// 	return unsafe.String(&b[0], len(b))
// }

// Contains returns true if the provided element is found in the tag slice.
func (t *T) Contains(s B) bool {
	for i := range t.Field {
		if equals(t.Field[i], s) {
			return true
		}
	}
	return false
}

// Equal checks that the provided tag list matches.
func (t *T) Equal(ta *T) bool {
	if len(t.Field) != len(ta.Field) {
		return false
	}
	for i := range t.Field {
		if !equals(t.Field[i], ta.Field[i]) {
			return false
		}
	}
	return true
}
