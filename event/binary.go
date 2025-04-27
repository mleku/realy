package event

import (
	"io"

	"realy.lol/chk"
	"realy.lol/ec/schnorr"
	"realy.lol/kind"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/timestamp"
	"realy.lol/varint"
)

// MarshalBinary writes a binary encoding of an event.
//
// [ 32 bytes Id ]
// [ 32 bytes Pubkey ]
// [ varint CreatedAt ]
// [ 2 bytes Kind ]
// [ varint Tags length ]
//
//	[ varint tag length ]
//	  [ varint tag element length ]
//	  [ tag element data ]
//	...
//
// [ varint Content length ]
// [ 64 bytes Sig ]
func (ev *T) MarshalBinary(w io.Writer) {
	_, _ = w.Write(ev.Id)
	_, _ = w.Write(ev.Pubkey)
	varint.Encode(w, uint64(ev.CreatedAt.V))
	varint.Encode(w, uint64(ev.Kind.K))
	varint.Encode(w, uint64(ev.Tags.Len()))
	for _, x := range ev.Tags.ToSliceOfTags() {
		varint.Encode(w, uint64(x.Len()))
		for _, y := range x.ToSliceOfBytes() {
			varint.Encode(w, uint64(len(y)))
			_, _ = w.Write(y)
		}
	}
	varint.Encode(w, uint64(len(ev.Content)))
	_, _ = w.Write(ev.Content)
	_, _ = w.Write(ev.Sig)
	return
}

func (ev *T) UnmarshalBinary(r io.Reader) (err error) {
	ev.Id = make([]byte, 32)
	if _, err = r.Read(ev.Id); chk.E(err) {
		return
	}
	ev.Pubkey = make([]byte, 32)
	if _, err = r.Read(ev.Pubkey); chk.E(err) {
		return
	}
	var ca uint64
	if ca, err = varint.Decode(r); chk.E(err) {
		return
	}
	ev.CreatedAt = timestamp.New(ca)
	var k uint64
	if k, err = varint.Decode(r); chk.E(err) {
		return
	}
	ev.Kind = kind.New(k)
	var nTags uint64
	if nTags, err = varint.Decode(r); chk.E(err) {
		return
	}
	ev.Tags = tags.NewWithCap(nTags)
	for range nTags {
		var nField uint64
		if nField, err = varint.Decode(r); chk.E(err) {
			return
		}
		t := tag.NewWithCap(nField)
		for range nField {
			var lenField uint64
			if lenField, err = varint.Decode(r); chk.E(err) {
				return
			}
			field := make([]byte, lenField)
			if _, err = r.Read(field); chk.E(err) {
				return
			}
			t = t.Append(field)
		}
		ev.Tags.AppendTags(t)
	}
	var cLen uint64
	if cLen, err = varint.Decode(r); chk.E(err) {
		return
	}
	ev.Content = make([]byte, cLen)
	if _, err = r.Read(ev.Content); chk.E(err) {
		return
	}
	ev.Sig = make([]byte, schnorr.SignatureSize)
	if _, err = r.Read(ev.Sig); chk.E(err) {
		return
	}
	return
}
