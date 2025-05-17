package ratel

import (
	eventstore "realy.lol/addresstag"
	"realy.lol/chk"
	"realy.lol/eventid"
	"realy.lol/hex"
	"realy.lol/ratel/keys"
	"realy.lol/ratel/keys/arb"
	"realy.lol/ratel/keys/id"
	"realy.lol/ratel/keys/kinder"
	"realy.lol/ratel/keys/pubkey"
	"realy.lol/ratel/prefixes"
)

// GetTagKeyPrefix returns tag index prefixes based on the initial field of a
// tag.
//
// There is 4 types of index tag keys:
//
// - TagAddr:   [ 8 ][ 2b Kind ][ 8b Pubkey ][ address/URL ][ 8b Serial ]
//
// - Tag32:     [ 7 ][ 8b Pubkey ][ 8b Serial ]
//
// - Tag:       [ 6 ][ address/URL ][ 8b Serial ]
//
// - TagEventId [ 8 bytes eventid.T prefix ][ 8 bytes Serial ]
//
// This function produces the initial bytes without the index.
func GetTagKeyPrefix(prf byte, tagValue []byte) (key []byte, err error) {
	if prf == 'e' {
		var eid []byte
		if eid, err = hex.DecAppend(eid, tagValue); chk.E(err) {
			return
		}
		key = prefixes.TagEventId.Key(id.New(eventid.NewWith(eid)))
	} else if k, pkb, d := eventstore.DecodeAddressTag(tagValue); len(pkb) == 32 {
		// store value in the new special "a" tag index
		var pk *pubkey.T
		if pk, err = pubkey.NewFromBytes(pkb); chk.E(err) {
			return
		}
		els := []keys.Element{kinder.New(k), pk}
		if len(d) > 0 {
			els = append(els, arb.New(d))
		}
		key = prefixes.TagAddr.Key(els...)
	} else if pkb, _ := hex.DecAppend(nil, tagValue); len(pkb) == 32 {
		// store value as bytes
		var pkk *pubkey.T
		if pkk, err = pubkey.NewFromBytes(pkb); chk.E(err) {
			return
		}
		key = prefixes.Tag32.Key(pkk)
	} else {
		// store whatever as utf-8
		if len(tagValue) > 0 {
			var a *arb.T
			a = arb.New(tagValue)
			key = prefixes.Tag.Key(a)
		} else {
			key = prefixes.Tag.Key()
		}
	}
	return
}
