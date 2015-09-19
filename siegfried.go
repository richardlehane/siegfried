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
//  s, err := siegfried.Load("pronom.sig")
//  if err != nil {
//  	log.Fatal(err)
//  }
//  f, err := os.Open("file")
//  if err != nil {
//  	log.Fatal(err)
//  }
//  defer f.Close()
//  c, err := s.Identify("filename", f)
//  if err != nil {
//  	log.Fatal(err)
//  }
//  for id := range c {
//  	fmt.Println(id)
//  }
package siegfried

import (
	"bytes"
	"compress/flate"
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
	"github.com/richardlehane/siegfried/pkg/core/persist"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
	"github.com/richardlehane/siegfried/pkg/core/textmatcher"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

// Siegfried structs are persisent objects that can be serialised to disk and
// used to identify file formats.
// They contain three matchers as well as a slice of identifiers. When identifiers
// are added to a Siegfried struct, they are registered with each matcher.
type Siegfried struct {
	C  time.Time    // signature create time
	em core.Matcher // extensionmatcher
	cm core.Matcher // containermatcher
	bm core.Matcher // bytematcher
	tm core.Matcher // textmatcher
	// mutatable fields follow
	ids     []core.Identifier // identifiers
	buffers *siegreader.Buffers
}

// New creates a new Siegfried struct. It initializes the three matchers.
//
// Example:
//  s := New()
//  p, err := pronom.New() // create a new PRONOM identifier
//  if err != nil {
//  	log.Fatal(err)
//  }
//  err = s.Add(p) // add the identifier to the Siegfried
//  if err != nil {
//  	log.Fatal(err)
//  }
//  err = s.Save("pronom.sig") // save the Siegfried
func New() *Siegfried {
	s := &Siegfried{}
	s.C = time.Now()
	s.em = extensionmatcher.New()
	s.cm = containermatcher.New()
	s.bm = bytematcher.New()
	s.tm = textmatcher.New()
	s.buffers = siegreader.New()
	return s
}

// Add adds an identifier to a Siegfried struct.
// The identifer is type switched to test if it is supported. At present, only PRONOM identifiers are supported
func (s *Siegfried) Add(i core.Identifier) error {
	switch i := i.(type) {
	default:
		return fmt.Errorf("siegfried: unknown identifier type %T", i)
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
		if err := i.Add(s.tm); err != nil {
			return err
		}
		s.ids = append(s.ids, i)
	}
	return nil
}

// Save persists a Siegfried struct to disk (path)
func (s *Siegfried) Save(path string) error {
	ls := persist.NewLoadSaver(nil)
	ls.SaveTime(s.C)
	s.em.Save(ls)
	s.cm.Save(ls)
	s.bm.Save(ls)
	s.tm.Save(ls)
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
	_, err = f.Write(append(config.Magic(), byte(config.Version()[0]), byte(config.Version()[1])))
	if err != nil {
		return err
	}
	z, err := flate.NewWriter(f, 1)
	if err != nil {
		return err
	}
	_, err = z.Write(ls.Bytes())
	z.Close()
	return err
}

// Load creates a Siegfried struct and loads content from path
func Load(path string) (*Siegfried, error) {
	errOpening := "siegfried: error opening signature file, got %v; try running `sf -update`"
	errNotSig := "siegfried: not a siegfried signature file; try running `sf -update`"
	errUpdateSig := "siegfried: signature file is incompatible with this version of sf; try running `sf -update`"
	fbuf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(errOpening, err)
	}
	if len(fbuf) < len(config.Magic())+2 {
		return nil, fmt.Errorf(errNotSig)
	}
	if string(fbuf[:len(config.Magic())]) != string(config.Magic()) {
		return nil, fmt.Errorf(errNotSig)
	}
	if major, minor := fbuf[len(config.Magic())], fbuf[len(config.Magic())+1]; major < byte(config.Version()[0]) || (major == byte(config.Version()[0]) && minor < byte(config.Version()[1])) {
		return nil, fmt.Errorf(errUpdateSig)
	}
	r := bytes.NewBuffer(fbuf[len(config.Magic())+2:])
	rc := flate.NewReader(r)
	defer rc.Close()
	buf, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf(errOpening, err)
	}
	ls := persist.NewLoadSaver(buf)
	return &Siegfried{
		C:  ls.LoadTime(),
		em: extensionmatcher.Load(ls),
		cm: containermatcher.Load(ls),
		bm: bytematcher.Load(ls),
		tm: textmatcher.Load(ls),
		ids: func() []core.Identifier {
			ids := make([]core.Identifier, ls.LoadTinyUInt())
			for i := range ids {
				ids[i] = core.LoadIdentifier(ls)
			}
			return ids
		}(),
		buffers: siegreader.New(),
	}, ls.Err
}

