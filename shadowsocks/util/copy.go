package util

import (
	"io"
	"os"

	"github.com/Qingluan/merkur/shadowsocks/bufferpool"
)

var bufSize = os.Getpagesize()

// Copy copy with default buf
func Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := bufferpool.Get(bufSize)
	defer bufferpool.Put(buf)

	return io.CopyBuffer(dst, src, buf)
}
