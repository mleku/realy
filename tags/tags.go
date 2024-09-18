package tags

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"

	"realy.lol/lol"
	"realy.lol/tag"
)

// T is a list of T - which are lists of string elements with ordering and no
// uniqueness constraint (not a set).
type T struct {
	t []*tag.T
}

func New(fields ...*tag.T) (t *T) {
	// t = &T{T: make([]*tag.T, 0, len(fields))}
	t = &T{}
	for _, field := range fields {
		t.t = append(t.t, field)
	}
	return
}

func NewWithCap(c int) (t *T) {
	return &T{t: make([]*tag.T, 0, c)}
}

func (t *T) F() (tt []*tag.T) {
	if t == nil {
		return []*tag.T{tag.New(B{})}
	}
	return t.t
}

func (t *T) N(i int) (tt *tag.T) {
	if t == nil {
		return tag.NewWithCap(0)
	}
	if len(t.t) <= i {
		return tag.New(B{})
	}
	return t.t[i]
}

func (t *T) AppendTo(n int, b ...B) (tt *T) {
	if t == nil {
		log.E.S(t, b)
		return
	}
	if t.Len() < n+1 {
		log.E.F("cannot append to nonexistent tags field %d with tags len %d",
			n, t.Len())
		fmt.Fprint(os.Stderr, lol.GetNLoc(7))
		return
	}
	for _, bb := range b {
		t.N(n).Append(bb)
		// t.T[n].Field = append(t.T[n].Field, bb)
	}
	return t
}

func (t *T) AddCap(i, c int) (tt *T) {
	if t == nil {
		log.E.F("cannot add capacity to index %d of nil tags", i)
		fmt.Fprint(os.Stderr, lol.GetNLoc(7))
		return t
	}
	if len(t.t) == 0 && i == 0 {
		t.t = append(t.t, tag.NewWithCap(c))
	}
	if len(t.t) == 1 && i == 1 {
		t.t = append(t.t, tag.NewWithCap(c))
	}
	if len(t.t) == 2 && i == 2 {
		t.t = append(t.t, tag.NewWithCap(c))
	}
	if len(t.t) <= i {
		log.I.Ln("len", t.Len(), "i", i)
		log.E.F("cannot add capacity to nonexistent tag field of tags %d of len %d",
			i, t.Len())
		fmt.Fprint(os.Stderr, lol.GetNLoc(7))
		return t
	}
	t.t[i] = tag.NewWithCap(c)
	return t
}

func (t *T) Value() (tt []*tag.T) {
	if t == nil {
		return []*tag.T{}
	}
	return t.t
}

func (t *T) ToStringSlice() (b [][]S) {
	b = make([][]S, 0, len(t.t))
	for i := range t.t {
		b = append(b, t.t[i].ToStringSlice())
	}
	return
}

func (t *T) Clone() (c *T) {
	c = &T{t: make([]*tag.T, len(t.t))}
	for i, field := range t.t {
		c.t[i] = field.Clone()
	}
	return
}

func (t *T) Equal(ta *T) bool {
	// sort them the same so if they are the same in content they compare the same.
	t1 := t.Clone()
	sort.Sort(t1)
	t2 := ta.Clone()
	sort.Sort(t2)
	for i := range t.t {
		if !t1.t[i].Equal(t2.t[i]) {
			return false
		}
	}
	return true
}

// Less returns which tag's first element is first lexicographically
func (t *T) Less(i, j int) (less bool) {
	a, b := t.t[i], t.t[j]
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
	t.t[i], t.t[j] = t.t[j], t.t[i]
}

func (t *T) Len() (l int) {
	if t == nil {
		return
	}
	if t.t != nil {
		return len(t.t)
	}
	return
}

// GetFirst gets the first tag in tags that matches the prefix, see [T.StartsWith]
func (t *T) GetFirst(tagPrefix *tag.T) *tag.T {
	for _, v := range t.t {
		if v.StartsWith(tagPrefix) {
			return v
		}
	}
	return nil
}

// GetLast gets the last tag in tags that matches the prefix, see [T.StartsWith]
func (t *T) GetLast(tagPrefix *tag.T) *tag.T {
	for i := len(t.t) - 1; i >= 0; i-- {
		v := t.t[i]
		if v.StartsWith(tagPrefix) {
			return v
		}
	}
	return nil
}

