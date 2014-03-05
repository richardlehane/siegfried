package siegreader

import (
	"bytes"
	"io"
	"os"
)

var readSz int64 = 4096

type Buffer struct {
	eof  chan struct{}
	src  io.Reader
	main []byte
	tail []byte
}

func New() *Buffer {
	return new(Buffer)
}

func (b *Buffer) grow(n int) int {
	m := b.Len()
	if len(b.buf)+n > cap(b.buf) {
		var buf []byte
		if b.buf == nil && n <= len(b.bootstrap) {
			buf = b.bootstrap[0:]
		} else if m+n <= cap(b.buf)/2 {
			copy(b.buf[:], b.buf[b.off:])
			buf = b.buf[:m]
		} else {
			// not enough space anywhere
			buf = makeSlice(2*cap(b.buf) + n)
			copy(buf, b.buf[b.off:])
		}
		b.buf = buf
		b.off = 0
	}
	b.buf = b.buf[0 : b.off+m+n]
	return b.off + m
}

func (b *Buffer) ReadFrom(i io.Reader) (int64, error) {
	file, ok := i.(os.File)

}

type Reader struct {
	i int
	b *Buffer
}

func (b *Buffer) NewReader() {

}

type ReverseReader struct {
	i int
	b []byte
}

func (r *reverseReader) ReadByte() (byte, error) {
	if r.eof != nil {
		<-r.eof
	}
	if r.i < 1 {
		return 0, io.EOF
	}
	r.i--
	return r.b[r.i], nil
}

func NewReverseReader(r *Reader, l int) *ReverseReader {
	return &reverseReader{len(buf), buf}
}
