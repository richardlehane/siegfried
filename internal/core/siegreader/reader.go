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

	"github.com/richardlehane/characterize"
)

// Reader implements the io.Reader, io.Seeker, io.ByteReader and io.ReaderAt interfaces
// The special thing about a siegreader.Reader is that you can have a bunch of them all reading independently from the one buffer.
//
// Example:
//  buffers := siegreader.New()
//  buffer := buffers.Get(underlying_io_reader)
//  rdr := siegreader.ReaderFrom(buffer)
//	second_rdr := siegreader.ReaderFrom(buffer)
//  limit_rdr := siegreader.LimitedReaderFrom(buffer, 4096)
//  reverse_rdr := siegreader.ReverseReaderFrom(buffer)
type Reader struct {
	i       int64
	j       int
	scratch []byte
	end     bool // buffer adjoins the end of the file
	*Buffer
}

// ReaderFrom returns a Reader reading from the Buffer.
func ReaderFrom(b *Buffer) *Reader {
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

// ReadByte implements the io.ByteReader interface.
// Checks the quit channel every 4096 bytes.
func (r *Reader) ReadByte() (byte, error) {
	if r.j >= len(r.scratch) {
		if r.end {
			return 0, io.EOF
		}
		// every slice len check on quit channel
		select {
		case <-r.Quit:
			return 0, io.EOF
		default:
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

// Read implements the io.Reader interface.
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

// ReadAt implements the io.ReaderAt interface.
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

// Seek implements the io.Seeker interface.
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
	success, err := r.CanSeek(offset, rev)
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
	*Buffer
}

// ReverseReaderFrom returns a ReverseReader reading from the Buffer.
func ReverseReaderFrom(b *Buffer) *ReverseReader {
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

// Read implements the io.Reader interface.
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

// ReadByte implements the io.ByteReader interface.
func (r *ReverseReader) ReadByte() (byte, error) {
	if r.j >= len(r.scratch) {
		if r.end {
			return 0, io.EOF
		}
		// every slice len check quit channel
		select {
		case <-r.Quit:
			return 0, io.EOF
		default:
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
	b := r.scratch[len(r.scratch)-r.j-1]
	r.i++
	r.j++
	return b, nil
}

// LimitReader allows you to set an early limit for the ByteReader.
// At limit, ReadByte() returns 0, io.EOF.
type LimitReader struct {
	limit int
	*Reader
}

// LimitReaderFrom returns a new LimitReader reading from Buffer.
func LimitReaderFrom(b *Buffer, l int) io.ByteReader {
	// A BOF reader may not have been used, trigger a fill if necessary.
	r := &Reader{0, 0, nil, false, b}
	if l < 0 {
		return r
	}
	return &LimitReader{l, r}
}

// ReadByte implements the io.ByteReader interface.
// Once limit is reached, returns 0, io.EOF.
func (l *LimitReader) ReadByte() (byte, error) {
	if l.i >= int64(l.limit) {
		return 0, io.EOF
	}
	return l.Reader.ReadByte()
}

// LimitReverseReader allows you to set an early limit for the ByteReader.
// At limit, REadByte() returns 0, io.EOF.
type LimitReverseReader struct {
	limit int
	*ReverseReader
}

// LimitReverseReaderFrom returns a new LimitReverseReader reading from Buffer.
func LimitReverseReaderFrom(b *Buffer, l int) io.ByteReader {
	if l < 0 {
		return &ReverseReader{0, 0, nil, false, b}
	}
	return &LimitReverseReader{l, &ReverseReader{0, 0, nil, false, b}}
}

// ReadByte implements the io.ByteReader interface.
// Once limit is reached, returns 0, io.EOF.
func (r *LimitReverseReader) ReadByte() (byte, error) {
	if r.i >= int64(r.limit) {
		return 0, io.EOF
	}
	return r.ReverseReader.ReadByte()
}

type nullReader struct{}

func (n nullReader) ReadByte() (byte, error) { return 0, io.EOF }

func (n nullReader) Read(b []byte) (int, error) { return 0, io.EOF }

type utf16Reader struct{ *Reader }

func (u *utf16Reader) ReadByte() (byte, error) {
	_, err := u.Reader.ReadByte()
	if err != nil {
		return 0, err
	}
	return u.Reader.ReadByte()
}

func utf16leReaderFrom(b *Buffer) *utf16Reader {
	r := ReaderFrom(b)
	r.ReadByte()
	return &utf16Reader{r}
}

func utf16beReaderFrom(b *Buffer) *utf16Reader {
	r := ReaderFrom(b)
	r.ReadByte()
	r.ReadByte()
	return &utf16Reader{r}
}

func TextReaderFrom(b *Buffer) io.ByteReader {
	switch b.Text() {
	case characterize.ASCII, characterize.UTF8, characterize.LATIN1, characterize.EXTENDED:
		return ReaderFrom(b)
	case characterize.UTF16BE:
		return utf16beReaderFrom(b)
	case characterize.UTF16LE:
		return utf16leReaderFrom(b)
	case characterize.UTF8BOM:
		r := ReaderFrom(b)
		for i := 0; i < 3; i++ {
			r.ReadByte()
		}
		return r
	case characterize.UTF7:
		r := ReaderFrom(b)
		for i := 0; i < 4; i++ {
			r.ReadByte()
		}
		return r
	}
	return nullReader{}
}

type reverseUTF16Reader struct {
	*ReverseReader
	first bool
}

func (u *reverseUTF16Reader) ReadByte() (byte, error) {
	if u.first {
		u.first = false
		return u.ReverseReader.ReadByte()
	}
	if _, err := u.ReverseReader.ReadByte(); err != nil {
		return 0, err
	}
	return u.ReverseReader.ReadByte()
}

func reverseUTF16leReaderFrom(b *Buffer) *reverseUTF16Reader {
	r := ReverseReaderFrom(b)
	r.ReadByte()
	return &reverseUTF16Reader{r, true}
}

func reverseUTF16beReaderFrom(b *Buffer) *reverseUTF16Reader {
	return &reverseUTF16Reader{ReverseReaderFrom(b), true}
}

func TextReverseReaderFrom(b *Buffer) io.ByteReader {
	switch b.Text() {
	case characterize.ASCII, characterize.UTF8, characterize.LATIN1, characterize.EXTENDED, characterize.UTF8BOM, characterize.UTF7:
		return ReaderFrom(b)
	case characterize.UTF16BE:
		return reverseUTF16beReaderFrom(b)
	case characterize.UTF16LE:
		return reverseUTF16leReaderFrom(b)
	}
	return nullReader{}
}
