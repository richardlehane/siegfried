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

// Signatures are just slices of frames
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

// Signatures can be divided into signature segments.
// This separation happens on wildcards or when the distance between frames is deemed too great.
// E.g. a signature of [BOF 0: "ABCD"][PREV 0-20: "EFG"][PREV Wild: "HI"][EOF 0: "XYZ"]
// has three segments:
// 1. [BOF 0: "ABCD"][PREV 0-20: "EFG"]
// 2. [PREV Wild: "HI"]
// 3. [EOF 0: "XYZ"]
// The Distance and Range options control the allowable distance and range between frames
// (i.e. a fixed offset of 5000 distant might be acceptable, where a range of 1-2000 might not be).
func (s Signature) Segment(dist, rng int) []Signature {
	if len(s) <= 1 {
		return []Signature{s}
	}
	segments := make([]Signature, 0, 1)
	segment := Signature{s[0]}
	for i, frame := range s[1:] {
		if frame.Linked(s[i], dist, rng) {
			segment = append(segment, frame)
		} else {
			segments = append(segments, segment)
			segment = Signature{frame}
		}
	}
	return append(segments, segment)
}

// turn a wild prev into a succ segment
func (s Signature) reverse(last bool) Signature {
	ret := make(Signature, len(s))
	for i := range s[:len(s)-1] {
		ret[i] = SwitchFrame(s[i+1], s[i].Pat())
	}
	typ := SUCC
	if last {
		typ = EOF
	}
	ret[len(ret)-1] = NewFrame(typ, s[len(s)-1].Pat())
	return ret
}

// wildcard previous segments are turned into wildcard succ/eof segments
func (s Signature) Mirror() Signature {
	bignum := 1000000
	segments := s.Segment(bignum, bignum)
	var hasWild bool
	for _, v := range segments {
		if v[0].Orientation() < SUCC && v[0].Min() == 0 && v[0].Max() == -1 {
			hasWild = true
		}
	}
	if !hasWild {
		return nil
	}
	ret := make(Signature, 0, len(s))
	for i, v := range segments {
		if v[0].Orientation() < SUCC && v[0].Min() == 0 && v[0].Max() == -1 {
			var last bool
			if i == len(segments)-1 {
				last = true
			}
			ret = append(ret, v.reverse(last)...)
		} else {
			ret = append(ret, v...)
		}
	}
	return ret
}
