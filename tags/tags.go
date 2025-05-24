// Package tags provides a set of tools for composing and searching lists of
// tag.T as well as marshal/unmarshal to JSON.
package tags

import (
	"bytes"
	"encoding/json"
	"sort"

	"golang.org/x/exp/constraints"

	"realy.lol/chk"
	"realy.lol/errorf"
	"realy.lol/log"
	"realy.lol/lol"
	"realy.lol/tag"
)

// T is a list of tag.T - which are lists of string elements with ordering and no uniqueness
// constraint (not a set).
type T struct {
	element []*tag.T
}

// New creates a new tags.T with a provided list of tag.T. These can be created with tag.New for
// the elements or other methods like tag.FromBytesSlice.
func New(fields ...*tag.T) (t *T) {
	// t = &T{T: make([]*tag.T, 0, len(fields))}
	t = &T{}
	for _, field := range fields {
		t.element = append(t.element, field)
	}
	return
}

// NewWithCap creates a tags.T with space pre-allocated for a number of tag.T elements.
func NewWithCap[V constraints.Integer](c V) (t *T) {
	return &T{element: make([]*tag.T, 0, c)}
}

// ToSliceOfTags converts a tags.T to a slice of tag.T.
func (t *T) ToSliceOfTags() (tt []*tag.T) {
	if t == nil {
		return []*tag.T{tag.New([]byte{})}
	}
	return t.element
}

// GetTagElement returns a specific ordered element of a tags.T.
func (t *T) GetTagElement(i int) (tt *tag.T) {
	if t == nil {
		return tag.NewWithCap(0)
	}
	if len(t.element) <= i {
		return tag.New([]byte{})
	}
	return t.element[i]
}

// AppendTo adds a new tag, at a given position, from a slice of slice of bytes.
func (t *T) AppendTo(n int, b ...[]byte) (tt *T) {
	if t == nil {
		log.E.S(t, b)
		return
	}
	// Ensure t.element has enough elements up to index n
	for len(t.element) <= n {
		t.element = append(t.element, &tag.T{}) // Append empty tag.T instances
	}
	// Now, t.element[n] is a valid tag.T to append to
	currentTag := t.element[n]
	for _, bb := range b {
		currentTag.Append(bb) // Append to the existing tag.T
	}
	return t
}

// AppendSlice just appends a slice of slices of bytes into the tags. Like AppendTo
// but without the position specifier. todo: this is a terribly constructed API innit.
func (t *T) AppendSlice(b ...[]byte) (tt *T) {
	t.element = append(t.element, tag.New(b...))
	return
}

// AddCap adds extra capacity to a tags.T.
func (t *T) AddCap(i, c int) (tt *T) {
	if t == nil {
		log.E.F("cannot add capacity to index %d of nil tags", i)
		log.I.F("nil tags %s", lol.GetNLoc(7))
		return t
	}

	n := i - len(t.element) + 1
	for range n {
		t.element = append(t.element, &tag.T{})
	}
	return t
}

// ToStringsSlice converts a tags.T to a slice of slice of strings.
func (t *T) ToStringsSlice() (b [][]string) {
	if t == nil {
		// log.I.F("nil tags %s", lol.GetNLoc(7))
		return nil
	}
	b = make([][]string, 0, len(t.element))
	for i := range t.element {
		b = append(b, t.element[i].ToStringSlice())
	}
	return
}

// Clone makes a copy of all of the elements of a tags.T into a new tags.T.
func (t *T) Clone() (c *T) {
	if t == nil {
		log.I.F("nil tags %s", lol.GetNLoc(7))
		return t
	}
	c = &T{element: make([]*tag.T, len(t.element))}
	for i, field := range t.element {
		c.element[i] = field.Clone()
	}
	return
}

func (t *T) Equal(ta *T) bool {
	// Handle nil cases:
	if t == nil && ta == nil {
		log.I.F("nil tags %s", lol.GetNLoc(7))
		return true
	}
	if t == nil || ta == nil {
		log.I.F("nil tags %s", lol.GetNLoc(7))
		return false // One is nil, the other isn't, so they are not equal
	}
	// sort them the same so if they are the same in content they compare the same.
	t1 := t.Clone()
	sort.Sort(t1)
	t2 := ta.Clone()
	sort.Sort(t2)
	for i := range t.element {
		if !t1.element[i].Equal(t2.element[i]) {
			return false
		}
	}
	return true
}

