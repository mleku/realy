// Package filter is a codec for nostr filters (queries) and includes tools for
// matching them to events, a canonical format scheme to enable compactly
// identifying subscription filters, and a simplified filter that leavse out the
// IDs and Search fields for use in the HTTP API.
package filter

import (
	"bytes"
	"encoding/binary"
	"sort"

	"lukechampine.com/frand"

	"realy.lol/chk"
	"realy.lol/ec/schnorr"
	"realy.lol/ec/secp256k1"
	"realy.lol/errorf"
	"realy.lol/event"
	"realy.lol/hex"
	"realy.lol/ints"
	"realy.lol/kind"
	"realy.lol/kinds"
	"realy.lol/realy/pointers"
	"realy.lol/sha256"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/text"
	"realy.lol/timestamp"
)

// T is the primary query form for requesting events from a nostr relay.
//
// The ordering of fields of filters is not specified as in the protocol there
// is no requirement to generate a hash for fast recognition of identical
// filters. However, for internal use in a relay, by applying a consistent sort
// order, this library will produce an identical JSON from the same *set* of
// fields no matter what order they were provided.
//
// This is to facilitate the deduplication of filters so an effective identical
// match is not performed on an identical filter.
type T struct {
	IDs     *tag.T       `json:"ids,omitempty"`
	Kinds   *kinds.T     `json:"kinds,omitempty"`
	Authors *tag.T       `json:"authors,omitempty"`
	Tags    *tags.T      `json:"-,omitempty"`
	Since   *timestamp.T `json:"since,omitempty"`
	Until   *timestamp.T `json:"until,omitempty"`
	Search  []byte       `json:"search,omitempty"`
	Limit   *uint        `json:"limit,omitempty"`
}

// New creates a new, reasonably initialized filter that will be ready for most uses without
// further allocations.
func New() (f *T) {
	return &T{
		IDs:     tag.NewWithCap(10),
		Kinds:   kinds.NewWithCap(10),
		Authors: tag.NewWithCap(10),
		Tags:    tags.New(),
		// Since:   timestamp.New(),
		// Until:   timestamp.New(),
		Search: nil,
	}
}

// Clone creates a new filter with all the same elements in them, because they
// are immutable, basically, except setting the Limit field as 1, because it is
// used in the subscription management code to act as a reference counter, and
// making a clone implicitly means 1 reference.
func (f *T) Clone() (clone *T) {
	lim := new(uint)
	*lim = 1
	_IDs := *f.IDs
	_Kinds := *f.Kinds
	_Authors := *f.Authors
	_Tags := *f.Tags.Clone()
	_Since := *f.Since
	_Until := *f.Until
	_Search := make([]byte, len(f.Search))
	copy(Search, f.Search)
	return &T{
		IDs:     &_IDs,
		Kinds:   &_Kinds,
		Authors: &_Authors,
		Tags:    &_Tags,
		Since:   &_Since,
		Until:   &_Until,
		Search:  _Search,
		Limit:   lim,
	}
}

var (

	// IDs is the JSON object key for IDs.
	IDs = []byte("ids")
	// Kinds is the JSON object key for Kinds.
	Kinds = []byte("kinds")
	// Authors is the JSON object key for Authors.
	Authors = []byte("authors")
	// Since is the JSON object key for Since.
	Since = []byte("since")
	// Until is the JSON object key for Until.
	Until = []byte("until")
	// Limit is the JSON object key for Limit.
	Limit = []byte("limit")
	// Search is the JSON object key for Search.
	Search = []byte("search")
)

