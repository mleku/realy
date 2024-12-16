package graph

import (
	"sync"
	"github.com/dgraph-io/badger/v4/options"
	"realy.lol/units"
	"github.com/dgraph-io/badger/v4"
	"realy.lol/lol"
	"fmt"
	"realy.lol/ratel/graph/prefixes"
	"encoding/binary"
	"errors"
	"realy.lol/ratel/keys/serial"
	"bytes"
)

// T is a badger based key/value store implementation for a filesystem-like data
// structure.
//
// Each node has a name, can be addressed with a standard posix style path, and
// then instead of just having one data blob inside it, it can have an arbitrary
// number of named fields.
//
// Basically it's a filesystem where files can also have a structure inside them
// that is addressable.
//
// The purpose of this database not for search as in a graph database, but
// rather to enable the creation of persistent states, and at any given node,
// including the root, one can generate a JSON form of the data, which can be
// used for API implementations. Essentially similar to the Windows Registry in
// its architecture.
//
// It should be fast enough to use as a persistent state of a render tree for a
// simulation modeling anything, as well as storing live configuration items.
//
// The reason why this is being added as an extension to the simple event store
// ratel is to facilitate later work with direct message based CLIs that
// interact with the relay and enable per-user configurations to activate
// features they may optionally want to use, or disable.
type T struct {
	cx
	wg      *sync.WaitGroup
	dataDir st
	// DB is the badger db
	*badger.DB
	// seq is the monotonic collision free index for raw event storage.
	seq *badger.Sequence
	// compression sets the compression to use, none/snappy/zstd
	compression              st
	blockCacheSize, logLevel no
	logger                   *logger
}

type Params struct {
	Path           st
	Ctx            cx
	WG             *sync.WaitGroup
	LogLevel       st
	Compression    st // none,snappy,zstd
	BlockCacheSize no
}

func New(p *Params) (g *T, err er) {
	g = &T{
		dataDir: p.Path,
	}
	if p.Ctx == nil {
		err = errorf.E("Context not provided for creating graph")
		return
	}
	g.cx = p.Ctx
	if p.WG == nil {
		err = errorf.E("WaitGroup not provided for creating graph")
		return
	}
	g.wg = p.WG
	switch p.Compression {
	case "", "none":
		g.compression = "none"
	case "snappy", "zstd":
		if g.blockCacheSize == 0 {
			g.blockCacheSize = 16 * units.Mb
		}
		g.compression = p.Compression
	default:
		err = errorf.E("unknown compression type '%s'", p.Compression)
		return
	}
	g.logLevel = lol.GetLogLevel(p.LogLevel)
	return
}

const Version = 1

func (g *T) bumpVersion(txn *badger.Txn, version uint16) er {
	buf := make(by, 2)
	binary.BigEndian.PutUint16(buf, version)
	return txn.Set(prefixes.Version.Key(), buf)
}

func (g *T) runMigrations() (err er) {
	return g.Update(func(txn *badger.Txn) (err er) {
		var version uint16
		var item *badger.Item
		item, err = txn.Get(prefixes.Version.Key())
		if errors.Is(err, badger.ErrKeyNotFound) {
			version = 0
		} else if chk.E(err) {
			return err
		} else {
			chk.E(item.Value(func(val by) (err er) {
				version = binary.BigEndian.Uint16(val)
				return
			}))
		}
		// do the migrations in increasing steps (there is no rollback)
		if version < Version {
			// if there is any data in the relay we will stop and notify the user, otherwise we
			// just set version to 1 and proceed
			prefix := prefixes.Node.Key()
			it := txn.NewIterator(badger.IteratorOptions{
				PrefetchValues: true,
				PrefetchSize:   100,
				Prefix:         prefix,
			})
			defer it.Close()
			hasAnyEntries := false
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				hasAnyEntries = true
				break
			}
			if hasAnyEntries {
				return fmt.Errorf(
					`your database is at version %d, but in order to migrate up to version %d you 
must manually export all the data and then import again:

run an old version of this software, export the data, then delete the database 
files, run the new version, import the data back into it`, Version, version)
			}
			chk.E(g.bumpVersion(txn, Version))
		}
		return nil
	})
}

func (g *T) Init() (err er) {
	// log.I.Ln("opening ratel graph store at", g.dataDir)
	opts := badger.DefaultOptions(g.dataDir)
	opts.BlockCacheSize = int64(g.blockCacheSize)
	opts.BlockSize = 128 * units.Mb
	opts.CompactL0OnClose = true
	opts.LmaxCompaction = true
	switch g.compression {
	case "none":
		opts.Compression = options.None
	case "snappy":
		opts.Compression = options.Snappy
	case "zstd":
		opts.Compression = options.ZSTD
	}
	g.logger = NewLogger(g.logLevel, g.dataDir)
	opts.Logger = g.logger
	if g.DB, err = badger.Open(opts); chk.E(err) {
		return err
	}
	log.T.Ln("getting graph sequence index", g.dataDir)
	if g.seq, err = g.DB.GetSequence(by("nodes"), 1000); chk.E(err) {
		return err
	}
	return
}

