package tests

import (
	"encoding/base64"

	"lukechampine.com/frand"
	"mleku.dev/event"
	"mleku.dev/hex"
	"mleku.dev/kind"
	"mleku.dev/p256k"
	"mleku.dev/timestamp"
)

func GenerateEvent(nsec B, maxSize int) (ev *event.T, binSize int, err E) {
	l := frand.Intn(maxSize * 6 / 8) // account for base64 expansion
	ev = &event.T{
		Kind:      kind.TextNote,
		CreatedAt: timestamp.Now(),
		Content:   B(base64.StdEncoding.EncodeToString(frand.Bytes(l))),
	}
	var sec B
	if _, err = hex.DecBytes(sec, nsec); chk.E(err) {
		return
	}
	signer := new(p256k.Signer)
	if err = signer.Generate(); chk.E(err) {
		return
	}
	if err = signer.InitSec(sec); chk.E(err) {
		return
	}
	if err = ev.Sign(signer); chk.E(err) {
		return
	}
	var bin []byte
	bin, err = ev.MarshalBinary(bin)
	binSize = len(bin)
	return
}