// Marshal a filter into raw JSON bytes, minified. The field ordering and sort of fields is
// canonicalized so that a hash can identify the same filter.
func (f *T) Marshal(dst []byte) (b []byte) {
	var err error
	_ = err
	var first bool
	// sort the fields so they come out the same
	f.Sort()
	// open parentheses
	dst = append(dst, '{')
	if f.IDs != nil && f.IDs.Len() > 0 {
		first = true
		dst = text.JSONKey(dst, IDs)
		dst = text.MarshalHexArray(dst, f.IDs.ToSliceOfBytes())
	}
	if f.Kinds.Len() > 0 {
		if first {
			dst = append(dst, ',')
		} else {
			first = true
		}
		dst = text.JSONKey(dst, Kinds)
		dst = f.Kinds.Marshal(dst)
	}
	if f.Authors.Len() > 0 {
		if first {
			dst = append(dst, ',')
		} else {
			first = true
		}
		dst = text.JSONKey(dst, Authors)
		dst = text.MarshalHexArray(dst, f.Authors.ToSliceOfBytes())
	}
	if f.Tags.Len() > 0 {
		// log.I.S(f.Tags)
		// if first {
		// 	dst = append(dst, ',')
		// } else {
		// 	first = true
		// }
		// tags are stored as tags with the initial element the "#a" and the rest the list in
		// each element of the tags list. eg:
		//
		//     [["#p","<pubkey1>","<pubkey3"],["#t","hashtag","stuff"]]
		//
		for _, tg := range f.Tags.ToSliceOfTags() {
			if tg == nil {
				// nothing here
				continue
			}
			if tg.Len() < 1 || len(tg.Key()) != 2 {
				// if there is no values, skip; the "key" field must be 2 characters long,
				continue
			}
			tKey := tg.ToSliceOfBytes()[0]
			if tKey[0] != '#' &&
				(tKey[1] < 'a' && tKey[1] > 'z' || tKey[1] < 'A' && tKey[1] > 'Z') {
				// first "key" field must begin with '#' and second be alpha
				continue
			}
			values := tg.ToSliceOfBytes()[1:]
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
	if len(f.Search) > 0 {
		if first {
			dst = append(dst, ',')
		} else {
			first = true
		}
		dst = text.JSONKey(dst, Search)
		dst = text.AppendQuote(dst, f.Search, text.NostrEscape)
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
	// close parentheses
	dst = append(dst, '}')
	b = dst
	return
}

// Serialize a filter.T into raw minified JSON bytes.
func (f *T) Serialize() (b []byte) { return f.Marshal(nil) }

// states of the unmarshaler
const (
	beforeOpen = iota
	openParen
	inKey
	inKV
	inVal
	betweenKV
	afterClose
)

// Unmarshal a filter from raw (minified) JSON bytes into the runtime format.
//
// todo: this may tolerate whitespace, not certain currently.
func (f *T) Unmarshal(b []byte) (r []byte, err error) {
	r = b[:]
	var key []byte
	var state int
	for ; len(r) >= 0; r = r[1:] {
		// log.I.ToSliceOfBytes("%c", rem[0])
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
				// tags start with # and have 1 letter
				l := len(key)
				if l != 2 {
					err = errorf.E("filter tag keys can only be # and one alpha character: '%s'\n%s",
						key, b)
					return
				}
				k := make([]byte, len(key))
				copy(k, key)
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
					// f.Tags.T = append(f.Tags.T, tag.New(ff...))
				default:
					// other types of tags can be anything
					var ff [][]byte
					if ff, r, err = text.UnmarshalStringArray(r); chk.E(err) {
						return
					}
					ff = append([][]byte{k}, ff...)
					f.Tags = f.Tags.AppendTags(tag.New(ff...))
					// f.Tags.T = append(f.Tags.T, tag.New(ff...))
				}
				state = betweenKV
			case IDs[0]:
				if len(key) < len(IDs) {
					goto invalid
				}
				var ff [][]byte
				if ff, r, err = text.UnmarshalHexArray(r, sha256.Size); chk.E(err) {
					return
				}
				f.IDs = tag.New(ff...)
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
			case Until[0]:
				if len(key) < len(Until) {
					goto invalid
				}
				u := ints.New(0)
				if r, err = u.Unmarshal(r); chk.E(err) {
					return
				}
				f.Until = timestamp.FromUnix(int64(u.N))
				state = betweenKV
			case Limit[0]:
				if len(key) < len(Limit) {
					goto invalid
				}
				l := ints.New(0)
				if r, err = l.Unmarshal(r); chk.E(err) {
					return
				}
				u := uint(l.N)
				f.Limit = &u
				state = betweenKV
			case Search[0]:
				if len(key) < len(Since) {
					goto invalid
				}
				switch key[1] {
				case Search[1]:
					if len(key) < len(Search) {
						goto invalid
					}
					var txt []byte
					if txt, r, err = text.UnmarshalQuoted(r); chk.E(err) {
						return
					}
					f.Search = txt
					// log.I.ToSliceOfBytes("\n%s\n%s", txt, rem)
					state = betweenKV
					// log.I.Ln("betweenKV")
				case Since[1]:
					if len(key) < len(Since) {
						goto invalid
					}
					s := ints.New(0)
					if r, err = s.Unmarshal(r); chk.E(err) {
						return
					}
					f.Since = timestamp.FromUnix(int64(s.N))
					state = betweenKV
					// log.I.Ln("betweenKV")
				}
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
			} else if r[0] == ',' {
				state = openParen
			} else if r[0] == '"' {
				state = inKey
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

// Matches checks a filter against an event and determines if the event matches the filter.
func (f *T) Matches(ev *event.T) bool {
	if ev == nil {
		return false
	}
	if f.IDs.Len() > 0 && !f.IDs.Contains(ev.Id) {
		return false
	}
	if f.Kinds.Len() > 0 && !f.Kinds.Contains(ev.Kind) {
		return false
	}
	if f.Authors.Len() > 0 && !f.Authors.Contains(ev.Pubkey) {
		return false
	}
	if f.Tags.Len() > 0 && !ev.Tags.Intersects(f.Tags) {
		return false
	}
	if f.Since.Int() != 0 && ev.CreatedAt.I64() < f.Since.I64() {
		return false
	}
	if f.Until.Int() != 0 && ev.CreatedAt.I64() > f.Until.I64() {
		return false
	}
	return true
}

// Fingerprint returns an 8 byte truncated sha256 hash of the filter in the canonical form
// created by Marshal.
//
// This hash is generated via the JSON encoded form of the filter, with the Limit field removed.
// This value should be set to zero after all results from a query of stored events, as per
// NIP-01.
func (f *T) Fingerprint() (fp uint64, err error) {
	lim := f.Limit
	f.Limit = nil
	var b []byte
	b = f.Marshal(b)
	h := sha256.Sum256(b)
	hb := h[:]
	fp = binary.LittleEndian.Uint64(hb)
	f.Limit = lim
	return
}

// Sort the fields of a filter so a fingerprint on a filter that has the same set of content
// produces the same fingerprint.
func (f *T) Sort() {
	if f.IDs != nil {
		sort.Sort(f.IDs)
	}
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

func arePointerValuesEqual[V comparable](a *V, b *V) bool {
	if a == nil && b == nil {
		return true
	}
	if a != nil && b != nil {
		return *a == *b
	}
	return false
}

// Equal checks a filter against another filter to see if they are the same filter.
func (f *T) Equal(b *T) bool {
	// sort the fields so they come out the same
	f.Sort()
	if !f.Kinds.Equals(b.Kinds) ||
		!f.IDs.Equal(b.IDs) ||
		!f.Authors.Equal(b.Authors) ||
		f.Tags.Len() != b.Tags.Len() ||
		!arePointerValuesEqual(f.Since, b.Since) ||
		!arePointerValuesEqual(f.Until, b.Until) ||
		!bytes.Equal(f.Search, b.Search) ||
		!f.Tags.Equal(b.Tags) {
		return false
	}
	return true
}

// GenFilter is a testing tool to create random arbitrary filters for tests.
func GenFilter() (f *T, err error) {
	f = New()
	n := frand.Intn(16)
	for _ = range n {
		id := make([]byte, sha256.Size)
		_, _ = frand.Read(id)
		f.IDs = f.IDs.Append(id)
	}
	n = frand.Intn(16)
	for _ = range n {
		f.Kinds.K = append(f.Kinds.K, kind.New(frand.Intn(65535)))
	}
	n = frand.Intn(16)
	for _ = range n {
		var sk *secp256k1.SecretKey
		if sk, err = secp256k1.GenerateSecretKey(); chk.E(err) {
			return
		}
		pk := sk.PubKey()
		f.Authors = f.Authors.Append(schnorr.SerializePubKey(pk))
	}
	a := frand.Intn(16)
	if a < n {
		n = a
	}
	for i := range n {
		p := make([]byte, 0, schnorr.PubKeyBytesLen*2)
		p = hex.EncAppend(p, f.Authors.B(i))
	}
	for b := 'a'; b <= 'z'; b++ {
		l := frand.Intn(6)
		if b == 'e' || b == 'p' {
			var idb [][]byte
			for range l {
				id := make([]byte, sha256.Size)
				_, _ = frand.Read(id)
				idb = append(idb, id)
			}
			idb = append([][]byte{{'#', byte(b)}}, idb...)
			f.Tags = f.Tags.AppendTags(tag.FromBytesSlice(idb...))
		} else {
			var idb [][]byte
			for range l {
				bb := make([]byte, frand.Intn(31)+1)
				_, _ = frand.Read(bb)
				id := make([]byte, 0, len(bb)*2)
				id = hex.EncAppend(id, bb)
				idb = append(idb, id)
			}
			idb = append([][]byte{{'#', byte(b)}}, idb...)
			f.Tags = f.Tags.AppendTags(tag.FromBytesSlice(idb...))
		}
	}
	tn := int(timestamp.Now().I64())
	f.Since = &timestamp.T{int64(tn - frand.Intn(10000))}
	f.Until = timestamp.Now()
	f.Search = []byte("token search text")
	return
}
