// Package buf implements a simple concurrent safe buffer pool for raw bytes.
package buf

import "sync"

var bufferPool = &sync.Pool{
	New: func() interface{} {
		buf := make(by, 32*1024)
		return &buf
	},
}

type Pool struct{}

func (bp Pool) Get() by  { return *(bufferPool.Get().(*by)) }
func (bp Pool) Put(b by) { bufferPool.Put(&b) }
