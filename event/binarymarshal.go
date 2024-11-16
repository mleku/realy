package event

import (
	"bytes"
	"encoding/binary"
	"github.com/golang/protobuf/proto"

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
	// log.I.S(w.Buf)
	w.Buf = append(w.Buf, pk...)
	return
}

func (w *Writer) WriteKind(k *kind.T) (err E) {
	// log.I.S(w.Buf)
	w.Buf = binary.LittleEndian.AppendUint16(w.Buf, k.ToU16())
	return
}

func (w *Writer) WriteCreatedAt(t *timestamp.T) (err E) {
	// log.I.S(w.Buf)
	w.Buf = appendUvarint(w.Buf, t.U64())
	return
}

// WriteTags encodes tags into binary form, including special handling for
// protocol defined a, e and p tags.
func (w *Writer) WriteTags(t *tags.T) (err E) {
	// log.I.F("writing tags %s", t.MarshalTo(nil))
	start := len(w.Buf)
	// first a varint for the number of tags
	// log.I.S(w.Buf)
	w.Buf = appendUvarint(w.Buf, uint64(t.Len()))
	// log.I.S(w.Buf)
	if t.F() == nil || len(t.F()) == 0 {
		// log.I.S(t.F())
		return
	}
	for i := range t.F() {
		var secondIsHex, secondIsDecimalHex bool
		// first the length of the tag
		// log.I.S(w.Buf)
		w.Buf = appendUvarint(w.Buf, uint64(t.N(i).Len()))
	scanning:
		for j := range t.N(i).F() {
			// we know from this first tag certain conditions that allow
			// data optimizations
			ts := t.N(i).B(j)
			// log.I.F("%0x,%s", len(ts), ts)
			switch {
			case j == 0:
				if len(ts) == 1 {
					// check if it is a special field with hex values
					for k := range HexInSecond {
						if ts[0] == HexInSecond[k] {
							secondIsHex = true
							// log.I.F("%d:%d '%s' second is hex", i, j, ts)
						}
					}
					for k := range DecimalHexInSecond {
						if ts[0] == DecimalHexInSecond[k] {
							secondIsDecimalHex = true
							// log.I.F("'%s' second is decimal:hex:string", ts)
						}
					}
				}
				// write first field length
				// log.I.S(w.Buf)
				w.Buf = appendUvarint(w.Buf, uint64(len(ts)))
				w.Buf = append(w.Buf, ts...)
			case j == 1:
				switch {
				case secondIsHex:
					// log.I.S(w.Buf)
					w.Buf = appendUvarint(w.Buf, uint64(32))
					if w.Buf, err = hex.DecAppend(w.Buf, ts); err != nil {
						// log.I.S(w.Buf)
						// the value MUST be hex by the spec
						return
					}
					continue scanning
				case secondIsDecimalHex:
					// first two fields are fixed length, 2 bytes for the kind and 32 bytes for
					// the event ID, then, a varint length prefix and the raw string of the
					// third field.
					// log.I.S(w.Buf)
					split := bytes.Split(t.N(i).B(j), B(":"))
					// log.I.S(split)
					// append the lengths accordingly
					// first is 2 bytes size
					k := kind.New(0)
					if _, err = k.UnmarshalJSON(split[0]); chk.T(err) {
						return
					}
					// write as little-endian uint16 (two bytes)
					w.Buf = binary.LittleEndian.AppendUint16(w.Buf, k.ToU16())
					if len(split) > 1 {
						// second is a 32 byte value encoded in hex
						if len(split[1]) != 64 {
							err = errorf.E("invalid length event ID in `a` tag: %d",
								len(split[1]))
							return
						}
						// write as 32 bytes binary
						if w.Buf, err = hex.DecAppend(w.Buf, split[1]); chk.E(err) {
							return
						}
						if len(split) > 2 {
							w.Buf = appendUvarint(w.Buf, uint64(len(split[2])))
							if len(split[2]) > 0 {
								// omit if there is no content for clarity
								w.Buf = append(w.Buf, split[2]...)
							}
							// log.I.S(w.Buf)
						}
						continue scanning
					}
				default:
					// log.I.S(w.Buf)
					w.Buf = appendUvarint(w.Buf, uint64(len(ts)))
					w.Buf = append(w.Buf, ts...)
				}
			case j > 1:
				// log.I.S(w.Buf)
				w.Buf = appendUvarint(w.Buf, uint64(len(ts)))
				w.Buf = append(w.Buf, ts...)
			}
		}
	}
	_ = start
	// log.I.S(w.Buf[start:])
	return
}

func (w *Writer) WriteContent(s B) (err error) {
	defer func() {
		if err != nil {
			log.I.S(uint64(len(s)), s)
		}
	}()
	// log.I.F("%d %0x '%s'", len(s), len(s), s)
	// log.I.S(w.Buf)
	w.Buf = appendUvarint(w.Buf, uint64(len(s)))
	// log.I.S(w.Buf)
	w.Buf = append(w.Buf, s...)
	// log.I.S(w.Buf)
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
	// log.I.F("writing binary event from %s", ev.Serialize())
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
	if err = w.WriteTags(ev.Tags); err != nil {
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

func (ev *T) oldMarshalBinary(dst B) (b B, err E) {
	w := NewBufForEvent(dst, ev)
	if err = w.WriteEvent(ev); err != nil {
		return
	}
	b = w.Bytes()
	return
}

func (ev *T) MarshalBinary(dst B) (b B, err E) {
	var pb B
	if pb, err = proto.Marshal(ev.ToProto()); chk.E(err) {
		return
	}
	b = append(dst, pb...)
	return
}

func (ev *T) bencMarshalBinary(dst B) (b B, err E) {
	be := ev.ToBenc()
	buf := make(B, be.Size())
	be.Marshal(buf)
	b = append(dst, buf...)
	return
}
