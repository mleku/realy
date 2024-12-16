package graph

import (
	"os"
	"path/filepath"
	"realy.lol/hex"
	"lukechampine.com/frand"
	"realy.lol/context"
	"realy.lol/interrupt"
	"sync"
	"testing"
	"github.com/dgraph-io/badger/v4"
	"fmt"
)

func TestWrite(t *testing.T) {
	path := filepath.Join(os.TempDir(), hex.Enc(frand.Bytes(8)))
	var err er
	var g *T
	c, cancel := context.Cancel(context.Bg())
	var wg sync.WaitGroup
	if g, err = New(&Params{
		Path:        path,
		Ctx:         c,
		WG:          &wg,
		LogLevel:    "trace",
		Compression: "zstd",
	}); chk.E(err) {
		return
	}
	if err = g.Init(); chk.E(err) {
		return
	}
	interrupt.AddHandler(func() { cancel() })
	if err = Write(g, "/path/to/one", "1"); chk.E(err) {
		return
	}
	fmt.Fprintln(os.Stderr)
	if err = Write(g, "/path/to/other/two", "2"); chk.E(err) {
		return
	}
	fmt.Fprintln(os.Stderr)
	if err = Write(g, "/path/four", "4"); chk.E(err) {
		return
	}
	fmt.Fprintln(os.Stderr)
	g.DB.View(func(txn *badger.Txn) (err er) {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			v, _ := it.Item().ValueCopy(nil)
			log.I.F("%0x %0x %s", it.Item().Key(), v, v)
		}
		return
	})
}
