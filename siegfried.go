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
	"github.com/richardlehane/siegfried/pkg/core/mimematcher"
	"github.com/richardlehane/siegfried/pkg/core/namematcher"
	"github.com/richardlehane/siegfried/pkg/core/persist"
	"github.com/richardlehane/siegfried/pkg/core/riffmatcher"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
	"github.com/richardlehane/siegfried/pkg/core/textmatcher"
	"github.com/richardlehane/siegfried/pkg/core/xmlmatcher"
	"github.com/richardlehane/siegfried/pkg/mimeinfo"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var ( // for side effect - register their patterns
	_ = pronom.Range{}
	_ = mimeinfo.Int8(0)
)

// Siegfried structs are persisent objects that can be serialised to disk and
// used to identify file formats.
// They contain three matchers as well as a slice of identifiers. When identifiers
// are added to a Siegfried struct, they are registered with each matcher.
type Siegfried struct {
	C  time.Time    // signature create time
	nm core.Matcher // namematcher
	mm core.Matcher // mimematcher
	cm core.Matcher // containermatcher
	xm core.Matcher // bytematcher
	rm core.Matcher // riffmatcher
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
	return &Siegfried{
		C:       time.Now(),
		buffers: siegreader.New(),
	}
}

// Add adds an identifier to a Siegfried struct.
func (s *Siegfried) Add(i core.Identifier) error {
	for _, v := range s.ids {
		if v.Name() == i.Name() {
			return fmt.Errorf("siegfried: identifiers must have unique names, you already have an identifier named %s. Use the -name flag to assign a new name e.g. `roy add -name richard`", i.Name())
		}
	}
	var err error
	if s.nm, err = i.Add(s.nm, core.NameMatcher); err != nil {
		return err
	}
	if s.mm, err = i.Add(s.mm, core.MIMEMatcher); err != nil {
		return err
	}
	if s.cm, err = i.Add(s.cm, core.ContainerMatcher); err != nil {
		return err
	}
	if s.xm, err = i.Add(s.xm, core.XMLMatcher); err != nil {
		return err
	}
	if s.rm, err = i.Add(s.rm, core.RIFFMatcher); err != nil {
		return err
	}
	if s.bm, err = i.Add(s.bm, core.ByteMatcher); err != nil {
		return err
	}
	if s.tm, err = i.Add(s.tm, core.TextMatcher); err != nil {
		return err
	}
	s.ids = append(s.ids, i)
	return nil
}

// Save persists a Siegfried struct to disk (path)
func (s *Siegfried) Save(path string) error {
	ls := persist.NewLoadSaver(nil)
	ls.SaveTime(s.C)
	namematcher.Save(s.nm, ls)
	mimematcher.Save(s.mm, ls)
	containermatcher.Save(s.cm, ls)
	xmlmatcher.Save(s.xm, ls)
	riffmatcher.Save(s.rm, ls)
	bytematcher.Save(s.bm, ls)
	textmatcher.Save(s.tm, ls)
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
		nm: namematcher.Load(ls),
		mm: mimematcher.Load(ls),
		cm: containermatcher.Load(ls),
		xm: xmlmatcher.Load(ls),
		rm: riffmatcher.Load(ls),
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
		str += fmt.Sprintf("  - %v: %v\n", id.Name(), id.Details())
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
		str += fmt.Sprintf("  - name    : '%v'\n    details : '%v'\n", id.Name(), id.Details())
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
		str += fmt.Sprintf("{\"name\":\"%s\",\"details\":\"%s\"}", id.Name(), id.Details())
	}
	str += "],"
	return str
}

// Fields returns a slice of the names of the fields in each identifier.
func (s *Siegfried) Fields() [][]string {
	ret := make([][]string, len(s.ids))
	for i, v := range s.ids {
		ret[i] = v.Fields()
	}
	return ret
}

