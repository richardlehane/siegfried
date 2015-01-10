package siegreader

// b) File
//    b i)   Satisifed with small read beginning
//    b ii)  Small enough for full read
//    b iii) Mmap
//    b iv) Too big for MMap - small buffers and expose ReaderAt

import (
	"io"
	"os"
)

const (
	initialRead = readSz * 4
)

type data interface {
	slice(offset int64, length int) ([]byte, error)
}

type file struct {
	bof  []byte
	sz   int64
	quit chan struct{}
	src  *os.File
	data
}

func (f *file) SetQuit(q chan struct{}) {
	f.quit = q
}

func (f *file) setSource(src io.Reader) error {
	if f == nil {
		return ErrNilBuffer
	}
	f.src = src.(*os.File)
	info, err := f.src.Stat()
	if err != nil {
		return err
	}
	f.sz = info.Size()
	i, err := f.src.Read(f.bof)
	if err == nil && i < initialRead {
		return io.EOF
	}
	return err
}

func (f *file) Size() int64 { return f.sz }

func (f *file) SizeNow() int64 { return f.sz }

func (f *file) Slice(off int64, length int) ([]byte, error) {
	// return EOF if offset is larger than the file size
	if off >= f.sz {
		return nil, io.EOF
	}
	// the slice falls entirely in the rem segment
	if off > int64(initialRead) {
		if f.data == nil {

		}
		return f.slice(off, length)
	}
	// shorten the length, if necessary
	var err error
	if off+int64(length) > f.sz {
		length = int(f.sz - off)
		err = io.EOF
	}
	// the slice falls entirely in the bof segment
	if off+int64(length) <= int64(initialRead) {
		return f.bof[int(off):length], err
	}
	// the slice spans the bof and rem segments, so copy to fresh slice
	ret := make([]byte, length)
	copy(ret, f.bof[int(off):])
	start := initialRead - int(off)
	if f.data == nil {

	}
	slc, _ := f.slice(off, length)
	copy(ret[start:], slc)
	return ret, err
}

func (f *file) EofSlice(off int64, l int) ([]byte, error) {
	return nil, nil
}

func (f *file) canSeek(off int64, whence bool) (bool, error) {
	if f.sz < off {
		return false, nil
	}
	return true, nil
}
