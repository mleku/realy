package event

import (
	"bytes"
	"encoding/binary"

	"realy.lol/ec/schnorr"
	"realy.lol/hex"
	"realy.lol/kind"
	"realy.lol/sha256"
	"realy.lol/tags"
	"realy.lol/timestamp"
)

const (
	ID = iota
	PubKey
	CreatedAt
	Kind
	Tags
	Content
	Signature
)

var FieldSizes = []int{
	ID:        sha256.Size,
	PubKey:    schnorr.PubKeyBytesLen,
	CreatedAt: binary.MaxVarintLen64,
	Kind:      2,
	Tags:      -1, // -1 indicates variable
	Content:   -1,
	Signature: schnorr.SignatureSize,
}

var appendUvarint = binary.AppendUvarint

// Writer the buffer should have its capacity pre-allocated but its length
// initialized to zero, and the length of the buffer acts as its effective
// cursor position.
type Writer struct {
	Buf B
}

func EstimateSize(ev *T) (size int) {
	// id
	size += FieldSizes[ID]
	// pubkey
	size += FieldSizes[PubKey]
	// created_at timestamp
	size += FieldSizes[CreatedAt]
	// kind is most efficient as 2 bytes fixed length integer
	size += FieldSizes[Kind]
	// number of tags is a byte because it is unlikely that this many will ever
	// be used
	size++
	// next a byte for the length of each tag list
	for i := range ev.Tags.F() {
		size++
		for _ = range ev.Tags.N(i).F() {
			// plus a varint16 for each tag length prefix (very often will be 1
			// byte, occasionally 2, but no more than this
			size += binary.MaxVarintLen16
			// and the length of the actual tag
			size += ev.Tags.N(i).Len()
		}
	}
	// length prefix of the content field
	size += binary.MaxVarintLen32
	// the length of the content field
	size += len(ev.Content)
	// and the signature
	size += FieldSizes[Signature]
	return
}

// HexInSecond is the list of first tag fields that the second is pure hex
var HexInSecond = B{'e', 'p'}

// DecimalHexInSecond is the list of first tag fields that have "decimal:hex:"
var DecimalHexInSecond = B{'a'}

func NewBufForEvent(dst B, ev *T) (buf *Writer) {
	return NewWriteBuffer(dst, EstimateSize(ev))
}

// NewWriteBuffer allocates a slice with zero length and capacity at the given
// length. Use with EstimateSize to get a buffer that will not require a
// secondary allocation step.
func NewWriteBuffer(dst B, l int) (buf *Writer) {
	return &Writer{Buf: append(dst, make(B, 0, l)...)}
}

func (w *Writer) Bytes() B { return w.Buf }
func (w *Writer) Len() int { return len(w.Buf) }

func (w *Writer) WriteID(id B) (err E) {
	if len(id) != sha256.Size {
		err = errorf.E("wrong size, require %d got %d", sha256.Size, len(id))
		return
	}
	w.Buf = append(w.Buf, id...)
	return
}

func (w *Writer) WritePubKey(pk B) (err E) {
	if len(pk) != schnorr.PubKeyBytesLen {
		err = errorf.E("wrong size, require %d got %d",
			schnorr.PubKeyBytesLen, len(pk))
		return
	}
	w.Buf = append(w.Buf, pk...)
	return
}

func (w *Writer) WriteKind(k *kind.T) (err E) {
	w.Buf = binary.LittleEndian.AppendUint16(w.Buf, k.ToU16())
	return
}

func (w *Writer) WriteCreatedAt(t *timestamp.T) (err E) {
	w.Buf = appendUvarint(w.Buf, t.U64())
	return
}

