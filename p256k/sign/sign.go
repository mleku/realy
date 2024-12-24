package sign

import (
	"realy.lol/signer"
	"realy.lol/bech32encoding"
	"realy.lol/hex"
	"realy.lol/p256k"
	"realy.lol/event"
)

func FromNsec[V st | by](sec V) (s signer.I, err er) {
	var sk by
	if sk, err = bech32encoding.DecodeNsec(sec); chk.E(err) {
		return
	}
	// log.I.F("%d %0x", len(sk), sk)
	sign := &p256k.Signer{}
	if err = sign.InitSec(sk); chk.E(err) {
		return
	}
	s = sign
	return
}

func FromHsec[V st | by](sec V) (s signer.I, err er) {
	var sk by
	if sk, err = hex.Dec(st(sec)); chk.E(err) {
		return
	}
	sign := &p256k.Signer{}
	if err = sign.InitSec(sk); chk.E(err) {
		return
	}
	s = sign
	return
}

func FromNpub[V st | by](pub V) (v signer.I, err er) {
	var pk by
	if pk, err = bech32encoding.DecodeNpub(pub); chk.E(err) {
		return
	}
	sign := &p256k.Signer{}
	if err = sign.InitPub(pk); chk.E(err) {
		return
	}
	v = sign
	return
}

func FromHpub[V st | by](pub V) (v signer.I, err er) {
	var pk by
	if pk, err = hex.Dec(st(pub)); chk.E(err) {
		return
	}
	// log.I.S(pk)
	sign := &p256k.Signer{}
	if err = sign.InitPub(pk); chk.E(err) {
		return
	}
	// log.I.S(sign)
	v = sign
	return
}

func SignEvent(s signer.I, ev *event.T) (res *event.T, err er) {
	res = ev
	// must set the pubkey first as it's part of the canonical encoding.
	res.PubKey = s.Pub()
	id := res.GetIDBytes()
	if res.Sig, err = s.Sign(id); chk.E(err) {
		return
	}
	ev.ID = id
	return
}
