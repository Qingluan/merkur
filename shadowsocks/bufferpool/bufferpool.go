package bufferpool

import (
	"sync"
)

// Pool pool
type Pool struct {
	size int
	pool *sync.Pool
}

// newPool create pool
func newPool(size int) *Pool {
	return &Pool{
		size: size,
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		},
	}
}

// Get get a buf
func (p *Pool) Get() []byte {
	return p.pool.Get().([]byte)
}

// Put release a buf
func (p *Pool) Put(x []byte) {
	p.pool.Put(x)
}

var poolMap map[int]*Pool

func init() {
	poolMap = make(map[int]*Pool)

	poolMap[8] = newPool(8)
	poolMap[16] = newPool(16)
	poolMap[32] = newPool(32)
	poolMap[64] = newPool(64)
	poolMap[128] = newPool(128)
	poolMap[256] = newPool(256)
	poolMap[512] = newPool(512)
	poolMap[1024] = newPool(1024)
	poolMap[2048] = newPool(2048)
	poolMap[4096] = newPool(4096)
	poolMap[8192] = newPool(8192)
}

// Get a buffer with size
func Get(size int) []byte {
	var p []byte

	switch {
	case size <= 8:
		p = poolMap[8].Get()
	case size <= 16:
		p = poolMap[16].Get()
	case size <= 32:
		p = poolMap[32].Get()
	case size <= 64:
		p = poolMap[64].Get()
	case size <= 128:
		p = poolMap[128].Get()
	case size <= 256:
		p = poolMap[256].Get()
	case size <= 512:
		p = poolMap[512].Get()
	case size <= 1024:
		p = poolMap[1024].Get()
	case size <= 2048:
		p = poolMap[2048].Get()
	case size <= 4096:
		p = poolMap[4096].Get()
	case size <= 8192:
		p = poolMap[8192].Get()
	default:
		p = make([]byte, size)
	}

	return p[:size]
}

// Put release buffer
func Put(p []byte) {
	c := cap(p)

	switch c {
	case 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192:
		poolMap[c].Put(p)
	}
}
