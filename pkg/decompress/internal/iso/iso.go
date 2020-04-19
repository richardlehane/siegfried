package iso

import (
	"fmt"
	"io"
)

type Reader interface {
	// TODO: Implement reader.
}

type ISOReader struct {
	// TODO: ISO Reader.
}

func NewISOReader(reader io.Reader) (*ISOReader, error) {
	isoReader := &ISOReader{}
	return isoReader, fmt.Errorf("Decompress: Implementing ISO extraction")
}
