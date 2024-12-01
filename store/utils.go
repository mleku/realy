package store

import (
	"bytes"
	"strconv"
	"strings"

	"realy.lol/hex"
	"realy.lol/tag"
)

func GetAddrTagElements(tagValue st) (k uint16, pkb by, d st) {
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

func TagSorter(a, b tag.T) no {
	if a.Len() < 2 {
		if b.Len() < 2 {
			return 0
		}
		return -1
	}
	if b.Len() < 2 {
		return 1
	}
	return bytes.Compare(a.B(1), b.B(1))
}

func Less(a, b tag.T) bo { return TagSorter(a, b) < 0 }
