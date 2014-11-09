// Copyright 2014 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package siegfried describes the layout of the Siegfried signature file.
// This signature file contains the siegfried object that performs identification
package siegfried

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/core/containermatcher"
	"github.com/richardlehane/siegfried/pkg/core/extensionmatcher"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

type Siegfried struct {
	V  Version
	em core.Matcher // extensionmatcher
	cm core.Matcher // containermatcher
	bm core.Matcher // bytematcher
	// mutatable fields follow
	ids    []core.Identifier // at present only one identifier (the PRONOM identifier) is used, but can add other identifiers e.g. for FILE sigs
	buffer *siegreader.Buffer
}

func New() *Siegfried {
	s := &Siegfried{}
	s.V = Version{config.Siegfried.SignatureVersion, time.Now()}
	s.em = extensionmatcher.New()
	s.cm = containermatcher.New()
	s.bm = bytematcher.New()
	s.buffer = siegreader.New()
	return s
}

func (s *Siegfried) Add(t identifierType) error {
	switch t {
	case Pronom:
		p, err := pronom.New()
		if err != nil {
			return err
		}
		if err := p.Add(core.ExtensionMatcher, s.em); err != nil {
			return err
		}
		if err := p.Add(core.ContainerMatcher, s.cm); err != nil {
			return err
		}
		if err := p.Add(core.ByteMatcher, s.bm); err != nil {
			return err
		}
		s.ids = append(s.ids, p)
	}
	return nil
}

func (s *Siegfried) Yaml() string {
	str := s.V.Yaml()
	for _, id := range s.ids {
		str += id.Yaml()
	}
	return str
}

func (s *Siegfried) Update(i int) bool {
	return i > s.V.Version
}

// Version info about the signature file
type Version struct {
	Version int
	Created time.Time
}

func (sv Version) Yaml() string {
	return fmt.Sprintf("---\nsiegfried   : %d.%d.%d\nscanDate    : %v\nsignatures  : %s\ncreated     : %v\nidentifiers : \n",
		config.Siegfried.Version[0], config.Siegfried.Version[1], config.Siegfried.Version[2],
		time.Now(),
		config.Siegfried.Signature,
		sv.Created)
}

type Header struct {
	SSize int                // sigversion
	BSize int                // bytematcher
	CSize int                // container
	ESize int                // extension matcher
	Ids   []IdentifierHeader // size and types of identifiers
}

type IdentifierHeader struct {
	Typ identifierType
	Sz  int
}

type identifierType int

// Register additional identifier types here
const (
	Pronom identifierType = iota
)

func identifierSz(ids []IdentifierHeader) int {
	var sz int
	for _, v := range ids {
		sz += v.Sz
	}
	return sz
}

func (s *Siegfried) Save(path string) error {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(s)
	if err != nil {
		return err
	}
	ssz := buf.Len()
	bsz, err := s.bm.Save(buf)
	if err != nil {
		return err
	}
	csz, err := s.cm.Save(buf)
	if err != nil {
		return err
	}
	esz, err := s.em.Save(buf)
	if err != nil {
		return err
	}
	ids := make([]IdentifierHeader, len(s.ids))
	for i, v := range s.ids {
		sz, err := v.Save(buf)
		if err != nil {
			return err
		}
		// add any additional identifiers to this type switch
		switch t := v.(type) {
		default:
			return fmt.Errorf("Siegfried: unexpected type for an identifier %T", t)
		case *pronom.Identifier:
			ids[i].Typ = Pronom
		}
		ids[i].Sz = sz
	}
	hbuf := new(bytes.Buffer)
	henc := gob.NewEncoder(hbuf)
	err = henc.Encode(Header{ssz, bsz, csz, esz, ids})
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.Write(hbuf.Bytes())
	if err != nil {
		return err
	}
	_, err = f.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func Load(path string) (*Siegfried, error) {
	c, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(c)
	dec := gob.NewDecoder(buf)
	var h Header
	err = dec.Decode(&h)
	if err != nil {
		return nil, err
	}
	iSize := identifierSz(h.Ids)
	sstart := len(c) - h.SSize - h.BSize - h.CSize - h.ESize - iSize
	bstart := len(c) - h.ESize - h.CSize - h.BSize - iSize
	cstart := len(c) - h.ESize - h.CSize - iSize
	estart := len(c) - h.ESize - iSize
	istart := len(c) - iSize
	sbuf := bytes.NewBuffer(c[sstart : sstart+h.SSize])
	bbuf := bytes.NewBuffer(c[bstart : bstart+h.BSize])
	cbuf := bytes.NewBuffer(c[cstart : cstart+h.CSize])
	ebuf := bytes.NewBuffer(c[estart : estart+h.ESize])
	sdec := gob.NewDecoder(sbuf)
	var s Siegfried
	err = sdec.Decode(&s)
	if err != nil {
		return nil, err
	}
	bm, err := bytematcher.Load(bbuf)
	if err != nil {
		return nil, err
	}
	cm, err := containermatcher.Load(cbuf)
	if err != nil {
		return nil, err
	}
	em, err := extensionmatcher.Load(ebuf)
	if err != nil {
		return nil, err
	}
	s.bm = bm
	s.cm = cm
	s.em = em
	s.ids = make([]core.Identifier, len(h.Ids))
	for i, v := range h.Ids {
		ibuf := bytes.NewBuffer(c[istart : istart+v.Sz])
		var id core.Identifier
		var err error
		switch v.Typ {
		default:
			return nil, fmt.Errorf("Siegfried: loading, unknown identifier type %d", v.Typ)
		case Pronom:
			id, err = pronom.Load(ibuf)
			if err != nil {
				return nil, err
			}
		}
		s.ids[i] = id
		istart += v.Sz
	}
	s.buffer = siegreader.New()
	return &s, nil
}

func (s *Siegfried) Identify(r io.Reader, n string) (chan core.Identification, error) {
	err := s.buffer.SetSource(r)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("Siegfried: error reading input, got %v", err)
	}
	res := make(chan core.Identification)
	recs := make([]core.Recorder, len(s.ids))
	for i, v := range s.ids {
		recs[i] = v.Recorder()
	}
	// Extension Matcher
	if len(n) > 0 {
		ems := s.em.Identify(n, nil)
		for v := range ems {
			for _, rec := range recs {
				if rec.Record(core.ExtensionMatcher, v) {
					break
				}
			}
		}
	}

	// Container Matcher
	if s.cm != nil {
		cms := s.cm.Identify(n, s.buffer)
		for v := range cms {
			for _, rec := range recs {
				if rec.Record(core.ContainerMatcher, v) {
					break
				}
			}
		}
	}
	satisfied := true
	for _, rec := range recs {
		if !rec.Satisfied() {
			satisfied = false
		}
	}
	// Byte Matcher
	if !satisfied {
		ids := s.bm.Identify("", s.buffer)
		for v := range ids {
			for _, rec := range recs {
				if rec.Record(core.ByteMatcher, v) {
					break
				}
			}
		}
	}
	go func() {
		for _, rec := range recs {
			rec.Report(res)
		}
		close(res)
	}()
	return res, nil
}

func (s *Siegfried) String() string {
	var str string
	str += s.em.String()
	str += s.cm.String()
	str += s.bm.String()
	return str
}
