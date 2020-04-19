package decompress

import (
	"fmt"
	"io"
	"time"

	"github.com/richardlehane/siegfried/pkg/decompress/internal/iso"
)

type isoDecompressor struct {
	p   string
	rdr *iso.Reader
}

func newISO(reader io.Reader, path string) (Decompressor, error) {
	_, err := iso.NewISOReader(reader)
	return &isoDecompressor{}, err
}

func (iso *isoDecompressor) close() {
	// TODO: not implemented...
}

func (iso *isoDecompressor) Next() error {
	return fmt.Errorf("Decompress (ISO) Next() isn't implemented")
}

func (iso *isoDecompressor) Reader() io.Reader {
	return nil
}

func (iso *isoDecompressor) Path() string {
	return ""
}

func (iso *isoDecompressor) MIME() string {
	return ""
}

func (iso *isoDecompressor) Size() int64 {
	return 0
}

func (iso *isoDecompressor) Mod() time.Time {
	return time.Now()
}

func (iso *isoDecompressor) Dirs() []string {
	return []string{""}
}
