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

package containermatcher

import (
	"encoding/binary"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/richardlehane/siegfried/internal/bytematcher"
	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/internal/siegreader"
	"github.com/richardlehane/siegfried/pkg/core"
)

type containerType int

const (
	Zip   containerType = iota // Zip container type e.g. for .docx etc.
	Mscfb                      // Mscfb container type  e.g. for .doc etc.
)

// Matcher is a slice of container matchers
type Matcher []*ContainerMatcher

// Load returns a container Matcher
func Load(ls *persist.LoadSaver) core.Matcher {
	if !ls.LoadBool() {
		return nil
	}
	ret := make(Matcher, ls.LoadTinyUInt())
	for i := range ret {
		ret[i] = loadCM(ls)
		ret[i].ctype = ctypes[ret[i].conType]
		ret[i].entryBufs = siegreader.New()
	}
	return ret
}

// Save encodes a container Matcher
func Save(c core.Matcher, ls *persist.LoadSaver) {
	if c == nil {
		ls.SaveBool(false)
		return
	}
	m := c.(Matcher)
	if m.total(-1) == 0 {
		ls.SaveBool(false)
		return
	}
	ls.SaveBool(true)
	ls.SaveTinyUInt(len(m))
	for _, v := range m {
		v.save(ls)
	}
}

type SignatureSet struct {
	Typ       containerType
	NameParts [][]string
	SigParts  [][]frames.Signature
}

func Add(c core.Matcher, ss core.SignatureSet, l priority.List) (core.Matcher, int, error) {
	var m Matcher
	if c == nil {
		m = Matcher{newZip(), newMscfb()}
	} else {
		m = c.(Matcher)
	}
	sigs, ok := ss.(SignatureSet)
	if !ok {
		return nil, 0, fmt.Errorf("container matcher error: cannot convert signature set to CM signature set")
	}
	err := m.addSigs(int(sigs.Typ), sigs.NameParts, sigs.SigParts, l)
	if err != nil {
		return nil, 0, err
	}
	return m, m.total(-1), nil
}

// calculate total number of signatures present in the matcher. Provide -1 to get the total sum, or supply an index of an individual matcher to exclude that matcher's total
func (m Matcher) total(i int) int {
	var t int
	for j, v := range m {
		// don't include the count for the ContainerMatcher in question
		if i > -1 && j == i {
			continue
		}
		t += len(v.parts)
	}
	return t
}

func (m Matcher) addSigs(i int, nameParts [][]string, sigParts [][]frames.Signature, l priority.List) error {
	if len(m) < i+1 {
		return fmt.Errorf("container: missing container matcher")
	}
	var err error
	if len(nameParts) != len(sigParts) {
		return fmt.Errorf("container: expecting equal name and persist parts")
	}
	// give as a starting index the current total of signatures in the matcher, except those in the ContainerMatcher in question
	m[i].startIndexes = append(m[i].startIndexes, m.total(i))
	for j, n := range nameParts {
		err = m[i].addSignature(n, sigParts[j])
		if err != nil {
			return err
		}
	}
	// commit all the ctests
	for _, v := range m[i].nameCTest {
		err = v.commit()
		if err != nil {
			return err
		}
	}
	for _, v := range m[i].globCtests {
		err = v.commit()
		if err != nil {
			return err
		}
	}
	m[i].priorities.Add(l, len(nameParts), 0, 0)
	return nil
}

func (m Matcher) String() string {
	var str string
	for _, c := range m {
		str += c.String()
	}
	return str
}

type ContainerMatcher struct {
	ctype
	startIndexes []int //  added to hits - these place all container matches in a single slice
	conType      containerType
	nameCTest    map[string]*cTest // map of literal paths to ctests
	globs        []string          // corresponds with globCtests
	globCtests   []*cTest          //
	parts        []int             // corresponds with each signature: represents the number of CTests for each sig
	priorities   *priority.Set
	extension    string
	entryBufs    *siegreader.Buffers
}

func loadCM(ls *persist.LoadSaver) *ContainerMatcher {
	ct := &ContainerMatcher{
		startIndexes: ls.LoadInts(),
		conType:      containerType(ls.LoadTinyUInt()),
		nameCTest:    loadCTests(ls),
		globs:        ls.LoadStrings(),
	}
	gcts := make([]*cTest, ls.LoadSmallInt())
	for i := range gcts {
		gcts[i] = loadCTest(ls)
	}
	ct.globCtests = gcts
	ct.parts = ls.LoadInts()
	ct.priorities = priority.Load(ls)
	ct.extension = ls.LoadString()
	return ct
}

func (c *ContainerMatcher) save(ls *persist.LoadSaver) {
	ls.SaveInts(c.startIndexes)
	ls.SaveTinyUInt(int(c.conType))
	saveCTests(ls, c.nameCTest)
	ls.SaveStrings(c.globs)
	ls.SaveSmallInt(len(c.globCtests))
	for _, v := range c.globCtests {
		saveCTest(ls, v)
	}
	ls.SaveInts(c.parts)
	c.priorities.Save(ls)
	ls.SaveString(c.extension)
}

