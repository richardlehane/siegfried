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

// Package frames describes the Frame interface.
// A set of standard frames are also defined in this package. These are: Fixed, Window, Wild and WildMin.
package frames

import (
	"encoding/gob"
	"strconv"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns"
)

func init() {
	gob.Register(Fixed{})
	gob.Register(Window{})
	gob.Register(Wild{})
	gob.Register(WildMin{})
}

// Frame encapsulates a pattern with offset information, mediating between the pattern and the bytestream.
type Frame interface {
	Match([]byte) (bool, []int)  // Match the byte sequence in a L-R direction. Return a boolean to indicate success. If true, return an offset for where a successive match by a related frame should begin.
	MatchR([]byte) (bool, []int) // Match the byte seqeuence in a reverse (R-L) direction. Return a boolean to indicate success. If true, return an offset for where a successive match by a related frame should begin.
	Equals(Frame) bool
	String() string
	Linked(Frame, int, int) bool // Is a frame linked to a preceding frame (by a preceding or succeding relationship) with an offset and range that is less than the supplied ints?
	Min() int                    // minimum offset
	Max() int                    // maximum offset. Return -1 for no limit (wildcard, *)
	Pat() patterns.Pattern

	// The following methods are inherited from the enclosed OffType
	Orientation() OffType
	SwitchOff() OffType

	// The following methods are inherited from the enclosed pattern
	Length() (int, int) // min and max lengths of the enclosed pattern
	NumSequences() int  // // the number of simple sequences that the enclosed pattern can be represented by. Return 0 if the pattern cannot be represented by a defined number of simple sequence (e.g. for an indirect offset pattern) or, if in your opinion, the number of sequences is unreasonably large.
	Sequences() []patterns.Sequence
}

type OffType int

const (
	BOF  OffType = iota // beginning of file offset
	PREV                // offset from previous frame
	SUCC                // offset from successive frame
	EOF                 // end of file offset
)

var OffString = [...]string{"B", "P", "S", "E"}

// Orientation returns the offset type of the frame which must be either BOF, PREV, SUCC or EOF
func (o OffType) Orientation() OffType {
	return o
}

// Switchoff returns a new offset type according to a given set of rules. These are:
// 	- PREV -> SUCC
// 	- SUCC and EOF -> PREV
// This is helpful when changing the orientation of a frame (for example to allow right-left searching)
func (o OffType) SwitchOff() OffType {
	switch o {
	case PREV:
		return SUCC
	case SUCC, EOF:
		return PREV
	default:
		return o
	}
}

// Generates Fixed, Window, Wild and WildMin frames. The offsets argument controls what type of frame is created:
// 	- for a Wild frame, give no offsets or give a max offset of < 0 and a min of < 1
// 	- for a WildMin frame, give one offset, or give a max offset of < 0 and a min of > 0
// 	- for a Fixed frame, give two offsets that are both >= 0 and that are equal to each other
// 	- for a Window frame, give two offsets that are both >= 0 and that are not equal to each other.
func NewFrame(typ OffType, pat patterns.Pattern, offsets ...int) Frame {
	switch len(offsets) {
	case 0:
		return Frame(Wild{typ, pat})
	case 1:
		if offsets[0] > 0 {
			return Frame(WildMin{typ, offsets[0], pat})
		} else {
			return Frame(Wild{typ, pat})
		}
	}
	if offsets[1] < 0 {
		if offsets[0] > 0 {
			return Frame(WildMin{typ, offsets[0], pat})
		} else {
			return Frame(Wild{typ, pat})
		}
	}
	if offsets[0] < 0 {
		offsets[0] = 0
	}
	if offsets[0] == offsets[1] {
		return Frame(Fixed{typ, offsets[0], pat})
	}
	return Frame(Window{typ, offsets[0], offsets[1], pat})
}

// SwitchFrame returns a new frame with a different orientation (for example to allow right-left searching)
func SwitchFrame(f Frame, p patterns.Pattern) Frame {
	return NewFrame(f.SwitchOff(), p, f.Min(), f.Max())
}

// BMHConvert converts the patterns within a slice of frames to BMH sequences if possible
func BMHConvert(fs []Frame, rev bool) []Frame {
	nfs := make([]Frame, len(fs))
	for i, f := range fs {
		nfs[i] = NewFrame(f.Orientation(), patterns.BMH(f.Pat(), rev), f.Min(), f.Max())
	}
	return nfs
}

