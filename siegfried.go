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
	"archive/zip"
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
	"github.com/richardlehane/siegfried/pkg/core/signature"
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
	ids     []core.Identifier // identifiers
	buffers *siegreader.Buffers
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
	s.buffers = siegreader.New()
	return s
}

// Save a Siegfried signature file
func (s *Siegfried) Save(path string) error {
	ls := signature.NewLoadSaver(nil)
	ls.SaveString("siegfried")
	ls.SaveTime(s.C)
	s.em.Save(ls)
	s.cm.Save(ls)
	s.bm.Save(ls)
	ls.SaveTinyUInt(len(s.ids))
	for _, i := range s.ids {
		i.Save(ls)
	}
	if ls.Err != nil {
		return ls.Err
	}
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.Write(ls.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// Save a Siegfried signature file
func (s *Siegfried) SaveC(path string) error {
	ls := signature.NewLoadSaver(nil)
	ls.SaveString("siegfried")
	ls.SaveTime(s.C)
	s.em.Save(ls)
	s.cm.Save(ls)
	s.bm.Save(ls)
	ls.SaveTinyUInt(len(s.ids))
	for _, i := range s.ids {
		i.Save(ls)
	}
	if ls.Err != nil {
		return ls.Err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	z := zip.NewWriter(f)
	zw, err := z.Create("siegfried")
	if err != nil {
		return err
	}
	_, err = zw.Write(ls.Bytes())
	if err != nil {
		return err
	}
	z.Close()
	return nil
}

// Load a Siegfried signature file
func LoadC(path string) (*Siegfried, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	rc, err := r.File[0].Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	buf, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("Siegfried: error opening signature file; got %s\nTry running `sf -update`", err)
	}
	ls := signature.NewLoadSaver(buf)
	if ls.LoadString() != "siegfried" {
		return nil, fmt.Errorf("Siegfried: not a siegfried signature file")
	}
	return &Siegfried{
		C:       ls.LoadTime(),
		em:      extensionmatcher.Load(ls),
		cm:      containermatcher.Load(ls),
		bm:      bytematcher.Load(ls),
		ids:     loadIDs(ls),
		buffers: siegreader.New(),
	}, ls.Err
}

// Load a Siegfried signature file
func Load(path string) (*Siegfried, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Siegfried: error opening signature file; got %s\nTry running `sf -update`", err)
	}
	ls := signature.NewLoadSaver(buf)
	if ls.LoadString() != "siegfried" {
		return nil, fmt.Errorf("Siegfried: not a siegfried signature file")
	}
	return &Siegfried{
		C:       ls.LoadTime(),
		em:      extensionmatcher.Load(ls),
		cm:      containermatcher.Load(ls),
		bm:      bytematcher.Load(ls),
		ids:     loadIDs(ls),
		buffers: siegreader.New(),
	}, ls.Err
}

func loadIDs(ls *signature.LoadSaver) []core.Identifier {
	ids := make([]core.Identifier, ls.LoadTinyUInt())
	for i := range ids {
		ids[i] = core.LoadIdentifier(ls)
	}
	return ids
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

// Yaml representation of a Siegfried
// This is the provenace block at the beginning of siegfried results and includes Yaml descriptions for each identifier.
func (s *Siegfried) Yaml() string {
	version := config.Version()
	str := fmt.Sprintf(
		"---\nsiegfried   : %d.%d.%d\nscandate    : %v\nsignature   : %s\ncreated     : %v\nidentifiers : \n",
		version[0], version[1], version[2],
		time.Now().Format(time.RFC3339),
		config.SignatureBase(),
		s.C.Format(time.RFC3339))
	for _, id := range s.ids {
		d := id.Describe()
		str += fmt.Sprintf("  - name    : '%v'\n    details : '%v'\n", d[0], d[1])
	}
	return str
}

func (s *Siegfried) Json() string {
	version := config.Version()
	str := fmt.Sprintf(
		"{\"siegfried\":\"%d.%d.%d\",\"scandate\":\"%v\",\"signature\":\"%s\",\"created\":\"%v\",\"identifiers\":[",
		version[0], version[1], version[2],
		time.Now().Format(time.RFC3339),
		config.SignatureBase(),
		s.C.Format(time.RFC3339))
	for i, id := range s.ids {
		if i > 0 {
			str += ","
		}
		d := id.Describe()
		str += fmt.Sprintf("{\"name\":\"%s\",\"details\":\"%s\"}", d[0], d[1])
	}
	str += "],"
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

// Identify identifies a stream or file object.
// It takes the name of the file/stream (if unknown, give an empty string) and an io.Reader
// It returns a channel of identifications and an error
func (s *Siegfried) Identify(n string, r io.Reader) (chan core.Identification, error) {
	buffer, err := s.buffers.Get(r)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	res := make(chan core.Identification)
	recs := make([]core.Recorder, len(s.ids))
	for i, v := range s.ids {
		recs[i] = v.Recorder()
	}
	// Extension Matcher
	if len(n) > 0 {
		ems, _ := s.em.Identify(n, nil) // we don't care about an error here
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
		cms, cerr := s.cm.Identify(n, buffer)
		for v := range cms {
			for _, rec := range recs {
				if rec.Record(core.ContainerMatcher, v) {
					break
				}
			}
		}
		err = cerr
	}
	satisfied := true
	for _, rec := range recs {
		if !rec.Satisfied() {
			satisfied = false
		}
	}
	// Byte Matcher
	if !satisfied {
		ids, _ := s.bm.Identify("", buffer) // we don't care about an error here
		for v := range ids {
			for _, rec := range recs {
				if rec.Record(core.ByteMatcher, v) {
					break
				}
			}
		}
	}
	s.buffers.Put(buffer)
	go func() {
		for _, rec := range recs {
			rec.Report(res)
		}
		close(res)
	}()
	return res, err
}
