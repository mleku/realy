# atomic

Simple wrappers for primitive types to enforce atomic access.

## Installation

```shell
$ go get -u github.com/mleku/nodl/pkg/atomic@latest
```

## Usage

The standard library's `sync/atomic` is powerful, but it's easy to forget which
variables must be accessed atomically. `github.com/mleku/nodl/pkg/atomic` preserves all the
functionality of the standard library, but wraps the primitive types to
provide a safer, more convenient API.

```go
var atom atomic.Uint32
atom.Store(42)
atom.Sub(2)
atom.CompareAndSwap(40, 11)
```

See the [documentation][doc] for a complete API specification.

## Development Status

Stable.

---

Released under the [MIT License](LICENSE.txt).