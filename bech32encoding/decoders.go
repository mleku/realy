package bech32encoding

import (
	"realy.lol/ec/bech32"
)

func DecodeNsec[V st | by](nsec V) (skb by, err er) {
	sks := by(nsec)
	var prefix, bits5 by
	if prefix, bits5, err = bech32.DecodeNoLimit(sks); chk.D(err) {
		return
	}
	if !equals(prefix, NsecHRP) {
		err = errorf.E("incorrect prefix for nsec: %s", prefix)
		return
	}
	if skb, err = bech32.ConvertBits(bits5, 5, 8,
		false); chk.D(err) {

		return
	}
	return
}

func DecodeNpub[V st | by](nsec V) (skb by, err er) {
	pks := by(nsec)
	var prefix, bits5 by
	if prefix, bits5, err = bech32.DecodeNoLimit(pks); chk.D(err) {
		return
	}
	if !equals(prefix, NpubHRP) {
		err = errorf.E("incorrect prefix for npub: %s", prefix)
		return
	}
	if skb, err = bech32.Convert5to8(bits5, false); chk.D(err) {
		return
	}
	return
}
