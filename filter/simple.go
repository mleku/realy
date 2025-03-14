package filter

import (
	"encoding/binary"
	"sort"

	"realy.lol/ec/schnorr"
	"realy.lol/event"
	"realy.lol/hex"
	"realy.lol/kinds"
	"realy.lol/sha256"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/text"
)

// S is a simplified filter that only covers the nip-01 REQ filter minus the
// separate and superseding ID list. The search field is from a different NIP,
// but it is a separate API for which reason it is also not here.
//
// The Since, Until and Limit fields are also omitted because the first two are
// short values able to be found in the URL parameters and the latter is not
// relevant because the filter API returns event IDs, not the whole events, and
// so the cost of delivering them is also substantially reduced. Likewise, the
// search process is simplified.
type S struct {
	Kinds   *kinds.T `json:"kinds,omitempty"`
	Authors *tag.T   `json:"authors,omitempty"`
	Tags    *tags.T  `json:"-,omitempty"`
}

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
func (s *S) Clone() (clone *S) {
	lim := new(uint)
	*lim = 1
	_Kinds := *s.Kinds
	_Authors := *s.Authors
	_Tags := *s.Tags.Clone()
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
func (s *S) Fingerprint() (fp uint64, err error) {
	var b []byte
	b = s.Marshal(b)
	h := sha256.Sum256(b)
	hb := h[:]
	fp = binary.LittleEndian.Uint64(hb)
	return
}

// Sort the fields of a filter so a fingerprint on a filter that has the same
// set of content produces the same fingerprint.
func (s *S) Sort() {
	if s.Kinds != nil {
		sort.Sort(s.Kinds)
	}
	if s.Authors != nil {
		sort.Sort(s.Authors)
	}
	if s.Tags != nil {
		sort.Sort(s.Tags)
	}
}

func (s *S) Marshal(dst []byte) (b []byte) {
	var err error
	_ = err
	var first bool
	// sort the fields so they come out the same
	s.Sort()
	// open parentheses
	dst = append(dst, '{')
	if s.Kinds.Len() > 0 {
		first = true
		dst = append(dst, ',')
		dst = text.JSONKey(dst, Kinds)
		dst = s.Kinds.Marshal(dst)
	}
	if s.Authors.Len() > 0 {
		if first {
			dst = append(dst, ',')
		} else {
			first = true
		}
		dst = text.JSONKey(dst, Authors)
		dst = text.MarshalHexArray(dst, s.Authors.ToByteSlice())
	}
	if s.Tags.Len() > 0 {
		// tags are stored as tags with the initial element the "#s" and the rest the
		// list in each element of the tags list. eg:
		//
		//     [["#p","<pubkey1>","<pubkey3"],["#t","hashtag","stuff"]]
		//
		for _, tg := range s.Tags.Value() {
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

func (s *S) Serialize() (b []byte) { return s.Marshal(nil) }

func (s *S) Unmarshal(b []byte) (r []byte, err error) {
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
					s.Tags = s.Tags.AppendTags(tag.New(ff...))
					// s.Tags.T = append(s.Tags.T, tag.New(ff...))
				default:
					// other types of tags can be anything
					var ff [][]byte
					if ff, r, err = text.UnmarshalStringArray(r); chk.E(err) {
						return
					}
					ff = append([][]byte{k}, ff...)
					s.Tags = s.Tags.AppendTags(tag.New(ff...))
					// s.Tags.T = append(s.Tags.T, tag.New(ff...))
				}
				state = betweenKV
			case Kinds[0]:
				if len(key) < len(Kinds) {
					goto invalid
				}
				s.Kinds = kinds.NewWithCap(0)
				if r, err = s.Kinds.Unmarshal(r); chk.E(err) {
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
				s.Authors = tag.New(ff...)
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

func (s *S) Matches(ev *event.T) bool {
	if ev == nil {
		// log.T.F("nil event")
		return false
	}
	if s.Kinds.Len() > 0 && !s.Kinds.Contains(ev.Kind) {
		// log.T.F("no matching kinds in filter\nEVENT %s\nFILTER %s", ev.ToObject().String(), s.ToObject().String())
		return false
	}
	if s.Authors.Len() > 0 && !s.Authors.Contains(ev.PubKey) {
		// log.T.F("no matching authors in filter\nEVENT %s\nFILTER %s", ev.ToObject().String(), s.ToObject().String())
		return false
	}
	if s.Tags.Len() > 0 && !ev.Tags.Intersects(s.Tags) {
		return false
	}
	return true
}

func (s *S) Equal(b *S) bool {
	if !s.Kinds.Equals(b.Kinds) ||
		!s.Authors.Equal(b.Authors) ||
		s.Tags.Len() != b.Tags.Len() ||
		!s.Tags.Equal(b.Tags) {
		return false
	}
	return true
}
