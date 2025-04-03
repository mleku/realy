package filter

import (
	"encoding/binary"
	"sort"

	"realy.lol/ec/schnorr"
	"realy.lol/event"
	"realy.lol/hex"
	"realy.lol/ints"
	"realy.lol/kinds"
	"realy.lol/realy/pointers"
	"realy.lol/sha256"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/text"
	"realy.lol/timestamp"
)

// S is a simplified filter that only covers the nip-01 REQ filter minus the
// separate and superseding Id list. The search field is from a different NIP,
// but it is a separate API for which reason it is also not here.
type S struct {
	Kinds   *kinds.T     `json:"kinds,omitempty"`
	Authors *tag.T       `json:"authors,omitempty"`
	Tags    *tags.T      `json:"-,omitempty"`
	Since   *timestamp.T `json:"since,omitempty"`
	Until   *timestamp.T `json:"until,omitempty"`
	Limit   *uint        `json:"limit,omitempty"`
}

// NewSimple creates a new, reasonably pre-allocated filter.S.
func NewSimple() (f *S) {
	return &S{
		Kinds:   kinds.NewWithCap(10),
		Authors: tag.NewWithCap(10),
		Tags:    tags.New(),
	}
}

// Clone creates a new filter with all the same elements in them, because they
// are immutable, basically, except setting the Limit field as 1, because it is
// used in the subscription management code to act as a reference counter, and
// making a clone implicitly means 1 reference.
func (f *S) Clone() (clone *S) {
	lim := new(uint)
	*lim = 1
	_Kinds := *f.Kinds
	_Authors := *f.Authors
	_Tags := *f.Tags.Clone()
	return &S{
		Kinds:   &_Kinds,
		Authors: &_Authors,
		Tags:    &_Tags,
	}
}

// Fingerprint returns an 8 byte truncated sha256 hash of the filter in the
// canonical form created by Marshal.
//
// The purpose of this fingerprint is enabling the creation of a map of filters
// that can be searched by making a canonical form of a provided input filter
// that will match the same filter. It achieves this by making all fields sorted
// in lexicographical order and from this a single 8 byte truncated hash can be
// used to identify if a received filter is the same filter.
func (f *S) Fingerprint() (fp uint64, err error) {
	var b []byte
	b = f.Marshal(b)
	h := sha256.Sum256(b)
	hb := h[:]
	fp = binary.LittleEndian.Uint64(hb)
	return
}

// Sort the fields of a filter so a fingerprint on a filter that has the same
// set of content produces the same fingerprint.
func (f *S) Sort() {
	if f.Kinds != nil {
		sort.Sort(f.Kinds)
	}
	if f.Authors != nil {
		sort.Sort(f.Authors)
	}
	if f.Tags != nil {
		sort.Sort(f.Tags)
	}
}

// Marshal a filter.S in canonical form with sorted fields.
func (f *S) Marshal(dst []byte) (b []byte) {
	var err error
	_ = err
	var first bool
	// sort the fields so they come out the same
	f.Sort()
	// open parentheses
	dst = append(dst, '{')
	if f.Kinds.Len() > 0 {
		if first {
			dst = append(dst, ',')
		} else {
			first = true
		}
		dst = text.JSONKey(dst, Kinds)
		dst = f.Kinds.Marshal(dst)
	}
	if f.Since != nil && f.Since.U64() > 0 {
		if first {
			dst = append(dst, ',')
		} else {
			first = true
		}
		dst = text.JSONKey(dst, Since)
		dst = f.Since.Marshal(dst)
	}
	if f.Until != nil && f.Until.U64() > 0 {
		if first {
			dst = append(dst, ',')
		} else {
			first = true
		}
		dst = text.JSONKey(dst, Until)
		dst = f.Until.Marshal(dst)
	}
	if pointers.Present(f.Limit) {
		if first {
			dst = append(dst, ',')
		} else {
			first = true
		}
		dst = text.JSONKey(dst, Limit)
		dst = ints.New(*f.Limit).Marshal(dst)
	}
	if f.Authors.Len() > 0 {
		if first {
			dst = append(dst, ',')
		} else {
			first = true
		}
		dst = text.JSONKey(dst, Authors)
		dst = text.MarshalHexArray(dst, f.Authors.ToByteSlice())
	}
	if f.Tags.Len() > 0 {
		// tags are stored as tags with the initial element the "#s" and the rest the
		// list in each element of the tags list. eg:
		//
		//     [["#p","<pubkey1>","<pubkey3"],["#t","hashtag","stuff"]]
		//
		for _, tg := range f.Tags.Value() {
			if tg == nil {
				// nothing here
				continue
			}
			if tg.Len() < 1 || len(tg.Key()) != 2 {
				// if there is no values, skip; the "key" field must be 2 characters long,
				continue
			}
			tKey := tg.F()[0]
			if tKey[0] != '#' &&
				(tKey[1] < 's' && tKey[1] > 'z' || tKey[1] < 'A' && tKey[1] > 'Z') {
				// first "key" field must begin with '#' and second be alpha
				continue
			}
			values := tg.F()[1:]
			if len(values) == 0 {
				continue
			}
			if first {
				dst = append(dst, ',')
			} else {
				first = true
			}
			// append the key
			dst = append(dst, '"', tg.B(0)[0], tg.B(0)[1], '"', ':')
			dst = append(dst, '[')
			for i, value := range values {
				dst = append(dst, '"')
				if tKey[1] == 'e' || tKey[1] == 'p' {
					// event and pubkey tags are binary 32 bytes
					dst = hex.EncAppend(dst, value)
				} else {
					dst = append(dst, value...)
				}
				dst = append(dst, '"')
				if i < len(values)-1 {
					dst = append(dst, ',')
				}
			}
			dst = append(dst, ']')
		}
	}
	// close parentheses
	dst = append(dst, '}')
	b = dst
	return
}

