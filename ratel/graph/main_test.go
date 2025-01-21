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
		t.Fatal(err)
	}
	if err = g.Init(); chk.E(err) {
		t.Fatal(err)
	}
	interrupt.AddHandler(func() { cancel() })
	if err = Write(g, "/path/to/one", "1"); chk.E(err) {
		t.Fatal(err)
	}
	fmt.Fprintln(os.Stderr)
	if err = Write(g, "/path/to/other/two", "2"); chk.E(err) {
		t.Fatal(err)
	}
	fmt.Fprintln(os.Stderr)
	if err = Write(g, "/path/four", "4"); chk.E(err) {
		t.Fatal(err)
	}
	fmt.Fprintln(os.Stderr)
	var b by
	if b, err = Read(g, "/path/four"); chk.E(err) {
		t.Fatal(err)
	}
	log.I.F("/path/four = '%s'", b)
	b = b[:0]
	fmt.Fprintln(os.Stderr)
	if b, err = Read(g, "/path/to/one"); chk.E(err) {
		t.Fatal(err)
	}
	log.I.F("/path/to/one = '%s'", b)
	b = b[:0]
	fmt.Fprintln(os.Stderr)
	b = b[:0]
	if b, err = Read(g, "/path/to/other/two"); chk.E(err) {
		t.Fatal(err)
	}
	log.I.F("/path/to/other/two = '%s'", b)
	fmt.Fprintln(os.Stderr)
	b = b[:0]
	if b, err = Read(g, "/path/to"); chk.E(err) {
		t.Fatal(err)
	}
	log.I.F("/path/to = '%s'", b)
	fmt.Fprintln(os.Stderr)
	b = b[:0]
	if b, err = Read(g, "/path/to/other"); chk.E(err) {
		t.Fatal(err)
	}
	log.I.F("/path/to/other = '%s'", b)
	b = b[:0]
	fmt.Fprintln(os.Stderr)
	if b, err = Read(g, "/path/to/one"); chk.E(err) {
		t.Fatal(err)
	}
	log.I.F("/path/to/one = '%s'", b)
	fmt.Fprintln(os.Stderr)
	g.DB.View(func(txn *badger.Txn) (err er) {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		log.I.F("content of db:")
		for it.Rewind(); it.Valid(); it.Next() {
			v, _ := it.Item().ValueCopy(nil)
			log.I.F("%0x [%s] %s", it.Item().Key(), it.Item().Key(), v)
		}
		return
	})
}
