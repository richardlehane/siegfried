package containermatcher

import (
	"archive/zip"
	"io"

	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type zipReader struct {
	idx int
	rdr *zip.Reader
	rc  io.ReadCloser
}

func (z *zipReader) Next() error {
	// scan past directories
	for ; z.idx < len(z.rdr.File) && z.rdr.File[z.idx].CompressedSize64 <= 0; z.idx++ {
	}
	if z.idx >= len(z.rdr.File) {
		return io.EOF
	}
	return nil
}
func (z *zipReader) Name() string {
	return z.rdr.File[z.idx].Name
}
func (z *zipReader) SetSource(b *siegreader.Buffer) error {
	var err error
	z.rc, err = z.rdr.File[z.idx].Open()
	if err != nil {
		return err
	}
	return b.SetSource(z.rc)
}
func (z *zipReader) Close() {
	if z.rc == nil {
		return
	}
	z.rc.Close()
}

func newZip(b *siegreader.Buffer) (Reader, error) {
	r, err := zip.NewReader(b.NewReader(), b.SizeNow())
	return &zipReader{rdr: r}, err
}
