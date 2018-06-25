// Copyright 2016 Richard Lehane. All rights reserved.
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

package riffmatcher

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/image/riff"

	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/internal/siegreader"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

type Matcher struct {
	riffs      map[riff.FourCC][]int
	priorities *priority.Set
}

func Load(ls *persist.LoadSaver) core.Matcher {
	le := ls.LoadSmallInt()
	if le == 0 {
		return nil
	}
	riffs := make(map[riff.FourCC][]int)
	for i := 0; i < le; i++ {
		k := riff.FourCC(ls.LoadFourCC())
		r := make([]int, ls.LoadSmallInt())
		for j := range r {
			r[j] = ls.LoadSmallInt()
		}
		riffs[k] = r
	}
	return &Matcher{
		riffs:      riffs,
		priorities: priority.Load(ls),
	}
}

func Save(c core.Matcher, ls *persist.LoadSaver) {
	if c == nil {
		ls.SaveSmallInt(0)
		return
	}
	m := c.(*Matcher)
	ls.SaveSmallInt(len(m.riffs))
	if len(m.riffs) == 0 {
		return
	}
	for k, v := range m.riffs {
		ls.SaveFourCC(k)
		ls.SaveSmallInt(len(v))
		for _, w := range v {
			ls.SaveSmallInt(w)
		}
	}
	m.priorities.Save(ls)
}

type SignatureSet [][4]byte

func Add(c core.Matcher, ss core.SignatureSet, p priority.List) (core.Matcher, int, error) {
	sigs, ok := ss.(SignatureSet)
	if !ok {
		return nil, -1, fmt.Errorf("RIFFmatcher: can't cast persist set")
	}
	if len(sigs) == 0 {
		return c, 0, nil
	}
	var m *Matcher
	if c == nil {
		m = &Matcher{
			riffs:      make(map[riff.FourCC][]int),
			priorities: &priority.Set{},
		}
	} else {
		m = c.(*Matcher)
	}
	var length int
	// unless it is a new matcher, calculate current length by iterating through all the result values
	if len(m.riffs) > 0 {
		for _, v := range m.riffs {
			for _, w := range v {
				if w > length {
					length = w
				}
			}
		}
		length++ // add one - because the result values are indexes
	}
	for i, v := range sigs {
		cc := riff.FourCC(v)
		_, ok := m.riffs[cc]
		if ok {
			m.riffs[cc] = append(m.riffs[cc], i+length)
		} else {
			m.riffs[cc] = []int{i + length}
		}
	}
	// add priorities
	m.priorities.Add(p, len(sigs), 0, 0)
	return m, length + len(sigs), nil
}

type result struct {
	idx int
	cc  riff.FourCC
}

func (r result) Index() int {
	return r.idx
}

func (r result) Basis() string {
	return "fourCC matches " + string(r.cc[:])
}

func (m Matcher) Identify(na string, b *siegreader.Buffer, hints ...core.Hint) (chan core.Result, error) {
	buf, err := b.Slice(0, 8)
	if err != nil || buf[0] != 'R' || buf[1] != 'I' || buf[2] != 'F' || buf[3] != 'F' {
		res := make(chan core.Result)
		close(res)
		return res, nil
	}
	rcc, rrdr, err := riff.NewReader(siegreader.ReaderFrom(b))
	if err != nil {
		res := make(chan core.Result)
		close(res)
		return res, nil
	}
	// now make structures for testing
	uniqs := make(map[riff.FourCC]bool)
	res := make(chan core.Result)
	waitset := m.priorities.WaitSet(hints...)
	// send and report if satisified
	send := func(cc riff.FourCC) bool {
		if config.Debug() {
			fmt.Fprintf(config.Out(), "riff match %s\n", string(cc[:]))
		}
		if uniqs[cc] {
			return false
		}
		uniqs[cc] = true
		for _, hit := range m.riffs[cc] {
			if waitset.Check(hit) {
				if config.Debug() {
					fmt.Fprintf(config.Out(), "sending riff match %s\n", string(cc[:]))
				}
				res <- result{hit, cc}
				if waitset.Put(hit) {
					return true
				}
			}
		}
		return false
	}
	// riff walk
	var descend func(*riff.Reader) bool
	descend = func(r *riff.Reader) bool {
		for {
			chunkID, chunkLen, chunkData, err := r.Next()
			if err != nil || send(chunkID) {
				return true
			}
			if chunkID == riff.LIST {
				listType, list, err := riff.NewListReader(chunkLen, chunkData)
				if err != nil || send(listType) {
					return true
				}
				if descend(list) {
					return true
				}
			}
		}
	}
	// go time
	go func() {
		if send(rcc) {
			close(res)
			return
		}
		descend(rrdr)
		close(res)
	}()
	return res, nil
}

func (m Matcher) String() string {
	keys := make([]string, 0, len(m.riffs))
	for k := range m.riffs {
		keys = append(keys, string(k[:]))
	}
	sort.Strings(keys)
	return fmt.Sprintf("RIFF matcher: %s\n", strings.Join(keys, ", "))
}
