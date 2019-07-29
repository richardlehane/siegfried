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

package namematcher

// todo: add a precise map[string][]int to take out bulk of globs which are exact names e.g. README

import (
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/internal/siegreader"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/reader"
)

type Matcher struct {
	extensions map[string][]int
	globs      []string // use filepath.Match(glob, name) https://golang.org/pkg/path/filepath/#Match
	globIdx    [][]int
}

func Load(ls *persist.LoadSaver) core.Matcher {
	if !ls.LoadBool() {
		return nil
	}
	le := ls.LoadSmallInt()
	var ext map[string][]int
	if le > 0 {
		ext = make(map[string][]int)
		for i := 0; i < le; i++ {
			k := ls.LoadString()
			r := make([]int, ls.LoadSmallInt())
			for j := range r {
				r[j] = ls.LoadSmallInt()
			}
			ext[k] = r
		}
	}
	globs := ls.LoadStrings()
	globIdx := make([][]int, ls.LoadSmallInt())
	for i := range globIdx {
		globIdx[i] = ls.LoadInts()
	}
	return &Matcher{
		extensions: ext,
		globs:      globs,
		globIdx:    globIdx,
	}
}

func Save(c core.Matcher, ls *persist.LoadSaver) {
	if c == nil {
		ls.SaveBool(false)
		return
	}
	m := c.(*Matcher)
	ls.SaveBool(true)
	ls.SaveSmallInt(len(m.extensions))
	for k, v := range m.extensions {
		ls.SaveString(k)
		ls.SaveSmallInt(len(v))
		for _, w := range v {
			ls.SaveSmallInt(int(w))
		}
	}
	ls.SaveStrings(m.globs)
	ls.SaveSmallInt(len(m.globIdx))
	for _, v := range m.globIdx {
		ls.SaveInts(v)
	}
}

type SignatureSet []string

func Add(c core.Matcher, ss core.SignatureSet, p priority.List) (core.Matcher, int, error) {
	var m *Matcher
	if c == nil {
		m = &Matcher{extensions: make(map[string][]int), globs: []string{}, globIdx: [][]int{}}
	} else {
		m = c.(*Matcher)
	}
	sigs, ok := ss.(SignatureSet)
	if !ok {
		return nil, -1, fmt.Errorf("Namematcher: can't cast persist set")
	}
	var length int
	// unless it is a new matcher, calculate current length by iterating through all the result values
	if len(m.extensions) > 0 || len(m.globs) > 0 {
		for _, v := range m.extensions {
			for _, w := range v {
				if int(w) > length {
					length = int(w)
				}
			}
		}
		for _, v := range m.globIdx {
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

func (m *Matcher) add(s string, fmt int) {
	// handle extension globs first
	if strings.HasPrefix(s, "*.") && strings.LastIndex(s, ".") == 1 {
		ext := strings.ToLower(strings.TrimPrefix(s, "*."))
		if _, ok := m.extensions[ext]; ok {
			m.extensions[ext] = append(m.extensions[ext], fmt)
		} else {
			m.extensions[ext] = []int{fmt}
		}
		return
	}
	for i, v := range m.globs {
		if v == s {
			m.globIdx[i] = append(m.globIdx[i], fmt)
			return
		}
	}
	m.globs = append(m.globs, s)
	m.globIdx = append(m.globIdx, []int{fmt})
}

// normalise returns a path's base name (e.g. README.txt) and extension (e.g. txt)
func normalise(s string) (string, string) {
	// check if this might be a URL (i.e. if source is from a WARC or ARC)
	i := strings.Index(s, "://")
	if i > 0 {
		// backup until hit first non-ASCII alpha char (so we can trim the string to start with scheme)
		for i > 0 {
			i--
			if s[i] < 65 || s[i] > 122 || (s[i] > 90 && s[i] < 97) {
				i++
				break
			}
		}
		u, err := url.Parse(s[i:])
		if err == nil && u.Scheme != "" {
			// make sure it really is a URL
			switch u.Scheme {
			case "http", "https", "ftp", "mailto", "file", "data", "irc":
				// grab the path (trims any trailing query string from the URL)
				s = u.Path
			}
		}
	}
	base := reader.Base(s)
	return base, strings.ToLower(strings.TrimPrefix(filepath.Ext(base), "."))
}

func (m *Matcher) Identify(s string, na *siegreader.Buffer, hints ...core.Hint) (chan core.Result, error) {
	var efmts, gfmts []int
	base, ext := normalise(s)
	var glob string
	if len(s) > 0 {
		efmts = m.extensions[ext]
		for i, g := range m.globs {
			if ok, _ := filepath.Match(g, base); ok {
				glob = g
				gfmts = m.globIdx[i]
				break
			}
		}
	}
	res := make(chan core.Result, len(efmts)+len(gfmts))
	for _, fmt := range efmts {
		res <- result{
			idx:     fmt,
			matches: ext,
		}
	}
	for _, fmt := range gfmts {
		res <- result{
			glob:    true,
			idx:     fmt,
			matches: glob,
		}
	}
	close(res)
	return res, nil
}

func (m *Matcher) String() string {
	var str string
	keys := make([]string, len(m.extensions))
	var i int
	for k := range m.extensions {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, v := range keys {
		str += fmt.Sprintf("%s: %v\n", v, m.extensions[v])
	}
	for i, v := range m.globs {
		str += fmt.Sprintf("%s: %v\n", v, m.globIdx[i])
	}
	return str
}

type result struct {
	glob    bool
	idx     int
	matches string
}

func (r result) Index() int {
	return r.idx
}

func (r result) Basis() string {
	if r.glob {
		return "glob match " + r.matches
	}
	return "extension match " + r.matches
}
