package kinds

import (
	"realy.lol/ints"
	"realy.lol/kind"
)

type T struct {
	K []*kind.T
}

func New(k ...*kind.T) *T { return &T{k} }
func NewWithCap(c int) *T { return &T{make([]*kind.T, 0, c)} }

func FromIntSlice(is []int) (k *T) {
	k = &T{}
	for i := range is {
		k.K = append(k.K, kind.New(uint16(is[i])))
	}
	return
}

func (k *T) Len() (l int) {
	if k == nil {
		return
	}
	return len(k.K)
}

func (k *T) Less(i, j int) bool { return k.K[i].K < k.K[j].K }

func (k *T) Swap(i, j int) {
	k.K[i].K, k.K[j].K = k.K[j].K, k.K[i].K
}

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

func (k *T) IsPrivileged() (priv bool) {
	for i := range k.K {
		if k.K[i].IsPrivileged() {
			return true
		}
	}
	return
}
