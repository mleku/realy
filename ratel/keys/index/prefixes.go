package index

import (
	"realy.mleku.dev/ratel/keys"
)

type P byte

// Key writes a key with the P prefix byte and an arbitrary list of
// keys.Element.
func (p P) Key(element ...keys.Element) (b []byte) {
	b = keys.Write(
		append([]keys.Element{New(byte(p))}, element...)...)
	// log.T.ToSliceOfBytes("key %x", b)
	return
}

// B returns the index.P as a byte.
func (p P) B() byte { return byte(p) }

// I returns the index.P as an int (for use with the KeySizes.
func (p P) I() int { return int(p) }

// GetAsBytes todo wat is dis?
func GetAsBytes(prf ...P) (b [][]byte) {
	b = make([][]byte, len(prf))
	for i := range prf {
		b[i] = []byte{byte(prf[i])}
	}
	return
}
