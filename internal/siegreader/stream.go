package siegreader

import (
	"io"
	"io/ioutil"
	"os"
	"sync"
)

type stream struct {
	b     *Buffer
	src   io.Reader
	sz    int64
	buf   []byte
	tf    *os.File // temp backing file - used when stream exceeds streamSz
	tfBuf []byte
	eofc  chan struct{}

	mu  sync.Mutex
	i   int // marks how much of buf we have filled
	eof bool
}

func newStream() interface{} {
	return &stream{buf: make([]byte, readSz*2), tfBuf: make([]byte, readSz)}
}

func (s *stream) setSource(src io.Reader, b *Buffer) error {
	s.b = b
	s.src = src
	s.sz = 0
	s.eofc = make(chan struct{})
	s.i = 0
	s.eof = false
	_, err := s.fill()
	return err
}

// close and delete temp file, if exists
func (s *stream) cleanUp() {
	if s.tf == nil {
		return
	}
	s.tf.Close()
	os.Remove(s.tf.Name())
	s.tf = nil
}

// Size returns the buffer's size, which is available immediately for files. Must wait for full read for streams.
func (s *stream) Size() int64 {
	select {
	case <-s.eofc:
		return s.sz
	case <-s.b.Quit:
		return 0
	}
}

// SizeNow is a non-blocking Size(). Will force a full read of a stream.
func (s *stream) SizeNow() int64 {
	var err error
	for _, err = s.fill(); err == nil; _, err = s.fill() {
	}
	return s.sz
}

func (s *stream) grow() error {
	if s.tf != nil { // return if we already have a temp file
		return nil
	}
	c := cap(s.buf) * 2
	if c > streamSz {
		if cap(s.buf) < streamSz {
			c = streamSz
		} else { // if we've exceeded streamSz, use a temp file to copy remainder
			var err error
			s.tf, err = ioutil.TempFile("", "siegfried")
			return err
		}
	}
	buf := make([]byte, c)
	copy(buf, s.buf[:s.i]) // don't care about unlocking as grow() is only called by fill()
	s.buf = buf
	return nil
}

func (s *stream) fill() (int64, error) {
	// have already scanned to the end of the stream
	if s.eof {
		return s.sz, io.EOF
	}
	// if we've run out of room in buf, & there is no backing file, grow the buffer
	if len(s.buf)-readSz < s.i && s.tf == nil {
		s.grow()
	}
	// now let's read
	var err error
	if s.tf != nil {
		// if we have a backing file, fill that
		var wi int64
		wi, err = io.CopyBuffer(s.tf, io.LimitReader(s.src, int64(readSz)), s.tfBuf)
		if wi < int64(readSz) && err == nil {
			err = io.EOF
		}
		// update s.sz
		s.sz += wi
	} else {
		// otherwise, fill the slice
		var i int
		i, err = io.ReadFull(s.src, s.buf[s.i:s.i+readSz])
		s.i += i
		s.sz += int64(i)
		if err == io.ErrUnexpectedEOF {
			err = io.EOF
		}
	}
	if err != nil {
		close(s.eofc)
		s.eof = true
		if err == io.EOF && s.sz == 0 {
			err = ErrEmpty
		}
	}
	return s.sz, err
}

// Slice returns a byte slice from the buffer that begins at offset off and has length l.
func (s *stream) Slice(off int64, l int) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var bound int64
	furthest := off + int64(l)
	if furthest > s.sz {
		for bound, err = s.fill(); furthest > bound && err == nil; bound, err = s.fill() {
		}
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	if err == io.EOF {
		// in the case of an empty file
		if s.sz == 0 {
			return nil, io.EOF
		}
		if furthest > s.sz {
			if off >= s.sz {
				return nil, io.EOF
			}
			l = int(s.sz - off)
		}
	}
	// slice is wholly in buf
	if off+int64(l) <= int64(len(s.buf)) {
		return s.buf[int(off) : int(off)+l], err
	}
	ret := make([]byte, l)
	// if slice crosses border, copy first bit from end of buf
	var ci int
	if off < int64(len(s.buf)) {
		ci = copy(ret, s.buf[int(off):])
		off = 0
	} else {
		off -= int64(len(s.buf))
	}
	_, rerr := s.tf.ReadAt(ret[ci:], off)
	if rerr != nil {
		err = rerr
	}
	return ret, err
}

// EofSlice returns a slice from the end of the buffer that begins at offset s and has length l.
// Blocks until the slice is available (which may be until the full stream is read).
func (s *stream) EofSlice(o int64, l int) ([]byte, error) {
	// block until the EOF is available or we quit
	select {
	case <-s.b.Quit:
		return nil, ErrQuit
	case <-s.eofc:
	}
	if o >= s.sz {
		return nil, io.EOF
	}
	var err error
	if o+int64(l) > s.sz {
		l = int(s.sz - o)
		err = io.EOF
	}
	slc, serr := s.Slice(s.sz-o-int64(l), l)
	if serr != nil {
		err = serr
	}
	return slc, err
}

// fill until a seek to a particular offset is possible, then return true, if it is impossible return false
func (s *stream) CanSeek(o int64, rev bool) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rev {
		var err error
		for _, err = s.fill(); err == nil; _, err = s.fill() {
		}
		if err != io.EOF {
			return false, err
		}
		if o >= s.sz {
			return false, nil
		}
		return true, nil
	}
	var err error
	var bound int64
	if o > s.sz {
		for bound, err = s.fill(); o > bound && err == nil; bound, err = s.fill() {
		}
	}
	if err == nil {
		return true, nil
	}
	if err == io.EOF {
		if o > s.sz {
			return false, err
		}
		return true, nil
	}
	return false, err
}
