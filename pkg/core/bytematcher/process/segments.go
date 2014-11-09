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

package process

import (
	"fmt"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
)

// Signatures are divided into signature segments.
// This separation happens on wildcards or when the distance between frames is deemed too great.
// E.g. a signature of [BOF 0: "ABCD"][PREV 0-20: "EFG"][PREV Wild: "HI"][EOF 0: "XYZ"]
// has three segments:
// 1. [BOF 0: "ABCD"][PREV 0-20: "EFG"]
// 2. [PREV Wild: "HI"]
// 3. [EOF 0: "XYZ"]
// The Distance and Range options control the allowable distance and range between frames
// (i.e. a fixed offset of 5000 distant might be acceptable, where a range of 1-2000 might not be).
func (p *Process) splitSegments(sig frames.Signature) []frames.Signature {
	if len(sig) < 2 {
		return []frames.Signature{sig}
	}
	segments := make([]frames.Signature, 0, 1)
	segment := frames.Signature{sig[0]}
	for i, frame := range sig[1:] {
		if frame.Linked(sig[i], p.Distance, p.Range) {
			segment = append(segment, frame)
		} else {
			segments = append(segments, segment)
			segment = frames.Signature{frame}
		}
	}
	return append(segments, segment)
}

type sigType int

const (
	bofZero   sigType = iota // fixed offset, zero length from BOF
	bofWindow                // offset is a window or fixed value greater than zero from BOF
	bofWild
	prev
	succ
	eofZero
	eofWindow
	eofWild
)

// Simple characterisation of a segment: is it relative to the BOF, or the EOF, or is it a prev/succ segment.
func characterise(seg frames.Signature) sigType {
	switch seg[len(seg)-1].Orientation() {
	case frames.SUCC:
		return succ
	case frames.EOF:
		off := seg[len(seg)-1].Max()
		switch {
		case off == 0:
			return eofZero
		case off < 0:
			return eofWild
		default:
			return eofWindow
		}
	}
	switch seg[0].Orientation() {
	case frames.PREV:
		return prev
	case frames.BOF:
		off := seg[0].Max()
		switch {
		case off == 0:
			return bofZero
		case off < 0:
			return bofWild
		}
	}
	return bofWindow
}

// the position of a key frame in a sequence: the length (minimum length in bytes), start and end indexes
type position struct {
	length int
	start  int
	end    int
}

func (p position) String() string {
	return fmt.Sprintf("POS Length: %d; Start: %d; End: %d", p.length, p.start, p.end)
}

func varLength(seg frames.Signature, max int) position {
	var cur int
	var current, greatest position
	num := seg[0].NumSequences()
	if num > 0 && num <= max && frames.NonZero(seg[0]) {
		current.length, _ = seg[0].Length()
		greatest = position{current.length, 0, 1}
		cur = num
	}
	if len(seg) > 1 {
		for i, f := range seg[1:] {
			if f.Linked(seg[i], 0, 0) {
				num = f.NumSequences()
				if num > 0 && num <= max {
					if current.length > 0 && cur*num <= max {
						l, _ := f.Length()
						current.length += l
						current.end = i + 2
						cur = cur * num
					} else {
						current.length, _ = f.Length()
						current.start, current.end = i+1, i+2
						cur = num
					}
				} else {
					current.length = 0
				}
			} else {
				num = f.NumSequences()
				if num > 0 && num <= max && frames.NonZero(seg[i+1]) {
					current.length, _ = f.Length()
					current.start, current.end = i+1, i+2
					cur = num
				} else {
					current.length = 0
				}
			}
			if current.length > greatest.length {
				greatest = current
			}
		}
	}
	return greatest
}

func bofLength(seg frames.Signature, max int) position {
	var cur int
	var pos position
	num := seg[0].NumSequences()
	if num > 0 && num <= max {
		pos.length, _ = seg[0].Length()
		pos.start, pos.end = 0, 1
		cur = num
	}
	if len(seg) > 1 {
		for i, f := range seg[1:] {
			if f.Linked(seg[i], 0, 0) {
				num = f.NumSequences()
				if num > 0 && num <= max {
					if pos.length > 0 && cur*num <= max {
						l, _ := f.Length()
						pos.length += l
						pos.end = i + 2
						cur = cur * num
						continue
					}
				}
			}
			break
		}
	}
	return pos
}

func eofLength(seg frames.Signature, max int) position {
	var cur int
	var pos position
	num := seg[len(seg)-1].NumSequences()
	if num > 0 && num <= max {
		pos.length, _ = seg[len(seg)-1].Length()
		pos.start, pos.end = len(seg)-1, len(seg)
		cur = num
	}
	if len(seg) > 1 {
		for i := len(seg) - 2; i > -1; i-- {
			f := seg[i]
			if seg[i+1].Linked(f, 0, 0) {
				num = f.NumSequences()
				if num > 0 && num <= max {
					if pos.length > 0 && cur*num <= max {
						l, _ := f.Length()
						pos.length += l
						pos.start = i
						cur = cur * num
						continue
					}
				}
			}
			break
		}
	}
	return pos
}