// Less returns which tag's first element is first lexicographically
func (t *T) Less(i, j int) (less bool) {
	if t == nil {
		log.I.F("nil tags %s", lol.GetNLoc(7))
		return
	}
	a, b := t.element[i], t.element[j]
	if a.Len() < 1 && b.Len() < 1 {
		return false // they are equal
	}
	if a.Len() < 1 || b.Len() < 1 {
		return a.Len() < b.Len()
	}
	if bytes.Compare(a.Key(), b.Key()) < 0 {
		return true
	}
	return
}

func (t *T) Swap(i, j int) {
	t.element[i], t.element[j] = t.element[j], t.element[i]
}

func (t *T) Len() (l int) {
	if t == nil {
		return
	}
	if t.element != nil {
		return len(t.element)
	}
	return
}

// GetFirst gets the first tag in tags that matches the prefix, see [T.StartsWith]
func (t *T) GetFirst(tagPrefix *tag.T) *tag.T {
	if t == nil {
		log.I.F("caller provided nil tag %v", lol.GetNLoc(4))
		return nil
	}
	for _, v := range t.element {
		if v.StartsWith(tagPrefix) {
			return v
		}
	}
	return nil
}

// GetLast gets the last tag in tags that matches the prefix, see [T.StartsWith]
func (t *T) GetLast(tagPrefix *tag.T) *tag.T {
	if t == nil {
		log.I.F("caller provided nil tag %v", lol.GetNLoc(4))
		return nil
	}
	for i := len(t.element) - 1; i >= 0; i-- {
		v := t.element[i]
		if v.StartsWith(tagPrefix) {
			return v
		}
	}
	return nil
}

// GetAll gets all the tags that match the prefix, see [T.StartsWith]
func (t *T) GetAll(tagPrefix *tag.T) (result *T) {
	// log.I.S("GetAll", tagPrefix, t)
	if t == nil {
		log.I.F("caller provided nil tag %v", lol.GetNLoc(4))
		return nil
	}
	for _, v := range t.element {
		if v.StartsWith(tagPrefix) {
			if result == nil {
				result = &T{element: make([]*tag.T, 0, len(t.element))}
			}
			result.element = append(result.element, v)
		}
	}
	return
}

// FilterOut removes all tags that match the prefix, see [T.StartsWith]
func (t *T) FilterOut(tagPrefix [][]byte) *T {
	filtered := &T{element: make([]*tag.T, 0, len(t.element))}
	for _, v := range t.element {
		if !v.StartsWith(tag.New(tagPrefix...)) {
			filtered.element = append(filtered.element, v)
		}
	}
	return filtered
}

// AppendUnique appends a tag if it doesn't exist yet, otherwise does nothing.
// the uniqueness comparison is done based only on the first 2 elements of the
// tag.
func (t *T) AppendUnique(tag *tag.T) *T {
	if t == nil {
		log.I.F("caller provided nil tag %v", lol.GetNLoc(4))
		return nil
	}
	n := tag.Len()
	if n > 2 {
		n = 2
	}
	if t.GetFirst(tag.Slice(0, n)) == nil {
		return &T{append(t.element, tag)}
	}
	return t
}

func (t *T) Append(ttt ...*T) (tt *T) {
	if t == nil {
		log.I.F("caller provided nil tag %v", lol.GetNLoc(4))
		return
	}
	for _, tf := range ttt {
		for _, v := range tf.element {
			t.element = append(t.element, v)
		}
	}
	return t
}

func (t *T) AppendTags(ttt ...*tag.T) (tt *T) {
	if t == nil {
		log.I.F("caller provided nil tag %v", lol.GetNLoc(4))
		return
	}
	t.element = append(t.element, ttt...)
	return t
}

// Scan parses a string or raw bytes that should be a string and embeds the values into the tags
// variable from which this method is invoked.
//
// todo: wut is this?
func (t *T) Scan(src any) (err error) {
	var jtags []byte
	switch v := src.(type) {
	case []byte:
		jtags = v
	case string:
		jtags = []byte(v)
	default:
		return errorf.E("couldn't scan tag, it's not a json string")
	}
	err = json.Unmarshal(jtags, &t)
	chk.E(err)
	return
}

