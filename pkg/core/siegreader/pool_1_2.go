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

// +build !go1.3

package siegreader

import (
	"io"
)

type buffers struct {
	fpool chan *file   // Pool of file buffers
	spool chan *stream // Pool of stream buffers
	*datas
}

func (b *buffers) get(src io.Reader) Buffer {
	return Buffer{}
}

func (b *buffers) put(i buffer) Buffer {
	return Buffer{}
}

// Data pool (used by file)

type mmap struct{}
type buffer struct{}

type datas struct {
	bfpool chan *bigfile
	mpool  chan *mmap
}

func (d *datas) get(sz int64) data {
	return &bigfile{}
}

func (d *datas) put(i data) {

}
