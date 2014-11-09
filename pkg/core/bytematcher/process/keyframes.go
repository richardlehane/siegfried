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

// positioning information: min/max offsets (in relation to BOF or EOF) and min/max lengths
type keyFramePos struct {
	// Minimum and maximum position
	PMin int
	PMax int
	// Minimum and maximum length
	LMin int
	LMax int
}

// Each segment in a signature is represented by a single keyFrame. A slice of keyFrames represents a full signature.
// The keyFrame includes the range of offsets that need to match for a successful hit.
// The segment (Seg) offsets are relative (to preceding/succeding segments or to BOF/EOF if the first or last segment).
// The keyframe (Key) offsets are absolute to the BOF or EOF.
type keyFrame struct {
	Typ frames.OffType // BOF|PREV|SUCC|EOF
	Seg keyFramePos    // positioning info for segment as a whole (min/max length and offset in relation to BOF/EOF/PREV/SUCC)
	Key keyFramePos    // positioning info for keyFrame portion of segment (min/max length and offset in relation to BOF/EOF)
}

func (kf keyFrame) String() string {
	return fmt.Sprintf("%s Min:%d Max:%d", frames.OffString[kf.Typ], kf.Seg.PMin, kf.Seg.PMax)
}

// A double index: the first int is for the signature's position within the set of all signatures,
// the second int is for the keyFrames position within the segments of the signature.
type KeyFrameID [2]int

func (kf KeyFrameID) String() string {
	return fmt.Sprintf("[%d:%d]", kf[0], kf[1])
}

// Turn a signature segment into a keyFrame and left and right frame slices.
// The left and right frame slices are converted into BMH sequences where possible
func toKeyFrame(seg frames.Signature, pos position) (keyFrame, []frames.Frame, []frames.Frame) {
	var left, right []frames.Frame
	var typ frames.OffType
	var segPos, keyPos keyFramePos
	segPos.LMin, segPos.LMax = calcLen(seg)
	keyPos.LMin, keyPos.LMax = calcLen(seg[pos.start:pos.end])
	// BOF and PREV segments
	if seg[0].Orientation() < frames.SUCC {
		typ, segPos.PMin, segPos.PMax = seg[0].Orientation(), seg[0].Min(), seg[0].Max()
		keyPos.PMin, keyPos.PMax = segPos.PMin, segPos.PMax
		for i, f := range seg[:pos.start+1] {
			if pos.start > i {
				min, max := f.Length()
				keyPos.PMin += min
				keyPos.PMin += seg[i+1].Min()
				if keyPos.PMax > -1 {
					keyPos.PMax += max
					keyPos.PMax += seg[i+1].Max()
				}
				left = append([]frames.Frame{frames.SwitchFrame(seg[i+1], f.Pat())}, left...)
			}
		}
		if pos.end < len(seg) {
			right = seg[pos.end:]
		}
		return keyFrame{typ, segPos, keyPos}, frames.BMHConvert(left, true), frames.BMHConvert(right, false)
	}
	// EOF and SUCC segments
	typ, segPos.PMin, segPos.PMax = seg[len(seg)-1].Orientation(), seg[len(seg)-1].Min(), seg[len(seg)-1].Max()
	keyPos.PMin, keyPos.PMax = segPos.PMin, segPos.PMax
	if pos.end < len(seg) {
		for i, f := range seg[pos.end:] {
			min, max := f.Length()
			keyPos.PMin += min
			keyPos.PMin += seg[pos.end+i-1].Min()
			if keyPos.PMax > -1 {
				keyPos.PMax += max
				keyPos.PMax += seg[pos.end+i-1].Max()
			}
			right = append(right, frames.SwitchFrame(seg[pos.end+i-1], f.Pat()))
		}
	}
	for _, f := range seg[:pos.start] {
		left = append([]frames.Frame{f}, left...)
	}
	return keyFrame{typ, segPos, keyPos}, frames.BMHConvert(left, true), frames.BMHConvert(right, false)
}

// calculate minimum and maximum lengths for a segment (slice of frames)
func calcLen(fs []frames.Frame) (int, int) {
	var min, max int
	for _, f := range fs {
		fmin, fmax := f.Length()
		min += fmin
		max += fmax
	}
	return min, max
}

func calcMinMax(min, max int, sp keyFramePos) (int, int) {
	min = min + sp.PMin + sp.LMin
	if max < 0 || sp.PMax < 0 {
		return min, -1
	}
	max = max + sp.PMax + sp.LMax
	return min, max
}

