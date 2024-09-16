package eventstore

import (
	"bytes"
	"strconv"
	"strings"

	"mleku.dev/hex"
	"mleku.dev/tag"
)

func GetAddrTagElements(tagValue S) (k uint16, pkb B, d S) {
	split := strings.Split(tagValue, ":")
	if len(split) == 3 {
		if pkb, _ = hex.Dec(split[1]); len(pkb) == 32 {
			if key, err := strconv.ParseUint(split[0], 10, 16); err == nil {
				return uint16(key), pkb, split[2]
			}
		}
	}
	return 0, nil, ""
}

func TagSorter(a, b tag.T) int {
	if len(a.Field) < 2 {
		if len(b.Field) < 2 {
			return 0
		}
		return -1
	}
	if len(b.Field) < 2 {
		return 1
	}
	return bytes.Compare(a.Field[1], b.Field[1])
}
