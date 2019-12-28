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

// Signature is just a slice of frames.
type Signature []Frame

func (s Signature) String() string {
	var str string
	for i, v := range s {
		if i > 0 {
			str += " | "
		}
		str += v.String()
	}
	return "(" + str + ")"
}

func (s Signature) OneEnough() bool {
	for _, f := range s {
		if f.Min != f.Max {
			return false
		}
	}
	return true
}

// Equals tests equality of two signatures.
func (s Signature) Equals(s1 Signature) bool {
	if len(s) != len(s1) {
		return false
	}
	for i, v := range s {
		if !v.Equals(s1[i]) {
			return false
		}
	}
	return true
}

// add ints together & if any are -1 (wildcard) then return -1
func addWilds(i ...int) int {
	var j int
	for _, k := range i {
		if k == -1 {
			return -1
		}
		j += k
	}
	return j
}

// return the min and max legal offsets for a frame. If it is a BOF sequence the first variable returned is true.
func (s Signature) position(idx int) (bool, int, int) {
	var min, max int
	f := s[idx]
	if f.Orientation() >= SUCC {
		for j := len(s) - 1; j >= 0; j-- {
			f = s[j]
			if f.Orientation() == EOF {
				min, max = 0, 0
			}
			if j == idx {
				return false, addWilds(min, f.Min), addWilds(max, f.Max)
			}
			minl, maxl := f.Length()
			min, max = addWilds(min, f.Min, minl), addWilds(max, f.Max, maxl)
		}
	}
	for i, f := range s {
		if f.Orientation() == BOF {
			min, max = 0, 0
		}
		if i == idx {
			return true, addWilds(min, f.Min), addWilds(max, f.Max)
		}
		minl, maxl := f.Length()
		min, max = addWilds(min, f.Min, minl), addWilds(max, f.Max, maxl)
	}
	// should not get here
	return false, -1, -1
}

// test whether two positions overlap. Positions are marked by BOF/EOF and min/max ints
func overlap(a bool, amin, amax int, b bool, bmin, bmax int) bool {
	if a != b {
		return false
	}
	if amax > -1 && (amax < bmin || amax < amin) {
		return false
	}
	if bmax > -1 && bmax < amin {
		return false
	}
	return true
}

// Contains tests whether a signature wholly contains the segments of another signature.
func (s Signature) Contains(s1 Signature) bool {
	if len(s1) > len(s) {
		return false
	}
	var numEquals int
	for i, f := range s {
		if idx := patterns.Index(f.Pattern, s1[numEquals].Pattern); idx >= 0 {
			a, amin, amax := s.position(i)
			amin += idx
			b, bmin, bmax := s1.position(numEquals)
			if overlap(a, amin, amax, b, bmin, bmax) {
				numEquals++
				if numEquals == len(s1) {
					break
				}
			}
		}
	}
	return numEquals == len(s1)
}

// turn a wild prev into a succ segment
func (s Signature) reverse(last bool, min int) Signature {
	ret := make(Signature, len(s))
	for i := range s[:len(s)-1] {
		ret[i] = SwitchFrame(s[i+1], s[i].Pattern)
	}
	typ := SUCC
	if last {
		typ = EOF
	}
	ret[len(ret)-1] = NewFrame(typ, s[len(s)-1].Pattern, min)
	return ret
}

// Mirror returns a signature in which wildcard previous segments are turned into wildcard succ/eof segments.
// If no wildcard previous segments are present, nil is returned.
func (s Signature) Mirror() Signature {
	segments := s.segment(-1, -1)
	var hasWild = -1
	for i, v := range segments {
		if v[0].Orientation() < SUCC && v[0].Max == -1 {
			if v[0].Orientation() < PREV && v[0].Min > 0 {
				hasWild = -1 // reset on BOF min wild
			} else {
				if hasWild < 0 {
					hasWild = i // get the first wild segment
				}
			}
		}
	}
	if hasWild < 0 {
		return nil
	}
	ret := make(Signature, 0, len(s))
	for i, v := range segments {
		if i >= hasWild && v[0].Orientation() < SUCC && v[0].Max == -1 {
			var last bool
			var min int
			if i == len(segments)-1 {
				last = true
			} else {
				next := segments[i+1][0]
				if next.Orientation() < SUCC {
					min = next.Min
				}
			}
			ret = append(ret, v.reverse(last, min)...)
		} else {
			ret = append(ret, v...)
		}
	}
	return ret
}
