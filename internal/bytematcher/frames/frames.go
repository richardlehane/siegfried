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
	"errors"
	"strconv"

	"github.com/richardlehane/siegfried/internal/bytematcher/patterns"
	"github.com/richardlehane/siegfried/internal/persist"
)

// Frame encapsulates a pattern with offset information, mediating between the pattern and the bytestream.
type Frame interface {
	Match([]byte) (bool, []int) // Match the enclosed pattern against the byte slice in a L-R direction. Return a boolean to indicate success. If true, return an offset for where a successive match by a related frame should begin.
	MatchN([]byte, int) (bool, int)
	MatchR([]byte) (bool, []int) // Match the enclosed pattern against the byte slice in a reverse (R-L) direction. Return a boolean to indicate success. If true, return an offset for where a successive match by a related frame should begin.
	MatchNR([]byte, int) (bool, int)
	Equals(Frame) bool // Equals tests equality of two frames.
	String() string
	Min() int                                // Min returns the minimum offset a frame can appear at
	Max() int                                // Max returns the maximum offset a frame can appear at. Returns -1 for no limit (wildcard, *)
	MaxMatches(l int) (int, int, int)        // MaxMatches returns the max number of times a frame can match, the maximum remaining slice length, and the minimum length of the pattern, given a byte slice of length 'l'
	Linked(Frame, int, int) (bool, int, int) // Linked tests whether a frame is linked to a preceding frame (by a preceding or succeding relationship) with an offset and range that is less than the supplied ints. Pass -1 as maxDistance to test linkage regardless of distance/range
	Pat() patterns.Pattern                   // Pat exposes the enclosed pattern
	Save(*persist.LoadSaver)                 // Save a frame to a LoadSaver. The first byte written should be the identifying byte provided to Register().

	// The following methods are inherited from the enclosed OffType
	Orientation() OffType
	SwitchOff() OffType

	// The following methods are inherited from the enclosed pattern
	Length() (int, int) // min and max lengths of the enclosed pattern
	NumSequences() int  // // the number of simple sequences that the enclosed pattern can be represented by. Return 0 if the pattern cannot be represented by a defined number of simple sequence (e.g. for an indirect offset pattern) or, if in your opinion, the number of sequences is unreasonably large.
	Sequences() []patterns.Sequence
}

// Loader is any function that accepts a persist.LoadSaver and returns a Frame.
type Loader func(*persist.LoadSaver) Frame

const (
	fixedLoader byte = iota
	windowLoader
	wildLoader
	wildMinLoader
)

var loaders = [8]Loader{loadFixed, loadWindow, loadWild, loadWildMin, nil, nil, nil, nil}

// Register an additional Loader.
// Must provide an integer between 4 and 7 (the first four loaders are taken by the standard four frames).
func Register(id byte, l Loader) {
	loaders[int(id)] = l
}

// Load a frame from a persist.LoadSaver
func Load(ls *persist.LoadSaver) Frame {
	id := ls.LoadByte()
	l := loaders[int(id)]
	if l == nil {
		if ls.Err == nil {
			ls.Err = errors.New("bad frame loader")
		}
		return nil
	}
	return l(ls)
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
		return Frame(Wild{typ, pat})
	case 1:
		if offsets[0] > 0 {
			return Frame(WildMin{typ, offsets[0], pat})
		}
		return Frame(Wild{typ, pat})
	}
	if offsets[1] < 0 {
		if offsets[0] > 0 {
			return Frame(WildMin{typ, offsets[0], pat})
		}
		return Frame(Wild{typ, pat})
	}
	if offsets[0] < 0 {
		offsets[0] = 0
	}
	if offsets[0] == offsets[1] {
		return Frame(Fixed{typ, offsets[0], pat})
	}
	return Frame(Window{typ, offsets[0], offsets[1], pat})
}

// SwitchFrame returns a new frame with a different orientation (for example to allow right-left searching).
func SwitchFrame(f Frame, p patterns.Pattern) Frame {
	return NewFrame(f.SwitchOff(), p, f.Min(), f.Max())
}

// BMHConvert converts the patterns within a slice of frames to BMH sequences if possible.
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

// TotalLength is sum of the maximum length of the enclosed pattern and the maximum offset.
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

// Match the enclosed pattern against the byte slice in a L-R direction. Return a boolean to indicate success. If true, return an offset for where a successive match by a related frame should begin.
func (f Fixed) Match(b []byte) (bool, []int) {
	if m, l := f.MatchN(b, 0); m {
		return true, []int{l}
	}
	return false, nil
}

