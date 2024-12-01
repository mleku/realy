package event

import (
	sch "realy.lol/ec/schnorr"
	k1 "realy.lol/ec/secp256k1"
	"realy.lol/p256k"
	"realy.lol/signer"
)

// Sign the event using the signer.I. Uses github.com/bitcoin-core/secp256k1 if available for much faster
// signatures.
func (ev *T) Sign(keys signer.I) (err er) {
	ev.ID = ev.GetIDBytes()
	if ev.Sig, err = keys.Sign(ev.ID); chk.E(err) {
		return
	}
	ev.PubKey = keys.Pub()
	return
}

// Verify an event is signed by the pubkey it contains. Uses
// github.com/bitcoin-core/secp256k1 if available for faster verification.
func (ev *T) Verify() (valid bo, err er) {
	keys := p256k.Signer{}
	if err = keys.InitPub(ev.PubKey); chk.E(err) {
		return
	}
	if valid, err = keys.Verify(ev.ID, ev.Sig); chk.T(err) {
		// check that this isn't because of a bogus ID
		id := ev.GetIDBytes()
		if !equals(id, ev.ID) {
			log.E.Ln("event ID incorrect")
			ev.ID = id
			err = nil
			if valid, err = keys.Verify(ev.ID, ev.Sig); chk.E(err) {
				return
			}
			err = errorf.W("event ID incorrect but signature is valid on correct ID")
		}
		return
	}
	return
}

// SignWithSecKey signs an event with a given *secp256xk1.SecretKey.
//
// Deprecated: use Sign and nostr.I and p256k.Signer / p256k.BTCECSigner
// implementations.
func (ev *T) SignWithSecKey(sk *k1.SecretKey,
	so ...sch.SignOption) (err er) {

	// sign the event.
	var sig *sch.Signature
	ev.ID = ev.GetIDBytes()
	if sig, err = sch.Sign(sk, ev.ID, so...); chk.D(err) {
		return
	}
	// we know secret key is good so we can generate the public key.
	ev.PubKey = sch.SerializePubKey(sk.PubKey())
	ev.Sig = sig.Serialize()
	return
}

// CheckSignature returns whether an event signature is authentic and matches
// the event ID and Pubkey.
//
// Deprecated: use Verify
func (ev *T) CheckSignature() (valid bo, err er) {
	// parse pubkey bytes.
	var pk *k1.PublicKey
	if pk, err = sch.ParsePubKey(ev.PubKey); chk.D(err) {
		err = errorf.E("event has invalid pubkey '%0x': %v", ev.PubKey, err)
		return
	}
	// parse signature bytes.
	var sig *sch.Signature
	if sig, err = sch.ParseSignature(ev.Sig); chk.D(err) {
		err = errorf.E("failed to parse signature:\n%d %s\n%v", len(ev.Sig),
			ev.Sig, err)
		return
	}
	// check signature.
	valid = sig.Verify(ev.GetIDBytes(), pk)
	return
}