// GetAll gets all the tags that match the prefix, see [T.StartsWith]
func (t *T) GetAll(tagPrefix *tag.T) *T {
	result := &T{t: make([]*tag.T, 0, len(t.t))}
	for _, v := range t.t {
		if v.StartsWith(tagPrefix) {
			result.t = append(result.t, v)
		}
	}
	return result
}

// FilterOut removes all tags that match the prefix, see [T.StartsWith]
func (t *T) FilterOut(tagPrefix []B) *T {
	filtered := &T{t: make([]*tag.T, 0, len(t.t))}
	for _, v := range t.t {
		if !v.StartsWith(tag.New(tagPrefix...)) {
			filtered.t = append(filtered.t, v)
		}
	}
	return filtered
}

// AppendUnique appends a tag if it doesn't exist yet, otherwise does nothing.
// the uniqueness comparison is done based only on the first 2 elements of the
// tag.
func (t *T) AppendUnique(tag *tag.T) *T {
	n := tag.Len()
	if n > 2 {
		n = 2
	}
	if t.GetFirst(tag.Slice(0, n)) == nil {
		return &T{append(t.t, tag)}
	}
	return t
}

func (t *T) Append(ttt ...*T) (tt *T) {
	if t == nil {
		t = NewWithCap(len(ttt))
	}
	for _, tf := range ttt {
		for _, v := range tf.t {
			t.t = append(t.t, v)
		}
	}
	return t
}

func (t *T) AppendTags(ttt ...*tag.T) (tt *T) {
	if t == nil {
		t = NewWithCap(len(ttt))
	}
	t.t = append(t.t, ttt...)
	return t
}

// Scan parses a string or raw bytes that should be a string and embeds the values into the tags variable from which
// this method is invoked.
//
// todo: wut is this?
func (t *T) Scan(src any) (err error) {
	var jtags []byte
	switch v := src.(type) {
	case B:
		jtags = v
	case S:
		jtags = []byte(v)
	default:
		return errors.New("couldn't scan tag, it's not a json string")
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
		// if either are empty there can't be a match (if caller wants to know if both are empty
		// that's not the same as an intersection).
		return
	}
	matches := len(f.t)
	for _, v := range f.t {
		for _, w := range t.t {
			if equals(v.FilterKey(), w.Key()) {
				// we have a matching tag key, and both have a first field, check if tag has any
				// of the subsequent values in the filter tag.
				for _, val := range v.F()[1:] {
					if equals(val, w.Value()) {
						matches--
					}
				}
			}
		}
	}
	return matches == 0
}

// // ContainsAny returns true if any of the strings given in `values` matches any of the tag
// // elements.
// func (t *T) ContainsAny(tagName B, values *tag.T) bool {
// 	for _, v := range t.T {
// 		if v.Len() < 2 {
// 			continue
// 		}
// 		if !equals(v.Key(), tagName) {
// 			continue
// 		}
// 		for _, candidate := range values.Field {
// 			if equals(v.Value(), candidate) {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }
//
// func (t *T) Contains(filterTags *T) (has bool) {
// 	for _, v := range filterTags.T {
// 		if t.ContainsAny(v.FilterKey(), v) {
// 			return true
// 		}
// 	}
// 	return
// }

// MarshalTo appends the JSON encoded byte of T as [][]string to dst. String escaping is as described in RFC8259.
func (t *T) MarshalTo(dst B) []byte {
	dst = append(dst, '[')
	for i, tt := range t.t {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst, _ = tt.MarshalJSON(dst)
	}
	dst = append(dst, ']')
	return dst
}

// func (t *T) String() string {
// 	buf := new(bytes.Buffer)
// 	buf.WriteByte('[')
// 	last := len(t.T) - 1
// 	for i := range t.T {
// 		_, _ = fmt.Fprint(buf, t.T[i])
// 		if i < last {
// 			buf.WriteByte(',')
// 		}
// 	}
// 	buf.WriteByte(']')
// 	return buf.String()
// }

func (t *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b = append(b, '[')
	if t == nil || t.t == nil {
		b = append(b, ']')
		return
	}
	if len(t.t) == 0 {
		b = append(b, '[', ']')
	}
	for i, s := range t.t {
		if i > 0 {
			b = append(b, ',')
		}
		b, _ = s.MarshalJSON(b)
	}
	b = append(b, ']')
	return
}

func (t *T) UnmarshalJSON(b B) (r B, err error) {
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
				if r, err = tt.UnmarshalJSON(r); chk.E(err) {
					return
				}
				t.t = append(t.t, tt)
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
