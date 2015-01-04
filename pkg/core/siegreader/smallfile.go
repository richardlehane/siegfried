// Copyright 2014 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package siegreader

import (
	"io"
	"os"
	"sync"
)

type sfprotected struct {
	sync.Mutex
	val     int
	eofRead bool
}

// Buffer wraps an io.Reader, buffering its contents in byte slices that will keep growing until IO.EOF.
// It supports multiple concurrent Readers, including Readers reading from the end of the stream (ReverseReaders)
type SmallFile struct {
	quit      chan struct{} // allows quittting - otherwise will block forever while awaiting EOF
	src       io.Reader
	buf, eof  []byte
	completec chan struct{} // signals when the file has been completely read, allows EOF scanning beyond the small buffer
	complete  bool          // marks that the file has been completely read
	sz        int64
	w         sfprotected // index of latest write
}

// New instatatiates a new Buffer with a buf size of 4096*3, and an end-of-file buf size of 4096
func NewSF() *SmallFile {
	sf := &SmallFile{}
	sf.buf, sf.eof = make([]byte, initialRead), make([]byte, readSz)
	return sf
}

func (sf *SmallFile) reset() {
	sf.completec = make(chan struct{})
	sf.complete = false
	sf.sz = 0
	sf.w.Lock()
	sf.w.val = 0
	sf.w.eofRead = false
	sf.w.Unlock()
}

// SetSource sets the buffer's source.
// Can be any io.Reader. If it is an os.File, will load EOF buffer early. Otherwise waits for a complete read.
// The source can be reset to recycle an existing Buffer.
// Siegreader blocks on EOF reads or Size() calls when the reader isn't a file or the stream isn't completely read. The quit channel overrides this block.
func (sf *SmallFile) SetSource(r io.Reader) error {
	if sf == nil {
		return ErrNilBuffer
	}
	sf.reset()
	sf.src = r
	file := r.(*os.File)
	info, err := file.Stat()
	if err != nil {
		return err
	}
	sf.sz = info.Size()
	if sf.sz > int64(initialRead) {
		sf.eof = sf.eof[:cap(sf.eof)]
	} else {
		sf.eof = sf.eof[:0]
	}
	_, err = sf.fill() // initial fill
	return err
}

func (sf *SmallFile) SetQuit(q chan struct{}) {
	sf.quit = q
}

// Size returns the buffer's size, which is available immediately for files. Must wait for full read for streams.
func (sf *SmallFile) Size() int64 {
	return sf.sz
}

// non-blocking Size(), for use with zip reader
func (sf *SmallFile) SizeNow() int64 {
	return sf.sz
}

func (sf *SmallFile) grow() {
	buf := make([]byte, int(sf.sz)) // safe to convert int because it is a small file
	copy(buf, sf.buf[:sf.w.val])    // don't care about unlocking as grow() is only called by fill()
	sf.buf = buf
}

// Rules for filling:
// - if we have a sz greater than 0, if there is stuff in the eof buffer, and if we are less than readSz from the end, copy across from the eof buffer
// - read readsz * 2 at a time
func (sf *SmallFile) fill() (int, error) {
	// if we've run out of room, grow the buffer
	if len(sf.buf)-readSz < sf.w.val {
		sf.grow()
	}
	// if we have an eof buffer, and we are near the end of the file, avoid an extra read by copying straight into the main buffer
	if len(sf.eof) > 0 && sf.w.eofRead && sf.w.val+readSz >= int(sf.sz) {
		close(sf.completec)
		sf.complete = true
		lr := int(sf.sz) - sf.w.val
		sf.w.val += copy(sf.buf[sf.w.val:sf.w.val+lr], sf.eof[readSz-lr:])
		return sf.w.val, io.EOF
	}
	// otherwise, let's read
	e := sf.w.val + readSz
	if e > len(sf.buf) {
		e = len(sf.buf)
	}
	i, err := sf.src.Read(sf.buf[sf.w.val:e])
	if i < readSz {
		err = io.EOF // Readers can give EOF or nil here
	}
	if err != nil {
		close(sf.completec)
		sf.complete = true
		if err == io.EOF {
			sf.w.val += i
			// if we haven't got an eof buf already
			if len(sf.eof) < readSz {
				sf.sz = int64(sf.w.val)
			}
		}
		return sf.w.val, err
	}
	sf.w.val += i
	return sf.w.val, nil
}

func (sf *SmallFile) fillEof() error {
	// return nil if file too small
	if len(sf.eof) < readSz {
		return nil
	}
	sf.w.Lock()
	defer sf.w.Unlock()
	if sf.w.eofRead {
		return nil // another reverse reader has already filled the buffer
	}
	rs := sf.src.(io.ReadSeeker)
	_, err := rs.Seek(0-int64(readSz), 2)
	if err != nil {
		return err
	}
	_, err = rs.Read(sf.eof)
	if err != nil {
		return err
	}
	_, err = rs.Seek(int64(sf.w.val), 0)
	if err != nil {
		return err
	}
	sf.w.eofRead = true
	return nil
}

// Return a slice from the buffer that begins at offset s and has length l
func (sf *SmallFile) Slice(off int64, l int, w bool) ([]byte, error) {
	if !w {
		return sf.eofSlice(off, l)
	}
	s := int(off)
	sf.w.Lock()
	defer sf.w.Unlock()
	var err error
	var bound int
	if s+l > sf.w.val && !sf.complete {
		for bound, err = sf.fill(); s+l > bound && err == nil; bound, err = sf.fill() {
		}
	}
	if err == nil && !sf.complete {
		return sf.buf[s : s+l], nil
	}
	if err == io.EOF || sf.complete {
		if s+l > sf.w.val {
			if s > sf.w.val {
				return nil, io.EOF
			}
			// in the case of an empty file
			if sf.Size() == 0 {
				return nil, io.EOF
			}
			return sf.buf[s:sf.w.val], io.EOF
		} else {
			return sf.buf[s : s+l], nil
		}
	}
	return nil, err
}

// Return a slice from the end of the buffer that begins at offset s and has length l.
// This will block until the slice is available (which may be until the full stream is read).
func (sf *SmallFile) eofSlice(off int64, l int) ([]byte, error) {
	s := int(off)
	var buf []byte
	if len(sf.eof) > 0 && s+l <= len(sf.eof) {
		buf = sf.eof
	} else {
		select {
		case <-sf.quit:
			return nil, ErrQuit
		case <-sf.completec:
		}
		buf = sf.buf[:int(sf.sz)]
		if s+l == len(buf) {
			return buf[:len(buf)-s], io.EOF
		}
	}
	if s+l > len(buf) {
		if s > len(buf) {
			return nil, io.EOF
		}
		return buf[:len(buf)-s], io.EOF
	}
	return buf[len(buf)-(s+l) : len(buf)-s], nil
}

// fill until a seek to a particular offset is possible, then return true, if it is impossible return false
func (sf *SmallFile) canSeek(o int64, w bool) (bool, error) {
	if !w {
		o = sf.sz - o
		if o < 0 {
			return false, nil
		}
	}
	sf.w.Lock()
	defer sf.w.Unlock()
	var err error
	var bound int
	if o > int64(sf.w.val) {
		for bound, err = sf.fill(); o > int64(bound) && err == nil; bound, err = sf.fill() {
		}
	}
	if err == nil {
		return true, nil
	}
	if err == io.EOF {
		if o > int64(sf.w.val) {
			return false, err
		}
		return true, nil
	}
	return false, err
}
