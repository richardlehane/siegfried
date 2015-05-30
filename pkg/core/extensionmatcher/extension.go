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

package extensionmatcher

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

type Matcher map[string][]Result

func Load(ls *persist.LoadSaver) Matcher {
	le := ls.LoadSmallInt()
	if le == 0 {
		return nil
	}
	ret := make(Matcher)
	for i := 0; i < le; i++ {
		k := ls.LoadString()
		r := make([]Result, ls.LoadSmallInt())
		for j := range r {
			r[j] = Result(ls.LoadSmallInt())
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

type Result int

func (r Result) Index() int {
	return int(r)
}

func (r Result) Basis() string {
	return "extension match"
}

func (e Matcher) Add(ss core.SignatureSet, p priority.List) (int, error) {
	sigs, ok := ss.(SignatureSet)
	if !ok {
		return -1, fmt.Errorf("Extension matcher: can't cast persist set as an EM ss")
	}
	var length int
	// unless it is a new matcher, calculate current length by iterating through all the Result values
	if len(e) > 0 {
		for _, v := range e {
			for _, w := range v {
				if int(w) > length {
					length = int(w)
				}
			}
		}
		length++ // add one - because the Result values are indexes
	}
	for i, v := range sigs {
		for _, w := range v {
			e.add(w, i+length)
		}
	}
	return length + len(sigs), nil
}

func (e Matcher) add(ext string, fmt int) {
	_, ok := e[ext]
	if ok {
		e[ext] = append(e[ext], Result(fmt))
		return
	}
	e[ext] = []Result{Result(fmt)}
}

func (e Matcher) Identify(name string, na siegreader.Buffer) (chan core.Result, error) {
	res := make(chan core.Result, 10)
	go func() {
		ext := filepath.Ext(name)
		if len(ext) > 0 {
			fmts, ok := e[strings.ToLower(strings.TrimPrefix(ext, "."))]
			if ok {
				for _, v := range fmts {
					res <- v
				}
			}
		}
		close(res)
	}()
	return res, nil
}

func (e Matcher) String() string {
	var str string
	var keys []string
	for k := range e {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, v := range keys {
		str += fmt.Sprintf("%v: %v\n", v, e[v])
	}
	return str
}
