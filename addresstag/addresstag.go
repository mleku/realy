package addresstag

import (
	"bytes"
	"strconv"

	"realy.lol/chk"
	"realy.lol/hex"
)

// DecodeAddressTag unpacks the contents of an `a` tag.
func DecodeAddressTag(tagValue []byte) (k uint16, pkb []byte, d []byte) {
	split := bytes.Split(tagValue, []byte(":"))
	if len(split) == 3 {
		var err error
		var key uint64
		if pkb, _ = hex.DecAppend(pkb, split[1]); len(pkb) == 32 {
			// todo: use ints package for this
			if key, err = strconv.ParseUint(string(split[0]), 10, 16); !chk.E(err) {
				return uint16(key), pkb, split[2]
			}
		}
	}
	return
}
