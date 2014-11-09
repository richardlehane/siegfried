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
	"strconv"
)

// Helper func to turn patterns into BMH sequences if possible
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

// BMH Sequence is an optimised version of the regular Sequence pattern.
// It is used behind the scenes in the Bytematcher package to speed up matching and should not be used directly in other packages (use the plain Sequence instead).
type BMHSequence struct {
	Seq   Sequence
	Shift [256]int
}

// Create a standard BMH sequence
func NewBMHSequence(s Sequence) *BMHSequence {
	var shift [256]int
	for i, _ := range shift {
		shift[i] = len(s)
	}
	last := len(s) - 1
	for i := 0; i < last; i++ {
		shift[s[i]] = last - i
	}
	return &BMHSequence{s, shift}
}

func (s *BMHSequence) Test(b []byte) (bool, int) {
	if len(b) < len(s.Seq) {
		return false, 0
	}
	for i := len(s.Seq) - 1; i > -1; i-- {
		if b[i] != s.Seq[i] {
			return false, s.Shift[b[len(s.Seq)-1]]
		}
	}
	return true, len(s.Seq)
}

func (s *BMHSequence) TestR(b []byte) (bool, int) {
	if len(b) < len(s.Seq) {
		return false, 0
	}
	if bytes.Equal(s.Seq, b[len(b)-len(s.Seq):]) {
		return true, len(s.Seq)
	}
	return false, 1
}

func (s *BMHSequence) Equals(pat Pattern) bool {
	seq2, ok := pat.(*BMHSequence)
	if ok {
		return bytes.Equal(s.Seq, seq2.Seq)
	}
	return false
}

func (s *BMHSequence) Length() (int, int) {
	return len(s.Seq), len(s.Seq)
}

func (s *BMHSequence) NumSequences() int {
	return 1
}

func (s *BMHSequence) Sequences() []Sequence {
	return []Sequence{s.Seq}
}

func (s *BMHSequence) ValidBytes(i int) []byte {
	if i < len(s.Seq) {
		return []byte{s.Seq[i]}
	}
	return []byte{}
}

func (s *BMHSequence) String() string {
	return "seq" + strconv.Itoa(len(s.Seq))
}

// RBMH Sequence is a variant of the BMH sequence designed for reverse (R-L) matching.
// It is used behind the scenes in the Bytematcher package to speed up matching and should not be used directly in other packages (use the plain Sequence instead).
type RBMHSequence struct {
	Seq   Sequence
	Shift [256]int
}

// Create a reverse matching BMH sequence (apply the BMH optimisation to TestR rather than Test)
func NewRBMHSequence(s Sequence) *RBMHSequence {
	var shift [256]int
	for i, _ := range shift {
		shift[i] = len(s)
	}
	last := len(s) - 1
	for i := 0; i < last; i++ {
		shift[s[last-i]] = last - i
	}
	return &RBMHSequence{s, shift}
}

func (s *RBMHSequence) Test(b []byte) (bool, int) {
	if len(b) < len(s.Seq) {
		return false, 0
	}
	if bytes.Equal(s.Seq, b[:len(s.Seq)]) {
		return true, len(s.Seq)
	}
	return false, 1
}

func (s *RBMHSequence) TestR(b []byte) (bool, int) {
	if len(b) < len(s.Seq) {
		return false, 0
	}
	for i, v := range b[len(b)-len(s.Seq):] {
		if v != s.Seq[i] {
			return false, s.Shift[b[len(b)-len(s.Seq)]]
		}
	}
	return true, len(s.Seq)
}

func (s *RBMHSequence) Equals(pat Pattern) bool {
	seq2, ok := pat.(*RBMHSequence)
	if ok {
		return bytes.Equal(s.Seq, seq2.Seq)
	}
	return false
}

func (s *RBMHSequence) Length() (int, int) {
	return len(s.Seq), len(s.Seq)
}

func (s *RBMHSequence) NumSequences() int {
	return 1
}

func (s *RBMHSequence) Sequences() []Sequence {
	return []Sequence{s.Seq}
}

func (s *RBMHSequence) ValidBytes(i int) []byte {
	if i < len(s.Seq) {
		return []byte{s.Seq[i]}
	}
	return []byte{}
}

func (s *RBMHSequence) String() string {
	return "seq" + strconv.Itoa(len(s.Seq))
}
