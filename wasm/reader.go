//go:build js

package main

import (
	"io"
	"syscall/js"
)

const bufSz int = 4096 * 8

type reader struct {
	sz   int64
	idx  int64 // index for Read calls
	off  int64 // offset current contents of buf
	l    int   // len current contents of buf
	file js.Value
	data js.Value // Uint8Array
	buf  []byte
}

func newReader() *reader {
	return &reader{
		buf: make([]byte, bufSz),
	}
}

func (r *reader) reset(v js.Value) {
	r.sz = int64(v.Get("size").Int())
	r.off = 0
	r.l = 0
	r.file = v
}

func (r *reader) IsSlicer() bool { return true }

func (r *reader) Size() int64 { return r.sz }

func (r *reader) fill(off int64) error {
	if off >= r.sz {
		return io.EOF
	}
	if off < 0 {
		off = 0 // could happen in a reverse fill
	}
	l := len(r.buf)
	if r.sz-off < int64(l) {
		l = int(r.sz - off)
	}
	blob := r.file.Call("slice", off, off+int64(l))
	arrbuf, err := await(blob.Call("arrayBuffer"))
	if err != nil {
		return err
	}
	r.data = js.Global().Get("Uint8Array").New(arrbuf)
	js.CopyBytesToGo(r.buf, r.data)
	r.off = off
	r.l = l
	return nil
}

// Slice returns a byte slice with size l from a given offset
func (r *reader) Slice(off int64, l int) ([]byte, error) {
	if off >= r.sz {
		return nil, io.EOF
	}
	var err error
	if l > int(r.sz-off) {
		l, err = int(r.sz-off), io.EOF
	}
	if l > len(r.buf) {
		r.buf = make([]byte, l)
	}
	if off < r.off || (off+int64(l)) > (r.off+int64(r.l)) { // need to fill
		if err1 := r.fill(off); err1 != nil {
			return nil, err1
		}
	}
	return r.buf[int(off-r.off) : int(off-r.off)+l], err
}

// Slice returns a byte slice with size l from a given offset from the end of the content of the record.
func (r *reader) EofSlice(off int64, l int) ([]byte, error) {
	if off >= r.sz {
		return nil, io.EOF
	}
	var err error
	if l > int(r.sz-off) {
		l, off, err = int(r.sz-off), 0, io.EOF
	} else {
		off = r.sz - off - int64(l)
	}
	if l > len(r.buf) {
		r.buf = make([]byte, l)
	}
	if off < r.off || (off+int64(l)) > (r.off+int64(r.l)) { // need to fill
		noff := off + int64(l) - int64(len(r.buf))
		if err1 := r.fill(noff); err1 != nil {
			return nil, err1
		}
	}
	return r.buf[int(off-r.off) : int(off-r.off)+l], err
}

func (r *reader) Read(p []byte) (int, error) {
	if r.idx >= r.sz {
		return 0, io.EOF
	}
	l := len(p)
	if int64(len(p)) > r.sz-r.idx {
		l = int(r.sz - r.idx)
	}
	buf, err := r.Slice(r.idx, l)
	l = copy(p, buf)
	r.idx += int64(l)
	return l, err
}
