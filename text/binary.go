package text

import (
	"encoding/binary"
)

// AppendBinary is a straight append with length prefix.
func AppendBinary(dst, src B) (b B) {
	// if an allocation or two may occur, do it all in one immediately.
	minLen := len(src) + len(dst) + binary.MaxVarintLen32
	if cap(dst) < minLen {
		tmp := make(B, 0, minLen)
		dst = append(tmp, dst...)
	}
	dst = binary.AppendUvarint(dst, uint64(len(src)))
	dst = append(dst, src...)
	b = dst
	return
}

// ExtractBinary decodes the data based on the length prefix and returns a the the
// remaining data from the provided slice.
func ExtractBinary(b B) (str, rem B, err error) {
	l, read := binary.Uvarint(b)
	if read < 1 {
		err = errorf.E("failed to read uvarint length prefix")
		return
	}
	if len(b) < int(l)+read {
		err = errorf.E("insufficient data in buffer, require %d have %d",
			int(l)+read, len(b))
		return
	}
	str = b[read : read+int(l)]
	rem = b[read+int(l):]
	return
}
