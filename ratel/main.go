// Package ratel is a badger DB based event store with optional cache management
// and capability to be used as a pruning cache along with a secondary larger
// event store.
package ratel

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"

	"realy.mleku.dev/context"
	"realy.mleku.dev/ratel/keys/serial"
	"realy.mleku.dev/ratel/prefixes"
	"realy.mleku.dev/store"
	"realy.mleku.dev/units"
)

// DefaultMaxLimit is set to a size that means the usual biggest batch of events sent to a
// client usually is at most about 256kb or so.
const DefaultMaxLimit = 512

// T is a badger event store database with layer2 and garbage collection.
type T struct {
	Ctx     context.T
	WG      *sync.WaitGroup
	dataDir string
	// DBSizeLimit is the number of bytes we want to keep the data store from exceeding.
	DBSizeLimit int
	// DBLowWater is the percentage of DBSizeLimit a GC run will reduce the used storage down
	// to.
	DBLowWater int
	// DBHighWater is the trigger point at which a GC run should start if exceeded.
	DBHighWater int
	// GCFrequency is the frequency of checks of the current utilisation.
	GCFrequency    time.Duration
	HasL2          bool
	BlockCacheSize int
	InitLogLevel   int
	Logger         *logger
	// DB is the badger db
	*badger.DB
	// seq is the monotonic collision free index for raw event storage.
	seq *badger.Sequence
	// Threads is how many CPU threads we dedicate to concurrent actions, flatten and GC mark
	Threads int
	// MaxLimit is a default limit that applies to a query without a limit, to avoid sending out
	// too many events to a client from a malformed or excessively broad filter.
	MaxLimit int
	// ActuallyDelete sets whether we actually delete or rewrite deleted entries with a modified
	// deleted prefix value (8th bit set)
	ActuallyDelete bool
	// Flatten should be set to true to trigger a flatten at close... this is mainly
	// triggered by running an import
	Flatten bool
	// UseCompact uses a compact encoding based on the canonical format (generate
	// hash of it to get Id field with the signature in raw binary after.
	UseCompact bool
	// Compression sets the compression to use, none/snappy/zstd
	Compression string
}

var _ store.I = (*T)(nil)

// BackendParams is the configurations used in creating a new ratel.T.
type BackendParams struct {
	Ctx                                context.T
	WG                                 *sync.WaitGroup
	HasL2, UseCompact                  bool
	BlockCacheSize, LogLevel, MaxLimit int
	Compression                        string // none,snappy,zstd
	Extra                              []int
}

// New configures a a new ratel.T event store.
func New(p BackendParams, params ...int) *T {
	return GetBackend(p.Ctx, p.WG, p.HasL2, p.UseCompact, p.BlockCacheSize, p.LogLevel,
		p.MaxLimit,
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
func GetBackend(Ctx context.T, WG *sync.WaitGroup, hasL2, useCompact bool,
	blockCacheSize, logLevel, maxLimit int, compression string, params ...int) (b *T) {
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

// Path returns the path where the database files are stored.
func (r *T) Path() string { return r.dataDir }

// SerialKey returns a key used for storing events, and the raw serial counter
// bytes to copy into index keys.
func (r *T) SerialKey() (idx []byte, ser *serial.T) {
	var err error
	var s []byte
	if s, err = r.SerialBytes(); chk.E(err) {
		panic(err)
	}
	ser = serial.New(s)
	return prefixes.Event.Key(ser), ser
}

// Serial returns the next monotonic conflict free unique serial on the database.
func (r *T) Serial() (ser uint64, err error) {
	if ser, err = r.seq.Next(); chk.E(err) {
	}
	// log.T.ToSliceOfBytes("serial %x", ser)
	return
}

// SerialBytes returns a new serial value, used to store an event record with a
// conflict-free unique code (it is a monotonic, atomic, ascending counter).
func (r *T) SerialBytes() (ser []byte, err error) {
	var serU64 uint64
	if serU64, err = r.Serial(); chk.E(err) {
		panic(err)
	}
	ser = make([]byte, serial.Len)
	binary.BigEndian.PutUint64(ser, serU64)
	return
}
