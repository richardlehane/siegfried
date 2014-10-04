package containermatcher

import "github.com/richardlehane/siegfried/pkg/core/siegreader"

type Reader interface {
	Next() error  // when finished, should return io.EOF
	Name() string // return name of the object with paths concatenated with / character
	SetSource(*siegreader.Buffer) error
	Close() // close files
	Quit()  // close down reader
}