// Serialize a filter directly to raw bytes.
func (f *S) Serialize() (b []byte) { return f.Marshal(nil) }

// Unmarshal a filter from JSON (minified) form.
//
// todo: maybe this tolerates whitespace?
func (f *S) Unmarshal(b []byte) (r []byte, err error) {
	r = b[:]
	var key []byte
	var state int
	for ; len(r) >= 0; r = r[1:] {
		// log.I.F("%c", rem[0])
		switch state {
		case beforeOpen:
			if r[0] == '{' {
				state = openParen
				// log.I.Ln("openParen")
			}
		case openParen:
			if r[0] == '"' {
				state = inKey
				// log.I.Ln("inKey")
			}
		case inKey:
			if r[0] == '"' {
				state = inKV
				// log.I.Ln("inKV")
			} else {
				key = append(key, r[0])
			}
		case inKV:
			if r[0] == ':' {
				state = inVal
			}
		case inVal:
			if len(key) < 1 {
				err = errorf.E("filter key zero length: '%s'\n'%s", b, r)
				return
			}
			switch key[0] {
			case '#':
				k := make([]byte, len(key))
				copy(k, key)
				// r = r[1:]
				switch key[1] {
				case 'e', 'p':
					// the tags must all be 64 character hexadecimal
					var ff [][]byte
					if ff, r, err = text.UnmarshalHexArray(r,
						sha256.Size); chk.E(err) {
						return
					}
					ff = append([][]byte{k}, ff...)
					f.Tags = f.Tags.AppendTags(tag.New(ff...))
					// s.Tags.T = append(s.Tags.T, tag.New(ff...))
				default:
					// other types of tags can be anything
					var ff [][]byte
					if ff, r, err = text.UnmarshalStringArray(r); chk.E(err) {
						return
					}
					ff = append([][]byte{k}, ff...)
					f.Tags = f.Tags.AppendTags(tag.New(ff...))
					// s.Tags.T = append(s.Tags.T, tag.New(ff...))
				}
				state = betweenKV
			case Kinds[0]:
				if len(key) < len(Kinds) {
					goto invalid
				}
				f.Kinds = kinds.NewWithCap(0)
				if r, err = f.Kinds.Unmarshal(r); chk.E(err) {
					return
				}
				state = betweenKV
			case Authors[0]:
				if len(key) < len(Authors) {
					goto invalid
				}
				var ff [][]byte
				if ff, r, err = text.UnmarshalHexArray(r, schnorr.PubKeyBytesLen); chk.E(err) {
					return
				}
				f.Authors = tag.New(ff...)
				state = betweenKV
			default:
				goto invalid
			}
			key = key[:0]
		case betweenKV:
			if len(r) == 0 {
				return
			}
			if r[0] == '}' {
				state = afterClose
				// log.I.Ln("afterClose")
				// rem = rem[1:]
			} else if r[0] == ',' {
				state = openParen
				// log.I.Ln("openParen")
			} else if r[0] == '"' {
				state = inKey
				// log.I.Ln("inKey")
			}
		}
		if len(r) == 0 {
			return
		}
		if r[0] == '}' {
			r = r[1:]
			return
		}
	}
invalid:
	err = errorf.E("invalid key,\n'%s'\n'%s'", string(b), string(r))
	return
}

// Matches checks if a filter.S matches an event.
func (f *S) Matches(ev *event.T) bool {
	if ev == nil {
		// log.T.F("nil event")
		return false
	}
	if f.Kinds.Len() > 0 && !f.Kinds.Contains(ev.Kind) {
		// log.T.F("no matching kinds in filter\nEVENT %s\nFILTER %s", ev.ToObject().String(), s.ToObject().String())
		return false
	}
	if f.Authors.Len() > 0 && !f.Authors.Contains(ev.Pubkey) {
		// log.T.F("no matching authors in filter\nEVENT %s\nFILTER %s", ev.ToObject().String(), s.ToObject().String())
		return false
	}
	if f.Tags.Len() > 0 && !ev.Tags.Intersects(f.Tags) {
		return false
	}
	return true
}

// Equal checks if two filters are the same filter.
func (f *S) Equal(b *S) bool {
	f.Sort()
	b.Sort()
	if !f.Kinds.Equals(b.Kinds) ||
		!f.Authors.Equal(b.Authors) ||
		f.Tags.Len() != b.Tags.Len() ||
		!f.Tags.Equal(b.Tags) {
		return false
	}
	return true
}
