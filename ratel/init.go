package ratel

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"

	"realy.mleku.dev/ratel/prefixes"
	"realy.mleku.dev/units"
)

// Init sets up the database with the loaded configuration.
func (r *T) Init(path string) (err error) {
	r.dataDir = path
	log.I.Ln("opening ratel event store at", r.Path())
	opts := badger.DefaultOptions(r.dataDir)
	opts.BlockCacheSize = int64(r.BlockCacheSize)
	opts.BlockSize = 128 * units.Mb
	opts.CompactL0OnClose = true
	opts.LmaxCompaction = true
	switch r.Compression {
	case "none":
		opts.Compression = options.None
	case "snappy":
		opts.Compression = options.Snappy
	case "zstd":
		opts.Compression = options.ZSTD
	}
	r.Logger = NewLogger(r.InitLogLevel, r.dataDir)
	opts.Logger = r.Logger
	if r.DB, err = badger.Open(opts); chk.E(err) {
		return err
	}
	log.T.Ln("getting event store sequence index", r.dataDir)
	if r.seq, err = r.DB.GetSequence([]byte("events"), 1000); chk.E(err) {
		return err
	}
	log.T.Ln("running migrations", r.dataDir)
	if err = r.runMigrations(); chk.E(err) {
		return log.E.Err("error running migrations: %w; %s", err, r.dataDir)
	}
	if r.DBSizeLimit > 0 {
		go r.GarbageCollector()
		// } else {
		// 	go r.GCCount()
	}
	return nil

}

const Version = 1

func (r *T) runMigrations() (err error) {
	return r.Update(func(txn *badger.Txn) (err error) {
		var version uint16
		var item *badger.Item
		item, err = txn.Get(prefixes.Version.Key())
		if errors.Is(err, badger.ErrKeyNotFound) {
			version = 0
		} else if chk.E(err) {
			return err
		} else {
			chk.E(item.Value(func(val []byte) (err error) {
				version = binary.BigEndian.Uint16(val)
				return
			}))
		}
		// do the migrations in increasing steps (there is no rollback)
		if version < Version {
			// if there is any data in the relay we will stop and notify the user, otherwise we
			// just set version to 1 and proceed
			prefix := prefixes.Id.Key()
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
				return fmt.Errorf("your database is at version %d, but in order to migrate up "+
					"to version 1 you must manually export all the events and then import "+
					"again:\n"+
					"run an old version of this software, export the data, then delete the "+
					"database files, run the new version, import the data back it", version)
			}
			chk.E(r.bumpVersion(txn, Version))
		}
		return nil
	})
}

func (r *T) bumpVersion(txn *badger.Txn, version uint16) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, version)
	return txn.Set(prefixes.Version.Key(), buf)
}
