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
	field []BS[B]
}

func (t *T) S(i int) (s S) {
	if t == nil {
		return
	}
	if t.Len() <= i {
		return
	}
	return S(t.field[i])
}

func (t *T) B(i int) (b B) {
	if t == nil {
		return
	}
	if t.Len() <= i {
		return
	}
	return B(t.field[i])
}

func (t *T) F() (b []B) {
	if t == nil {
		return []B{}
	}
	b = make([]B, t.Len())
	for i := range t.field {
		b[i] = t.B(i)
	}
	return
}

func (t *T) Len() int {
	if t == nil {
		return 0
	}
	return len(t.field)
}

func (t *T) Less(i, j int) bool {
	var cursor N
	for len(t.field[i]) < cursor-1 && len(t.field[j]) < cursor-1 {
		if bytes.Compare(t.field[i], t.field[j]) < 0 {
			return true
		}
		cursor++
	}
	return false
}

func (t *T) Swap(i, j int) { t.field[i], t.field[j] = t.field[j], t.field[i] }

func NewWithCap(c int) *T { return &T{make([]BS[B], 0, c)} }

func New[V S | B](fields ...V) (t *T) {
	t = &T{field: make([]BS[B], len(fields))}
	for i, field := range fields {
		t.field[i] = B(field)
	}
	return
}

func FromBytesSlice(fields ...B) (t *T) {
	t = &T{field: make([]BS[B], len(fields))}
	for i, field := range fields {
		t.field[i] = field
	}
	return
}

// Clone makes a new tag.T with the same members.
func (t *T) Clone() (c *T) {
	c = &T{field: make([]BS[B], 0, len(t.field))}
	for _, f := range t.field {
		l := len(f)
		b := make([]byte, l)
		copy(b, f)
		c.field = append(c.field, b)
	}
	return
}

func (t *T) Append(b ...B) (tt *T) {
	if t == nil {
		// we are propagating back this to tt if t was nil, else it appends
		t = &T{make([]BS[B], 0, len(t.field))}
	}
	for _, bb := range b {
		t.field = append(t.field, bb)
	}
	return t
}
func (t *T) Cap() int                { return cap(t.field) }
func (t *T) Clear()                  { t.field = t.field[:0] }
func (t *T) Slice(start, end int) *T { return &T{t.field[start:end]} }

func (t *T) ToByteSlice() (b []B) {
	for i := range t.field {
		b = append(b, t.field[i])
	}
	return
}

func (t *T) ToStringSlice() (b []S) {
	b = make([]S, 0, len(t.field))
	for i := range t.field {
		b = append(b, S(t.field[i]))
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
		if !equals(prefix.field[i], t.field[i]) {
			return false
		}
	}
	// check last element just for a prefix
	return bytes.HasPrefix(t.field[prefixLen-1], prefix.field[prefixLen-1])
}

// Key returns the first element of the tags.
func (t *T) Key() B {
	if t == nil {
		return nil
	}
	if t.Len() > Key {
		return t.field[Key]
	}
	return nil
}

// FilterKey returns the first element of a filter tag (the key) with the # removed
func (t *T) FilterKey() B {
	if t == nil {
		return nil
	}
	if len(t.field) > Key {
		return t.field[Key][1:]
	}
	return nil
}

// Value returns the second element of the tag.
func (t *T) Value() B {
	if t == nil {
		return nil
	}
	if len(t.field) > Value {
		return t.field[Value]
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
		len(t.field) >= Relay {

		return normalize.URL(B(t.field[Relay]))
	}
	return
}

// MarshalJSON appends the JSON form to the passed bytes.
func (t *T) MarshalJSON(dst B) (b B, err error) {
	dst = append(dst, '[')
	for i, s := range t.field {
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
			t.field = append(t.field, text.NostrUnescape(b[quoteStart:i]))
		}
	}
	if !openedBracket || inQuotes {
		log.I.F("\n%v\n%s", t, r)
		return nil, errorf.E("tag: failed to parse tag")
	}
	log.I.S(t.field)
	return
}

// func (t *T) String() string {
// 	b, _ := t.MarshalJSON(nil)
// 	return unsafe.String(&b[0], len(b))
// }

// Contains returns true if the provided element is found in the tag slice.
func (t *T) Contains(s B) (b bool) {
	// var isHex bool
	// if t.Len() > 1 && (t.S(0) == "e" || t.S(0) == "p") {
	// 	isHex = true
	// }
	// if isHex {
	// 	o := "contains,["
	// 	for _, i := range t.field {
	// 		o += "\""
	// 		o += hex.Enc(i)
	// 		o += "\","
	// 	}
	// } else {
	// 	log.I.F("contains %v\n%v", s, t.field)
	// }
	for i := range t.field {
		if equals(t.field[i], s) {
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
		if !equals(t.field[i], ta.field[i]) {
			return false
		}
	}
	return true
}
