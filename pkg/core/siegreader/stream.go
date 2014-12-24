package siegreader

//import "sync"

import "io"

//type protected struct {
//	sync.Mutex
//	val int
//}

// Scenarios:
// a) Stream - just copy into a big buffer as at present (... but if there is a MaxBof??)
type stream struct {
	buf  []byte
	w    protected
	eofc chan struct{}
	b    *Buffer
}

// Size returns the buffer's size, which is available immediately for files. Must wait for full read for streams.
func (s *stream) Size() int {
	select {
	case <-s.eofc:
		return len(s.buf)
	case <-s.b.quit:
		return 0
	}
}

// non-blocking Size(), for use with zip reader
func (s *stream) SizeNow() int64 {
	var err error
	for _, err = s.fill(); err == nil; _, err = s.fill() {
	}
	return int64(len(s.buf))
}

func (s *stream) grow() {
	// Rules for growing:
	// - if we need to grow, we have passed the initial read and can assume we will need whole file so, if we have a sz grow to it straight away
	// - otherwise, double capacity each time
	var buf []byte
	buf = make([]byte, cap(s.buf)*2)
	copy(buf, s.buf[:s.w.val]) // don't care about unlocking as grow() is only called by fill()
	s.buf = buf
}

// Rules for filling:
// - if we have a sz greater than 0, if there is stuff in the eof buffer, and if we are less than readSz from the end, copy across from the eof buffer
// - read readsz * 2 at a time
func (s *stream) fill() (int, error) {
	// if we've run out of room, grow the buffer
	if len(s.buf)-readSz < s.w.val {
		s.grow()
	}
	// otherwise, let's read
	e := s.w.val + readSz
	if e > len(s.buf) {
		e = len(s.buf)
	}
	i, err := s.b.src.Read(s.buf[s.w.val:e])
	if i < readSz {
		err = io.EOF // Readers can give EOF or nil here
	}
	if err != nil {
		close(s.eofc)
		if err == io.EOF {
			s.w.val += i
		}
		return s.w.val, err
	}
	s.w.val += i
	return s.w.val, nil
}

func (s *stream) fillEof() error {
	// return nil
	return nil
}

// Return a slice from the buffer that begins at offset s and has length l
func (s *stream) Slice(o, l int) ([]byte, error) {
	s.w.Lock()
	defer s.w.Unlock()
	var err error
	var bound int
	if o+l > s.w.val {
		for bound, err = s.fill(); o+l > bound && err == nil; bound, err = s.fill() {
		}
	}
	if err == nil {
		return s.buf[o : o+l], nil
	}
	if err == io.EOF {
		if o+l > s.w.val {
			if o > s.w.val {
				return nil, io.EOF
			}
			// in the case of an empty file
			if len(s.buf) == 0 {
				return nil, io.EOF
			}
			return s.buf[o:s.w.val], io.EOF
		} else {
			return s.buf[o : o+l], nil
		}
	}
	return nil, err
}

// Return a slice from the end of the buffer that begins at offset s and has length l.
// This will block until the slice is available (which may be until the full stream is read).
func (s *stream) EofSlice(o, l int) ([]byte, error) {
	// block until the EOF is available or we quit
	select {
	case <-s.b.quit:
		return nil, ErrQuit
	case <-s.eofc:
	}
	if o+l > len(s.buf) {
		if o > len(s.buf) {
			return nil, io.EOF
		}
		return s.buf[:len(s.buf)-o], io.EOF
	}
	return s.buf[len(s.buf)-(o+l) : len(s.buf)-o], nil
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
	s.w.Lock()
	defer s.w.Unlock()
	var err error
	var bound int
	if o > int64(s.w.val) {
		for bound, err = s.fill(); o > int64(bound) && err == nil; bound, err = s.fill() {
		}
	}
	if err == nil {
		return true, nil
	}
	if err == io.EOF {
		if o > int64(s.w.val) {
			return false, err
		}
		return true, nil
	}
	return false, err
}