// update the absolute positional information (distance from the BOF or EOF)
// for keyFrames based on the other keyFrames in the signature
func updatePositions(ks []keyFrame) {
	var min, max int
	// first forwards, for BOF and PREV
	for i := range ks {
		if ks[i].Typ == frames.BOF {
			min, max = calcMinMax(0, 0, ks[i].Seg)
		}
		if ks[i].Typ == frames.PREV {
			ks[i].Key.PMin = min + ks[i].Key.PMin
			if max > -1 && ks[i].Key.PMax > -1 {
				ks[i].Key.PMax = max + ks[i].Key.PMax
			} else {
				ks[i].Key.PMax = -1
			}
			min, max = calcMinMax(min, max, ks[i].Seg)
		}
	}
	// now backwards for EOF and SUCC
	min, max = 0, 0
	for i := len(ks) - 1; i >= 0; i-- {
		if ks[i].Typ == frames.EOF {
			min, max = calcMinMax(0, 0, ks[i].Seg)
		}
		if ks[i].Typ == frames.SUCC {
			ks[i].Key.PMin = min + ks[i].Key.PMin
			if max > -1 && ks[i].Key.PMax > -1 {
				ks[i].Key.PMax = max + ks[i].Key.PMax
			} else {
				ks[i].Key.PMax = -1
			}
			min, max = calcMinMax(min, max, ks[i].Seg)
		}
	}
}

// for doing a running totally of the maxBOF:
// is the maxBOF we already have, further from the BOF than the maxBOF of the current signature?
func maxBOF(max int, ks []keyFrame) int {
	if max < 0 {
		return -1
	}
	for _, v := range ks {
		if v.Typ < frames.SUCC {
			if v.Key.PMax < 0 {
				return -1
			}
			if v.Key.PMax+v.Key.LMax > max {
				max = v.Key.PMax + v.Key.LMax
			}
		}
	}
	return max
}

func maxEOF(max int, ks []keyFrame) int {
	if max < 0 {
		return -1
	}
	for _, v := range ks {
		if v.Typ > frames.PREV {
			if v.Key.PMax < 0 {
				return -1
			}
			if v.Key.PMax+v.Key.LMax > max {
				max = v.Key.PMax + v.Key.LMax
			}
		}
	}
	return max
}

// quick check performed before applying a keyFrame ID
func (kf keyFrame) Check(o int) bool {
	if kf.Key.PMin > o {
		return false
	}
	if kf.Key.PMax == -1 {
		return true
	}
	if kf.Key.PMax < o {
		return false
	}
	return true
}

// proper segment check before committing an incomplete keyframe (necessary when there are left or right tests)
func (kf keyFrame) CheckSeg(o int) bool {
	if kf.Seg.PMin > o {
		return false
	}
	if kf.Seg.PMax == -1 {
		return true
	}
	if kf.Seg.PMax < o {
		return false
	}
	return true
}

// test two key frames (current and previous) to see if they are connected and, if so, at what offsets
func (kf keyFrame) CheckRelated(prevKf keyFrame, thisOff, prevOff [][2]int) ([][2]int, bool) {
	// quick test for wild kf
	if prevKf.Seg.PMax == -1 && prevKf.Seg.PMin == 0 {
		return thisOff, true
	}
	switch kf.Typ {
	case frames.BOF:
		return thisOff, true
	case frames.EOF, frames.SUCC:
		if prevKf.Typ == frames.SUCC {
			ret := make([][2]int, 0, len(thisOff))
			success := false
			for _, v := range thisOff {
				for _, v1 := range prevOff {
					dif := v[0] - v1[0] - v1[1]
					if dif > -1 {
						if dif < prevKf.Seg.PMin || (prevKf.Seg.PMax > -1 && dif > prevKf.Seg.PMax) {
							continue
						} else {
							ret = append(ret, v)
							success = true
						}
					}
				}
			}
			return ret, success
		} else {
			return thisOff, true
		}
	default:
		ret := make([][2]int, 0, len(thisOff))
		success := false
		for _, v := range thisOff {
			for _, v1 := range prevOff {
				dif := v[0] - v1[0] - v1[1] // current offset, minus previous offset, minus previous length
				if dif > -1 {
					if dif < kf.Seg.PMin || (kf.Seg.PMax > -1 && dif > kf.Seg.PMax) {
						continue
					} else {
						ret = append(ret, v)
						success = true
					}
				}
			}
		}
		return ret, success
	}
}
