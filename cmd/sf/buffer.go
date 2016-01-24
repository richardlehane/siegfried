// Copyright 2016 Richard Lehane. All rights reserved.
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

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/richardlehane/siegfried"
)

type buffer struct {
	*bytes.Buffer
	buf []byte
	sz  int
}

var fileBuffer *buffer

const retries = 3

func identifyBuffer(w writer, s *siegfried.Siegfried, path string, sz int64, mod string) {
	fileBuffer.Reset()
	fileBuffer.sz = int(sz)
	if fileBuffer.sz < 0 {
		writeError(w, path, sz, mod, fmt.Errorf("file too large to buffer"))
		return
	}
	if fileBuffer.Cap() < fileBuffer.sz {
		fileBuffer.Buffer = bytes.NewBuffer(make([]byte, 0, fileBuffer.sz))
	}
	// recover from bytes.ErrTooLarge panic
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			writeError(w, path, sz, mod, fmt.Errorf("file too large to buffer"))
			return
		} else {
			panic(e)
		}
	}()
	for i := 0; i < retries; i++ {
		if i > 0 && throttle != nil {
			<-throttle.C
		}
		f, err := os.Open(path)
		if err != nil {
			f, err = retryOpen(path, err)
			if err != nil {
				continue
			}
		}
		n, _ := fileBuffer.ReadFrom(f)
		f.Close()
		if n == sz {
			break
		}
		fileBuffer.Reset() // try again
	}
	fileBuffer.buf = fileBuffer.Bytes()
	if len(fileBuffer.buf) != fileBuffer.sz {
		writeError(w, path, sz, mod, fmt.Errorf("failed buffering; expected to read %d, got %d", fileBuffer.sz, len(fileBuffer.buf)))
		return
	}
	identifyRdr(w, s, fileBuffer, sz, path, "", mod)
}

func (b *buffer) IsSlicer() bool { return true }

func (b *buffer) Slice(off int64, l int) ([]byte, error) {
	o := int(off)
	if o < 0 || o >= len(b.buf) || l < 1 {
		return nil, io.EOF
	}
	if o+l > len(b.buf) {
		return b.buf[o:], io.EOF
	}
	return b.buf[o : o+l], nil
}

func (b *buffer) EofSlice(off int64, l int) ([]byte, error) {
	o := int(off)
	if o < 0 || o >= len(b.buf) || l < 1 {
		return nil, io.EOF
	}
	if o+l > len(b.buf) {
		return b.buf[:len(b.buf)-o], io.EOF
	}
	return b.buf[len(b.buf)-o-l : len(b.buf)-o], nil
}

func (b *buffer) Size() int64 { return int64(b.sz) }
