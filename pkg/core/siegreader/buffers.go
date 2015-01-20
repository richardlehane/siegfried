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
)

func (b *Buffers) Get(src io.Reader) (Buffer, error) {
	f, ok := src.(*os.File)
	if !ok {
		stream := b.spool.Get().(*stream)
		stream.setSource(src)
		return stream, nil
	}
	buf := b.fpool.Get().(*file)
	err := buf.setSource(f, b.fdatas)
	return buf, err
}

func (b *Buffers) Put(i Buffer) {
	switch i.(type) {
	default:
		panic("Siegreader: unknown buffer type")
	case *stream:
		b.spool.Put(i)
	case *file:
		b.fdatas.put(i.(*file).data)
		b.fpool.Put(i)
	}
}

// Data pool (used by file)
func (d *datas) get(f *file) data {
	if mmapable(f.sz) {
		m := d.mpool.Get().(*mmap)
		m.setSource(f)
		return m
	} else if f.sz <= int64(smallFileSz) {
		sf := d.sfpool.Get().(*smallfile)
		sf.setSource(f)
		return sf
	}
	bf := d.bfpool.Get().(*bigfile)
	bf.setSource(f)
	return bf
}

func (d *datas) put(i data) {
	if i == nil {
		return
	}
	switch i.(type) {
	default:
		panic("Siegreader: unknown data type")
	case *bigfile:
		d.bfpool.Put(i)
	case *smallfile:
		d.sfpool.Put(i)
	case *mmap:
		i.(*mmap).reset()
		d.mpool.Put(i)
	}
	return
}