func (f Fixed) MatchN(b []byte, n int) (bool, int) {
	if n > 0 || f.Off >= len(b) {
		return false, -1
	}
	if success, length := f.Test(b[f.Off:]); success {
		return true, f.Off + length
	}
	return false, -1
}

// MatchR matches the enclosed pattern against the byte slice in a reverse (R-L) direction. Return a boolean to indicate success. If true, return an offset for where a successive match by a related frame should begin.
func (f Fixed) MatchR(b []byte) (bool, []int) {
	if m, l := f.MatchNR(b, 0); m {
		return true, []int{l}
	}
	return false, nil
}

func (f Fixed) MatchNR(b []byte, n int) (bool, int) {
	if n > 0 || f.Off >= len(b) {
		return false, -1
	}
	if success, length := f.TestR(b[:len(b)-f.Off]); success {
		return true, f.Off + length
	}
	return false, -1
}

// Equals tests equality of two frames
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

// Min returns the minimum offset a frame can appear at.
func (f Fixed) Min() int {
	return f.Off
}

// Max returns the maximum offset a frame can appear at. Returns -1 for no limit (wildcard, *).
func (f Fixed) Max() int {
	return f.Off
}

// MaxMatches returns the max number of times a frame can match, given a byte slice of length 'l', and the maximum remaining slice length
func (f Fixed) MaxMatches(l int) (int, int, int) {
	min, _ := f.Length()
	rem := l - min - f.Off
	if rem >= 0 {
		return 1, rem, min
	}
	return 0, 0, 0
}

// Linked tests whether a frame is linked to a preceding frame (by a preceding or succeding relationship) with an offset and range that is less than the supplied ints.
func (f Fixed) Linked(prev Frame, maxDistance, maxRange int) (bool, int, int) {
	switch f.OffType {
	case PREV:
		if maxDistance < 0 {
			return true, maxDistance, maxRange
		}
		if f.Off > maxDistance {
			return false, 0, 0
		}
		return true, maxDistance - f.Off, maxRange
	case SUCC, EOF:
		if prev.Orientation() != SUCC || prev.Max() < 0 {
			return false, 0, 0
		}
		if maxDistance < 0 {
			return true, maxDistance, maxRange
		}
		if prev.Max() > maxDistance || prev.Max()-prev.Min() > maxRange {
			return false, 0, 0
		}
		return true, maxDistance - prev.Max(), maxRange - (prev.Max() - prev.Min())
	default:
		return false, 0, 0
	}
}

// Pat exposes the enclosed pattern.
func (f Fixed) Pat() patterns.Pattern {
	return f.Pattern
}

// Save frame to a LoadSaver.
func (f Fixed) Save(ls *persist.LoadSaver) {
	ls.SaveByte(fixedLoader)
	ls.SaveByte(byte(f.OffType))
	ls.SaveInt(f.Off)
	f.Pattern.Save(ls)
}

func loadFixed(ls *persist.LoadSaver) Frame {
	return Fixed{
		OffType(ls.LoadByte()),
		ls.LoadInt(),
		patterns.Load(ls),
	}
}

// Window frames are at a range of offsets e.g. e.g. 1-1500 from the BOF, EOF or a preceding or succeeding frame.
type Window struct {
	OffType
	MinOff int
	MaxOff int
	patterns.Pattern
}

// Match the enclosed pattern against the byte slice in a L-R direction. Return a boolean to indicate success. If true, return an offset for where a successive match by a related frame should begin.
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
			min++ // TODO: why not += length?? - is that only reliable for fail? Check patterns
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

func (w Window) MatchN(b []byte, n int) (bool, int) {
	var i int
	min, max := w.MinOff, w.MaxOff
	_, m := w.Length()
	max += m
	if max > len(b) {
		max = len(b)
	}
	for min < max {
		success, length := w.Test(b[min:max])
		if success {
			if i == n {
				return true, min + length
			}
			i++
			min++ // TODO: why not += length?? - is that only reliable for fail? Check patterns
		} else {
			if length == 0 {
				break
			}
			min += length
		}
	}
	return false, -1
}

// MatchR matches the enclosed pattern against the byte slice in a reverse (R-L) direction. Return a boolean to indicate success. If true, return an offset for where a successive match by a related frame should begin.
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

func (w Window) MatchNR(b []byte, n int) (bool, int) {
	var i int
	min, max := w.MinOff, w.MaxOff
	_, m := w.Length()
	max += m
	if max > len(b) {
		max = len(b)
	}
	for min < max {
		success, length := w.TestR(b[len(b)-max : len(b)-min])
		if success {
			if i == n {
				return true, min + length
			}
			i++
			min++
		} else {
			if length == 0 {
				break
			}
			min += length
		}
	}
	return false, -1
}