// Identify identifies a stream or file object.
// It takes the name of the file/stream (if unknown, give an empty string) and an io.Reader
// It returns a channel of identifications and an error.
func (s *Siegfried) Identify(r io.Reader, name, mime string) (chan core.Identification, error) {
	buffer, err := s.buffers.Get(r)
	if err == io.EOF {
		err = nil
	}
	if err != nil && err != siegreader.ErrEmpty {
		return nil, fmt.Errorf("siegfried: error reading file; got %v", err)
	}
	res := make(chan core.Identification)
	recs := make([]core.Recorder, len(s.ids))
	for i, v := range s.ids {
		recs[i] = v.Recorder()
		if name != "" {
			recs[i].Active(core.NameMatcher)
		}
		if mime != "" {
			recs[i].Active(core.MIMEMatcher)
		}
		if err == nil {
			recs[i].Active(core.XMLMatcher)
			recs[i].Active(core.TextMatcher)
		}
	}
	// Name Matcher
	if len(name) > 0 && s.nm != nil {
		nms, _ := s.nm.Identify(name, nil) // we don't care about an error here
		for v := range nms {
			for _, rec := range recs {
				if rec.Record(core.NameMatcher, v) {
					break
				}
			}
		}
	}
	// MIME Matcher
	if len(mime) > 0 && s.mm != nil {
		mms, _ := s.mm.Identify(mime, nil) // we don't care about an error here
		for v := range mms {
			for _, rec := range recs {
				if rec.Record(core.MIMEMatcher, v) {
					break
				}
			}
		}
	}
	// Container Matcher
	if s.cm != nil {
		if config.Debug() {
			fmt.Fprintln(config.Out(), ">>START CONTAINER MATCHER")
		}
		cms, cerr := s.cm.Identify(name, buffer)
		for v := range cms {
			for _, rec := range recs {
				if rec.Record(core.ContainerMatcher, v) {
					break
				}
			}
		}
		if err == nil {
			err = cerr
		}
	}
	satisfied := true
	// XML Matcher
	if s.xm != nil {
		for _, rec := range recs {
			if ok, _ := rec.Satisfied(core.XMLMatcher); !ok {
				satisfied = false
				break
			}
		}
		if !satisfied {
			if config.Debug() {
				fmt.Fprintln(config.Out(), ">>START XML MATCHER")
			}
			xms, xerr := s.xm.Identify("", buffer)
			for v := range xms {
				for _, rec := range recs {
					if rec.Record(core.XMLMatcher, v) {
						break
					}
				}
			}
			if err == nil {
				err = xerr
			}
		}
	}
	satisfied = true
	// RIFF Matcher
	if s.rm != nil {
		for _, rec := range recs {
			if ok, _ := rec.Satisfied(core.RIFFMatcher); !ok {
				satisfied = false
				break
			}
		}
		if !satisfied {
			if config.Debug() {
				fmt.Fprintln(config.Out(), ">>START RIFF MATCHER")
			}
			rms, rerr := s.rm.Identify("", buffer)
			for v := range rms {
				for _, rec := range recs {
					if rec.Record(core.RIFFMatcher, v) {
						break
					}
				}
			}
			if err == nil {
				err = rerr
			}
		}
	}
	satisfied = true
	exclude := make([]int, 0, len(recs))
	for _, rec := range recs {
		ok, ex := rec.Satisfied(core.ByteMatcher)
		if !ok {
			satisfied = false
		} else {
			exclude = append(exclude, ex)
		}
	}
	// Byte Matcher
	if s.bm != nil && !satisfied {
		if config.Debug() {
			fmt.Fprintln(config.Out(), ">>START BYTE MATCHER")
		}
		ids, _ := s.bm.Identify("", buffer, exclude...) // we don't care about an error here
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
		if ok, _ := rec.Satisfied(core.TextMatcher); !ok {
			satisfied = false
			break
		}
	}
	// Text Matcher
	if s.tm != nil && !satisfied {
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
// which strikes.
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
// This prevents unnecessary double-up of IO e.g. when unzipping files post-identification.
func (s *Siegfried) Buffer() *siegreader.Buffer {
	last := s.buffers.Last()
	last.Quit = make(chan struct{}) // may have already closed the quit channel
	return last
}

// Update checks whether a Siegfried struct is due for update, by testing whether the time given is after the time
// the signature was created.
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
	case core.NameMatcher:
		return s.nm.String()
	case core.MIMEMatcher:
		return s.mm.String()
	case core.ContainerMatcher:
		return s.cm.String()
	}
	return fmt.Sprintf("Identifiers\n%s\nByte Matcher\n%sText Matcher\n%s",
		func() string {
			var str string
			for _, i := range s.ids {
				str += i.String()
			}
			return str
		}(),
		s.bm.String(),
		s.tm.String())
}
