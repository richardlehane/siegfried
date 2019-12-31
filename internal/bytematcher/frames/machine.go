// Copyright 2019 Richard Lehane. All rights reserved.
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

import (
	"github.com/richardlehane/siegfried/internal/bytematcher/patterns"
	"github.com/richardlehane/siegfried/internal/persist"
)

func init() {
	patterns.Register(machineLoader, loadMachine)
	patterns.Register(blockLoader, loadBlock)
}

const (
	machineLoader byte = iota + 12 // mimeinfo patterns start at 16
	blockLoader
)

func machinify(seg Signature) Signature {
	seg = Blockify(seg)
	switch seg.Characterise() {
	case BOFZero, BOFWindow, BOFWild:
		return Signature{NewFrame(BOF, Machine(seg), 0, 0)}
	case EOFZero, EOFWindow, EOFWild:
		return Signature{NewFrame(EOF, Machine(seg), 0, 0)}
	default: //todo handle Prev and Succ wild
	}
	return seg
}

// A Machine is a segment of a signature that implements the patterns interface
type Machine Signature

func (m Machine) Test(b []byte) ([]int, int) {
	var iter int
	offs := make([]int, len(m))
	for {
		if iter < 0 {
			return nil, 1
		}
		if offs[iter] >= len(b) {
			iter--
			continue
		}
		length, adv := m[iter].MatchN(b[offs[iter]:], 0)
		if length < 0 {
			iter--
			continue
		}
		// success!
		if iter == len(offs)-1 {
			offs[iter] += length
			break
		}
		offs[iter+1] = offs[iter] + length
		offs[iter] += adv
		iter++
	}
	return []int{offs[iter]}, 1
}

func (m Machine) TestR(b []byte) ([]int, int) {
	iter := len(m) - 1
	offs := make([]int, len(m))
	for {
		if iter >= len(m) {
			return nil, 0
		}
		if offs[iter] >= len(b) {
			iter++
			continue
		}
		length, adv := m[iter].MatchNR(b[:len(b)-offs[iter]], 0)
		if length < 0 {
			iter++
			continue
		}
		// success!
		if iter == 0 {
			offs[iter] += length
			break
		}
		offs[iter-1] = offs[iter] + length
		offs[iter] += adv
		iter--
	}
	return []int{offs[iter]}, 1
}

func (m Machine) Equals(pat patterns.Pattern) bool {
	m2, ok := pat.(Machine)
	if !ok || len(m) != len(m2) {
		return false
	}
	for i, f := range m {
		if !f.Equals(m2[i]) {
			return false
		}
	}
	return true
}

func (m Machine) Length() (int, int) {
	var min, max int
	for _, f := range m {
		pmin, pmax := f.Length()
		min += f.Min
		min += pmin
		max += f.Max
		max += pmax
	}
	return min, max
}

// Machines are used where sequence matching inefficient
func (m Machine) NumSequences() int              { return 0 }
func (m Machine) Sequences() []patterns.Sequence { return nil }

func (m Machine) String() string {
	var str string
	for i, v := range m {
		if i > 0 {
			str += " | "
		}
		str += v.String()
	}
	return "m {" + str + "}"
}

func (m Machine) Save(ls *persist.LoadSaver) {
	ls.SaveByte(machineLoader)
	ls.SaveSmallInt(len(m))
	for _, f := range m {
		f.Save(ls)
	}
}

func loadMachine(ls *persist.LoadSaver) patterns.Pattern {
	m := make(Machine, ls.LoadSmallInt())
	for i := range m {
		m[i] = Load(ls)
	}
	return m
}
