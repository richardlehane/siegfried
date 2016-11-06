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

package frames

import "github.com/richardlehane/siegfried/internal/bytematcher/patterns"

// Sequencer turns sequential frames into a set of plain byte sequences. The set represents possible choices.
type Sequencer func(Frame) [][]byte

// NewSequencer creates a Sequencer (reversed if true).
func NewSequencer(rev bool) Sequencer {
	var ret [][]byte
	return func(f Frame) [][]byte {
		var s []patterns.Sequence
		if rev {
			s = f.Sequences()
			for i := range s {
				s[i] = s[i].Reverse()
			}
		} else {
			s = f.Sequences()
		}
		ret = appendSeq(ret, s)
		return ret
	}
}

func appendSeq(b [][]byte, s []patterns.Sequence) [][]byte {
	var c [][]byte
	if len(b) == 0 {
		c = make([][]byte, len(s))
		for i, seq := range s {
			c[i] = make([]byte, len(seq))
			copy(c[i], []byte(seq))
		}
	} else {
		c = make([][]byte, len(b)*len(s))
		iter := 0
		for _, seq := range s {
			for _, orig := range b {
				c[iter] = make([]byte, len(orig)+len(seq))
				copy(c[iter], orig)
				copy(c[iter][len(orig):], []byte(seq))
				iter++
			}
		}
	}
	return c
}
