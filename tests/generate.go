package tests

import (
	"encoding/base64"

	"lukechampine.com/frand"

	"realy.lol/event"
	"realy.lol/kind"
	"realy.lol/p256k"
	"realy.lol/timestamp"
)

func GenerateEvent(maxSize no) (ev *event.T, binSize no, err er) {
	l := frand.Intn(maxSize * 6 / 8) // account for base64 expansion
	ev = &event.T{
		Kind:      kind.TextNote,
		CreatedAt: timestamp.Now(),
		Content:   by(base64.StdEncoding.EncodeToString(frand.Bytes(l))),
	}
	signer := new(p256k.Signer)
	if err = signer.Generate(); chk.E(err) {
		return
	}
	if err = ev.Sign(signer); chk.E(err) {
		return
	}
	var bin by
	bin = ev.Marshal(bin)
	binSize = len(bin)
	return
}
