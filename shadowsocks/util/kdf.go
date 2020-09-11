package util

import (
	"crypto/md5"
	"math"

	"github.com/Qingluan/merkur/shadowsocks/bufferpool"
)

// KDF generate key from password
func KDF(password string, size int) []byte {
	count := int(math.Ceil(float64(size) / float64(md5.Size)))

	r := bufferpool.Get(count * md5.Size)
	defer bufferpool.Put(r)

	copy(r, MD5([]byte(password)))

	d := bufferpool.Get(md5.Size + len(password))
	defer bufferpool.Put(d)

	start := 0
	for i := 1; i < count; i++ {
		start += md5.Size
		copy(d[:md5.Size], r[start-md5.Size:start])
		copy(d[md5.Size:], password)
		copy(r[start:start+md5.Size], MD5(d))
	}

	key := make([]byte, size)
	copy(key, r[:size])

	return key
}