// Equals tests equality of two frames
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

// Min returns the minimum offset a frame can appear at.
func (w Window) Min() int {
	return w.MinOff
}

// Max returns the maximum offset a frame can appear at. Returns -1 for no limit (wildcard, *).
func (w Window) Max() int {
	return w.MaxOff
}

// MaxMatches returns the max number of times a frame can match, given a byte slice of length 'l', and the maximum remaining slice length
// TODO: this is *wrong* because it presumes a pattern can't overlap i.e. in AAAAAA, the string AA can match at 5 positions, not 3
func (w Window) MaxMatches(l int) (int, int, int) {
	min, _ := w.Length()
	rem := l - min - w.MinOff
	if rem < 0 {
		return 0, 0, 0
	}
	if w.MaxOff+min > l {
		return rem/min + 1, rem, min
	}
	return (w.MaxOff + min - w.MinOff) / min, rem, min
}

// Linked tests whether a frame is linked to a preceding frame (by a preceding or succeding relationship) with an offset and range that is less than the supplied ints.
func (w Window) Linked(prev Frame, maxDistance, maxRange int) (bool, int, int) {
	switch w.OffType {
	case PREV:
		if maxDistance < 0 {
			return true, maxDistance, maxRange
		}
		if w.MaxOff > maxDistance || w.MaxOff-w.MinOff > maxRange {
			return false, 0, 0
		}
		return true, maxDistance - w.MaxOff, maxRange - (w.MaxOff - w.MinOff)
	case SUCC, EOF:
		if prev.Orientation() != SUCC || prev.Max() < 0 {
			return false, 0, 0
		}
		if maxDistance < 0 {
			return true, maxDistance, maxRange
		}
		if prev.Max() > maxDistance || prev.Max()-prev.Min() > maxRange {
			return false, 0, 0
		}
		return true, maxDistance - prev.Max(), maxRange - (prev.Max() - prev.Min())
	default:
		return false, 0, 0
	}
}

// Pat exposes the enclosed pattern.
func (w Window) Pat() patterns.Pattern {
	return w.Pattern
}

// Save frame to a LoadSaver.
func (w Window) Save(ls *persist.LoadSaver) {
	ls.SaveByte(windowLoader)
	ls.SaveByte(byte(w.OffType))
	ls.SaveInt(w.MinOff)
	ls.SaveInt(w.MaxOff)
	w.Pattern.Save(ls)
}

func loadWindow(ls *persist.LoadSaver) Frame {
	return Window{
		OffType(ls.LoadByte()),
		ls.LoadInt(),
		ls.LoadInt(),
		patterns.Load(ls),
	}
}

// Wild frames can be at any offset (i.e. 0 to the end of the file) relative to the BOF, EOF or a preceding or succeeding frame.
type Wild struct {
	OffType
	patterns.Pattern
}

// Match the enclosed pattern against the byte slice in a L-R direction. Return a boolean to indicate success. If true, return an offset for where a successive match by a related frame should begin.
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

func (w Wild) MatchN(b []byte, n int) (bool, int) {
	var i int
	min, max := 0, len(b)
	for min < max {
		success, length := w.Test(b[min:])
		if success {
			if i == n {
				return true, min + length
			}
			i++
			min++
		} else {
			if length == 0 {
				break
			}
			min += length
		}
	}
	return false, -1
}

// MatchR matches the enclosed pattern against the byte slice in a reverse (R-L) direction. Return a boolean to indicate success. If true, return an offset for where a successive match by a related frame should begin.
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

func (w Wild) MatchNR(b []byte, n int) (bool, int) {
	var i int
	min, max := 0, len(b)
	for min < max {
		success, length := w.TestR(b[:len(b)-min])
		if success {
			if i == n {
				return true, min + length
			}
			i++
			min++
		} else {
			if length == 0 {
				break
			}
			min += length
		}
	}
	return false, -1
}

// Equals tests equality of two frames.
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

// Min returns the minimum offset a frame can appear at.
func (w Wild) Min() int {
	return 0
}

// Max returns the maximum offset a frame can appear at. Returns -1 for no limit (wildcard, *).
func (w Wild) Max() int {
	return -1
}

// MaxMatches returns the max number of times a frame can match, given a byte slice of length 'l', and the maximum remaining slice length
func (w Wild) MaxMatches(l int) (int, int, int) {
	min, _ := w.Length()
	rem := l - min
	if rem < 0 {
		return 0, 0, 0
	}
	return rem/min + 1, rem, min
}

