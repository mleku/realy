// Package filters is a set of tools for working with multiple nostr filters.
package filters

import (
	"realy.lol/event"
	"realy.lol/filter"
)

// T is a wrapper around an array of pointers to filter.T.
type T struct {
	F []*filter.T
}

// Make a new filters.T.
func Make(l int) *T { return &T{F: make([]*filter.T, l)} }

// GetFingerprints returns a collection of fingerprints (64 bit digest) of a set of filters.T.
func (f *T) GetFingerprints() (fps []uint64, err error) {
	for _, ff := range f.F {
		var fp uint64
		if fp, err = ff.Fingerprint(); chk.E(err) {
			continue
		}
		fps = append(fps, fp)
	}
	return
}

// Len returns the number of elements in a filters.T.
func (f *T) Len() int { return len(f.F) }

// New creates a new filters.T out of a variadic list of filter.T.
func New(ff ...*filter.T) (f *T) { return &T{F: ff} }

// Match checks if a set of filters.T matches on an event.T.
func (f *T) Match(event *event.T) bool {
	for _, f := range f.F {
		if f.Matches(event) {
			return true
		}
	}
	return false
}

// String returns a canonical sorted string of a slice of filters.T.
func (f *T) String() (s string) {
	return string(f.Marshal(nil))
}

// Marshal a filters.T into raw bytes, and append it to a provided slice, and return the result.
func (f *T) Marshal(dst []byte) (b []byte) {
	var err error
	_ = err
	b = dst
	b = append(b, '[')
	end := len(f.F) - 1
	for i := range f.F {
		b = f.F[i].Marshal(b)
		if i < end {
			b = append(b, ',')
		}
	}
	b = append(b, ']')
	return
}

// Unmarshal a filters.T in JSON (minified) form and store it in the provided filters.T.
func (f *T) Unmarshal(b []byte) (r []byte, err error) {
	r = b[:]
	if len(r) < 1 {
		err = errorf.E("cannot unmarshal nothing")
		return
	}
	for len(r) > 0 {
		switch r[0] {
		case '[':
			if len(r) > 1 && r[1] == ']' {
				r = r[1:]
				return
			}
			r = r[1:]
			ffa := filter.New()
			if r, err = ffa.Unmarshal(r); chk.E(err) {
				return
			}
			f.F = append(f.F, ffa)
			// continue
		case ',':
			r = r[1:]
			if len(r) > 1 && r[1] == ']' {
				r = r[1:]
				return
			}
			ffa := filter.New()
			if r, err = ffa.Unmarshal(r); chk.E(err) {
				return
			}
			f.F = append(f.F, ffa)
		// next
		case ']':
			r = r[1:]
			// the end
			return
		}
	}
	return
}

// GenFilters creates an arbitrary number of fake filters for tests.
func GenFilters(n int) (ff *T, err error) {
	ff = &T{}
	for _ = range n {
		var f *filter.T
		if f, err = filter.GenFilter(); chk.E(err) {
			return
		}
		ff.F = append(ff.F, f)
	}
	return
}
