package siegreader

import (
	"bytes"
	"io"
	"os"
	"sync"
)

var readSz int = 8192

type protected struct {
	sync.RWMutex
	sz int64
	w  int
}

// Buffer

type Buffer struct {
	eofc  chan struct{} // signals if EOF bytes are available. When EOF bytes are available, this chan is closed
	contc chan struct{} // communicate with the fill goroutine. Nil the chan to halt, send an empty struct to continue.
	src   io.Reader

	prot protected

	buf []byte
	eof []byte
}

func New() *Buffer {
	b := new(Buffer)
	b.buf, b.tail = make([]byte, readSz*2), make([]byte, readSz)
	b.cont = make(chan struct{})
	go b.fill()
	return b
}

func (b *Buffer) reset() {
	b.eof = make(chan struct{})
	b.buf = b.buf[:readSz*2]
	b.prot.Lock()
	b.w, b.sz = 0, 0
	b.prot.Unlock()
}

// Set the buffer's source.
// Can be any io.Reader. If it is an os.File, will load EOF buffer early. Otherwise waits for a complete read.
func (b *Buffer) SetSource(i io.Reader) error {
	b.reset()
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
	return nil
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
	// - if we have a sz greater than 0, if there is stuff in the eof buffer, and if we are less than readSz from the end, copy across from the eof buffer
	// - read readsz * 2 at a time

	for {
		<-b.cont // wait for a continue signal
		if len(b.buf)-readSz < b.w {
			b.buf.grow(readSz)
		}
		i, err := b.src.Read(b.buf[w : w+readSz])
		if err != nil {
			if err == io.EOF {
				switch {
				case <-b.eof:
					return
				}
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