func (g *T) Close() { chk.E(g.DB.Close()) }

// Serial returns the next monotonic conflict free unique serial on the database.
func (g *T) Serial() (ser uint64) {
	var err er
	if ser, err = g.seq.Next(); chk.E(err) {
		panic(err)
	}
	return
}

// SerialBytes returns a new serial value, used to store an event record with a
// conflict-free unique code (it is a monotonic, atomic, ascending counter).
func (g *T) SerialBytes() (ser by) {
	var serU64 uint64
	serU64 = g.Serial()
	ser = make(by, serial.Len)
	binary.BigEndian.PutUint64(ser, serU64)
	return
}

func ToSerialBytes(u uint64) (b by) {
	b = make(by, serial.Len)
	binary.BigEndian.PutUint64(b, u)
	return
}

func ToPath[V st | by](path []V) (b by) {
	if len(path) == 0 {
		b = by{'/'}
		return
	}
	for _, p := range path {
		b = append(b, '/')
		b = append(b, p...)
	}
	return
}

func Read[V by | st](g *T, path V) (b by, err er) {
	log.I.F("Read %s", path)
	pb := by(path)
	if !bytes.HasPrefix(pb, by{'/'}) {
		err = errorf.E("path must begin with the root /, '%s'", path)
		return
	}
	split := bytes.Split(pb, by{'/'})[1:]
	parent := serial.Make(0)
	var found bo
	for i, v := range split {
		_ = i
		// log.I.F("%s", v)
		if err = g.DB.View(func(txn *badger.Txn) (err er) {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prf := prefixes.Node.Key(parent)
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				var val by
				if val, err = item.ValueCopy(nil); chk.E(err) {
					return
				}
				k := item.KeyCopy(nil)
				if !bytes.HasPrefix(k, prf) {
					continue
				}
				if equals(v, val) {
					// log.I.F("%0x %d %s == %s", k, i, val, v)
					parent = serial.FromKey(k)
					found = true
					return
				}
			}
			return
		}); chk.E(err) {
			return
		}
	}
	if found {
		valKey := prefixes.Value.Key(parent)
		if err = g.DB.View(func(txn *badger.Txn) (err er) {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			// log.I.S(valKey)
			for it.Seek(valKey); it.ValidForPrefix(valKey); it.Next() {
				// k := it.Item().Key()
				// log.I.S(k, valKey, it.Item().Key())
				b, err = it.Item().ValueCopy(nil)
				// log.I.S(valKey, b)
				return

			}
			return
		}); chk.E(err) {
			return
		}
		// the value should now be returned
		return
	} else {
		// if there is no value it's as though there is no value
	}
	return
}

func Write[V by | st](g *T, path, b V) (err er) {
	// log.I.F("Write %s %s", path, b)
	pb := by(path)
	if !bytes.HasPrefix(pb, by{'/'}) {
		err = errorf.E("path must begin with the root /, '%s'", path)
		return
	}
	split := bytes.Split(pb, by{'/'})[1:]
	parent := serial.Make(0)
	for i, v := range split {
		var found bo
		if err = g.DB.View(func(txn *badger.Txn) (err er) {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prf := prefixes.Node.Key(parent)
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				var val by
				if val, err = item.ValueCopy(nil); chk.E(err) {
					return
				}
				k := item.KeyCopy(nil)
				if !bytes.HasPrefix(k, prf) {
					continue
				}
				if equals(v, val) {
					log.I.F("found %0x %d %s == %s", k, i, val, v)
					parent = serial.FromKey(k)
					found = true
					return
				}
			}
			return
		}); chk.E(err) {
			return
		}
		if found {
			continue
		}
		ser := g.Serial()
		if ser == 0 {
			ser = g.Serial()
		}
		seri := serial.Make(ser)
		if err = g.DB.Update(func(txn *badger.Txn) (err er) {
			k := prefixes.Node.Key(parent, seri)
			log.I.F("writing %0x %s", k, v)
			if err = txn.Set(k, v); chk.E(err) {
				return
			}
			return
		}); chk.E(err) {
			return
		}
		copy(parent.Val, seri.Val)
	}
	if err = g.DB.Update(func(txn *badger.Txn) (err er) {
		// write the value
		prf := prefixes.Value.Key(parent)
		log.I.F("writing %0x `%s`", prf, b)
		if err = txn.Set(prf, by(b)); chk.E(err) {
			return
		}
		return
	}); chk.E(err) {
		return
	}
	return
}
