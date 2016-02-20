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
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/persist"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type Matcher map[[2]string][]int

type SignatureSet [][2]string // slice of root, namespace (both optional)

func Load(ls *persist.LoadSaver) Matcher {
	return nil
}

func (m Matcher) Save(ls *persist.LoadSaver) {}

func New() Matcher {
	return make(Matcher)
}

func (m Matcher) Add(ss core.SignatureSet, p priority.List) (int, error) {
	return 0, nil
}

func (m Matcher) Identify(s string, na *siegreader.Buffer) (chan core.Result, error) {
	return nil, nil
}

func (m Matcher) String() string { return "" }
