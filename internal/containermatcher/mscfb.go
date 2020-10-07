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
	"strings"

	"github.com/richardlehane/mscfb"
	"github.com/richardlehane/siegfried/internal/siegreader"
)

type mscfbReader struct {
	rdr   *mscfb.Reader
	entry *mscfb.File
}

func mscfbRdr(b *siegreader.Buffer) (Reader, error) {
	m, err := mscfb.New(siegreader.ReaderFrom(b))
	if err != nil {
		return nil, err
	}
	return &mscfbReader{rdr: m}, nil
}

func (m *mscfbReader) Next() error {
	var err error
	m.entry, err = m.rdr.Next()
	return err
}

func (m *mscfbReader) Name() string {
	if m.entry == nil {
		return ""
	}
	return strings.Join(append(m.entry.Path, m.entry.Name), "/")
}

func (m *mscfbReader) SetSource(b *siegreader.Buffers) (*siegreader.Buffer, error) {
	return b.Get(m.entry)
}

func (m *mscfbReader) Close() {}

func (m *mscfbReader) IsDir() bool {
	if m.entry == nil {
		return false
	}
	return m.entry.FileInfo().IsDir()
}
