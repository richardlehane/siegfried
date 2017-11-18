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

// Package siegreader implements multiple independent Readers (and ReverseReaders) from a single Buffer.
//
// Example:
//   buffers := siegreader.New()
//   buffer, err := buffers.Get(io.Reader)
//   if err != nil {
//     log.Fatal(err)
//   }
//   rdr := siegreader.ReaderFrom(buffer)
//	 second_rdr := siegreader.ReaderFrom(buffer)
//   brdr := siegreader.LimitReaderFrom(buffer, -1)
//   rrdr, err := siegreader.LimitReverseReaderFrom(buffer, 16000)
//   i, err := rdr.Read(slc)
//   i2, err := second_rdr.Read(slc2)
//   i3, err := rrdr.ReadByte()
package siegreader

import (
	"errors"
	"io"

	"github.com/richardlehane/characterize"
)

var (
	ErrEmpty     = errors.New("empty source")
	ErrQuit      = errors.New("siegreader: quit chan closed while awaiting EOF")
	ErrNilBuffer = errors.New("siegreader: attempt to SetSource on a nil buffer")
)

const (
	readSz      int = 4096
	initialRead     = readSz * 2
	eofSz           = readSz * 2
	wheelSz         = readSz * 16
	smallFileSz     = readSz * 16 // 65536
	streamSz        = smallFileSz * 1024
)

type bufferSrc interface {
	Slice(off int64, l int) ([]byte, error)
	EofSlice(off int64, l int) ([]byte, error)
	Size() int64
	SizeNow() int64
	CanSeek(off int64, rev bool) (bool, error)
}

// Buffer allows multiple readers to read from the same source.
// Readers include reverse (from EOF) and limit readers.
type Buffer struct {
	Quit   chan struct{} // when this channel is closed, readers will return io.EOF
	texted bool
	text   characterize.CharType
	bufferSrc
}

// Bytes returns a byte slice for a full read of the buffered file or stream.
// Returns nil on error
func (b *Buffer) Bytes() []byte {
	sz := b.SizeNow()
	// check for int overflow
	if int64(int(sz)) != sz {
		return nil
	}
	s, err := b.Slice(0, int(sz))
	if err != nil {
		return nil
	}
	return s
}

// Text returns the CharType of the first 4096 bytes of the Buffer.
func (b *Buffer) Text() characterize.CharType {
	if b.texted {
		return b.text
	}
	b.texted = true
	buf, err := b.Slice(0, readSz)
	if err == nil || err == io.EOF {
		b.text = characterize.Detect(buf)
	}
	return b.text
}
