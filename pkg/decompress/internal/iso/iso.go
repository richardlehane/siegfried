package iso

/* The ISO package enables the reading of ISO9660 packages in Siegfried.

The initial implementation wraps the hooklift/iso9660 library but we can begin
to implement our own based on what the Siegfried package asks us to deliver.
*/

import (
	"os"

	"github.com/hooklift/iso9660"
)

// ISOReader is an alias for the iso9660.Reader to provide better compatibility
// moving forward.
type ISOReader = iso9660.Reader

// NewISOReader returns an ISO reader.
func NewISOReader(path string) (*ISOReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	isoReader, err := iso9660.NewReader(file)
	if err != nil {
		return nil, err
	}
	return isoReader, nil
}
