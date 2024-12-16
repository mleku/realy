package prefixes

import (
	"realy.lol/ratel/keys/index"
)

const (
	// Version is the key that stores the version number, the value is a 16-bit
	// integer (2 bytes)
	//
	//   [ 255 ][ 2 byte/16 bit version code ]
	Version index.P = 255
)

const (
	// Node is a node in the graph. It is associated with a serial number and a
	// human-readable name is found in the value field. The serial 0 is the root
	// node of the graph, from which point all paths are iterations from. It
	// uniquely does not have the parent, and the value is empty, and it doesn't
	// exist, because it doesn't need to, all direct descendants of root have 9 zero
	// bytes at the beginning of their keys.
	//
	// key:   [ 0 ][ parent serial number ][ 64-bit serial number ]
	// value: [ up to 256 byte utf-8 string ]
	Node index.P = iota

	// Value stores an arbitrary blob of data, the nature of which must be known by
	// the caller as the graph only understands returning a blob of data associated
	// with a node found by walking the path provided in the query. The type of the
	// data itself is defined by the self-describing data within it or not at all.
	//
	// key:   [ 1 ][ 64-bit serial of Node ]
	// value: [ arbitrary binary data ]
	Value
)
