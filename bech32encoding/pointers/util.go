package pointers
import (
	"bytes"

	"realy.lol/context"
	"realy.lol/lol"
)

type (
	bo  = bool
	by  = []byte
	st  = string
	er  = error
	no  = int
	i8  = int8
	i16 = int16
	i32 = int32
	i64 = int64
	u8  = uint8
	u16 = uint16
	u32 = uint32
	u64 = uint64
	cx  = context.T
)

var (
	log, chk, errorf = lol.Main.Log, lol.Main.Check, lol.Main.Errorf
	equals           = bytes.Equal
)
