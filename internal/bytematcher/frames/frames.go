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
	"strconv"

	"github.com/richardlehane/siegfried/internal/bytematcher/patterns"
	"github.com/richardlehane/siegfried/internal/persist"
)

// Frame encapsulates a pattern with offset information, mediating between the pattern and the bytestream.
type Frame struct {
	Min int
	Max int
	OffType
	patterns.Pattern
}

// OffType is the type of offset
type OffType uint8

// Four offset types are supported
const (
	BOF  OffType = iota // beginning of file offset
	PREV                // offset from previous frame
	SUCC                // offset from successive frame
	EOF                 // end of file offset
)

// OffString is an exported array of strings representing each of the four offset types
var OffString = [...]string{"B", "P", "S", "E"}

// Orientation returns the offset type of the frame which must be either BOF, PREV, SUCC or EOF
func (o OffType) Orientation() OffType {
	return o
}

// SwitchOff returns a new offset type according to a given set of rules. These are:
// 	- PREV -> SUCC
// 	- SUCC and EOF -> PREV
// This is helpful when changing the orientation of a frame (for example to allow right-left searching).
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

// NewFrame generates Fixed, Window, Wild and WildMin frames. The offsets argument controls what type of frame is created:
// 	- for a Wild frame, give no offsets or give a max offset of < 0 and a min of < 1
// 	- for a WildMin frame, give one offset, or give a max offset of < 0 and a min of > 0
// 	- for a Fixed frame, give two offsets that are both >= 0 and that are equal to each other
// 	- for a Window frame, give two offsets that are both >= 0 and that are not equal to each other.
func NewFrame(typ OffType, pat patterns.Pattern, offsets ...int) Frame {
	switch len(offsets) {
	case 0:
		return Frame{0, -1, typ, pat}
	case 1:
		if offsets[0] > 0 {
			return Frame{offsets[0], -1, typ, pat}
		}
		return Frame{0, -1, typ, pat}
	}
	if offsets[1] < 0 {
		if offsets[0] > 0 {
			return Frame{offsets[0], -1, typ, pat}
		}
		return Frame{0, -1, typ, pat}
	}
	if offsets[0] < 0 {
		offsets[0] = 0
	}
	return Frame{offsets[0], offsets[1], typ, pat}
}

// SwitchFrame returns a new frame with a different orientation (for example to allow right-left searching).
func SwitchFrame(f Frame, p patterns.Pattern) Frame {
	return NewFrame(f.SwitchOff(), p, f.Min, f.Max)
}

