package core

import (
	"io"
	"sync"

	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type Siegfried struct {
	identifiers []Identifier
	buffer      *siegreader.Buffer
	name        string
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

func (s *Siegfried) Identify(r io.Reader, n string) (chan Identification, error) {
	err := s.buffer.SetSource(r)
	if err != nil {
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
