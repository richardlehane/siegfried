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

package containermatcher

import (
	"archive/zip"
	"io"
	"strings"

	"github.com/richardlehane/siegfried/internal/siegreader"
)

type zipReader struct {
	idx int
	rdr *zip.Reader
	rc  io.ReadCloser
}

func (z *zipReader) Next() error {
	z.idx++
	if z.idx >= len(z.rdr.File) {
		return io.EOF
	}
	return nil
}

func (z *zipReader) Name() string {
	return strings.TrimSuffix(z.rdr.File[z.idx].Name, "\x00") // non-spec zip files may have null terminated strings
}

func (z *zipReader) SetSource(bufs *siegreader.Buffers) (*siegreader.Buffer, error) {
	var err error
	z.rc, err = z.rdr.File[z.idx].Open()
	if err != nil {
		return nil, err
	}
	return bufs.Get(z.rc)
}

func (z *zipReader) Close() {
	if z.rc == nil {
		return
	}
	z.rc.Close()
}

func (z *zipReader) IsDir() bool {
	if z.idx < len(z.rdr.File) {
		return z.rdr.File[z.idx].FileHeader.FileInfo().IsDir()
	}
	return false
}

func zipRdr(b *siegreader.Buffer) (Reader, error) {
	r, err := zip.NewReader(siegreader.ReaderFrom(b), b.SizeNow())
	return &zipReader{idx: -1, rdr: r}, err
}
