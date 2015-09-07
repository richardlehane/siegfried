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
	"fmt"
	"io"
)

/*
//   rdr := siegreader.ReaderFrom(buffer)
//	 second_rdr := siegreader.ReaderFrom(buffer)
//   brdr := siegreader.ByteReaderFrom(buffer, -1)
//   rrdr := siegreader.ReverseByteReaderFrom(buffer, 16000)
*/

// Reader implements the io.Reader, io.Seeker, io.ByteReader and io.ReaderAt interfaces
// The special thing about a siegreader.Reader is that you can have a bunch of them all reading independently from the one buffer.
type Reader struct {
	i       int64
	j       int
	scratch []byte
	end     bool // buffer adjoins the end of the file
	Buffer
}

func ReaderFrom(b Buffer) *Reader {
	// A BOF reader may not have been used, trigger a fill if necessary.
	return &Reader{0, 0, nil, false, b}
}

func (r *Reader) setBuf(o int64) error {
	var err error
	r.scratch, err = r.Slice(o, readSz)
	if err == io.EOF {
		r.end = true
	}
	return err
}

func (r *Reader) ReadByte() (byte, error) {
	if r.j >= len(r.scratch) {
		if r.end {
			return 0, io.EOF
		}
		err := r.setBuf(r.i)
		if err != nil && err != io.EOF {
			return 0, err
		}
		if len(r.scratch) == 0 {
			return 0, io.EOF
		}
		r.j = 0
	}
	b := r.scratch[r.j]
	r.i++
	r.j++
	return b, nil
}

func (r *Reader) Read(b []byte) (int, error) {
	var slc []byte
	var err error
	if len(b) > len(r.scratch)-r.j {
		slc, err = r.Slice(r.i, len(b))
		if err != nil {
			if err != io.EOF {
				return 0, err
			}
			r.end = true
		}
	} else {
		slc = r.scratch[r.j : r.j+len(b)]
	}
	n := copy(b, slc)
	r.i += int64(n)
	r.j += n
	return len(slc), err
}

func (r *Reader) ReadAt(b []byte, off int64) (int, error) {
	var slc []byte
	var err error
	// if b is already covered by the scratch slice
	if off >= r.i-int64(r.j) && off+int64(len(b)) <= r.i-int64(r.j+len(r.scratch)) {
		s := int(off-r.i) - r.j
		slc = r.scratch[s : s+len(b)]
	} else {
		slc, err = r.Slice(off, len(b))
		if err != nil {
			if err != io.EOF {
				return 0, err
			}
			r.end = true
		}
	}
	copy(b, slc)
	return len(slc), err
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	var rev bool
	switch whence {
	case 0:
	case 1:
		offset = offset + int64(r.i)
	case 2:
		rev = true
	default:
		return 0, fmt.Errorf("Siegreader: Seek error, whence value must be one of 0,1,2 got %v", whence)
	}
	success, err := r.canSeek(offset, rev)
	if success {
		if rev {
			offset = r.Size() - offset
		}
		d := offset - r.i
		r.i = offset
		r.j += int(d) // add the jump distance to r.j PROBLEM - WHAT IF r.j < 0!!
		return offset, err
	}
	return 0, err
}

// ReverseReader implements the io.Reader and io.ByteReader interfaces, but for each it does so from the end of the io source working backwards.
// Like Readers, you can have multiple ReverseReaders all reading independently from the same buffer.
type ReverseReader struct {
	i       int64
	j       int
	scratch []byte
	end     bool // if buffer is adjacent to the BOF, i.e. we have scanned all the way back to the beginning
	Buffer
}

func ReverseReaderFrom(b Buffer) *ReverseReader {
	return &ReverseReader{0, 0, nil, false, b}
}

func (r *ReverseReader) setBuf(o int64) error {
	var err error
	r.scratch, err = r.EofSlice(o, readSz)
	if err == io.EOF {
		r.end = true
	}
	return err
}

func (r *ReverseReader) Read(b []byte) (int, error) {
	if r.i == 0 {
		r.setBuf(0)
	}
	var slc []byte
	var err error
	if len(b) > len(r.scratch)-r.j {
		slc, err = r.EofSlice(r.i, len(b))
		if err != nil {
			if err != io.EOF {
				return 0, err
			}
			r.end = true
		}
	} else {
		slc = r.scratch[len(r.scratch)-len(b) : len(r.scratch)-r.j]
	}
	n := copy(b, slc)
	r.i += int64(n)
	r.j += n
	return len(slc), err
}

func (r *ReverseReader) ReadByte() (byte, error) {
	if r.j >= len(r.scratch) {
		if r.end {
			return 0, io.EOF
		}
		err := r.setBuf(r.i)
		if err != nil && err != io.EOF {
			return 0, err
		}
		r.j = 0
	}
	b := r.scratch[len(r.scratch)-r.j-1]
	r.i++
	r.j++
	return b, nil
}

type LimitReader struct {
	limit int
	*Reader
}

func LimitReaderFrom(b Buffer, l int) *LimitReader {
	// A BOF reader may not have been used, trigger a fill if necessary.
	r := &Reader{0, 0, nil, false, b}
	b.setLimit()
	return &LimitReader{l, r}
}

func (l *LimitReader) ReadByte() (byte, error) {
	if l.limit >= 0 && l.i >= int64(l.limit) {
		l.reachedLimit()
		l.limit = -1 // only run once
		return 0, io.EOF
	}
	if l.j >= len(l.scratch) {
		if l.end {
			l.reachedLimit()
			return 0, io.EOF
		}
		if l.hasQuit() {
			l.reachedLimit()
			return 0, io.EOF
		}
		err := l.setBuf(l.i)
		if err != nil && err != io.EOF {
			l.reachedLimit()
			return 0, err
		}
		if len(l.scratch) == 0 {
			l.reachedLimit()
			return 0, io.EOF
		}
		l.j = 0
	}
	b := l.scratch[l.j]
	l.i++
	l.j++
	return b, nil
}

type LimitReverseReader struct {
	limit int
	*ReverseReader
}

func LimitReverseReaderFrom(b Buffer, l int) *LimitReverseReader {
	// fill the EOF now, if possible and not already done
	return &LimitReverseReader{l, &ReverseReader{0, 0, nil, false, b}}
}

func (r *LimitReverseReader) ReadByte() (byte, error) {
	if r.i >= int64(r.limit) {
		return 0, io.EOF
	}
	if r.j >= len(r.scratch) {
		if r.end {
			return 0, io.EOF
		}
		if r.limit < 0 && r.i >= int64(eofSz) {
			r.waitLimit()
		}
		if r.hasQuit() {
			return 0, io.EOF
		}
		err := r.setBuf(r.i)
		if err != nil && err != io.EOF || len(r.scratch) == 0 {
			return 0, err
		}
		r.j = 0
	}
	b := r.scratch[len(r.scratch)-r.j-1]
	r.i++
	r.j++
	return b, nil
}
