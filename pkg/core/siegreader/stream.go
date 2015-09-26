package siegreader

import (
	"io"
	"sync"
)

type stream struct {
	src  io.Reader
	buf  []byte
	eofc chan struct{}
	quit chan struct{}

	limited bool
	limit   chan struct{}

	mu  sync.Mutex
	i   int
	eof bool
}

func (s *stream) Stream() bool { return true }

func newStream() interface{} {
	return &stream{buf: make([]byte, readSz*2)}
}

func (s *stream) setSource(src io.Reader) {
	s.src = src
	s.eofc = make(chan struct{})
	s.quit = nil
	s.limited = false
	s.limit = nil
	s.i = 0
	s.eof = false
}

func (s *stream) SetQuit(q chan struct{}) {
	s.quit = q
}

func (s *stream) setLimit() {
	s.limited = true
	s.limit = make(chan struct{})
}

func (s *stream) waitLimit() {
	if s.limited {
		<-s.limit
	}
}

func (s *stream) hasQuit() bool {
	select {
	case <-s.quit:
		return true
	default:
	}
	return false
}

func (s *stream) reachedLimit() {
	close(s.limit)
}

// Size returns the buffer's size, which is available immediately for files. Must wait for full read for streams.
func (s *stream) Size() int64 {
	select {
	case <-s.eofc:
		return int64(s.i)
	case <-s.quit:
		return 0
	}
}

// non-blocking Size(), for use with zip reader
func (s *stream) SizeNow() int64 {
	var err error
	for _, err = s.fill(); err == nil; _, err = s.fill() {
	}
	return int64(s.i)
}

func (s *stream) grow() {
	// Rules for growing:
	// - if we need to grow, we have passed the initial read and can assume we will need whole file so, if we have a sz grow to it straight away
	// - otherwise, double capacity each time
	var buf []byte
	buf = make([]byte, cap(s.buf)*2)
	copy(buf, s.buf[:s.i]) // don't care about unlocking as grow() is only called by fill()
	s.buf = buf
}

// Rules for filling:
// - if we have a sz greater than 0, if there is stuff in the eof buffer, and if we are less than readSz from the end, copy across from the eof buffer
// - read readsz * 2 at a time
func (s *stream) fill() (int, error) {
	if s.eof {
		return s.i, io.EOF
	}
	// if we've run out of room, grow the buffer
	if len(s.buf)-readSz < s.i {
		s.grow()
	}
	// otherwise, let's read
	var i, j int
	var err error
	for {
		j, err = s.src.Read(s.buf[s.i+i : s.i+readSz])
		i += j
		if i >= readSz || err != nil {
			break
		}
	}
	if err != nil {
		close(s.eofc)
		s.eof = true
		if err == io.EOF {
			s.i += i
		}
		return s.i, err
	}
	s.i += i
	return s.i, nil
}

// Return a slice from the buffer that begins at offset off and has length l
func (s *stream) Slice(off int64, l int) ([]byte, error) {
	o := int(off)
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var bound int
	if o+l > s.i {
		for bound, err = s.fill(); o+l > bound && err == nil; bound, err = s.fill() {
		}
	}
	if err == nil {
		return s.buf[o : o+l], nil
	}
	if err == io.EOF {
		if o+l > s.i {
			if o > s.i {
				return nil, io.EOF
			}
			// in the case of an empty file
			if len(s.buf) == 0 {
				return nil, io.EOF
			}
			return s.buf[o:s.i], io.EOF
		} else {
			return s.buf[o : o+l], nil
		}
	}
	return nil, err
}

// Return a slice from the end of the buffer that begins at offset s and has length l.
// This will block until the slice is available (which may be until the full stream is read).
func (s *stream) EofSlice(o int64, l int) ([]byte, error) {
	// block until the EOF is available or we quit
	select {
	case <-s.quit:
		return nil, ErrQuit
	case <-s.eofc:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if int(o)+l > s.i {
		if int(o) >= s.i {
			return nil, io.EOF
		}
		return s.buf[:s.i-int(o)], io.EOF
	}
	return s.buf[s.i-(int(o)+l) : s.i-int(o)], nil
}

// fill until a seek to a particular offset is possible, then return true, if it is impossible return false
func (s *stream) canSeek(o int64, rev bool) (bool, error) {
	if rev {
		var err error
		for _, err = s.fill(); err == nil; _, err = s.fill() {
		}
		if err != io.EOF {
			return false, err
		}
		if int64(len(s.buf))-o < 0 {
			return false, nil
		}
		return true, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var bound int
	if o > int64(s.i) {
		for bound, err = s.fill(); o > int64(bound) && err == nil; bound, err = s.fill() {
		}
	}
	if err == nil {
		return true, nil
	}
	if err == io.EOF {
		if o > int64(s.i) {
			return false, err
		}
		return true, nil
	}
	return false, err
}
