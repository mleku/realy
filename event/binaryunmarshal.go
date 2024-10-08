package event

import (
	"encoding/binary"
	"fmt"
	"io"

	"realy.lol/ec/schnorr"
	"realy.lol/hex"
	"realy.lol/kind"
	"realy.lol/sha256"
	"realy.lol/tags"
	"realy.lol/timestamp"
)

// Reader is a control structure for reading and writing buffers.
//
// It keeps track of the current cursor position, and each read
// function increments it to reflect the position of the next field in the data.
//
// All strings extracted from a Reader will be directly converted to strings
// using unsafe.String and will be garbage collected only once these strings
// fall out of scope.
//
// Thus the buffers cannot effectively be reused, the memory can only be reused
// via GC processing. This avoids data copy as the content fields are the
// biggest in the event.T structure and dominate the size of the whole event
// anyway, so either way this is done there is a tradeoff. This can be mitigated
// by changing the event.T to be a []byte instead. Or alternatively, copy the
// binary buffer out and the database can manage recycling this buffer.
type Reader struct {
	Pos int
	Buf B
}

// NewReadBuffer returns a new buffer containing the provided slice.
func NewReadBuffer(b []byte) (buf *Reader) {
	return &Reader{Buf: b}
}

func (r *Reader) ReadID() (id B, err E) {
	end := r.Pos + FieldSizes[ID]
	if len(r.Buf) < end {
		err = io.EOF
		return
	}
	id = r.Buf[r.Pos:end]
	r.Pos = end
	return
}

func (r *Reader) ReadPubKey() (pk B, err E) {
	end := r.Pos + FieldSizes[PubKey]
	if len(r.Buf) < end {
		err = io.EOF
		return
	}
	pk = r.Buf[r.Pos:end]
	r.Pos = end
	return
}

func (r *Reader) ReadCreatedAt() (t *timestamp.T, err error) {
	n, advance := binary.Uvarint(r.Buf[r.Pos:])
	if advance <= 0 {
		err = io.EOF
		return
	}
	r.Pos += advance
	tt := timestamp.T(n)
	t = &tt
	return
}

func (r *Reader) ReadKind() (k *kind.T, err error) {
	end := r.Pos + 2
	if len(r.Buf) < end {
		err = io.EOF
		return
	}
	k = &kind.T{K: binary.LittleEndian.Uint16(r.Buf[r.Pos:])}
	r.Pos = end
	return
}

func (r *Reader) ReadTags() (t *tags.T, err error) {
	// first get the count of tags
	vi, read := binary.Uvarint(r.Buf[r.Pos:])
	if read < 1 {
		err = io.EOF
		return
	}
	nTags := int(vi)
	var end int
	r.Pos += read
	t = tags.NewWithCap(nTags)
	// t = &tags.T{T: make([]*tag.T, nTags)}
	// t = make(tags.T, nTags)
	// iterate through the individual tags
	for i := 0; i < nTags; i++ {
		vi, read = binary.Uvarint(r.Buf[r.Pos:])
		if read < 1 {
			err = io.EOF
			return
		}
		lenTag := int(vi)
		r.Pos += read
		t.AddCap(i, lenTag)
		// t.T[i] = tag.NewWithCap(lenTag)
		// extract the individual tag strings
		var secondIsHex, secondIsDecimalHex bool
	reading:
		for j := 0; j < lenTag; j++ {
			// get the length prefix
			vi, read = binary.Uvarint(r.Buf[r.Pos:])
			if read < 1 {
				err = io.EOF
				log.I.S()
				return
			}
			r.Pos += read
			// now read it off
			end = r.Pos + int(vi)
			if len(r.Buf) < end {
				err = io.EOF
				return
			}
			// we know from this first tag certain conditions that allow
			// data optimizations
			switch {
			case j == 0:
				if vi != 1 {
					break
				}
				for k := range HexInSecond {
					if r.Buf[r.Pos] == HexInSecond[k] {
						secondIsHex = true
					}
				}
				for k := range DecimalHexInSecond {
					if r.Buf[r.Pos] == DecimalHexInSecond[k] {
						secondIsDecimalHex = true
					}
				}
			case j == 1:
				switch {
				case secondIsHex:
					hh := make(B, 0, sha256.Size*2)
					hh = hex.EncAppend(hh, r.Buf[r.Pos:end])
					t.AppendTo(i, hh)
					// t.N(i).Field = append(t.T[i].Field, make(B, 0, sha256.Size*2))
					// t.N(i).Field[j] = hex.EncAppend(t.N(i).B(j), r.Buf[r.Pos:end])
					r.Pos = end
					continue reading
				case secondIsDecimalHex:
					var k uint16
					var pk []byte
					fieldEnd := r.Pos + 2
					if fieldEnd > end {
						err = io.EOF
						return
					}
					k = binary.LittleEndian.Uint16(r.Buf[r.Pos:fieldEnd])
					r.Pos += 2
					fieldEnd += schnorr.PubKeyBytesLen
					if fieldEnd > end {
						err = log.E.Err("%v got %d expect %d",
							io.EOF, fieldEnd, end)
						return
					}
					pk = r.Buf[r.Pos:fieldEnd]
					r.Pos = fieldEnd
					t.AppendTo(i, B(fmt.Sprintf("%d:%0x:%s",
						k, hex.Enc(pk), string(r.Buf[r.Pos:end]))))
					// t.N(i).Field = append(t.N(i).Field, B(fmt.Sprintf("%d:%0x:%s",
					// 	k,
					// 	hex.Enc(pk),
					// 	string(r.Buf[r.Pos:end]))))
					r.Pos = end
				}
			}
			t.AppendTo(i, r.Buf[r.Pos:r.Pos+int(vi)])
			// t.N(i).Field = append(t.N(i).Field, r.Buf[r.Pos:r.Pos+int(vi)])
			r.Pos = end
		}
	}
	return
}

func (r *Reader) ReadContent() (s B, err error) {
	// get the length prefix
	vi, n := binary.Uvarint(r.Buf[r.Pos:])
	if n < 1 {
		err = io.EOF
		return
	}
	r.Pos += n
	end := r.Pos + int(vi)
	if end > len(r.Buf) {
		err = log.E.Err("expect %d got %d", end, len(r.Buf))
		return
	}
	// extract the string
	s = r.Buf[r.Pos : r.Pos+int(vi)]
	r.Pos = end
	return
}

func (r *Reader) ReadSignature() (sig B, err error) {
	end := r.Pos + FieldSizes[Signature]
	if len(r.Buf) < end {
		err = io.EOF
		return
	}
	sig = r.Buf[r.Pos:end]
	r.Pos = end
	return
}

func (r *Reader) ReadEvent() (ev *T, err error) {
	ev = &T{}
	if ev.ID, err = r.ReadID(); chk.E(err) {
		return
	}
	if ev.PubKey, err = r.ReadPubKey(); chk.E(err) {
		return
	}
	if ev.CreatedAt, err = r.ReadCreatedAt(); chk.E(err) {
		return
	}
	if ev.Kind, err = r.ReadKind(); chk.E(err) {
		return
	}
	if ev.Tags, err = r.ReadTags(); chk.E(err) {
		return
	}
	if ev.Content, err = r.ReadContent(); chk.E(err) {
		return
	}
	if ev.Sig, err = r.ReadSignature(); chk.E(err) {
		return
	}
	return
}

func (ev *T) UnmarshalBinary(b B) (r B, err E) {
	er := &Reader{Buf: b}
	var re *T
	if re, err = er.ReadEvent(); chk.E(err) {
		return
	}
	*ev = *re
	r = er.Buf[er.Pos:]
	return
}