// String representation of a Siegfried struct
func (s *Siegfried) String() string {
	str := fmt.Sprintf(
		"%s (%v)\nidentifiers: \n",
		config.Signature(),
		s.C.Format(time.RFC3339))
	for _, id := range s.ids {
		d := id.Describe()
		str += fmt.Sprintf("  - %v: %v\n", d[0], d[1])
	}
	return str
}

// YAML representation of a Siegfried struct.
// This is the provenace block at the beginning of sf results and includes descriptions for each identifier.
func (s *Siegfried) YAML() string {
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

// JSON representation of a Siegfried struct.
// This is the provenace block at the beginning of sf results and includes descriptions for each identifier.
func (s *Siegfried) JSON() string {
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

// Identify identifies a stream or file object.
// It takes the name of the file/stream (if unknown, give an empty string) and an io.Reader
// It returns a channel of identifications and an error
func (s *Siegfried) Identify(n string, r io.Reader) (chan core.Identification, error) {
	buffer, err := s.buffers.Get(r)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("siegfried: error reading file; got %v", err)
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
		if config.Debug() {
			fmt.Println(">>START CONTAINER MATCHER")
		}
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
		if !rec.Satisfied(core.ByteMatcher) {
			satisfied = false
		}
	}
	// Byte Matcher
	if !satisfied {
		if config.Debug() {
			fmt.Println(">>START BYTE MATCHER")
		}
		ids, _ := s.bm.Identify("", buffer) // we don't care about an error here
		for v := range ids {
			for _, rec := range recs {
				if rec.Record(core.ByteMatcher, v) {
					break
				}
			}
		}
	}
	satisfied = true
	for _, rec := range recs {
		if !rec.Satisfied(core.TextMatcher) {
			satisfied = false
			break
		}
	}
	// Text Matcher
	if !satisfied {
		ids, _ := s.tm.Identify("", buffer) // we don't care about an error here
		for v := range ids {
			for _, rec := range recs {
				if rec.Record(core.TextMatcher, v) {
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

// Blame checks with the byte matcher to see what identification results subscribe to a particular result or test
// tree index. It can be used when identifying in a debug mode to check which identification results trigger
// which strikes
func (s *Siegfried) Blame(idx, ct int, cn string) string {
	matcher := "BYTE MATCHER"
	matcherType := core.ByteMatcher
	var ttis []int
	if cn != "" {
		matcher = "CONTAINER MATCHER"
		matcherType = core.ContainerMatcher
		cm := s.cm.(containermatcher.Matcher)
		ttis = cm.InspectTestTree(ct, cn, idx)
		res := make([]string, len(ttis))
		for i, v := range ttis {
			for _, id := range s.ids {
				if ok, str := id.Recognise(matcherType, v); ok {
					res[i] = str
				}
				break
			}
		}
		ttiNames := "not recognised"
		if len(res) > 0 {
			ttiNames = strings.Join(res, ",")
		}
		return fmt.Sprintf("%s\nHits at %d: %s (identifies hits reported by -debug)", matcher, idx, ttiNames)
	}
	bm := s.bm.(*bytematcher.Matcher)
	ttis = bm.InspectTestTree(idx)
	res := make([]string, len(ttis)+1)
	for _, id := range s.ids {
		if ok, str := id.Recognise(matcherType, idx); ok {
			res[0] = str
		}
		for i, v := range ttis {
			if ok, str := id.Recognise(matcherType, v); ok {
				res[i+1] = str
			}
		}
	}
	resName := res[0]
	if resName == "" {
		resName = "not recognised"
	}
	ttiNames := "not recognised"
	if len(res) > 1 {
		ttiNames = strings.Join(res[1:], ",")
	}
	return fmt.Sprintf("%s\nResults at %d: %s (identifies results reported by -slow)\nHits at %d: %s (identifies hits reported by -debug)", matcher, idx, resName, idx, ttiNames)

}

// Buffer returns the last buffer inspected
// This prevents unnecessary double-up of IO e.g. when unzipping files post-identification
func (s *Siegfried) Buffer() siegreader.Buffer {
	last := s.buffers.Last()
	last.SetQuit(make(chan struct{})) // may have already closed the quit channel
	return last
}

// Update checks whether a Siegfried struct is due for update, by testing whether the time given is after the time
// the signature was created
func (s *Siegfried) Update(t string) bool {
	tm, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return false
	}
	return tm.After(s.C)
}

// Inspect returns a string containing detail about the various matchers in the Siegfried struct.
func (s *Siegfried) Inspect(t core.MatcherType) string {
	switch t {
	case core.ByteMatcher:
		return s.bm.String()
	case core.ExtensionMatcher:
		return s.em.String()
	case core.ContainerMatcher:
		return s.cm.String()
	}
	return fmt.Sprintf("Identifiers\n%s\nExtension Matcher\n%s\nContainer Matcher\n%s\nByte Matcher\n%sText Matcher\n%s",
		func() string {
			var str string
			for _, i := range s.ids {
				str += i.String()
			}
			return str
		}(),
		s.em.String(),
		s.cm.String(),
		s.bm.String(),
		s.tm.String())
}
