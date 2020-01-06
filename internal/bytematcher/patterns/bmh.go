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

package patterns

import (
	"bytes"

	"github.com/richardlehane/siegfried/internal/persist"
)

// BMH turns patterns into BMH sequences if possible.
func BMH(p Pattern, rev bool) Pattern {
	s, ok := p.(Sequence)
	if !ok {
		return p
	}
	if rev {
		return NewRBMHSequence(s)
	}
	return NewBMHSequence(s)
}

// BMHSequence is an optimised version of the regular Sequence pattern.
// It is used behind the scenes in the Bytematcher package to speed up matching and should not be used directly in other packages (use the plain Sequence instead).
type BMHSequence struct {
	Seq     Sequence
	advance int
	Shift   [256]int
}

// NewBMHSequence turns a Sequence into a BMHSequence.
func NewBMHSequence(s Sequence) *BMHSequence {
	var shift [256]int
	for i := range shift {
		shift[i] = len(s)
	}
	last := len(s) - 1
	for i := 0; i < last; i++ {
		shift[s[i]] = last - i
	}
	return &BMHSequence{s, Overlap(s), shift}
}

// Test bytes against the pattern.
func (s *BMHSequence) Test(b []byte) ([]int, int) {
	if len(b) < len(s.Seq) {
		return nil, 0
	}
	for i := len(s.Seq) - 1; i > -1; i-- {
		if b[i] != s.Seq[i] {
			return nil, s.Shift[b[len(s.Seq)-1]]
		}
	}
	return []int{len(s.Seq)}, s.advance
}

// Test bytes against the pattern in reverse.
func (s *BMHSequence) TestR(b []byte) ([]int, int) {
	if len(b) < len(s.Seq) {
		return nil, 0
	}
	if bytes.Equal(s.Seq, b[len(b)-len(s.Seq):]) {
		return []int{len(s.Seq)}, s.advance
	}
	return nil, 1
}

// Equals reports whether a pattern is identical to another pattern.
func (s *BMHSequence) Equals(pat Pattern) bool {
	seq2, ok := pat.(*BMHSequence)
	if ok {
		return bytes.Equal(s.Seq, seq2.Seq)
	}
	return false
}

// Length returns a minimum and maximum length for the pattern.
func (s *BMHSequence) Length() (int, int) {
	return len(s.Seq), len(s.Seq)
}

// NumSequences reports how many plain sequences are needed to represent this pattern.
func (s *BMHSequence) NumSequences() int {
	return 1
}

// Sequences converts the pattern into a slice of plain sequences.
func (s *BMHSequence) Sequences() []Sequence {
	return []Sequence{s.Seq}
}

func (s *BMHSequence) String() string {
	return "seq " + Stringify(s.Seq)
}

// Save persists the pattern.
func (s *BMHSequence) Save(ls *persist.LoadSaver) {
	ls.SaveByte(bmhLoader)
	ls.SaveBytes(s.Seq)
	ls.SaveSmallInt(s.advance)
	for _, v := range s.Shift {
		ls.SaveSmallInt(v)
	}
}

func loadBMH(ls *persist.LoadSaver) Pattern {
	bmh := &BMHSequence{}
	bmh.Seq = Sequence(ls.LoadBytes())
	bmh.advance = ls.LoadSmallInt()
	for i := range bmh.Shift {
		bmh.Shift[i] = ls.LoadSmallInt()
	}
	return bmh
}

// RBMHSequence is a variant of the BMH sequence designed for reverse (R-L) matching.
// It is used behind the scenes in the Bytematcher package to speed up matching and should not be used directly in other packages (use the plain Sequence instead).
type RBMHSequence struct {
	Seq     Sequence
	advance int
	Shift   [256]int
}

// NewRBMHSequence create a reverse matching BMH sequence (apply the BMH optimisation to TestR rather than Test).
func NewRBMHSequence(s Sequence) *RBMHSequence {
	var shift [256]int
	for i := range shift {
		shift[i] = len(s)
	}
	last := len(s) - 1
	for i := 0; i < last; i++ {
		shift[s[last-i]] = last - i
	}
	return &RBMHSequence{s, Overlap(s), shift}
}

// Test bytes against the pattern.
func (s *RBMHSequence) Test(b []byte) ([]int, int) {
	if len(b) < len(s.Seq) {
		return nil, 0
	}
	if bytes.Equal(s.Seq, b[:len(s.Seq)]) {
		return []int{len(s.Seq)}, s.advance
	}
	return nil, 1
}

// Test bytes against the pattern in reverse.
func (s *RBMHSequence) TestR(b []byte) ([]int, int) {
	if len(b) < len(s.Seq) {
		return nil, 0
	}
	for i, v := range b[len(b)-len(s.Seq):] {
		if v != s.Seq[i] {
			return nil, s.Shift[b[len(b)-len(s.Seq)]]
		}
	}
	return []int{len(s.Seq)}, s.advance
}

// Equals reports whether a pattern is identical to another pattern.
func (s *RBMHSequence) Equals(pat Pattern) bool {
	seq2, ok := pat.(*RBMHSequence)
	if ok {
		return bytes.Equal(s.Seq, seq2.Seq)
	}
	return false
}

// Length returns a minimum and maximum length for the pattern.
func (s *RBMHSequence) Length() (int, int) {
	return len(s.Seq), len(s.Seq)
}

// NumSequences reports how many plain sequences are needed to represent this pattern.
func (s *RBMHSequence) NumSequences() int {
	return 1
}

// Sequences converts the pattern into a slice of plain sequences.
func (s *RBMHSequence) Sequences() []Sequence {
	return []Sequence{s.Seq}
}

func (s *RBMHSequence) String() string {
	return "seq " + Stringify(s.Seq)
}

// Save persists the pattern.
func (s *RBMHSequence) Save(ls *persist.LoadSaver) {
	ls.SaveByte(rbmhLoader)
	ls.SaveBytes(s.Seq)
	ls.SaveSmallInt(s.advance)
	for _, v := range s.Shift {
		ls.SaveSmallInt(v)
	}
}

func loadRBMH(ls *persist.LoadSaver) Pattern {
	rbmh := &RBMHSequence{}
	rbmh.Seq = Sequence(ls.LoadBytes())
	rbmh.advance = ls.LoadSmallInt()
	for i := range rbmh.Shift {
		rbmh.Shift[i] = ls.LoadSmallInt()
	}
	return rbmh
}