// NonZero checks whether, when converted to simple byte sequences, this frame's pattern is all 0 bytes.
func NonZero(f Frame) bool {
	for _, seq := range f.Sequences() {
		allzeros := true
		for _, b := range seq {
			if b != 0 {
				allzeros = false
			}
		}
		if allzeros {
			return false
		}
	}
	return true
}

// Total length is sum of the maximum length of the enclosed pattern and the maximum offset.
func TotalLength(f Frame) int {
	_, l := f.Length()
	return l + f.Max()
}

// Fixed frames are at a fixed offset e.g. 0 or 10 from the BOF, EOF or a preceding or succeeding frame.
type Fixed struct {
	OffType
	Off int
	patterns.Pattern
}

func (f Fixed) Match(b []byte) (bool, []int) {
	if f.Off >= len(b) {
		return false, nil
	}
	if success, length := f.Test(b[f.Off:]); success {
		return true, []int{f.Off + length}
	}
	return false, nil
}

func (f Fixed) MatchR(b []byte) (bool, []int) {
	if f.Off >= len(b) {
		return false, nil
	}
	if success, length := f.TestR(b[:len(b)-f.Off]); success {
		return true, []int{f.Off + length}
	}
	return false, nil
}

func (f Fixed) Equals(frame Frame) bool {
	f1, ok := frame.(Fixed)
	if ok {
		if f.OffType == f1.OffType && f.Off == f1.Off {
			return f.Pattern.Equals(f1.Pattern)
		}
	}
	return false
}

func (f Fixed) String() string {
	return "F " + OffString[f.OffType] + ":" + strconv.Itoa(f.Off) + " " + f.Pattern.String()
}

func (f Fixed) Min() int {
	return f.Off
}

func (f Fixed) Max() int {
	return f.Off
}

func (f Fixed) Linked(prev Frame, maxDistance, maxRange int) bool {
	switch f.OffType {
	case PREV:
		if f.Off > maxDistance {
			return false
		}
		return true
	case SUCC, EOF:
		if prev.Orientation() != SUCC || prev.Max() < 0 || prev.Max() > maxDistance {
			return false
		}
		return true
	default:
		return false
	}
}

func (f Fixed) Pat() patterns.Pattern {
	return f.Pattern
}

// Window frames are at a range of offsets e.g. e.g. 1-1500 from the BOF, EOF or a preceding or succeeding frame.
type Window struct {
	OffType
	MinOff int
	MaxOff int
	patterns.Pattern
}

func (w Window) Match(b []byte) (bool, []int) {
	ret := make([]int, 0, 1)
	min, max := w.MinOff, w.MaxOff
	_, m := w.Length()
	max += m
	if max > len(b) {
		max = len(b)
	}
	for min < max {
		success, length := w.Test(b[min:max])
		if success {
			ret = append(ret, min+length)
			min++
		} else {
			if length == 0 {
				break
			}
			min += length
		}
	}
	if len(ret) > 0 {
		return true, ret
	}
	return false, nil
}

func (w Window) MatchR(b []byte) (bool, []int) {
	ret := make([]int, 0, 1)
	min, max := w.MinOff, w.MaxOff
	_, m := w.Length()
	max += m
	if max > len(b) {
		max = len(b)
	}
	for min < max {
		success, length := w.TestR(b[len(b)-max : len(b)-min])
		if success {
			ret = append(ret, min+length)
			min++
		} else {
			if length == 0 {
				break
			}
			min += length
		}
	}
	if len(ret) > 0 {
		return true, ret
	}
	return false, nil
}

func (w Window) Equals(frame Frame) bool {
	w1, ok := frame.(Window)
	if ok {
		if w.OffType == w1.OffType && w.MinOff == w1.MinOff && w.MaxOff == w1.MaxOff {
			return w.Pattern.Equals(w1.Pattern)
		}
	}
	return false
}

func (w Window) String() string {
	return "WW " + OffString[w.OffType] + ":" + strconv.Itoa(w.MinOff) + "-" + strconv.Itoa(w.MaxOff) + " " + w.Pattern.String()
}

func (w Window) Min() int {
	return w.MinOff
}

func (w Window) Max() int {
	return w.MaxOff
}

