package siegreader

import (
	"io"
	"os"
	"sync"
)

type file struct {
	peek [initialRead]byte
	sz   int64
	quit chan struct{}
	src  *os.File
	once *sync.Once
	data

	limited bool
	limit   chan struct{}

	pool *datas // link to the data pool
}

type data interface {
	slice(offset int64, length int) []byte
	eofSlice(offset int64, length int) []byte
}

func (f *file) setLimit() {
	f.limited = true
	f.limit = make(chan struct{})
}

func (f *file) waitLimit() {
	if f.limited {
		select {
		case <-f.limit:
		case <-f.quit:
		}
	}
}

func (f *file) hasQuit() bool {
	select {
	case <-f.quit:
		return true
	default:
	}
	return false
}

func (f *file) reachedLimit() {
	close(f.limit)
}
func newFile() interface{} { return &file{once: &sync.Once{}} }

func (f *file) setSource(src *os.File, p *datas) error {
	// reset
	f.once = &sync.Once{}
	f.data = nil
	f.limit = nil
	f.limited = false

	f.pool = p
	f.src = src
	info, err := f.src.Stat()
	if err != nil {
		return err
	}
	f.sz = info.Size()
	i, err := f.src.Read(f.peek[:])
	if err == nil && i < initialRead {
		return io.EOF
	}
	return err
}

func (f *file) SetQuit(q chan struct{}) { f.quit = q }

func (f *file) Size() int64 { return f.sz }

func (f *file) SizeNow() int64 { return f.sz }

func (f *file) canSeek(off int64, whence bool) (bool, error) {
	if f.sz < off {
		return false, nil
	}
	return true, nil
}

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
	f.once.Do(func() { f.data = f.pool.get(f) })
	ret := f.slice(off, l)
	return ret, err
}

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
	if int(f.sz-off) <= initialRead {
		return f.peek[int(f.sz-off)-l : int(f.sz-off)], err
	}
	f.once.Do(func() { f.data = f.pool.get(f) })
	ret := f.eofSlice(off, l)
	return ret, err
}
