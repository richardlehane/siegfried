package siegreader

import (
	"bytes"
	"io"
	"os"
	"sync"
)

var readSz int = 8192

// Buffer

type Buffer struct {
	eofc  chan struct{} // signals if EOF bytes are available. When EOF bytes are available, this chan is closed
	contc chan struct{} // communicate with the fill goroutine. Nil the chan to halt, send an empty struct to continue.
	src   io.Reader
	sz    int64 // size of the source, if known

	w  int          // current write position
	wm sync.RWMutex // prevents race between multiple goroutines reading current write position

	buf []byte
	eof []byte
}

func New() *Buffer {
	b := new(Buffer)
	b.buf, b.tail = make([]byte, readSz*2), make([]byte, readSz)
	return b
}

func (b *Buffer) grow() {
	// Rules for growing:
	// - if we need to grow, we have passed the BOF and can assume we will need whole file so, if we have a sz grow to it straight away
	// - otherwise, double capacity each time
	var buf []byte
	if sz > 0 {
		buf = make([]byte, int(sz))
	} else {
		buf = make([]byte, cap(b.buf)*2)
	}
	copy(buf, b.buf[:b.w])
	b.buf = buf
}

func (b *Buffer) fill() {
	// Rules for filling:
	// - if we have a sz greater than 0, if there is stuff in the eof buffer, and if we are less than readz from the end, copy across from the eof buffer
	// - read readsz * 2 at a time

	for {
		if b.cont == nil {
			return
		}
		<-b.cont
		if len(b.buf)-readSz < b.w {
			b.buf.grow(readSz)
		}
		i, err := b.src.Read(b.buf[w : w+readSz])
		if err != nil {
			if err == io.EOF {
				close(b.eof)
			}
			return
		}
	}
}

func (b *Buffer) fillEof() {
	if len(b.eof) < int(readSz) {
		return
	}
	_, err := b.src.Seek(readSz, 2)
	if err != nil {
		panic(err)
	}
	i, err := b.src.Read(b.eof)
	if err != nil {
		if err == io.EOF {
			if int64(i) < readSz {
				b.eof = b.eof[:i]
			}
		} else {
			panic(err)
		}
	}
	_, err = b.src.Seek(int64(b.w), 0)
	if err != nil {
		panic(err)
	}
	close(b.eofc)
}

func (b *Buffer) Reset() {
	if b.cont != nil {
		b.cont = nil // make sure we aren't firing up multiple fill() goroutines
	}
	b.eof, b.cont = make(chan struct{}), make(chan struct{})
	b.buf = b.buf[:readSz*2]
	b.w, b.sz = 0, 0
}

func (b *Buffer) ReadFrom(i io.Reader) error {
	b.Reset()
	b.src = i
	file, ok := i.(os.File)
	if ok {
		info, err := file.Stat()
		if err != nil {
			return 0, err
		}
		b.sz = info.Size()
		if b.sz > readSz*3 {
			b.tail = b.tail[:int(readSz)]
		} else {
			b.tail = b.tail[:0]
		}
	} else {
		b.sz = 0
		b.tail = b.tail[:0]
	}
	go b.fill()
	return nil
}

// Slice

func (b *Buffer) Slice(s, e int) ([]byte, error) {
	// block until the slice is available

}

// Reader

type Reader struct {
	i int
	b *Buffer
}

func (r *Buffer) NewReader() *Reader {
	// A BOF reader may not have been used, trigger a fill if necessary.
	return &Reader{0, b}
}

func (r *Reader) Read(b []byte) (int, error) {

}

func (r *Reader) ReadByte() (byte, error) {
	if b.w+readSz > len(b.buf) {
		b.cont <- struct{}{}
	}
}

func (r *Reader) ReadAt(b []byte, off int64) (int, error) {
	// block until the read is possible
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	// block until the seek is possible

}

// BOF Reader

type BOFReader struct {
	i int
	b *Buffer
}

func (b *Buffer) NewBOFReader() *BOFReader {
	// fill the BOF now, if not already done
	return &BOFReader{0, b}
}

func (r *BOFReader) Read(b []byte) (int, error) {

}

func (r *BOFReader) ReadByte() (byte, error) {

}

// Reverse Reader

type EOFReader struct {
	i int
	b *Buffer
}

func (b *Buffer) NewEOFReader() *ReverseReader {
	// fill the EOF now, if possible and not already done
	return &ReverseReader{0, b}
}

func (r *ReverseReader) Read(b []byte) (int, error) {

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
