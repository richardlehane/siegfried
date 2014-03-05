package siegreader

import (
	"bytes"
	"io"
	"os"
)

var readSz int64 = 4096

type Buffer struct {
	eof  chan struct{}
	cont chan struct{}
	quit chan struct{}
	src  io.Reader
	w    int
	buf  []byte
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

func (b *Buffer) fill() error {
	for {
		select {
		case <-b.quit:
			return nil
		case <-b.cont:
		}
		if len(b.buf)-readSz < b.w {
			b.buf.grow(readSz)
		}
		i, err := b.src.Read(b.buf[w : w+readSz])
		if err != nil {
			return err
		}
	}
}

func (b *Buffer) Slice(s, e int) []byte {

}

func (b *Buffer) ReadFrom(i io.Reader) (int64, error) {
	file, ok := i.(os.File)
	if ok {
		info, err := file.Stat()
		if err != nil {
			return 0, nil
		}
		l := info.Size()
		if int(l) > readSz {

		}

	} else {

	}

}

type Reader struct {
	i int
	b *Buffer
}

func (b *Buffer) NewReader() *Reader {
	return &Reader{0, b}
}


func (b *Buffer) ReadByte() *Reader {
	if b.w + readSz > len(b.buf) {
		b.cont <- 
	}

type ReverseReader struct {
	i int
	b *Buffer
}

func (b *Buffer) NewReverseReader() *ReverseReader {
	return &ReverseReader{0, b}
}

func (r *ReverseReader) ReadByte() (byte, error) {
	// block if a stream, not a file
	if r.eof != nil {
		<-r.eof
	}

	if r.i > readSz {
		return 0, io.EOF
	}
	r.i++
	if len(b.tail) > 0 {
		return b.tail[len(b.tail)-r.i], nil
	} else {
		return b.buf[len(b.buf)-r.i], nil
	}
}
