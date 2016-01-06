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

package textmatcher

import (
	"github.com/richardlehane/characterize"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/persist"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type Matcher int

func Load(ls *persist.LoadSaver) *Matcher {
	m := Matcher(ls.LoadSmallInt())
	return &m
}

func (m *Matcher) Save(ls *persist.LoadSaver) {
	ls.SaveSmallInt(int(*m))
}

func New() *Matcher {
	m := Matcher(0)
	return &m
}

type SignatureSet struct{}

func (m *Matcher) Add(ss core.SignatureSet, p priority.List) (int, error) {
	*m++
	return int(*m), nil
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

func (m *Matcher) Identify(na string, buf *siegreader.Buffer) (chan core.Result, error) {
	if *m > 0 {
		tt := buf.Text()
		if tt != characterize.DATA {
			res := make(chan core.Result, *m)
			for i := 1; i < int(*m)+1; i++ {
				res <- result{
					idx:   i,
					basis: "text match " + tt.String(),
				}
			}
			close(res)
			return res, nil
		}
	}
	res := make(chan core.Result)
	close(res)
	return res, nil
}

func (m *Matcher) String() string {
	return "text matcher"
}
