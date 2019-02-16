// Copyright 2015 Richard Lehane. All rights reserved.
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

package webarchive

import (
	"errors"
	"io"
	"time"
)

var (
	ErrReset         = errors.New("webarchive: attempted reset on nil MultiReader, use NewReader() first")
	ErrNotWebarchive = errors.New("webarchive: not a valid ARC or WARC file")
	ErrVersionBlock  = errors.New("webarchive: invalid ARC version block")
	ErrARCHeader     = errors.New("webarchive: invalid ARC header")
	ErrNotSlicer     = errors.New("webarchive: underlying reader must be a slicer to expose Slice and EOFSlice methods")
	ErrWARCHeader    = errors.New("webarchive: invalid WARC header")
	ErrWARCRecord    = errors.New("webarchive: error parsing WARC record")
	ErrDiscard       = errors.New("webarchive: failed to do full read during discard")
)

// Record represents both ARC and WARC records.
type Record interface {
	Header
	Content
}

// Header represents the common header fields shared by ARC and WARC records.
type Header interface {
	URL() string
	Date() time.Time
	MIME() string
	Fields() map[string][]string
	// private methods - used by DecodePayload
	transferEncodings() []string
	encodings() []string
}

// Content represents the content portion of a WARC or ARC record.
type Content interface {
	Size() int64
	Read(p []byte) (n int, err error)
	Slice(off int64, l int) ([]byte, error)
	EofSlice(off int64, l int) ([]byte, error)
	// private method -used by DecodePayload
	peek(i int) ([]byte, error)
}

// Reader represents the common methods shared by ARC, WARC and Multi readers.
type Reader interface {
	Reset(io.Reader) error
	Next() (Record, error)
	NextPayload() (Record, error) // skip non-resonse/resource records; merge continuations; strip non-body content from record
	Close() error
}

// MultiReader is the concrete type returned by webarchive.NewReader.
// A MultiReader can represent both a WARC or ARC reader (or both if ARC and WARC files are given to the same Reader using Reset).
//
// Example:
//  f, _ := os.Open("examples/IAH-20080430204825-00000-blackbook.arc")
//  rdr, _ := NewReader(f)
//  f.Close()
//  f, _ = os.Open("examples/IAH-20080430204825-00000-blackbook.warc.gz")
//  rdr.Reset(f)
//  rdr.Close()
//  f.Close()
type MultiReader struct {
	r *reader
	a *ARCReader
	w *WARCReader
	Reader
}

// Reset allows re-use of a Multireader.
// A Multireader created with a WARC file can be reset with an ARC file, and vice versa.
func (m *MultiReader) Reset(r io.Reader) error {
	if m == nil {
		return ErrReset
	}
	err := m.r.reset(r)
	if err != nil {
		return err
	}
	if m.w == nil {
		m.w, err = newWARCReader(m.r)
	} else {
		err = m.w.reset()
	}
	if err == nil {
		m.Reader = m.w
		return nil
	}
	if m.a == nil {
		m.a, err = newARCReader(m.r)
	} else {
		err = m.a.reset()
	}
	if err == nil {
		m.Reader = m.a
		return nil
	}
	return ErrNotWebarchive
}

// NewReader returns a new webarchive Reader reading from the io.Reader.
// The supplied io.Reader can be a WARC, ARC, WARC.GZ or ARC.GZ file.
func NewReader(r io.Reader) (Reader, error) {
	rdr, err := newReader(r)
	if err != nil {
		return nil, err
	}
	w, err := newWARCReader(rdr)
	if err != nil {
		a, err := newARCReader(rdr)
		if err != nil {
			return nil, ErrNotWebarchive
		}
		return &MultiReader{r: rdr, a: a, Reader: a}, nil
	}
	return &MultiReader{r: rdr, w: w, Reader: w}, nil
}