func (w Window) Linked(prev Frame, maxDistance, maxRange int) bool {
	switch w.OffType {
	case PREV:
		if w.MaxOff > maxDistance || w.MaxOff-w.MinOff > maxRange {
			return false
		}
		return true
	case SUCC, EOF:
		if prev.Orientation() != SUCC || prev.Max() < 0 || prev.Max() > maxDistance || prev.Max()-prev.Min() > maxRange {
			return false
		}
		return true
	default:
		return false
	}
}

func (w Window) Pat() patterns.Pattern {
	return w.Pattern
}

// Wild frames can be at any offset (i.e. 0 to the end of the file) relative to the BOF, EOF or a preceding or succeeding frame.
type Wild struct {
	OffType
	patterns.Pattern
}

func (w Wild) Match(b []byte) (bool, []int) {
	ret := make([]int, 0, 1)
	min, max := 0, len(b)
	for min < max {
		success, length := w.Test(b[min:])
		if success {
			ret = append(ret, min+length)
			min++
		} else {
			if length == 0 {
				break
			}
			min += length
		}
	}
	if len(ret) > 0 {
		return true, ret
	}
	return false, nil
}

func (w Wild) MatchR(b []byte) (bool, []int) {
	ret := make([]int, 0, 1)
	min, max := 0, len(b)
	for min < max {
		success, length := w.TestR(b[:len(b)-min])
		if success {
			ret = append(ret, min+length)
			min++
		} else {
			if length == 0 {
				break
			}
			min += length
		}
	}
	if len(ret) > 0 {
		return true, ret
	}
	return false, nil
}

func (w Wild) Equals(frame Frame) bool {
	w1, ok := frame.(Wild)
	if ok {
		if w.OffType == w1.OffType {
			return w.Pattern.Equals(w1.Pattern)
		}
	}
	return false
}

func (w Wild) String() string {
	return "WL " + OffString[w.OffType] + " " + w.Pattern.String()
}

func (w Wild) Min() int {
	return 0
}

func (w Wild) Max() int {
	return -1
}

func (w Wild) Linked(prev Frame, maxDistance, maxRange int) bool {
	switch w.OffType {
	case SUCC, EOF:
		if prev.Orientation() != SUCC || prev.Max() < 0 || prev.Max() > maxDistance || prev.Max()-prev.Min() > maxRange {
			return false
		}
		return true
	default:
		return false
	}
}

func (w Wild) Pat() patterns.Pattern {
	return w.Pattern
}

// WildMin frames have a minimum but no maximum offset (e.g. 200-*) relative to the BOF, EOF or a preceding or succeeding frame.
type WildMin struct {
	OffType
	MinOff int
	patterns.Pattern
}

func (w WildMin) Match(b []byte) (bool, []int) {
	ret := make([]int, 0, 1)
	min, max := w.MinOff, len(b)
	for min < max {
		success, length := w.Test(b[min:])
		if success {
			ret = append(ret, min+length)
			min++
		} else {
			if length == 0 {
				break
			}
			min += length
		}
	}
	if len(ret) > 0 {
		return true, ret
	}
	return false, nil
}

func (w WildMin) MatchR(b []byte) (bool, []int) {
	ret := make([]int, 0, 1)
	min, max := w.MinOff, len(b)
	for min < max {
		success, length := w.TestR(b[:len(b)-min])
		if success {
			ret = append(ret, min+length)
			min++
		} else {
			if length == 0 {
				break
			}
			min += length
		}
	}
	if len(ret) > 0 {
		return true, ret
	}
	return false, nil
}

func (w WildMin) Equals(frame Frame) bool {
	w1, ok := frame.(WildMin)
	if ok {
		if w.OffType == w1.OffType && w.MinOff == w1.MinOff {
			return w.Pattern.Equals(w1.Pattern)
		}
	}
	return false
}

func (w WildMin) String() string {
	return "WM " + OffString[w.OffType] + ":" + strconv.Itoa(w.MinOff) + " " + w.Pattern.String()
}

func (w WildMin) Min() int {
	return w.MinOff
}

func (w WildMin) Max() int {
	return -1
}

func (w WildMin) Linked(prev Frame, maxDistance, maxRange int) bool {
	switch w.OffType {
	case SUCC, EOF:
		if prev.Orientation() != SUCC || prev.Max() < 0 || prev.Max() > maxDistance || prev.Max()-prev.Min() > maxRange {
			return false
		}
		return true
	default:
		return false
	}
}

func (w WildMin) Pat() patterns.Pattern {
	return w.Pattern
}
