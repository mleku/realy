package text

import (
	"encoding/binary"
)

// AppendBinary is a straight append with length prefix.
func AppendBinary(dst, src by) (b by) {
	// if an allocation or two may occur, do it all in one immediately.
	minLen := len(src) + len(dst) + binary.MaxVarintLen32
	if cap(dst) < minLen {
		tmp := make(by, 0, minLen)
		dst = append(tmp, dst...)
	}
	dst = binary.AppendUvarint(dst, uint64(len(src)))
	dst = append(dst, src...)
	b = dst
	return
}

// ExtractBinary decodes the data based on the length prefix and returns a the the
// remaining data from the provided slice.
func ExtractBinary(b by) (str, rem by, err er) {
	l, read := binary.Uvarint(b)
	if read < 1 {
		err = errorf.E("failed to read uvarint length prefix")
		return
	}
	if len(b) < no(l)+read {
		err = errorf.E("insufficient data in buffer, require %d have %d",
			no(l)+read, len(b))
		return
	}
	str = b[read : read+no(l)]
	rem = b[read+no(l):]
	return
}