// Linked tests whether a frame is linked to a preceding frame (by a preceding or succeding relationship) with an offset and range that is less than the supplied ints.
func (w Wild) Linked(prev Frame, maxDistance, maxRange int) (bool, int, int) {
	switch w.OffType {
	case SUCC, EOF:
		if prev.Orientation() != SUCC || prev.Max() < 0 {
			return false, 0, 0
		}
		if maxDistance < 0 {
			return true, maxDistance, maxRange
		}
		if prev.Max() > maxDistance || prev.Max()-prev.Min() > maxRange {
			return false, 0, 0
		}
		return true, maxDistance - prev.Max(), maxRange - (prev.Max() - prev.Min())
	default:
		return false, 0, 0
	}
}

// Pat exposes the enclosed pattern.
func (w Wild) Pat() patterns.Pattern {
	return w.Pattern
}

// Save frame to a LoadSaver.
func (w Wild) Save(ls *persist.LoadSaver) {
	ls.SaveByte(wildLoader)
	ls.SaveByte(byte(w.OffType))
	w.Pattern.Save(ls)
}

func loadWild(ls *persist.LoadSaver) Frame {
	return Wild{
		OffType(ls.LoadByte()),
		patterns.Load(ls),
	}
}

// WildMin frames have a minimum but no maximum offset (e.g. 200-*) relative to the BOF, EOF or a preceding or succeeding frame.
type WildMin struct {
	OffType
	MinOff int
	patterns.Pattern
}

// Match the enclosed pattern against the byte slice in a L-R direction. Return a boolean to indicate success. If true, return an offset for where a successive match by a related frame should begin.
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

func (w WildMin) MatchN(b []byte, n int) (bool, int) {
	var i int
	min, max := w.MinOff, len(b)
	for min < max {
		success, length := w.Test(b[min:])
		if success {
			if i == n {
				return true, min + length
			}
			i++
			min++
		} else {
			if length == 0 {
				break
			}
			min += length
		}
	}
	return false, -1
}

// MatchR matches the enclosed pattern against the byte slice in a reverse (R-L) direction. Return a boolean to indicate success. If true, return an offset for where a successive match by a related frame should begin.
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

func (w WildMin) MatchNR(b []byte, n int) (bool, int) {
	var i int
	min, max := w.MinOff, len(b)
	for min < max {
		success, length := w.TestR(b[:len(b)-min])
		if success {
			if i == n {
				return true, min + length
			}
			i++
			min++
		} else {
			if length == 0 {
				break
			}
			min += length
		}
	}
	return false, -1
}

// Equals tests equality of two frames.
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

// Min returns the minimum offset a frame can appear at.
func (w WildMin) Min() int {
	return w.MinOff
}

// Max returns the maximum offset a frame can appear at. Returns -1 for no limit (wildcard, *).
func (w WildMin) Max() int {
	return -1
}

// MaxMatches returns the max number of times a frame can match, given a byte slice of length 'l', and the maximum remaining slice length
func (w WildMin) MaxMatches(l int) (int, int, int) {
	min, _ := w.Length()
	rem := l - min - w.MinOff
	if rem < 0 {
		return 0, 0, 0
	}
	return rem/min + 1, rem, min
}

// Linked tests whether a frame is linked to a preceding frame (by a preceding or succeding relationship) with an offset and range that is less than the supplied ints.
func (w WildMin) Linked(prev Frame, maxDistance, maxRange int) (bool, int, int) {
	switch w.OffType {
	case SUCC, EOF:
		if prev.Orientation() != SUCC || prev.Max() < 0 {
			return false, 0, 0
		}
		if maxDistance < 0 {
			return true, maxDistance, maxRange
		}
		if prev.Max() > maxDistance || prev.Max()-prev.Min() > maxRange {
			return false, 0, 0
		}
		return true, maxDistance - prev.Max(), maxRange - (prev.Max() - prev.Min())
	default:
		return false, 0, 0
	}
}

// Pat exposes the enclosed pattern.
func (w WildMin) Pat() patterns.Pattern {
	return w.Pattern
}

// Save frame to a LoadSaver.
func (w WildMin) Save(ls *persist.LoadSaver) {
	ls.SaveByte(wildMinLoader)
	ls.SaveByte(byte(w.OffType))
	ls.SaveInt(w.MinOff)
	w.Pattern.Save(ls)
}

func loadWildMin(ls *persist.LoadSaver) Frame {
	return WildMin{
		OffType(ls.LoadByte()),
		ls.LoadInt(),
		patterns.Load(ls),
	}
}