// BMHConvert converts the patterns within a slice of frames to BMH sequences if possible.
func BMHConvert(fs []Frame, rev bool) []Frame {
	nfs := make([]Frame, len(fs))
	for i, f := range fs {
		nfs[i] = NewFrame(f.Orientation(), patterns.BMH(f.Pattern, rev), f.Min, f.Max)
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

// TotalLength is sum of the maximum length of the enclosed pattern and the maximum offset.
func TotalLength(f Frame) int {
	// a wild frame has no total length
	if f.Max < 0 {
		return -1
	}
	_, l := f.Length()
	return l + f.Max
}

// Match the enclosed pattern against the byte slice in a L-R direction.
// Returns a slice of offsets for where a successive match by a related frame should begin.
func (f Frame) Match(b []byte) []int {
	ret := make([]int, 0, 1)
	min, max := f.Min, f.Max
	if max < 0 || max > len(b) {
		max = len(b)
	}
	for min <= max {
		lengths, adv := f.Test(b[min:])
		for _, l := range lengths {
			ret = append(ret, min+l)
		}
		if adv < 1 {
			break
		}
		min += adv
	}
	return ret
}

// For the nth match (per above), return the offset for successive match by related frame and bytes that can advance to make a successive test by this frame.
func (f Frame) MatchN(b []byte, n int) (int, int) {
	var i int
	min, max := f.Min, f.Max
	if max < 0 || max > len(b) {
		max = len(b)
	}
	for min <= max {
		lengths, adv := f.Test(b[min:])
		for _, l := range lengths {
			if i == n {
				return min + l, min + adv
			}
			i++
		}
		if adv < 1 {
			break
		}
		min += adv
	}
	return -1, 0
}

// Match the enclosed pattern against the byte slice in a reverse (R-L) direction. Returns a slice of offsets for where a successive match by a related frame should begin.
func (f Frame) MatchR(b []byte) []int {
	ret := make([]int, 0, 1)
	min, max := f.Min, f.Max
	if max < 0 || max > len(b) {
		max = len(b)
	}
	for min <= max {
		lengths, adv := f.TestR(b[:len(b)-min])
		for _, l := range lengths {
			ret = append(ret, min+l)
		}
		if adv < 1 {
			break
		}
		min += adv
	}
	return ret
}

// For the nth match (per above), return the offset for successive match by related frame and bytes that can advance to make a successive test by this frame.
func (f Frame) MatchNR(b []byte, n int) (int, int) {
	var i int
	min, max := f.Min, f.Max
	if max < 0 || max > len(b) {
		max = len(b)
	}
	for min <= max {
		lengths, adv := f.TestR(b[:len(b)-min])
		for _, l := range lengths {
			if i == n {
				return min + l, min + adv
			}
			i++
		}
		if adv < 1 {
			break
		}
		min += adv
	}
	return -1, 0
}

func (f Frame) Equals(f1 Frame) bool {
	if f.Min == f1.Min && f.Max == f1.Max && f.OffType == f1.OffType && f.Pattern.Equals(f1.Pattern) {
		return true
	}
	return false
}

func (f Frame) String() string {
	var rng string
	if f.Min == f.Max {
		rng = strconv.Itoa(f.Min)
	} else {
		if f.Max < 0 {
			rng = strconv.Itoa(f.Min) + "..*"
		}
		rng = strconv.Itoa(f.Min) + ".." + strconv.Itoa(f.Max)
	}
	return OffString[f.OffType] + ":" + rng + " " + f.Pattern.String()
}

// MaxMatches returns:
//   - the max number of times a frame can match, given a byte slice of length 'l'
//   - the maximum remaining slice length
//   - the minimum length of a successful pattern match
func (f Frame) MaxMatches(l int) (int, int, int) {
	min, _ := f.Length()
	rem := l - min - f.Min
	if rem < 0 && l >= 0 {
		return 0, 0, 0
	}
	// handle fixed
	if f.Min == f.Max || (l < 0 && f.Max < 0) {
		return 1, rem, min
	}
	var ov int
	if f.OffType <= PREV {
		ov = patterns.Overlap(f.Pattern)
	} else {
		ov = patterns.OverlapR(f.Pattern)
	}
	if f.Max < 0 || (l > 0 && f.Max+min > l) {
		return rem/ov + 1, rem, min
	}
	return (f.Max-f.Min)/ov + 1, rem, min
}

// Linked tests whether a frame is linked to a preceding frame (by a preceding or succeding relationship) with an offset and range that is less than the supplied ints.
// If -1 is given for maxDistance & maxRange, then will check if frame is linked to a preceding frame via a PREV or SUCC relationship.
// If -1 is given for maxDistance, but not maxRange, then will check if frame linked without regard to distance (only range),
// this is useful because if give maxRange of 0 you can tell if it is a fixed relationship
func (f Frame) Linked(prev Frame, maxDistance, maxRange int) (bool, int, int) {
	switch f.OffType {
	case PREV:
		if maxDistance < 0 && f.Max > -1 && (maxRange < 0 || f.Max-f.Min <= maxRange) {
			return true, maxDistance, maxRange
		}
		if f.Max < 0 || f.Max > maxDistance || f.Max-f.Min > maxRange {
			return false, 0, 0
		}
		return true, maxDistance - f.Max, maxRange - (f.Max - f.Min)
	case SUCC, EOF:
		if prev.Orientation() != SUCC || prev.Max < 0 {
			return false, 0, 0
		}
		if maxDistance < 0 && (maxRange < 0 || prev.Max-prev.Min <= maxRange) {
			return true, maxDistance, maxRange
		}
		if prev.Max > maxDistance || prev.Max-prev.Min > maxRange {
			return false, 0, 0
		}
		return true, maxDistance - prev.Max, maxRange - (prev.Max - prev.Min)
	default:
		return false, 0, 0
	}
}

func (f Frame) Save(ls *persist.LoadSaver) {
	ls.SaveInt(f.Min)
	ls.SaveInt(f.Max)
	ls.SaveByte(byte(f.OffType))
	f.Pattern.Save(ls)
}

func Load(ls *persist.LoadSaver) Frame {
	return Frame{
		ls.LoadInt(),
		ls.LoadInt(),
		OffType(ls.LoadByte()),
		patterns.Load(ls),
	}
}
