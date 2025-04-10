package addresstag

import (
	"strconv"
	"strings"

	"realy.mleku.dev/hex"
)

// DecodeAddressTag unpacks the contents of an `a` tag.
func DecodeAddressTag(tagValue string) (k uint16, pkb []byte, d string) {
	split := strings.Split(tagValue, ":")
	if len(split) == 3 {
		if pkb, _ = hex.Dec(split[1]); len(pkb) == 32 {
			if key, err := strconv.ParseUint(split[0], 10, 16); err == nil {
				return uint16(key), pkb, split[2]
			}
		}
	}
	return
}
