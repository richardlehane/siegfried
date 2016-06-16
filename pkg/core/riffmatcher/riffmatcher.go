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

	"golang.org/x/image/riff"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/persist"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type Matcher map[riff.FourCC][]int

func Load(ls *persist.LoadSaver) core.Matcher {
	le := ls.LoadSmallInt()
	if le == 0 {
		return nil
	}
	ret := make(Matcher)
	for i := 0; i < le; i++ {
		k := riff.FourCC(ls.LoadFourCC())
		r := make([]int, ls.LoadSmallInt())
		for j := range r {
			r[j] = ls.LoadSmallInt()
		}
		ret[k] = r
	}
	return ret
}

func Save(c core.Matcher, ls *persist.LoadSaver) {
	if c == nil {
		ls.SaveSmallInt(0)
		return
	}
	m := c.(Matcher)
	ls.SaveSmallInt(len(m))
	for k, v := range m {
		ls.SaveFourCC(k)
		ls.SaveSmallInt(len(v))
		for _, w := range v {
			ls.SaveSmallInt(w)
		}
	}
}

type SignatureSet [][4]byte

func Add(c core.Matcher, ss core.SignatureSet, p priority.List) (core.Matcher, int, error) {
	var m Matcher
	if c == nil {
		m = make(Matcher)
	} else {
		m = c.(Matcher)
	}
	sigs, ok := ss.(SignatureSet)
	if !ok {
		return nil, -1, fmt.Errorf("RIFFmatcher: can't cast persist set")
	}
	var length int
	// unless it is a new matcher, calculate current length by iterating through all the result values
	if len(m) > 0 {
		for _, v := range m {
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
		_, ok := m[cc]
		if ok {
			m[cc] = append(m[cc], i+length)
		} else {
			m[cc] = []int{i + length}
		}
	}
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
	return "FourCC matches " + string(r.cc[:])
}

func (m Matcher) Identify(na string, b *siegreader.Buffer, exclude ...int) (chan core.Result, error) {
	buf, err := b.Slice(0, 8)
	if err != nil || buf[0] != 'R' || buf[1] != 'I' || buf[2] != 'F' || buf[3] != 'F' {
		res := make(chan core.Result)
		close(res)
		return res, nil
	}
	cc, rrdr, err := riff.NewReader(siegreader.ReaderFrom(b))
	if err != nil {
		res := make(chan core.Result)
		close(res)
		return res, nil
	}
	uniqs := make(map[riff.FourCC]struct{})
	uniqs[cc] = struct{}{}
	descend(rrdr, uniqs)
	var l int
	for k, _ := range uniqs {
		l += len(m[k])
	}
	res := make(chan core.Result, l)
	for k, _ := range uniqs {
		for _, h := range m[k] {
			res <- result{h, k}
		}
	}
	close(res)
	return res, nil
}

func descend(r *riff.Reader, uniqs map[riff.FourCC]struct{}) {
	for {
		chunkID, chunkLen, chunkData, err := r.Next()
		if err != nil {
			return
		}
		if chunkID == riff.LIST {
			listType, list, err := riff.NewListReader(chunkLen, chunkData)
			if err != nil {
				return
			}
			uniqs[listType] = struct{}{}
			descend(list, uniqs)
			continue
		}
		uniqs[chunkID] = struct{}{}
	}
}

func (m Matcher) String() string {
	return "RIFF matcher"
}
