package core

import "github.com/richardlehane/siegfried/pkg/core/siegreader"

type Siegfried struct {
	identifiers []Identifier
	buffer      *siegreader.Buffer
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

func (s *Siegfried) Identify(r io.Reader) (chan Identification, error) {
	err := s.buffer.ReadFrom(r)
	if err != nil {
		return err, nil
	}
	ret := make(chan Identification)
	for _, v := range s.identifiers {
		go v.Identify(s.buffer, ret)
	}
	return ret, nil
}
