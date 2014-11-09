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
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type Matcher map[string][]Result

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
		return -1, fmt.Errorf("Extension matcher: can't cast signature set as an EM ss")
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

func (e Matcher) Identify(name string, na *siegreader.Buffer) chan core.Result {
	res := make(chan core.Result, 10)
	go func() {
		ext := filepath.Ext(name)
		if len(ext) > 0 {
			fmts, ok := e[strings.TrimPrefix(ext, ".")]
			if ok {
				for _, v := range fmts {
					res <- v
				}
			}
		}
		close(res)
	}()
	return res
}

func Load(r io.Reader) (core.Matcher, error) {
	e := New()
	dec := gob.NewDecoder(r)
	err := dec.Decode(&e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (e Matcher) Save(w io.Writer) (int, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(e)
	if err != nil {
		return 0, err
	}
	sz := buf.Len()
	_, err = buf.WriteTo(w)
	if err != nil {
		return 0, err
	}
	return sz, nil
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
