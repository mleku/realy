package ratel

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"

	"realy.lol/ratel/keys/serial"
	"realy.lol/store"
	"realy.lol/units"
	"realy.lol/ratel/keys/prefixes"
)

const DefaultMaxLimit = 512

type T struct {
	Ctx     cx
	WG      *sync.WaitGroup
	dataDir st
	// DBSizeLimit is the number of bytes we want to keep the data store from exceeding.
	DBSizeLimit no
	// DBLowWater is the percentage of DBSizeLimit a GC run will reduce the used storage down
	// to.
	DBLowWater no
	// DBHighWater is the trigger point at which a GC run should start if exceeded.
	DBHighWater no
	// GCFrequency is the frequency of checks of the current utilisation.
	GCFrequency    time.Duration
	HasL2          bo
	BlockCacheSize no
	InitLogLevel   no
	Logger         *logger
	// DB is the badger db
	*badger.DB
	// seq is the monotonic collision free index for raw event storage.
	seq *badger.Sequence
	// Threads is how many CPU threads we dedicate to concurrent actions, flatten and GC mark
	Threads no
	// MaxLimit is a default limit that applies to a query without a limit, to avoid sending out
	// too many events to a client from a malformed or excessively broad filter.
	MaxLimit no
	// ActuallyDelete sets whether we actually delete or rewrite deleted entries with a modified
	// deleted prefix value (8th bit set)
	ActuallyDelete bo
	// Flatten should be set to true to trigger a flatten at close... this is mainly
	// triggered by running an import
	Flatten bo
	// UseCompact uses a compact encoding based on the canonical format (generate
	// hash of it to get ID field with the signature in raw binary after.
	UseCompact bo
	// Compression sets the compression to use, none/snappy/zstd
	Compression st
}

var _ store.I = (*T)(nil)

type BackendParams struct {
	Ctx                                cx
	WG                                 *sync.WaitGroup
	HasL2, UseCompact                  bo
	BlockCacheSize, LogLevel, MaxLimit no
	Compression                        st // none,snappy,zstd
	Extra                              []no
}

func New(p BackendParams, params ...no) *T {
	return GetBackend(p.Ctx, p.WG, p.HasL2, p.UseCompact, p.BlockCacheSize, p.LogLevel, p.MaxLimit,
		p.Compression, params...)
}

// GetBackend returns a reasonably configured badger.Backend.
//
// The variadic params correspond to DBSizeLimit, DBLowWater, DBHighWater and
// GCFrequency as an integer multiplier of number of seconds.
//
// Note that the cancel function for the context needs to be managed by the
// caller.
//
// Deprecated: use New instead.
func GetBackend(Ctx cx, WG *sync.WaitGroup, hasL2, useCompact bo,
	blockCacheSize, logLevel, maxLimit no, compression st, params ...no) (b *T) {
	var sizeLimit, lw, hw, freq = 0, 50, 66, 3600
	switch len(params) {
	case 4:
		freq = params[3]
		fallthrough
	case 3:
		hw = params[2]
		fallthrough
	case 2:
		lw = params[1]
		fallthrough
	case 1:
		sizeLimit = params[0] * units.Gb
	}
	// if unset, assume a safe maximum limit for unlimited filters.
	if maxLimit == 0 {
		maxLimit = 512
	}
	b = &T{
		Ctx:            Ctx,
		WG:             WG,
		DBSizeLimit:    sizeLimit,
		DBLowWater:     lw,
		DBHighWater:    hw,
		GCFrequency:    time.Duration(freq) * time.Second,
		HasL2:          hasL2,
		BlockCacheSize: blockCacheSize,
		InitLogLevel:   logLevel,
		MaxLimit:       maxLimit,
		UseCompact:     useCompact,
		Compression:    compression,
	}
	return
}

func (r *T) Path() st { return r.dataDir }

// SerialKey returns a key used for storing events, and the raw serial counter
// bytes to copy into index keys.
func (r *T) SerialKey() (idx by, ser *serial.T) {
	var err er
	var s by
	if s, err = r.SerialBytes(); chk.E(err) {
		panic(err)
	}
	ser = serial.New(s)
	return prefixes.Event.Key(ser), ser
}

// Serial returns the next monotonic conflict free unique serial on the database.
func (r *T) Serial() (ser uint64, err er) {
	if ser, err = r.seq.Next(); chk.E(err) {
	}
	// log.T.F("serial %x", ser)
	return
}

// SerialBytes returns a new serial value, used to store an event record with a
// conflict-free unique code (it is a monotonic, atomic, ascending counter).
func (r *T) SerialBytes() (ser by, err er) {
	var serU64 uint64
	if serU64, err = r.Serial(); chk.E(err) {
		panic(err)
	}
	ser = make(by, serial.Len)
	binary.BigEndian.PutUint64(ser, serU64)
	return
}
