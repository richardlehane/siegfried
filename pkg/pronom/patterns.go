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

// Define custom patterns (implementing the siegfried.Pattern interface) for the different patterns allowed by the PRONOM spec.
package pronom

import (
	"bytes"
	"encoding/gob"
	"strconv"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns"
)

func init() {
	gob.Register(NotSequence{})
	gob.Register(&Range{})
	gob.Register(&NotRange{})
}

type NotSequence []byte

func (ns NotSequence) Test(b []byte) (bool, int) {
	if len(b) < len(ns) {
		return false, 0
	}
	if bytes.Equal(ns, b[:len(ns)]) {
		return false, 1
	}
	return true, len(ns)
}

func (ns NotSequence) TestR(b []byte) (bool, int) {
	if len(b) < len(ns) {
		return false, 0
	}
	if bytes.Equal(ns, b[len(b)-len(ns):]) {
		return false, 1
	}
	return true, len(ns)
}

func (ns NotSequence) Equals(pat patterns.Pattern) bool {
	ns2, ok := pat.(NotSequence)
	if ok {
		return bytes.Equal(ns, ns2)
	}
	return false
}

func (ns NotSequence) Length() (int, int) {
	return len(ns), len(ns)
}

func (ns NotSequence) NumSequences() int {
	if len(ns) == 1 {
		return 255
	}
	return 0
}

func (ns NotSequence) Sequences() []patterns.Sequence {
	num := ns.NumSequences()
	seqs := make([]patterns.Sequence, num)
	if num < 1 {
		return seqs
	}
	b := int(ns[0])
	for i := 0; i < b; i++ {
		seqs[i] = patterns.Sequence{byte(i)}
	}
	for i := b + 1; i < 256; i++ {
		seqs[i-1] = patterns.Sequence{byte(i)}
	}
	return seqs
}

func (ns NotSequence) ValidBytes(i int) []byte {
	if i < len(ns) {
		b := ns[i]
		ret := make([]byte, 255)
		for j := 0; j < 256; j++ {
			if j < int(b) {
				ret[j] = byte(j)
			} else if j > int(b) {
				ret[j-1] = byte(j)
			}
		}
		return ret
	}
	return []byte{}
}

func (ns NotSequence) String() string {
	return "ns" + strconv.Itoa(len(ns))
}

type Range struct {
	From, To []byte
}

func (r Range) Test(b []byte) (bool, int) {
	if len(b) < len(r.From) || len(b) < len(r.To) {
		return false, 0
	}
	if bytes.Compare(r.From, b[:len(r.From)]) < 1 {
		if bytes.Compare(r.To, b[:len(r.To)]) > -1 {
			return true, len(r.From)
		}
	}
	return false, 1
}

func (r Range) TestR(b []byte) (bool, int) {
	if len(b) < len(r.From) || len(b) < len(r.To) {
		return false, 0
	}
	if bytes.Compare(r.From, b[len(b)-len(r.From):]) < 1 {
		if bytes.Compare(r.To, b[len(b)-len(r.To):]) > -1 {
			return true, len(r.From)
		}
	}
	return false, 1
}

func (r Range) Equals(pat patterns.Pattern) bool {
	rng, ok := pat.(Range)
	if ok {
		if bytes.Equal(rng.From, r.From) {
			if bytes.Equal(rng.To, r.To) {
				return true
			}
		}
	}
	return false
}

func (r Range) Length() (int, int) {
	return len(r.From), len(r.From)
}

func (r Range) NumSequences() int {
	l := len(r.From)
	if l > 2 || l < 1 {
		return 0
	}
	if l == 2 {
		if r.To[0]-r.From[0] > 1 {
			return 0
		}
		return 256*int(r.To[0]-r.From[0]) + int(r.To[1]) - int(r.From[1]) + 1
	}
	return int(r.To[0]-r.From[0]) + 1
}

func (r Range) Sequences() []patterns.Sequence {
	num := r.NumSequences()
	seqs := make([]patterns.Sequence, num)
	if num < 1 {
		return seqs
	}
	if len(r.From) == 2 {
		if r.From[0] == r.To[0] {
			for i := 0; i < num; i++ {
				seqs[i] = patterns.Sequence{r.From[0], r.From[1] + byte(i)}
			}
			return seqs
		}
		max := 256 - int(r.From[1])
		for i := 0; i < max; i++ {
			seqs[i] = patterns.Sequence{r.From[0], r.From[1] + byte(i)}
		}
		for i := 0; max < num; max++ {
			seqs[max] = patterns.Sequence{r.To[0], byte(0 + i)}
			i++
		}
		return seqs
	}
	for i := 0; i < num; i++ {
		seqs[i] = patterns.Sequence{r.From[0] + byte(i)}
	}
	return seqs
}

func (r Range) ValidBytes(i int) []byte {
	var f, t byte
	if i < len(r.From) {
		f = r.From[i]
		t = r.To[i]
		ret := make([]byte, int(t-f+1))
		for i, _ := range ret {
			ret[i] = f + byte(i)
		}
		return ret
	}
	return []byte{}
}

func (r Range) String() string {
	return "r" + strconv.Itoa(len(r.From))
}

// Implemented for completeness, have not seen any of these actually used
type NotRange struct {
	From, To []byte
}

func (nr NotRange) Test(b []byte) (bool, int) {
	if len(b) < len(nr.From) || len(b) < len(nr.To) {
		return false, 0
	}
	if bytes.Compare(nr.From, b) < 1 {
		if bytes.Compare(nr.To, b) > -1 {
			return false, 1
		}
	}
	return true, len(nr.From)
}

func (nr NotRange) TestR(b []byte) (bool, int) {
	if len(b) < len(nr.From) || len(b) < len(nr.To) {
		return false, 0
	}
	if bytes.Compare(nr.From, b[len(b)-len(nr.From):]) < 1 {
		if bytes.Compare(nr.To, b[len(b)-len(nr.To):]) > -1 {
			return false, 1
		}
	}
	return true, len(nr.From)
}

func (nr NotRange) Equals(pat patterns.Pattern) bool {
	not, ok := pat.(NotRange)
	if ok {
		if bytes.Equal(nr.From, not.From) {
			if bytes.Equal(nr.To, not.To) {
				return true
			}
		}
	}
	return false
}

func (n NotRange) Length() (int, int) {
	return len(n.From), len(n.From)
}

func (nr NotRange) NumSequences() int {
	_, l := nr.Length()
	if l != 1 {
		return 0
	}
	return int(nr.From[0] + 255 - nr.To[0])
}

func (nr NotRange) Sequences() []patterns.Sequence {
	num := nr.NumSequences()
	seqs := make([]patterns.Sequence, num)
	if num < 1 {
		return seqs
	}
	j := 1
	for i := 0; i < num; i++ {
		if i < int(nr.From[0]) {
			seqs[i] = patterns.Sequence{byte(i)}
		} else {
			seqs[i] = patterns.Sequence{nr.To[0] + byte(j)}
			j++
		}
	}
	return seqs
}

func (nr NotRange) ValidBytes(i int) []byte {
	var f, t byte
	if i < len(nr.From) {
		f = nr.From[i]
		t = nr.To[i]
		ret := make([]byte, int(255-t+f))
		for i, _ := range ret {
			if i < int(f) {
				ret[i] = byte(i)
			} else {
				ret[i] = t - f + byte(i+1)
			}
		}
		return ret
	}
	return []byte{}
}

func (nr NotRange) String() string {
	return "nr" + strconv.Itoa(len(nr.From))
}
