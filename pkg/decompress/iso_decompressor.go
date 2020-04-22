package decompress

/* ISO Decompresser satisfies the Decompress interface:

	type Decompressor interface {
		Next() error // when finished, should return io.EOF
		Reader() io.Reader
		Path() string
		MIME() string
		Size() int64
		Mod() time.Time
		Dirs() []string
	}

The type (isoDecompressor) is used as a convenience to maintain state as the
Decompresser is called from outside of the package. This gives us a lot of
freedom here.
*/

import (
	"io"
	"strings"
	"time"

	"github.com/richardlehane/siegfried/pkg/decompress/internal/iso"
)

type isoDecompressor struct {
	path      string
	name      string
	isoReader *iso.ISOReader
	reader    io.Reader
	size      int64
	mod       time.Time
}

func newISO(reader io.Reader, path string) (Decompressor, error) {
	// TODO: We're not using the reader, the iso9660 library requires a
	// readseeker. We can probably rectify this when we create our own
	// reader/s.
	isoReader, err := iso.NewISOReader(path)
	return &isoDecompressor{
		path:      path,
		isoReader: isoReader,
	}, err
}

func (iso *isoDecompressor) close() {
	// Reset the reader to ensure state is not left behind. TODO: Decide if
	// this is needed/necessary.
	if iso.reader != nil {
		iso.reader = nil
		iso.name = ""
		iso.size = 0
		iso.mod = time.Time{}
	}
}

func (iso *isoDecompressor) Next() error {
	iso.close()
	for {
		file, err := iso.isoReader.Next()
		if err != nil {
			if err == io.EOF {
				return err
			}
		}
		if file.IsDir() {
			continue
		}
		iso.name = strings.TrimLeft(strings.Trim(file.Name(), "."), "/")
		iso.size = file.Size()
		iso.mod = file.ModTime()
		iso.reader = file.Sys().(io.Reader)
		return nil
	}
	return nil
}

// Reader returns an io.Reader for Siegfried to perform in-memory analysis.
func (iso *isoDecompressor) Reader() io.Reader {
	return iso.reader
}

// Path returns the path within the ISO9660 object.
func (iso *isoDecompressor) Path() string {
	return Arcpath(iso.path, iso.name)
}

// Mime is a field available for ARC/WARC type records only at this point.
func (iso *isoDecompressor) MIME() string {
	return ""
}

// Size returns the size of the file.
func (iso *isoDecompressor) Size() int64 {
	return iso.size
}

// Mod returns the modified time from the file.
func (iso *isoDecompressor) Mod() time.Time {
	return iso.mod
}

// Dirs ...
func (iso *isoDecompressor) Dirs() []string {
	return []string{""}
}
