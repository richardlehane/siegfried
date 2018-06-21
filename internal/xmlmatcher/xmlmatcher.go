// Copyright 2015 Richard Lehane. All rights reserved.
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

package xmlmatcher

import (
	"fmt"

	"github.com/richardlehane/xmldetect"

	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/internal/siegreader"
	"github.com/richardlehane/siegfried/pkg/core"
)

type Matcher map[[2]string][]int

type SignatureSet [][2]string // slice of root, namespace (both optional)

func Load(ls *persist.LoadSaver) core.Matcher {
	le := ls.LoadSmallInt()
	if le == 0 {
		return nil
	}
	ret := make(Matcher)
	for i := 0; i < le; i++ {
		k := [2]string{ls.LoadString(), ls.LoadString()}
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
		ls.SaveString(k[0])
		ls.SaveString(k[1])
		ls.SaveSmallInt(len(v))
		for _, w := range v {
			ls.SaveSmallInt(w)
		}
	}
}

func Add(c core.Matcher, ss core.SignatureSet, p priority.List) (core.Matcher, int, error) {
	var m Matcher
	if c == nil {
		m = make(Matcher)
	} else {
		m = c.(Matcher)
	}
	sigs, ok := ss.(SignatureSet)
	if !ok {
		return nil, -1, fmt.Errorf("Xmlmatcher: can't cast persist set")
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
		_, ok := m[v]
		if ok {
			m[v] = append(m[v], i+length)
		} else {
			m[v] = []int{i + length}
		}
	}
	return m, length + len(sigs), nil
}

func (m Matcher) Identify(s string, b *siegreader.Buffer, hints ...core.Hint) (chan core.Result, error) {
	rdr := siegreader.TextReaderFrom(b)
	_, root, ns, err := xmldetect.Root(rdr)
	if err != nil {
		res := make(chan core.Result)
		close(res)
		return res, nil
	}
	both := m[[2]string{root, ns}]
	var nsonly []int
	var rootonly []int
	if ns != "" {
		nsonly = m[[2]string{"", ns}]
		rootonly = m[[2]string{root, ""}]
	}
	res := make(chan core.Result, len(both)+len(rootonly)+len(nsonly))
	for _, v := range both {
		res <- makeResult(v, root, ns)
	}
	for _, v := range rootonly {
		res <- makeResult(v, root, ns)
	}
	for _, v := range nsonly {
		res <- makeResult(v, root, ns)
	}
	close(res)
	return res, nil
}

func makeResult(idx int, root, ns string) result {
	switch {
	case root == "":
		return result{idx, "xml match with ns " + ns}
	case ns == "":
		return result{idx, "xml match with root " + root}
	}
	return result{idx, fmt.Sprintf("xml match with root %s and ns %s", root, ns)}
}

type result struct {
	idx   int
	basis string
}

func (r result) Index() int {
	return r.idx
}

func (r result) Basis() string {
	return r.basis
}

func (m Matcher) String() string {
	var str string
	for k, v := range m {
		str += fmt.Sprintf("%s %s: %v\n", k[0], k[1], v)
	}
	return str
}
