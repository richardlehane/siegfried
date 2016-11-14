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

package main

import (
	"io"
	"path/filepath"
	"strings"
	"time"

	"archive/tar"
	"archive/zip"
	"compress/gzip"

	"github.com/richardlehane/characterize"
	"github.com/richardlehane/webarchive"

	"github.com/richardlehane/siegfried/internal/siegreader"
)

type decompressor interface {
	next() error // when finished, should return io.EOF
	reader() io.Reader
	path() string
	mime() string
	size() int64
	mod() string
	dirs() []string
}

type zipD struct {
	idx     int
	p       string
	rdr     *zip.Reader
	rc      io.ReadCloser
	written map[string]bool
}

func newZip(ra io.ReaderAt, path string, sz int64) (decompressor, error) {
	zr, err := zip.NewReader(ra, sz)
	return &zipD{idx: -1, p: path, rdr: zr}, err
}

func (z *zipD) close() {
	if z.rc == nil {
		return
	}
	z.rc.Close()
}

func (z *zipD) next() error {
	z.close() // close the previous entry, if any
	// proceed
	z.idx++
	// scan past directories
	for ; z.idx < len(z.rdr.File) && z.rdr.File[z.idx].FileInfo().IsDir(); z.idx++ {
	}
	if z.idx >= len(z.rdr.File) {
		return io.EOF
	}
	var err error
	z.rc, err = z.rdr.File[z.idx].Open()
	return err
}

func (z *zipD) reader() io.Reader {
	return z.rc
}

func (z *zipD) path() string {
	return arcpath(z.p, filepath.FromSlash(characterize.ZipName(z.rdr.File[z.idx].Name)))
}

func (z *zipD) mime() string {
	return ""
}

func (z *zipD) size() int64 {
	return int64(z.rdr.File[z.idx].UncompressedSize64)
}

func (z *zipD) mod() string {
	return z.rdr.File[z.idx].ModTime().Format(time.RFC3339)
}

func (z *zipD) dirs() []string {
	if z.written == nil {
		z.written = make(map[string]bool)
	}
	return dirs(z.p, characterize.ZipName(z.rdr.File[z.idx].Name), z.written)
}

type tarD struct {
	p       string
	hdr     *tar.Header
	rdr     *tar.Reader
	written map[string]bool
}

func newTar(r io.Reader, path string) (decompressor, error) {
	return &tarD{p: path, rdr: tar.NewReader(r)}, nil
}

func (t *tarD) next() error {
	var err error
	// scan past directories
	for t.hdr, err = t.rdr.Next(); err == nil && t.hdr.FileInfo().IsDir(); t.hdr, err = t.rdr.Next() {
	}
	return err
}

func (t *tarD) reader() io.Reader {
	return t.rdr
}

func (t *tarD) path() string {
	return arcpath(t.p, filepath.FromSlash(t.hdr.Name))
}

func (t *tarD) mime() string {
	return ""
}

func (t *tarD) size() int64 {
	return t.hdr.Size
}

func (t *tarD) mod() string {
	return t.hdr.ModTime.Format(time.RFC3339)
}

func (t *tarD) dirs() []string {
	if t.written == nil {
		t.written = make(map[string]bool)
	}
	return dirs(t.p, t.hdr.Name, t.written)
}

type gzipD struct {
	sz   int64
	p    string
	read bool
	rdr  *gzip.Reader
}

func newGzip(b *siegreader.Buffer, path string) (decompressor, error) {
	b.Quit = make(chan struct{}) // in case a stream with a closed quit channel, make a new one
	_ = b.SizeNow()              // in case a stream, force full read
	buf, err := b.EofSlice(0, 4) // gzip stores uncompressed size in last 4 bytes of the stream
	if err != nil {
		return nil, err
	}
	sz := int64(uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24)
	g, err := gzip.NewReader(siegreader.ReaderFrom(b))
	return &gzipD{sz: sz, p: path, rdr: g}, err
}

func (g *gzipD) next() error {
	if g.read {
		g.rdr.Close()
		return io.EOF
	}
	g.read = true
	return nil
}

func (g *gzipD) reader() io.Reader {
	return g.rdr
}

func (g *gzipD) path() string {
	name := g.rdr.Name
	if len(name) == 0 {
		switch filepath.Ext(g.p) {
		case ".gz", ".z", ".gzip", ".zip":
			name = strings.TrimSuffix(filepath.Base(g.p), filepath.Ext(g.p))
		default:
			name = filepath.Base(g.p)
		}
	}
	return arcpath(g.p, name)
}

func (g *gzipD) mime() string {
	return ""
}

func (g *gzipD) size() int64 {
	return g.sz
}

func (g *gzipD) mod() string {
	return g.rdr.ModTime.Format(time.RFC3339)
}

func (t *gzipD) dirs() []string {
	return nil
}

func trimWebPath(p string) string {
	d, f := filepath.Split(p)
	clean := strings.TrimSuffix(d, string(filepath.Separator))
	_, f1 := filepath.Split(clean)
	if f == strings.TrimSuffix(f1, filepath.Ext(clean)) {
		return clean
	}
	return p
}

type wa struct {
	p   string
	rec webarchive.Record
	rdr webarchive.Reader
}

func newARC(r io.Reader, path string) (decompressor, error) {
	arcReader, err := webarchive.NewARCReader(r)
	return &wa{p: trimWebPath(path), rdr: arcReader}, err
}

func newWARC(r io.Reader, path string) (decompressor, error) {
	warcReader, err := webarchive.NewWARCReader(r)
	return &wa{p: trimWebPath(path), rdr: warcReader}, err
}

func (w *wa) next() error {
	var err error
	w.rec, err = w.rdr.NextPayload()
	return err
}

func (w *wa) reader() io.Reader {
	return webarchive.DecodePayload(w.rec)
}

func (w *wa) path() string {
	return arcpath(w.p, w.rec.Date().Format(webarchive.ARCTime)+"/"+w.rec.URL())
}

func (w *wa) mime() string {
	return w.rec.MIME()
}

func (w *wa) size() int64 {
	return w.rec.Size()
}

func (w *wa) mod() string {
	return w.rec.Date().Format(time.RFC3339)
}

func (w *wa) dirs() []string {
	return nil
}

func dirs(path, name string, written map[string]bool) []string {
	ds := strings.Split(filepath.ToSlash(name), "/")
	if len(ds) > 1 {
		var ret []string
		for _, p := range ds[:len(ds)-1] {
			path = path + string(filepath.Separator) + p
			if !written[path] {
				ret = append(ret, path)
				written[path] = true
			}
		}
		return ret
	}
	return nil
}

// per https://github.com/richardlehane/siegfried/issues/81
// construct paths for compressed objects acc. to KDE hash notation
func arcpath(base, path string) string {
	if *droido {
		return base + string(filepath.Separator) + path
	}
	return base + "#" + path
}
