// Package buf implements a simple concurrent safe buffer pool for raw bytes.
package buf

import "sync"

var bufferPool = &sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 32*1024)
		return &buf
	},
}

type Pool struct{}

func (bp Pool) Get() []byte  { return *(bufferPool.Get().(*[]byte)) }
func (bp Pool) Put(b []byte) { bufferPool.Put(&b) }
