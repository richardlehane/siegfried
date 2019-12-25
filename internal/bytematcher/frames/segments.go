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

import "fmt"

type SigType int

const (
	Unknown   SigType = iota
	BOFZero           // fixed offset, zero length from BOF
	BOFWindow         // offset is a window or fixed value greater than zero from BOF
	BOFWild
	Prev
	Succ
	EOFZero
	EOFWindow
	EOFWild
)

// Simple characterisation of a segment: is it relative to the BOF, or the EOF, or is it a prev/succ segment.
func Characterise(seg Signature) SigType {
	if len(seg) == 0 {
		return Unknown
	}
	switch seg[len(seg)-1].Orientation() {
	case SUCC:
		return Succ
	case EOF:
		off := seg[len(seg)-1].Max
		switch {
		case off == 0:
			return EOFZero
		case off < 0:
			return EOFWild
		default:
			return EOFWindow
		}
	}
	switch seg[0].Orientation() {
	case PREV:
		return Prev
	case BOF:
		off := seg[0].Max
		switch {
		case off == 0:
			return BOFZero
		case off < 0:
			return BOFWild
		}
	}
	return BOFWindow
}

// position of a key frame in a segment: the length (minimum length in bytes), start and end indexes.
// The keyframe can span multiple frames in the segment (if they are immediately adjacent and can make sequences)
// which is why there is a start and end index
// If length is 0, the segment goes to the frame matcher
type Position struct {
	Length int
	Start  int
	End    int
}

func (p Position) String() string {
	return fmt.Sprintf("POS Length: %d; Start: %d; End: %d", p.Length, p.Start, p.End)
}

func VarLength(seg Signature, max int) Position {
	var cur int
	var current, greatest Position
	num := seg[0].NumSequences()
	if num > 0 && num <= max && NonZero(seg[0]) {
		current.Length, _ = seg[0].Length()
		greatest = Position{current.Length, 0, 1}
		cur = num
	}
	if len(seg) > 1 {
		for i, f := range seg[1:] {
			if lnk, _, _ := f.Linked(seg[i], 0, 0); lnk {
				num = f.NumSequences()
				if num > 0 && num <= max {
					if current.Length > 0 && cur*num <= max {
						l, _ := f.Length()
						current.Length += l
						current.End = i + 2
						cur = cur * num
					} else {
						current.Length, _ = f.Length()
						current.Start, current.End = i+1, i+2
						cur = num
					}
				} else {
					current.Length = 0
				}
			} else {
				num = f.NumSequences()
				if num > 0 && num <= max && NonZero(seg[i+1]) {
					current.Length, _ = f.Length()
					current.Start, current.End = i+1, i+2
					cur = num
				} else {
					current.Length = 0
				}
			}
			if current.Length > greatest.Length {
				greatest = current
			}
		}
	}
	return greatest
}

func BOFLength(seg Signature, max int) Position {
	var cur int
	var pos Position
	num := seg[0].NumSequences()
	if num > 0 && num <= max {
		pos.Length, _ = seg[0].Length()
		pos.Start, pos.End = 0, 1
		cur = num
	}
	if len(seg) > 1 {
		for i, f := range seg[1:] {
			if lnk, _, _ := f.Linked(seg[i], 0, 0); lnk {
				num = f.NumSequences()
				if num > 0 && num <= max {
					if pos.Length > 0 && cur*num <= max {
						l, _ := f.Length()
						pos.Length += l
						pos.End = i + 2
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

func EOFLength(seg Signature, max int) Position {
	var cur int
	var pos Position
	num := seg[len(seg)-1].NumSequences()
	if num > 0 && num <= max {
		pos.Length, _ = seg[len(seg)-1].Length()
		pos.Start, pos.End = len(seg)-1, len(seg)
		cur = num
	}
	if len(seg) > 1 {
		for i := len(seg) - 2; i >= 0; i-- {
			f := seg[i]
			if lnk, _, _ := seg[i+1].Linked(f, 0, 0); lnk {
				num = f.NumSequences()
				if num > 0 && num <= max {
					if pos.Length > 0 && cur*num <= max {
						l, _ := f.Length()
						pos.Length += l
						pos.Start = i
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
