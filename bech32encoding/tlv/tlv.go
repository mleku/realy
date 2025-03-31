// Package tlv implements a simple Type Length Value encoder for nostr NIP-19
// bech32 encoded entities. The format is generic and could also be used for any
// TLV use case where fields are less than 255 bytes.
package tlv

import (
	"io"
)

const (
	Default byte = iota
	Relay
	Author
	Kind
)

// ReadEntry reads a TLV value from a bech32 encoded nostr entity.
func ReadEntry(buf io.Reader) (typ uint8, value []byte) {
	var err error
	t := make([]byte, 1)
	if _, err = buf.Read(t); err != nil {
		return
	}
	typ = t[0]
	l := make([]byte, 1)
	if _, err = buf.Read(l); err != nil {
		return
	}
	length := int(l[0])
	value = make([]byte, length)
	if _, err = buf.Read(value); err != nil {
		// nil value signals end of data or error
		value = nil
	}
	return
}

// WriteEntry writes a TLV value for a bech32 encoded nostr entity.
func WriteEntry(buf io.Writer, typ uint8, value []byte) {
	buf.Write(append([]byte{typ, byte(len(value))}, value...))
}