// Intersects returns true if a filter tags.T has a match. This means the second character of
// the filter tag key matches, (ignoring the stupid # prefix in the filter) and one of the
// following values in the tag matches the first tag of this tag.
func (t *T) Intersects(f *T) (has bool) {
	if t == nil || f == nil {
		log.I.F("caller provided nil tag %v", lol.GetNLoc(4))
		// if either are empty there can't be a match (if caller wants to know if both are empty
		// that's not the same as an intersection).
		return
	}
	matches := len(f.element)
	for _, v := range f.element {
		for _, w := range t.element {
			if bytes.Equal(v.FilterKey(), w.Key()) {
				// we have a matching tag key, and both have a first field, check if tag has any
				// of the subsequent values in the filter tag.
				for _, val := range v.ToSliceOfBytes()[1:] {
					if bytes.Equal(val, w.Value()) {
						matches--
					}
				}
			}
		}
	}
	return matches == 0
}

// ContainsProtectedMarker returns true if an event may only be published to the relay by a user
// authed with the same pubkey as in the event. This is for implementing relayinfo.NIP70.
func (t *T) ContainsProtectedMarker() (does bool) {
	if t == nil {
		log.I.F("caller provided nil tag %v", lol.GetNLoc(4))
		return false
	}
	for _, v := range t.element {
		if bytes.Equal(v.Key(), []byte("-")) {
			return true
		}
	}
	return
}

// ContainsAny returns true if any of the strings given in `values` matches any of the tag
// elements.
func (t *T) ContainsAny(tagName []byte, values *tag.T) bool {
	if t == nil {
		log.I.F("caller provided nil tag %v", lol.GetNLoc(4))
		return false
	}
	if len(tagName) < 1 {
		return false
	}
	for _, v := range t.element {
		if v.Len() < 2 {
			continue
		}
		if !bytes.Equal(v.Key(), tagName) {
			continue
		}
		for _, candidate := range values.ToSliceOfBytes() {
			if bytes.Equal(v.Value(), candidate) {
				return true
			}
		}
	}
	return false
}

func (t *T) Contains(filterTags *T) (has bool) {
	if t == nil {
		log.I.F("caller provided nil tag %v", lol.GetNLoc(4))
		return false
	}
	for _, v := range filterTags.element {
		if t.ContainsAny(v.FilterKey(), v) {
			return true
		}
	}
	return
}

// MarshalTo appends the JSON encoded byte of T as [][]string to dst. String escaping is as described in RFC8259.
func (t *T) MarshalTo(dst []byte) []byte {
	dst = append(dst, '[')
	for i, tt := range t.element {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = tt.Marshal(dst)
	}
	dst = append(dst, ']')
	return dst
}

// Marshal encodes a tags.T appended to a provided byte slice in JSON form.
func (t *T) Marshal(dst []byte) (b []byte) {
	b = dst
	b = append(b, '[') // Start with opening bracket for outer array

	if t == nil || t.element == nil || len(t.element) == 0 {
		b = append(b, ']') // For nil or empty, just append closing bracket
		return
	}

	for i, s := range t.element {
		if i > 0 {
			b = append(b, ',')
		}
		// Assume s.Marshal correctly marshals a single tag.T to a JSON array of strings
		b = s.Marshal(b)
	}
	b = append(b, ']') // End with closing bracket for outer array
	return
}

// Unmarshal a tags.T from a provided byte slice and return what remains after the end of the
// array.
func (t *T) Unmarshal(b []byte) (r []byte, err error) {
	r = b[:]
	for len(r) > 0 {
		switch r[0] {
		case '[':
			r = r[1:]
			goto inTags
		case ',':
			r = r[1:]
			// next
		case ']':
			r = r[1:]
			// the end
			return
		default:
			r = r[1:]
		}
	inTags:
		for len(r) > 0 {
			switch r[0] {
			case '[':
				tt := &tag.T{}
				if r, err = tt.Unmarshal(r); chk.E(err) {
					return
				}
				t.element = append(t.element, tt)
			case ',':
				r = r[1:]
				// next
			case ']':
				r = r[1:]
				// the end
				return
			default:
				r = r[1:]
			}
		}
	}
	return
}
