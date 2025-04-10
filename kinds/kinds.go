// Package kinds is a set of helpers for dealing with lists of kind numbers
// including comparisons and encoding.
package kinds

import (
	"realy.mleku.dev/ints"
	"realy.mleku.dev/kind"
)

// T is an array of kind.T, used in filter.T and filter.S for searches.
type T struct {
	K []*kind.T
}

// New creates a new kinds.T, if no parameter is given it just creates an empty zero kinds.T.
func New(k ...*kind.T) *T { return &T{k} }

// NewWithCap creates a new empty kinds.T with a given slice capacity.
func NewWithCap(c int) *T { return &T{make([]*kind.T, 0, c)} }

// FromIntSlice converts a []int into a kinds.T.
func FromIntSlice(is []int) (k *T) {
	k = &T{}
	for i := range is {
		k.K = append(k.K, kind.New(uint16(is[i])))
	}
	return
}

// Len returns the number of elements in a kinds.T.
func (k *T) Len() (l int) {
	if k == nil {
		return
	}
	return len(k.K)
}

// Less returns which of two elements of a kinds.T is lower.
func (k *T) Less(i, j int) bool { return k.K[i].K < k.K[j].K }

// Swap switches the position of two kinds.T elements.
func (k *T) Swap(i, j int) {
	k.K[i].K, k.K[j].K = k.K[j].K, k.K[i].K
}

// ToUint16 returns a []uint16 version of the kinds.T.
func (k *T) ToUint16() (o []uint16) {
	for i := range k.K {
		o = append(o, k.K[i].ToU16())
	}
	return
}

// Clone makes a new kind.T with the same members.
func (k *T) Clone() (c *T) {
	c = &T{K: make([]*kind.T, len(k.K))}
	for i := range k.K {
		c.K[i] = k.K[i]
	}
	return
}

// Contains returns true if the provided element is found in the kinds.T.
//
// Note that the request must use the typed kind.T or convert the number thus.
// Even if a custom number is found, this codebase does not have the logic to
// deal with the kind so such a search is pointless and for which reason static
// typing always wins. No mistakes possible with known quantities.
func (k *T) Contains(s *kind.T) bool {
	for i := range k.K {
		if k.K[i].Equal(s) {
			return true
		}
	}
	return false
}

// Equals checks that the provided kind.T matches.
func (k *T) Equals(t1 *T) bool {
	if len(k.K) != len(t1.K) {
		return false
	}
	for i := range k.K {
		if k.K[i] != t1.K[i] {
			return false
		}
	}
	return true
}

// Marshal renders the kinds.T into a JSON array of integers.
func (k *T) Marshal(dst []byte) (b []byte) {
	b = dst
	b = append(b, '[')
	for i := range k.K {
		b = k.K[i].Marshal(b)
		if i != len(k.K)-1 {
			b = append(b, ',')
		}
	}
	b = append(b, ']')
	return
}

// Unmarshal decodes a provided JSON array of integers into a kinds.T.
func (k *T) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	var openedBracket bool
	for ; len(r) > 0; r = r[1:] {
		if !openedBracket && r[0] == '[' {
			openedBracket = true
			continue
		} else if openedBracket {
			if r[0] == ']' {
				// done
				return
			} else if r[0] == ',' {
				continue
			}
			kk := ints.New(0)
			if r, err = kk.Unmarshal(r); chk.E(err) {
				return
			}
			k.K = append(k.K, kind.New(kk.Uint16()))
			if r[0] == ']' {
				r = r[1:]
				return
			}
		}
	}
	if !openedBracket {
		log.I.F("\n%v\n%s", k, r)
		return nil, errorf.E("kinds: failed to unmarshal\n%s\n%s\n%s", k,
			b, r)
	}
	return
}

// IsPrivileged returns true if any of the elements of a kinds.T are privileged (ie, they should
// be privacy protected).
func (k *T) IsPrivileged() (priv bool) {
	for i := range k.K {
		if k.K[i].IsPrivileged() {
			return true
		}
	}
	return
}
