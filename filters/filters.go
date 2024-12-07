package filters

import (
	"realy.lol/event"
	"realy.lol/filter"
)

type T struct {
	F []*filter.T
}

func Make(l no) *T { return &T{F: make([]*filter.T, l)} }

func (f *T) GetFingerprints() (fps []uint64, err er) {
	for _, ff := range f.F {
		var fp uint64
		if fp, err = ff.Fingerprint(); chk.E(err) {
			continue
		}
		fps = append(fps, fp)
	}
	return
}

func (f *T) Len() no { return len(f.F) }

func New(ff ...*filter.T) (f *T) { return &T{F: ff} }

func (f *T) Match(event *event.T) bo {
	for _, f := range f.F {
		if f.Matches(event) {
			return true
		}
	}
	return false
}

func (f *T) String() (s st) { return st(f.Marshal(nil)) }

func (f *T) Marshal(dst by) (b by) {
	var err er
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

func (f *T) Unmarshal(b by) (r by, err er) {
	r = b[:]
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

func GenFilters(n no) (ff *T, err er) {
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
