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

// Buffers is a combined pool of stream, external and file buffers
type Buffers struct {
	spool *pool // Pool of stream Buffers
	fpool *pool // Pool of file Buffers
	epool *pool // Pool of external buffers

	fdatas *datas // file datas
}

// New creates a new pool of stream, external and file buffers
func New() *Buffers {
	return &Buffers{
		spool: newPool(newStream),
		fpool: newPool(newFile),
		epool: newPool(newExternal),
		fdatas: &datas{
			newPool(newBigFile),
			newPool(newSmallFile),
			newPool(newMmap),
		},
	}
}

// Get returns a Buffer reading from the provided io.Reader.
// Get returns a Buffer backed by a stream, external or file
// source buffer depending on the type of reader.
// Source buffers are re-cycled where possible.
func (b *Buffers) Get(src io.Reader) (*Buffer, error) {
	f, ok := src.(*os.File)
	if ok {
		stat, err := f.Stat()
		if err != nil || stat.Mode()&os.ModeType != 0 {
			ok = false
		}
	}
	if !ok {
		e, ok := src.(source)
		if !ok || !e.IsSlicer() {
			stream := b.spool.get().(*stream)
			buf := &Buffer{}
			err := stream.setSource(src, buf)
			buf.bufferSrc = stream
			return buf, err
		}
		ext := b.epool.get().(*external)
		err := ext.setSource(e)
		return &Buffer{bufferSrc: ext}, err
	}
	fbuf := b.fpool.get().(*file)
	err := fbuf.setSource(f, b.fdatas)
	return &Buffer{bufferSrc: fbuf}, err
}

// Put returns a Buffer to the pool for re-cycling.
func (b *Buffers) Put(i *Buffer) {
	switch v := i.bufferSrc.(type) {
	default:
		panic("Siegreader: unknown buffer type")
	case *stream:
		v.cleanUp()
		b.spool.put(v)
	case *file:
		b.fdatas.put(v.data)
		b.fpool.put(v)
	case *external:
		b.epool.put(v)
	}
}

// data pool (used by file)
// pool of big files, small files, and mmap files
type datas struct {
	bfpool *pool
	sfpool *pool
	mpool  *pool
}

func (d *datas) get(f *file) data {
	if mmapable(f.sz) {
		m := d.mpool.get().(*mmap)
		if err := m.setSource(f); err == nil {
			return m
		}
		d.mpool.put(m) // replace on error and get big file instead
	}
	if f.sz <= int64(smallFileSz) {
		sf := d.sfpool.get().(*smallfile)
		sf.setSource(f)
		return sf
	}
	bf := d.bfpool.get().(*bigfile)
	bf.setSource(f)
	return bf
}

func (d *datas) put(i data) {
	if i == nil {
		return
	}
	switch v := i.(type) {
	default:
		panic("Siegreader: unknown data type")
	case *bigfile:
		d.bfpool.put(v)
	case *smallfile:
		d.sfpool.put(v)
	case *mmap:
		v.reset()
		d.mpool.put(v)
	}
	return
}
