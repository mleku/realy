package event

import (
	"bytes"

	sch "realy.mleku.dev/ec/schnorr"
	k1 "realy.mleku.dev/ec/secp256k1"
	"realy.mleku.dev/p256k"
	"realy.mleku.dev/signer"
)

// Sign the event using the signer.I. Uses github.com/bitcoin-core/secp256k1 if
// available for much faster signatures.
//
// Note that this only populates the Pubkey, Id and Sig. The caller must
// set the CreatedAt timestamp as intended.
func (ev *T) Sign(keys signer.I) (err error) {
	ev.Pubkey = keys.Pub()
	ev.Id = ev.GetIDBytes()
	if ev.Sig, err = keys.Sign(ev.Id); chk.E(err) {
		return
	}
	return
}

// Verify an event is signed by the pubkey it contains. Uses
// github.com/bitcoin-core/secp256k1 if available for faster verification.
func (ev *T) Verify() (valid bool, err error) {
	keys := p256k.Signer{}
	if err = keys.InitPub(ev.Pubkey); chk.E(err) {
		return
	}
	if valid, err = keys.Verify(ev.Id, ev.Sig); chk.T(err) {
		// check that this isn't because of a bogus Id
		id := ev.GetIDBytes()
		if !bytes.Equal(id, ev.Id) {
			log.E.Ln("event Id incorrect")
			ev.Id = id
			err = nil
			if valid, err = keys.Verify(ev.Id, ev.Sig); chk.E(err) {
				return
			}
			err = errorf.W("event Id incorrect but signature is valid on correct Id")
		}
		return
	}
	return
}

// SignWithSecKey signs an event with a given *secp256xk1.SecretKey.
//
// Deprecated: use Sign method of event.T and signer.I instead.
func (ev *T) SignWithSecKey(sk *k1.SecretKey,
	so ...sch.SignOption) (err error) {

	// sign the event.
	var sig *sch.Signature
	ev.Id = ev.GetIDBytes()
	if sig, err = sch.Sign(sk, ev.Id, so...); chk.D(err) {
		return
	}
	// we know secret key is good so we can generate the public key.
	ev.Pubkey = sch.SerializePubKey(sk.PubKey())
	ev.Sig = sig.Serialize()
	return
}

// CheckSignature returns whether an event signature is authentic and matches
// the event Id and Pubkey.
//
// Deprecated: use Verify
func (ev *T) CheckSignature() (valid bool, err error) {
	// parse pubkey bytes.
	var pk *k1.PublicKey
	if pk, err = sch.ParsePubKey(ev.Pubkey); chk.D(err) {
		err = errorf.E("event has invalid pubkey '%0x': %v", ev.Pubkey, err)
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
