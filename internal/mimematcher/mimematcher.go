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

package mimematcher

import (
	"fmt"
	"sort"
	"strings"

	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/internal/siegreader"
	"github.com/richardlehane/siegfried/pkg/core"
)

// Matcher matches provided MIME-types against MIME-types associated with formats.
// This is an extra signal for identification akin to a file extension.
// It is used, for example, for web archive files (WARC) where you have declared
// MIME-types which you might want to verify.
type Matcher map[string][]int

// Load returns a MIMEMatcher
func Load(ls *persist.LoadSaver) core.Matcher {
	le := ls.LoadSmallInt()
	if le == 0 {
		return nil
	}
	ret := make(Matcher)
	for i := 0; i < le; i++ {
		k := ls.LoadString()
		r := make([]int, ls.LoadSmallInt())
		for j := range r {
			r[j] = ls.LoadSmallInt()
		}
		ret[k] = r
	}
	return ret
}

// Save encodes a MIMEMatcher
func Save(c core.Matcher, ls *persist.LoadSaver) {
	if c == nil {
		ls.SaveSmallInt(0)
		return
	}
	m := c.(Matcher)
	ls.SaveSmallInt(len(m))
	for k, v := range m {
		ls.SaveString(k)
		ls.SaveSmallInt(len(v))
		for _, w := range v {
			ls.SaveSmallInt(w)
		}
	}
}

// SignatureSet for a MIMEMatcher is a slice of MIME-types
type SignatureSet []string

// Add adds a set of MIME-type signatures to a MIMEMatcher
func Add(c core.Matcher, ss core.SignatureSet, p priority.List) (core.Matcher, int, error) {
	var m Matcher
	if c == nil {
		m = make(Matcher)
	} else {
		m = c.(Matcher)
	}
	sigs, ok := ss.(SignatureSet)
	if !ok {
		return nil, -1, fmt.Errorf("MIMEmatcher: bad signature set")
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
		m.add(v, i+length)
	}
	return m, length + len(sigs), nil
}

func (m Matcher) add(s string, fmt int) {
	_, ok := m[s]
	if ok {
		m[s] = append(m[s], fmt)
		return
	}
	m[s] = []int{fmt}
}

// Identify tests the supplied MIME-type against the MIMEMatcher. The Buffer is not used.
func (m Matcher) Identify(s string, na *siegreader.Buffer, hints ...core.Hint) (chan core.Result, error) {
	var (
		fmts, tfmts []int
		idx         int
	)
	if len(s) > 0 {
		fmts = m[s]
		idx = strings.LastIndex(s, ";")
		if idx > 0 {
			tfmts = m[s[:idx]]
		}
	}
	res := make(chan core.Result, len(fmts)+len(tfmts))
	for _, v := range fmts {
		res <- Result{
			idx:  v,
			mime: s,
		}
	}
	for _, v := range tfmts {
		res <- Result{
			idx:     v,
			Trimmed: true,
			mime:    s[:idx],
		}
	}
	close(res)
	return res, nil
}

// String representation of a MIMEMatcher
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

// Result reports a MIME-type match. If Trimmed is true, then the supplied MIME-type
// was trimmed of text following a ";" before matching
type Result struct {
	idx     int
	Trimmed bool
	mime    string
}

// Index of the MIME-type match
func (r Result) Index() int {
	return r.idx
}

// Basis for a MIME-type match is always just that the mime matched
func (r Result) Basis() string {
	return "mime match " + r.mime
}
