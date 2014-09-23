package containermatcher

import "github.com/richardlehane/siegfried/core/siegreader"

type Reader interface {
	Next() error  // when finished, should return io.EOF
	Name() string // return name of the object with paths concatenated with / character
	SetSource(*siegreader.Buffer) error
	Quit() // close any unclosed files
}
