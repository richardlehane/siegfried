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

package stringmatcher

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/persist"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type Matcher map[string][]result

func Load(ls *persist.LoadSaver) Matcher {
	le := ls.LoadSmallInt()
	if le == 0 {
		return nil
	}
	ret := make(Matcher)
	for i := 0; i < le; i++ {
		k := ls.LoadString()
		r := make([]result, ls.LoadSmallInt())
		for j := range r {
			r[j] = result(ls.LoadSmallInt())
		}
		ret[k] = r
	}
	return ret
}

func (m Matcher) Save(ls *persist.LoadSaver) {
	ls.SaveSmallInt(len(m))
	for k, v := range m {
		ls.SaveString(k)
		ls.SaveSmallInt(len(v))
		for _, w := range v {
			ls.SaveSmallInt(int(w))
		}
	}
}

func New() Matcher {
	return make(Matcher)
}

type SignatureSet [][]string

func (m Matcher) Add(ss core.SignatureSet, p priority.List) (int, error) {
	sigs, ok := ss.(SignatureSet)
	if !ok {
		return -1, fmt.Errorf("Stringmatcher: can't cast persist set")
	}
	var length int
	// unless it is a new matcher, calculate current length by iterating through all the result values
	if len(m) > 0 {
		for _, v := range m {
			for _, w := range v {
				if int(w) > length {
					length = int(w)
				}
			}
		}
		length++ // add one - because the result values are indexes
	}
	for i, v := range sigs {
		for _, w := range v {
			m.add(w, i+length)
		}
	}
	return length + len(sigs), nil
}

func (m Matcher) add(s string, fmt int) {
	_, ok := m[s]
	if ok {
		m[s] = append(m[s], result(fmt))
		return
	}
	m[s] = []result{result(fmt)}
}

func (m Matcher) Identify(s string, na siegreader.Buffer) (chan core.Result, error) {
	if len(s) > 0 {
		if fmts, ok := m[s]; ok {
			res := make(chan core.Result, len(fmts))
			for _, v := range fmts {
				res <- v
			}
			close(res)
			return res, nil
		}
	}
	res := make(chan core.Result)
	close(res)
	return res, nil
}

func (m Matcher) String() string {
	var str string
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, v := range keys {
		str += fmt.Sprintf("%v: %v\n", v, m[v])
	}
	return str
}

type result int

func (r result) Index() int {
	return int(r)
}

func (r result) Basis() string {
	return "string match"
}

// Extension Matcher | MIME Matcher

func NormaliseExt(s string) string {
	ext := filepath.Base(s)
	idx := strings.LastIndex(ext, "?") // to get ext from URL paths, get rid of params
	if idx > -1 && strings.Index(ext[:idx], ".") > -1 {
		ext = ext[:idx]
	}
	return strings.ToLower(strings.TrimPrefix(filepath.Ext(ext), "."))
}

func NormaliseMIME(s string) string {
	idx := strings.LastIndex(s, ";")
	if idx > 0 {
		return s[:idx]
	}
	return s
}

type ExtResult struct{ core.Result }

func (_ ExtResult) Basis() string {
	return "extension match"
}

type MIMEResult struct{ core.Result }

func (_ MIMEResult) Basis() string {
	return "MIME match"
}
