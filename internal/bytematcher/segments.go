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

package bytematcher

import (
	"fmt"

	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
)

type sigType int

const (
	unknown   sigType = iota
	bofZero           // fixed offset, zero length from BOF
	bofWindow         // offset is a window or fixed value greater than zero from BOF
	bofWild
	prev
	succ
	eofZero
	eofWindow
	eofWild
)

// Simple characterisation of a segment: is it relative to the BOF, or the EOF, or is it a prev/succ segment.
func characterise(seg frames.Signature) sigType {
	if len(seg) == 0 {
		return unknown
	}
	switch seg[len(seg)-1].Orientation() {
	case frames.SUCC:
		return succ
	case frames.EOF:
		off := seg[len(seg)-1].Max
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
		off := seg[0].Max
		switch {
		case off == 0:
			return bofZero
		case off < 0:
			return bofWild
		}
	}
	return bofWindow
}

// position of a key frame in a segment: the length (minimum length in bytes), start and end indexes.
// The keyframe can span multiple frames in the segment (if they are immediately adjacent and can make sequences)
// which is why there is a start and end index
// If length is 0, the segment goes to the frame matcher
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
			if lnk, _, _ := f.Linked(seg[i], 0, 0); lnk {
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
			if lnk, _, _ := f.Linked(seg[i], 0, 0); lnk {
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
		for i := len(seg) - 2; i >= 0; i-- {
			f := seg[i]
			if lnk, _, _ := seg[i+1].Linked(f, 0, 0); lnk {
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
