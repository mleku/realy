package ratel

import (
	"strings"

	"realy.lol/ec/schnorr"
	"realy.lol/hex"
	"realy.lol/ratel/keys"
	"realy.lol/ratel/keys/arb"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/kinder"
	"realy.lol/ratel/keys/pubkey"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/tag/atag"
)

// GetTagKeyElements generates tag indexes from a tag key, tag value, created_at
// timestamp and the event serial.
func GetTagKeyElements(tagKey, tagValue st, CA *createdat.T,
	ser *serial.T) (prf index.P, elems []keys.Element, err er) {

	var pkb by
	// first check if it might be a public key, fastest test
	if len(tagValue) == 2*schnorr.PubKeyBytesLen {
		// this could be a pubkey
		pkb, err = hex.Dec(tagValue)
		if err == nil {
			// it's a pubkey
			var pkk keys.Element
			if pkk, err = pubkey.NewFromBytes(pkb); chk.E(err) {
				return
			}
			prf, elems = prefixes.Tag32, keys.Make(pkk, ser)
			return
		} else {
			err = nil
		}
	}
	// check for a tag
	if tagKey == "a" && strings.Count(tagValue, ":") == 2 {
		a := &atag.T{}
		var rem by
		if rem, err = a.Unmarshal(by(tagValue)); chk.E(err) {
			return
		}
		if len(rem) > 0 {
			log.I.S("remainder", tagKey, tagValue, rem)
		}
		prf = prefixes.TagAddr
		var pk *pubkey.T
		if pk, err = pubkey.NewFromBytes(a.PubKey); chk.E(err) {
			return
		}
		elems = keys.Make(kinder.New(a.Kind.K), pk, arb.New(a.DTag), CA,
			ser)
		return
		// todo: leaving this here in case bugz, note to remove this later
		// // this means we will get 3 pieces here
		// split := strings.Split(tagValue, ":")
		// // middle element should be a public key so must be 64 hex ciphers
		// if len(split[1]) != schnorr.PubKeyBytesLen*2 {
		// 	return
		// }
		// var k uint16
		// var d string
		// if pkb, err = hex.Dec(split[1]); !chk.E(err) {
		// 	var kin uint64
		// 	if kin, err = strconv.ParseUint(split[0], 10, 16); err == nil {
		// 		k = uint16(kin)
		// 		d = split[2]
		// 		var pk *pubkey.T
		// 		if pk, err = pubkey.NewFromBytes(pkb); chk.E(err) {
		// 			return
		// 		}
		// 		prf = prefixes.TagAddr
		// 		elems = keys.Make(kinder.New(k), pk, arb.NewFromString(d), CA,
		// 			ser)
		// 		return
		// 	}
		// }
	}
	// store whatever as utf-8
	prf = prefixes.Tag
	elems = keys.Make(arb.New(tagValue), CA, ser)
	return
}
