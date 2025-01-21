package main

import (
	"os"
	"io/fs"
	"strings"
	"bytes"
)

var content = `
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
`

func main() {
	fsys := os.DirFS(".")
	fs.WalkDir(fsys, ".", func(p st, d fs.DirEntry, err er) (err2 er) {
		if strings.HasSuffix(p, "util.go") {
			log.I.F("%s", p)
			var b by
			if b, err2 = os.ReadFile(p); chk.E(err) {
				panic(err)
			}
			split := bytes.Split(b, by("\n"))
			for i := range split {
				// put the content after this
				if bytes.HasPrefix(split[i], by("package")) {
					// this can't be the end of teh file tho, so this can't have bounds error.
					o := bytes.Join(split[:i+1], by("\n"))
					o = append(o, by(content)...)
					if err2 = os.WriteFile(p, o, 0660); chk.E(err) {
						panic(err)
					}
				}
			}
		}
		return
	})
}
