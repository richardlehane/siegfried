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
	spool  *pool // Pool of stream Buffers
	fpool  *pool // Pool of file Buffers
	epool  *pool // Pool of external buffers
	fdatas *datas
	last   *pool
}

// New creates a new pool of stream, external and file buffers
func New() *Buffers {
	return &Buffers{
		newPool(newStream),
		newPool(newFile),
		newPool(newExternal),
		&datas{
			newPool(newBigFile),
			newPool(newSmallFile),
			newPool(newMmap),
		},
		nil,
	}
}

// Get returns a Buffer reading from the provided io.Reader.
// Get returns a Buffer backed by a stream, external or file
// source buffer depending on the type of reader.
// Source buffers are re-cycled where possible.
func (b *Buffers) Get(src io.Reader) (*Buffer, error) {
	f, ok := src.(*os.File)
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
	switch i.bufferSrc.(type) {
	default:
		panic("Siegreader: unknown buffer type")
	case *stream:
		b.spool.put(i.bufferSrc)
		b.last = b.spool
	case *file:
		b.fdatas.put(i.bufferSrc.(*file).data)
		b.fpool.put(i.bufferSrc)
		b.last = b.fpool
	case *external:
		b.epool.put(i.bufferSrc)
		b.last = b.epool
	}
}

// Last retrieves the last Buffer returned to the pool with Put.
func (b *Buffers) Last() *Buffer {
	if b.last == nil {
		return nil
	}
	last := &Buffer{bufferSrc: b.last.get().(bufferSrc)}
	if str, ok := last.bufferSrc.(*stream); ok {
		str.b = last
	}
	return last
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
		d.mpool.put(m)
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
	switch i.(type) {
	default:
		panic("Siegreader: unknown data type")
	case *bigfile:
		d.bfpool.put(i)
	case *smallfile:
		d.sfpool.put(i)
	case *mmap:
		//i.(*mmap).reset() test resetting on setsource rather than put
		d.mpool.put(i)
	}
	return
}
