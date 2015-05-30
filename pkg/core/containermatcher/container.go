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

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/persist"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type containerType int

const (
	Zip containerType = iota
	Mscfb
)

type Matcher []*ContainerMatcher

func Load(ls *persist.LoadSaver) Matcher {
	ret := make(Matcher, ls.LoadTinyUInt())
	for i := range ret {
		ret[i] = loadCM(ls)
		ret[i].ctype = ctypes[ret[i].CType]
		ret[i].entryBufs = siegreader.New()
	}
	return ret
}

func (m Matcher) Save(ls *persist.LoadSaver) {
	ls.SaveTinyUInt(len(m))
	for _, v := range m {
		v.save(ls)
	}
}

func New() Matcher {
	m := make(Matcher, 2)
	m[0] = newZip()
	m[1] = newMscfb()
	return m
}

type SignatureSet struct {
	Typ       containerType
	NameParts [][]string
	SigParts  [][]frames.Signature
}

func (m Matcher) Add(ss core.SignatureSet, l priority.List) (int, error) {
	sigs, ok := ss.(SignatureSet)
	if !ok {
		return 0, fmt.Errorf("Container matcher error: cannot convert persist set to CM persist set")
	}
	err := m.addSigs(int(sigs.Typ), sigs.NameParts, sigs.SigParts, l)
	if err != nil {
		return 0, err
	}
	return m.total(-1), nil
}

// calculate total number of persists present in the matcher. Provide -1 to get the total sum, or supply an index of an individual matcher to exclude that matcher's total
func (m Matcher) total(i int) int {
	var t int
	for j, v := range m {
		// don't include the count for the ContainerMatcher in question
		if i > -1 && j == i {
			continue
		}
		t += len(v.Parts)
	}
	return t
}

func (m Matcher) addSigs(i int, nameParts [][]string, sigParts [][]frames.Signature, l priority.List) error {
	if len(m) < i+1 {
		return fmt.Errorf("Container: missing container matcher")
	}
	var err error
	if len(nameParts) != len(sigParts) {
		return fmt.Errorf("Container: expecting equal name and persist parts")
	}
	// give as a starting index the current total of persists in the matcher, except those in the ContainerMatcher in question
	m[i].Sindexes = append(m[i].Sindexes, m.total(i))
	prev := len(m[i].Parts)
	for j, n := range nameParts {
		err = m[i].addSignature(n, sigParts[j])
		if err != nil {
			return err
		}
	}
	m[i].Priorities.Add(l, len(nameParts))
	for _, v := range m[i].NameCTest {
		err := v.commit(l, prev)
		if err != nil {
			return err
		}
	}
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
	Sindexes   []int // start indexes, added to hits - these place all container matches in a single slice
	CType      containerType
	NameCTest  map[string]*CTest
	Parts      []int // corresponds with each persist: represents the number of CTests for each sig
	Priorities *priority.Set
	extension  string
	entryBufs  *siegreader.Buffers
}

func loadCM(ls *persist.LoadSaver) *ContainerMatcher {
	return &ContainerMatcher{
		Sindexes:   ls.LoadInts(),
		CType:      containerType(ls.LoadTinyUInt()),
		NameCTest:  loadCTests(ls),
		Parts:      ls.LoadInts(),
		Priorities: priority.Load(ls),
		extension:  ls.LoadString(),
	}
}

func (c *ContainerMatcher) save(ls *persist.LoadSaver) {
	ls.SaveInts(c.Sindexes)
	ls.SaveTinyUInt(int(c.CType))
	saveCTests(ls, c.NameCTest)
	ls.SaveInts(c.Parts)
	c.Priorities.Save(ls)
	ls.SaveString(c.extension)
}