func (c *ContainerMatcher) String() string {
	str := "\nContainer matcher:\n"
	str += fmt.Sprintf("Type: %d\n", c.conType)
	str += fmt.Sprintf("Priorities: %v\n", c.priorities)
	str += fmt.Sprintf("Parts: %v\n", c.parts)
	str += fmt.Sprintf("%d literal tests, %d glob tests\n", len(c.nameCTest), len(c.globCtests))
	for k, v := range c.nameCTest {
		str += "-----------\n"
		str += fmt.Sprintf("Name: %v\n", k)
		str += fmt.Sprintf("Satisfied: %v\n", v.satisfied)
		str += fmt.Sprintf("Unsatisfied: %v\n", v.unsatisfied)
		if v.bm == nil {
			str += "Bytematcher: None\n"
		} else {
			str += "Bytematcher:\n" + v.bm.String()
		}
	}
	for i, v := range c.globs {
		str += "-----------\n"
		str += fmt.Sprintf("Glob: %v\n", v)
		str += fmt.Sprintf("Satisfied: %v\n", c.globCtests[i].satisfied)
		str += fmt.Sprintf("Unsatisfied: %v\n", c.globCtests[i].unsatisfied)
		if c.globCtests[i].bm == nil {
			str += "Bytematcher: None\n"
		} else {
			str += "Bytematcher:\n" + c.globCtests[i].bm.String()
		}
	}
	return str
}

type ctype struct {
	trigger func([]byte) bool
	rdr     func(*siegreader.Buffer) (Reader, error)
}

var ctypes = []ctype{
	{
		zipTrigger,
		zipRdr, // see zip.go
	},
	{
		mscfbTrigger,
		mscfbRdr, // see mscfb.go
	},
}

func zipTrigger(b []byte) bool {
	return binary.LittleEndian.Uint32(b[:4]) == 0x04034B50
}

func newZip() *ContainerMatcher {
	return &ContainerMatcher{
		ctype:      ctypes[0],
		conType:    Zip,
		nameCTest:  make(map[string]*cTest),
		priorities: &priority.Set{},
		extension:  "zip",
		entryBufs:  siegreader.New(),
	}
}

func mscfbTrigger(b []byte) bool {
	return binary.LittleEndian.Uint64(b) == 0xE11AB1A1E011CFD0
}

func newMscfb() *ContainerMatcher {
	return &ContainerMatcher{
		ctype:      ctypes[1],
		conType:    Mscfb,
		nameCTest:  make(map[string]*cTest),
		priorities: &priority.Set{},
		entryBufs:  siegreader.New(),
	}
}

func (c *ContainerMatcher) addSignature(nameParts []string, sigParts []frames.Signature) error {
	if len(nameParts) != len(sigParts) {
		return errors.New("container matcher: nameParts and sigParts must be equal")
	}
	c.parts = append(c.parts, len(nameParts))
outer:
	for i, nm := range nameParts {
		if nm != "[Content_Types].xml" && strings.ContainsAny(nm, "*?[]") {
			// is glob pattern is valid
			if _, err := filepath.Match(nm, ""); err == nil {
				// do we already have this glob?
				for i, v := range c.globs {
					if nm == v {
						c.globCtests[i].add(sigParts[i], len(c.parts)-1)
						continue outer
					}
				}
				c.globs = append(c.globs, nm)
				ct := &cTest{}
				ct.add(sigParts[i], len(c.parts)-1)
				c.globCtests = append(c.globCtests, ct)
				continue
			}
		}
		ct, ok := c.nameCTest[nm]
		if !ok {
			ct = &cTest{}
			c.nameCTest[nm] = ct
		}
		ct.add(sigParts[i], len(c.parts)-1)
	}
	return nil
}

// a container test is a the basic element of container matching
type cTest struct {
	satisfied   []int              // satisfied persists are immediately matched: i.e. a name without a required bitstream
	unsatisfied []int              // unsatisfied persists depend on bitstreams as well as names matching
	buffer      []frames.Signature // temporary - used while creating CTests
	bm          core.Matcher       // bytematcher
}

func loadCTests(ls *persist.LoadSaver) map[string]*cTest {
	ret := make(map[string]*cTest)
	l := ls.LoadSmallInt()
	for i := 0; i < l; i++ {
		ret[ls.LoadString()] = loadCTest(ls)
	}
	return ret
}

func saveCTests(ls *persist.LoadSaver, ct map[string]*cTest) {
	ls.SaveSmallInt(len(ct))
	for k, v := range ct {
		ls.SaveString(k)
		saveCTest(ls, v)
	}
}

func loadCTest(ls *persist.LoadSaver) *cTest {
	return &cTest{
		satisfied:   ls.LoadInts(),
		unsatisfied: ls.LoadInts(),
		bm:          bytematcher.Load(ls),
	}
}

func saveCTest(ls *persist.LoadSaver, ct *cTest) {
	ls.SaveInts(ct.satisfied)
	ls.SaveInts(ct.unsatisfied)
	bytematcher.Save(ct.bm, ls)
}

func (ct *cTest) add(s frames.Signature, t int) {
	if s == nil {
		ct.satisfied = append(ct.satisfied, t)
		return
	}
	ct.unsatisfied = append(ct.unsatisfied, t)
	ct.buffer = append(ct.buffer, s)
}

// call for each key after all signatures added
func (ct *cTest) commit() error {
	if ct.buffer == nil {
		return nil
	}
	var err error
	ct.bm, _, err = bytematcher.Add(ct.bm, bytematcher.SignatureSet(ct.buffer), nil) // don't need to add priorities
	ct.buffer = nil
	return err
}

func (m Matcher) InspectTestTree(ct int, nm string, idx int) []int {
	for _, c := range m {
		if c.conType == containerType(ct) {
			if ctst, ok := c.nameCTest[nm]; ok {
				bmt := ctst.bm.(*bytematcher.Matcher).InspectTestTree(idx)
				ret := make([]int, len(bmt))
				for i, v := range bmt {
					s, _ := c.priorities.Index(ctst.unsatisfied[v])
					ret[i] = ctst.unsatisfied[v] + c.startIndexes[s]
				}
				return ret
			}
			return nil
		}
	}
	return nil
}