// WriteTags encodes tags into binary form, including special handling for
// protocol defined a, e and p tags.
//
// todo: currently logging of incorrect a tag second section hex encoding as an
//
//	event ID is disabled because of a wrong a tag in the test events cache.
func (w *Writer) WriteTags(t *tags.T) (err E) {
	// first a byte for the number of tags
	w.Buf = appendUvarint(w.Buf, uint64(t.Len()))
	for i := range t.F() {
		var secondIsHex, secondIsDecimalHex bool
		// first the length of the tag
		w.Buf = appendUvarint(w.Buf, uint64(t.N(i).Len()))
	scanning:
		for j := range t.N(i).F() {
			// we know from this first tag certain conditions that allow
			// data optimizations
			ts := t.N(i).B(j)
			switch {
			case j == 0 && len(ts) == 1:
				for k := range HexInSecond {
					if ts[0] == HexInSecond[k] {
						secondIsHex = true
					}
				}
				for k := range DecimalHexInSecond {
					if ts[0] == DecimalHexInSecond[k] {
						secondIsDecimalHex = true
						// log.I.Ln("second is decimal:hex:string")
					}
				}
			case j == 1:
				switch {
				case secondIsHex:
					w.Buf = appendUvarint(w.Buf, uint64(32))
					if w.Buf, err = hex.DecAppend(w.Buf, ts); chk.E(err) {
						// the value MUST be hex by the spec
						log.W.Ln(t.N(i))
						return
					}
					continue scanning
				case secondIsDecimalHex:
					split := bytes.Split(t.N(i).B(j), B(":"))
					// append the lengths accordingly
					// first is 2 bytes size
					var n int
					k := kind.New(0)
					if _, err = k.UnmarshalJSON(split[0]); chk.E(err) {
						return
					}
					// second is a 32 byte value encoded in hex
					if len(split[1]) != 64 {
						err = errorf.E("invalid length event ID in `a` tag: %d",
							len(split[1]))
						return
					}
					if len(split) > 2 {
						// prepend with the appropriate length prefix (we don't need
						// a separate length prefix for the string component)
						w.Buf = appendUvarint(w.Buf, uint64(2+32+len(split[2])))
						// encode a 16 bit kind value
						w.Buf = binary.LittleEndian.
							AppendUint16(w.Buf, uint16(n))
						// encode the 32 byte binary value
						if w.Buf, err = hex.DecAppend(w.Buf, split[1]); chk.E(err) {
							return
						}
						w.Buf = append(w.Buf, split[2]...)
					}
					continue scanning
				}
			}
			w.Buf = appendUvarint(w.Buf, uint64(len(ts)))
			w.Buf = append(w.Buf, ts...)
		}
	}
	return
}

func (w *Writer) WriteContent(s B) (err error) {
	w.Buf = appendUvarint(w.Buf, uint64(len(s)))
	w.Buf = append(w.Buf, s...)
	return
}

func (w *Writer) WriteSignature(sig B) (err error) {
	if len(sig) != schnorr.SignatureSize {
		err = errorf.E("wrong size, require %d got %d",
			schnorr.SignatureSize, len(sig))
		return
	}
	w.Buf = append(w.Buf, sig...)
	return
}

func (w *Writer) WriteEvent(ev *T) (err error) {
	if err = w.WriteID(ev.ID); chk.E(err) {
		return
	}
	if err = w.WritePubKey(ev.PubKey); chk.E(err) {
		return
	}
	if err = w.WriteCreatedAt(ev.CreatedAt); chk.E(err) {
		return
	}
	if err = w.WriteKind(ev.Kind); chk.E(err) {
		return
	}
	if err = w.WriteTags(ev.Tags); chk.E(err) {
		return
	}
	if err = w.WriteContent(ev.Content); chk.E(err) {
		return
	}
	if err = w.WriteSignature(ev.Sig); chk.E(err) {
		return
	}
	return
}

func (ev *T) MarshalBinary(dst B) (b B, err E) {
	w := NewBufForEvent(dst, ev)
	if err = w.WriteEvent(ev); err != nil {
		return
	}
	b = w.Bytes()
	return
}
