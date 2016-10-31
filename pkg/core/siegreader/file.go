package siegreader

import (
	"io"
	"os"
	"sync"
)

type file struct {
	peek [initialRead]byte
	sz   int64
	src  *os.File
	once *sync.Once
	data

	pool *datas // link to the data pool
}

func newFile() interface{} { return &file{once: &sync.Once{}} }

type data interface {
	slice(offset int64, length int) []byte
	eofSlice(offset int64, length int) []byte
}

func (f *file) setSource(src *os.File, p *datas) error {
	// reset
	f.once = &sync.Once{}
	f.data = nil
	f.pool = p
	f.src = src
	info, err := f.src.Stat()
	if err != nil {
		return err
	}
	f.sz = info.Size()
	i, err := f.src.Read(f.peek[:])
	if i < initialRead && (err == nil || err == io.EOF) {
		if i == 0 {
			return ErrEmpty
		}
		if err == nil {
			return io.EOF
		}
	}
	return err
}

// Size returns the buffer's size, which is available immediately for files. Must wait for full read for streams.
func (f *file) Size() int64 { return f.sz }

// SizeNow is a non-blocking Size().
func (f *file) SizeNow() int64 { return f.sz }

func (f *file) CanSeek(off int64, whence bool) (bool, error) {
	if f.sz < off {
		return false, nil
	}
	return true, nil
}

// Slice returns a byte slice from the buffer that begins at offset off and has length l.
func (f *file) Slice(off int64, l int) ([]byte, error) {
	// return EOF if offset is larger than the file size
	if off >= f.sz {
		return nil, io.EOF
	}
	// shorten the length, if necessary
	var err error
	if off+int64(l) > f.sz {
		l = int(f.sz - off)
		err = io.EOF
	}
	// the slice falls entirely in the bof segment
	if off+int64(l) <= int64(initialRead) {
		return f.peek[int(off) : int(off)+l], err
	}
	f.once.Do(func() {
		f.data = f.pool.get(f)
	})
	ret := f.slice(off, l)
	return ret, err
}

// EofSlice returns a slice from the end of the buffer that begins at offset s and has length l.
// May block until the slice is available.
func (f *file) EofSlice(off int64, l int) ([]byte, error) {
	// if the offset is larger than the file size, it is invalid
	if off >= f.sz {
		return nil, io.EOF
	}
	// shorten the length, if necessary
	var err error
	if off+int64(l) > f.sz {
		l = int(f.sz - off)
		err = io.EOF
	}
	// the slice falls entirely in the bof segment
	if f.sz-off <= int64(initialRead) {
		return f.peek[int(f.sz-off)-l : int(f.sz-off)], err
	}
	f.once.Do(func() {
		f.data = f.pool.get(f)
	})
	ret := f.eofSlice(off, l)
	return ret, err
}
