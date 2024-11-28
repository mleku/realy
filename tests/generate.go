package tests

import (
	"encoding/base64"

	"lukechampine.com/frand"

	"realy.lol/event"
	"realy.lol/kind"
	"realy.lol/p256k"
	"realy.lol/timestamp"
)

func GenerateEvent(maxSize int) (ev *event.T, binSize int, err E) {
	l := frand.Intn(maxSize * 6 / 8) // account for base64 expansion
	ev = &event.T{
		Kind:      kind.TextNote,
		CreatedAt: timestamp.Now(),
		Content:   B(base64.StdEncoding.EncodeToString(frand.Bytes(l))),
	}
	signer := new(p256k.Signer)
	if err = signer.Generate(); chk.E(err) {
		return
	}
	if err = ev.Sign(signer); chk.E(err) {
		return
	}
	var bin []byte
	bin, err = ev.MarshalJSON(bin)
	binSize = len(bin)
	return
}
