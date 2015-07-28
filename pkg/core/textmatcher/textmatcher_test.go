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
	"bytes"
	"fmt"
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type unit struct {
	label   string
	rdr     *bytes.Buffer
	expect  string
	results int
}

var suite = []unit{
	{
		label:   "bindata",
		rdr:     bytes.NewBuffer([]byte{0, 1, 50, 255}),
		expect:  "nada",
		results: 0,
	},
	{
		label:   "utf8",
		rdr:     bytes.NewBuffer([]byte("ᚠᛇᚻ᛫ᛒᛦᚦ᛫ᚠᚱᚩᚠ")),
		expect:  "text match UTF-8 Unicode",
		results: 3,
	},
	{
		label:   "ascii",
		rdr:     bytes.NewBuffer([]byte("hello world")),
		expect:  "text match ASCII",
		results: 3,
	},
}

var testMatcher *Matcher

func new(i int) (*Matcher, error) {
	m := New()
	for j := 1; j < i+1; j++ {
		idx, _ := m.Add(SignatureSet{}, nil)
		if idx != j {
			return nil, fmt.Errorf("Error adding signature set, expecting index %d got %d", j, idx)
		}
	}
	return m, nil
}

func TestNewMatcher(t *testing.T) {
	m, err := new(5)
	if err != nil {
		t.Fatal(err)
	}
	if *m != 5 {
		t.Fatalf("Expecting a matcher equalling %d, got %d", 5, *m)
	}
}

func TestSuite(t *testing.T) {
	ids := 3
	m, _ := new(ids)
	bufs := siegreader.New()
	for _, u := range suite {
		buf, _ := bufs.Get(u.rdr)
		res, _ := m.Identify("", buf)
		var i int
		for r := range res {
			i++
			if r.Index() != i || r.Basis() != u.expect {
				t.Fatalf("Expecting result %d for %s, got %d with %s", i, u.label, r.Index(), r.Basis())
			}
		}
		if i != u.results {
			t.Fatalf("Expecting a total of %d results, got %d", u.results, i)
		}
	}
}