func (c *ContainerMatcher) String() string {
	str := "\nContainer matcher:\n"
	str += fmt.Sprintf("Type: %d\n", c.CType)
	str += fmt.Sprintf("Priorities: %v\n", c.Priorities)
	str += fmt.Sprintf("Parts: %v\n", c.Parts)
	for k, v := range c.NameCTest {
		str += "-----------\n"
		str += fmt.Sprintf("Name: %v\n", k)
		str += fmt.Sprintf("Satisfied: %v\n", v.Satisfied)
		str += fmt.Sprintf("Unsatisfied: %v\n", v.Unsatisfied)
		if v.BM == nil {
			str += "Bytematcher: None\n"
		} else {
			str += "Bytematcher:\n" + v.BM.String()
		}
	}
	return str
}

type ctype struct {
	trigger func([]byte) bool
	rdr     func(siegreader.Buffer) (Reader, error)
}

var ctypes = []ctype{
	ctype{
		zipTrigger,
		zipRdr, // see zip.go
	},
	ctype{
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
		CType:      Zip,
		NameCTest:  make(map[string]*CTest),
		Priorities: &priority.Set{},
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
		CType:      Mscfb,
		NameCTest:  make(map[string]*CTest),
		Priorities: &priority.Set{},
		entryBufs:  siegreader.New(),
	}
}

func (c *ContainerMatcher) addSignature(nameParts []string, sigParts []frames.Signature) error {
	if len(nameParts) != len(sigParts) {
		return errors.New("Container matcher: nameParts and sigParts must be equal")
	}
	c.Parts = append(c.Parts, len(nameParts))
	for i, nm := range nameParts {
		ct, ok := c.NameCTest[nm]
		if !ok {
			ct = &CTest{}
			c.NameCTest[nm] = ct
		}
		ct.add(sigParts[i], len(c.Parts)-1)
	}
	return nil
}

// a container test is a the basic element of container matching
type CTest struct {
	Satisfied   []int              // satisfied persists are immediately matched: i.e. a name without a required bitstream
	Unsatisfied []int              // unsatisfied persists depend on bitstreams as well as names matching
	buffer      []frames.Signature // temporary - used while creating CTests
	BM          *bytematcher.Matcher
}

//map[string]*CTest

func loadCTests(ls *persist.LoadSaver) map[string]*CTest {
	ret := make(map[string]*CTest)
	l := ls.LoadSmallInt()
	for i := 0; i < l; i++ {
		ret[ls.LoadString()] = &CTest{
			Satisfied:   ls.LoadInts(),
			Unsatisfied: ls.LoadInts(),
			BM:          bytematcher.Load(ls),
		}
	}
	return ret
}

func saveCTests(ls *persist.LoadSaver, ct map[string]*CTest) {
	ls.SaveSmallInt(len(ct))
	for k, v := range ct {
		ls.SaveString(k)
		ls.SaveInts(v.Satisfied)
		ls.SaveInts(v.Unsatisfied)
		v.BM.Save(ls)
	}
}

func (ct *CTest) add(s frames.Signature, t int) {
	if s == nil {
		ct.Satisfied = append(ct.Satisfied, t)
		return
	}
	// if we haven't created a BM for this node yet, do it now
	if ct.BM == nil {
		ct.BM = bytematcher.New()
		ct.BM.SetLowMem()
	}
	ct.Unsatisfied = append(ct.Unsatisfied, t)
	ct.buffer = append(ct.buffer, s)
}

// call for each key after all persists added
func (ct *CTest) commit(p priority.List, prev int) error {
	if ct.buffer == nil {
		return nil
	}
	// don't set priorities if any of the persists are identical
	var dupes bool
	for i, v := range ct.buffer {
		if i == len(ct.buffer)-1 {
			break
		}
		for _, v2 := range ct.buffer[i+1:] {
			if v.Equals(v2) {
				dupes = true
				break
			}
		}
	}
	if dupes {
		_, err := ct.BM.Add(bytematcher.SignatureSet(ct.buffer), nil)
		ct.buffer = nil
		return err
	}
	_, err := ct.BM.Add(bytematcher.SignatureSet(ct.buffer), p.Subset(ct.Unsatisfied[len(ct.Unsatisfied)-len(ct.buffer):], prev))
	ct.buffer = nil
	return err
}
