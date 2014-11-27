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

// Package siegfried identifies file formats
//
// Example:
//  s, err := siegfried.Load("pronom.gob")
//  if err != nil {
//  	// handle err
//  }
//  f, _ := os.Open("file")
//  defer f.Close()
//  c, err := s.Identify("filename", f)
//  if err != nil {
//  	// handle err
//  }
//  for id := range c {
//  	fmt.Print(id.Yaml())
//  }
package siegfried

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/core/containermatcher"
	"github.com/richardlehane/siegfried/pkg/core/extensionmatcher"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

// Siegfried structs are persisent objects that can be serialised to disk (using encoding/gob).
// The private fields are the three matchers (extension, container, byte) and the identifiers.
type Siegfried struct {
	C  time.Time    // signature create time
	em core.Matcher // extensionmatcher
	cm core.Matcher // containermatcher
	bm core.Matcher // bytematcher
	// mutatable fields follow
	ids    []core.Identifier // identifiers
	buffer *siegreader.Buffer
}

// New creates a new Siegfried. It sets the create time to time.Now() and initializes the three matchers
//
// Example:
//  s := New() // create a new Siegfried
//  p, err := pronom.New() // create a new PRONOM identifier
//  if err != nil {
//  	// handle err
//  }
//  err = s.Add(p) // add the identifier to the Siegfried
//  if err != nil {
//  	// handle err
//  }
//  err = s.Save("signature.gob") // save the Siegfried
//  if err != nil {
//  	// handle err
//  }
func New() *Siegfried {
	s := &Siegfried{}
	s.C = time.Now()
	s.em = extensionmatcher.New()
	s.cm = containermatcher.New()
	s.bm = bytematcher.New()
	s.buffer = siegreader.New()
	return s
}

// String representation of a Siegfried
func (s *Siegfried) String() string {
	str := "IDENTIFIERS\n"
	for _, i := range s.ids {
		str += i.String()
	}
	str += "\nEXTENSION MATCHER\n"
	str += s.em.String()
	str += "\nCONTAINER MATCHER\n"
	str += s.cm.String()
	str += "\nBYTE MATCHER\n"
	str += s.bm.String()
	return str
}

// InspectTTI checks with the byte matcher to see what identification results subscribe to a particular test
// tree index. It can be used when identifying in a debug mode to check which identification results trigger
// which strikes
func (s *Siegfried) InspectTTI(tti int) string {
	bm := s.bm.(*bytematcher.Matcher)
	idxs := bm.InspectTTI(tti)
	if idxs == nil {
		return "No test tree at this index"
	}
	res := make([]string, len(idxs))
	for i, v := range idxs {
		for _, id := range s.ids {
			ok, str := id.Recognise(core.ByteMatcher, v)
			if ok {
				res[i] = str
				break
			}
		}
	}
	return "Test tree indexes match:\n" + strings.Join(res, "\n")
}

// Add adds an identifier to a Siegfried.
// The identifer is type switched to test if it is supported. At present, only PRONOM identifiers are supported
func (s *Siegfried) Add(i core.Identifier) error {
	switch i := i.(type) {
	default:
		return fmt.Errorf("Siegfried: unknown identifier type %T", i)
	case *pronom.Identifier:
		if err := i.Add(s.em); err != nil {
			return err
		}
		if err := i.Add(s.cm); err != nil {
			return err
		}
		if err := i.Add(s.bm); err != nil {
			return err
		}
		s.ids = append(s.ids, i)
	}
	return nil
}

// YAML representation of a Siegfried
func (s *Siegfried) Yaml() string {
	version := config.Version()
	str := fmt.Sprintf(
		"---\nsiegfried   : %d.%d.%d\nscandate    : %v\nsignature   : %s\ncreated     : %v\nidentifiers : \n",
		version[0], version[1], version[2],
		time.Now().Format(time.RFC3339),
		config.SignatureBase(),
		s.C.Format(time.RFC3339))
	for _, id := range s.ids {
		str += id.Yaml()
	}
	return str
}

// Update checks whether a Siegfried is due for update, by testing whether the time given is after the time
// the signature was created
func (s *Siegfried) Update(t string) bool {
	tm, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return false
	}
	return tm.After(s.C)
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
		return nil, fmt.Errorf("Siegfried: error opening signature file; got %s\nTry running `sf -update`", err)
	}
	buf := bytes.NewBuffer(c)
	dec := gob.NewDecoder(buf)
	var h Header
	err = dec.Decode(&h)
	if err != nil {
		return nil, fmt.Errorf("Siegfried: error reading signature file; got %s\nTry running `sf -update`", err)
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
		return nil, fmt.Errorf("Siegfried: error loading bytematcher; got %s", err)
	}
	cm, err := containermatcher.Load(cbuf)
	if err != nil {
		return nil, fmt.Errorf("Siegfried: error loading containermatcher; got %s", err)
	}
	em, err := extensionmatcher.Load(ebuf)
	if err != nil {
		return nil, fmt.Errorf("Siegfried: error loading extensionmatcher; got %s", err)
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
				return nil, fmt.Errorf("Siegfried: loading PRONOM identifier; got %s", err)
			}
		}
		s.ids[i] = id
		istart += v.Sz
	}
	s.buffer = siegreader.New()
	return &s, nil
}

func (s *Siegfried) Identify(n string, r io.Reader) (chan core.Identification, error) {
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
