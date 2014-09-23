// Package core defines the Siegfried struct and Identifier/Identification interfaces.
// The packages within core (bytematcher and namematcher) provide a toolkit for building identifiers based on different signature formats (such as PRONOM).
package core

import (
	"io"
	"strings"
	"sync"

	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type Siegfried struct {
	identifiers []Identifier // at present only one identifier (the PRONOM identifier) is used, but can add other identifiers e.g. for FILE sigs
	buffer      *siegreader.Buffer
	name        string
}

// Identifiers can be defined for different signature formats. E.g. there is a PRONOM identifier that implements the TNA's format.
type Identifier interface {
	Identify(*siegreader.Buffer, string, chan Identification, *sync.WaitGroup)
	Version() string
	Details() string // long test to describe the identifier (e.g. name, date created etc.)
	Update(i int) bool
}

// Identifications are sent by identifiers when a format matches
type Identification interface {
	String() string      // short text that should be displayed to indicate the format match
	Details() string     // long text that should be displayed to indicate the format match
	Confidence() float64 // how certain is this identification?
}

func NewSiegfried() *Siegfried {
	s := new(Siegfried)
	s.identifiers = make([]Identifier, 0, 1)
	s.buffer = siegreader.New()
	return s
}

func (s *Siegfried) AddIdentifier(i Identifier) {
	s.identifiers = append(s.identifiers, i)
}

// Identify applies the set of identifiers to a reader and file name. If the file name is not known, use an empty string instead.
func (s *Siegfried) Identify(r io.Reader, n string) (chan Identification, error) {
	err := s.buffer.SetSource(r)
	if err != nil && err != io.EOF {
		return nil, err
	}
	s.name = n
	ret := make(chan Identification)
	go s.identify(ret)
	return ret, nil
}

func (s *Siegfried) identify(ret chan Identification) {
	var wg sync.WaitGroup
	for _, v := range s.identifiers {
		wg.Add(1)
		go v.Identify(s.buffer, s.name, ret, &wg)
	}
	wg.Wait()
	close(ret)
}

func (s *Siegfried) String() string {
	ids := make([]string, len(s.identifiers))
	for i, v := range s.identifiers {
		ids[i] = v.Details()
	}
	return strings.Join(ids, "\n")
}

type Result struct {
	Index int
	Basis string
}
