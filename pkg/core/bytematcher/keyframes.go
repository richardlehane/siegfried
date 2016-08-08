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

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/persist"
	"github.com/richardlehane/siegfried/pkg/core/priority"
)

// positioning information: min/max offsets (in relation to BOF or EOF) and min/max lengths
type keyFramePos struct {
	// Minimum and maximum position
	pMin int64
	pMax int64
	// Minimum and maximum length
	lMin int
	lMax int
}

// Each segment in a signature is represented by a single keyFrame. A slice of keyFrames represents a full signature.
// The keyFrame includes the range of offsets that need to match for a successful hit.
// The segment (Seg) offsets are relative (to preceding/succeding segments or to BOF/EOF if the first or last segment).
// The keyframe (Key) offsets are absolute to the BOF or EOF.
type keyFrame struct {
	typ frames.OffType // BOF|PREV|SUCC|EOF
	seg keyFramePos    // relative positioning info for segment as a whole (min/max length and offset in relation to BOF/EOF/PREV/SUCC)
	key keyFramePos    // absolute positioning info for keyFrame portion of segment (min/max length and offset in relation to BOF/EOF)
}

func loadKeyFrames(ls *persist.LoadSaver) [][]keyFrame {
	kfs := make([][]keyFrame, ls.LoadSmallInt())
	for i := range kfs {
		kfs[i] = make([]keyFrame, ls.LoadSmallInt())
		for j := range kfs[i] {
			kfs[i][j].typ = frames.OffType(ls.LoadByte())
			kfs[i][j].seg.pMin = int64(ls.LoadInt())
			kfs[i][j].seg.pMax = int64(ls.LoadInt())
			kfs[i][j].seg.lMin = ls.LoadSmallInt()
			kfs[i][j].seg.lMax = ls.LoadSmallInt()
			kfs[i][j].key.pMin = int64(ls.LoadInt())
			kfs[i][j].key.pMax = int64(ls.LoadInt())
			kfs[i][j].key.lMin = ls.LoadSmallInt()
			kfs[i][j].key.lMax = ls.LoadSmallInt()
		}
	}
	return kfs
}

func saveKeyFrames(ls *persist.LoadSaver, kfs [][]keyFrame) {
	ls.SaveSmallInt(len(kfs))
	for _, v := range kfs {
		ls.SaveSmallInt(len(v))
		for _, kf := range v {
			ls.SaveByte(byte(kf.typ))
			ls.SaveInt(int(kf.seg.pMin))
			ls.SaveInt(int(kf.seg.pMax))
			ls.SaveSmallInt(kf.seg.lMin)
			ls.SaveSmallInt(kf.seg.lMax)
			ls.SaveInt(int(kf.key.pMin))
			ls.SaveInt(int(kf.key.pMax))
			ls.SaveSmallInt(kf.key.lMin)
			ls.SaveSmallInt(kf.key.lMax)
		}
	}

}

func (kf keyFrame) String() string {
	return fmt.Sprintf("%s Seg Min:%d Seg Max:%d; Abs Min:%d Abs Max:%d", frames.OffString[kf.typ], kf.seg.pMin, kf.seg.pMax, kf.key.pMin, kf.key.pMax)
}

// A double index: the first int is for the signature's position within the set of all signatures,
// the second int is for the keyFrames position within the segments of the signature.
type keyFrameID [2]int

func (kf keyFrameID) String() string {
	return fmt.Sprintf("[%d:%d]", kf[0], kf[1])
}

type kfFilter struct {
	idx int
	fdx int
	kfs []keyFrameID
	nfs []keyFrameID
}

func (k *kfFilter) Next() int {
	if k.idx >= len(k.kfs) {
		return -1
	}
	k.idx++
	return k.kfs[k.idx-1][0]
}

func (k *kfFilter) Mark(t bool) {
	if t {
		k.nfs[k.fdx] = k.kfs[k.idx-1]
		k.fdx++
	}
}

func filterKF(kfs []keyFrameID, ws *priority.WaitSet) []keyFrameID {
	f := &kfFilter{kfs: kfs, nfs: make([]keyFrameID, len(kfs))}
	ws.ApplyFilter(f)
	return f.nfs[:f.fdx]
}

// Turn a signature segment into a keyFrame and left and right frame slices.
// The left and right frame slices are converted into BMH sequences where possible
func toKeyFrame(seg frames.Signature, pos position) (keyFrame, []frames.Frame, []frames.Frame) {
	var left, right []frames.Frame
	var typ frames.OffType
	var segPos, keyPos keyFramePos
	segPos.lMin, segPos.lMax = calcLen(seg)
	keyPos.lMin, keyPos.lMax = calcLen(seg[pos.start:pos.end])
	// BOF and PREV segments
	if seg[0].Orientation() < frames.SUCC {
		typ, segPos.pMin, segPos.pMax = seg[0].Orientation(), int64(seg[0].Min()), int64(seg[0].Max())
		keyPos.pMin, keyPos.pMax = segPos.pMin, segPos.pMax
		for i, f := range seg[:pos.start+1] {
			if pos.start > i {
				min, max := f.Length()
				keyPos.pMin += int64(min)
				keyPos.pMin += int64(seg[i+1].Min())
				if keyPos.pMax > -1 {
					keyPos.pMax += int64(max)
					keyPos.pMax += int64(seg[i+1].Max())
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
	typ, segPos.pMin, segPos.pMax = seg[len(seg)-1].Orientation(), int64(seg[len(seg)-1].Min()), int64(seg[len(seg)-1].Max())
	keyPos.pMin, keyPos.pMax = segPos.pMin, segPos.pMax
	if pos.end < len(seg) {
		for i, f := range seg[pos.end:] {
			min, max := f.Length()
			keyPos.pMin += int64(min)
			keyPos.pMin += int64(seg[pos.end+i-1].Min())
			if keyPos.pMax > -1 {
				keyPos.pMax += int64(max)
				keyPos.pMax += int64(seg[pos.end+i-1].Max())
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
	if fs[0].Orientation() < frames.SUCC {
		for i, f := range fs {
			fmin, fmax := f.Length()
			min += fmin
			max += fmax
			if i > 0 {
				min += f.Min()
				max += f.Max()
			}
		}
		return min, max
	}
	for i := len(fs) - 1; i > -1; i-- {
		f := fs[i]
		fmin, fmax := f.Length()
		min += fmin
		max += fmax
		if i < len(fs)-1 {
			min += f.Min()
			max += f.Max()
		}
	}
	return min, max
}

func calcMinMax(min, max int64, sp keyFramePos) (int64, int64) {
	min = min + sp.pMin + int64(sp.lMin)
	if max < 0 || sp.pMax < 0 {
		return min, -1
	}
	max = max + sp.pMax + int64(sp.lMax)
	return min, max
}

// update the absolute positional information (distance from the BOF or EOF)
// for keyFrames based on the other keyFrames in the signature
func updatePositions(ks []keyFrame) {
	var min, max int64
	// first forwards, for BOF and PREV
	for i := range ks {
		if ks[i].typ == frames.BOF {
			min, max = calcMinMax(0, 0, ks[i].seg)
			// Apply max bof
			if config.MaxBOF() > 0 {
				if ks[i].key.pMax < 0 || ks[i].key.pMax > int64(config.MaxBOF()) {
					ks[i].key.pMax = int64(config.MaxBOF())
				}
			}
		}
		if ks[i].typ == frames.PREV {
			ks[i].key.pMin = min + ks[i].key.pMin
			if max > -1 && ks[i].key.pMax > -1 {
				ks[i].key.pMax = max + ks[i].key.pMax
			} else {
				ks[i].key.pMax = -1
			}
			min, max = calcMinMax(min, max, ks[i].seg)
			// Apply max bof
			if config.MaxBOF() > 0 {
				if ks[i].key.pMax < 0 || ks[i].key.pMax > int64(config.MaxBOF()) {
					ks[i].key.pMax = int64(config.MaxBOF())
				}
			}
		}
	}
	// now backwards for EOF and SUCC
	min, max = 0, 0
	for i := len(ks) - 1; i >= 0; i-- {
		if ks[i].typ == frames.EOF {
			min, max = calcMinMax(0, 0, ks[i].seg)
			// apply max eof
			if config.MaxEOF() > 0 {
				if ks[i].key.pMax < 0 || ks[i].key.pMax > int64(config.MaxEOF()) {
					ks[i].key.pMax = int64(config.MaxEOF())
				}
			}
		}
		if ks[i].typ == frames.SUCC {
			ks[i].key.pMin = min + ks[i].key.pMin
			if max > -1 && ks[i].key.pMax > -1 {
				ks[i].key.pMax = max + ks[i].key.pMax
			} else {
				ks[i].key.pMax = -1
			}
			min, max = calcMinMax(min, max, ks[i].seg)
			// apply max eof
			if config.MaxEOF() > 0 {
				if ks[i].key.pMax < 0 || ks[i].key.pMax > int64(config.MaxEOF()) {
					ks[i].key.pMax = int64(config.MaxEOF())
				}
			}
		}
	}
}

// for doing a running total of the firstBOF and firstEOF:
func firstBOFandEOF(bof int, eof int, ks []keyFrame) (int, int) {
	if bof < 0 {
		return bof, eof
	}
	b := getMax(-1, func(t frames.OffType) bool { return t == frames.BOF }, ks, true)
	if b < 0 || b > bof {
		e := getMax(-1, func(t frames.OffType) bool { return t == frames.EOF }, ks, true)
		if e < 0 {
			if b < 0 {
				return -1, -1
			}
			return b, eof
		}
		if e > eof {
			if b < 0 || b > e {
				return bof, e
			}
			return b, eof
		}
	}
	return bof, eof
}

func getMax(max int, t func(frames.OffType) bool, ks []keyFrame, localMin bool) int {
	for _, v := range ks {
		if t(v.typ) {
			if v.key.pMax < 0 {
				if !localMin {
					return -1
				}
				continue
			}
			this := int(v.key.pMax) + v.key.lMax
			if localMin {
				if max < 0 || this < max {
					max = this
				}
			} else if this > max {
				max = this
			}
		}
	}
	return max
}

// for doing a running total of the maxBOF:
// is the maxBOF we already have, further from the BOF than the maxBOF of the current signature?
func maxBOF(max int, ks []keyFrame) int {
	if max < 0 {
		return max
	}
	return getMax(max, func(t frames.OffType) bool { return t < frames.SUCC }, ks, false)
}

func maxEOF(max int, ks []keyFrame) int {
	if max < 0 {
		return max
	}
	return getMax(max, func(t frames.OffType) bool { return t > frames.PREV }, ks, false)
}

func crossOver(a, b keyFrame) bool {
	if a.key.pMax == -1 {
		return true
	}
	if a.key.pMax+int64(a.key.lMax) > b.key.pMin {
		return true
	}
	return false
}

// quick check performed before applying a keyFrame ID
func (kf keyFrame) check(o int64) bool {
	if kf.key.pMin > o {
		return false
	}
	if kf.key.pMax == -1 {
		return true
	}
	if kf.key.pMax < o {
		return false
	}
	return true
}

// can we gather just a single hit for this keyframe?
func oneEnough(id int, kfs []keyFrame) bool {
	kf := kfs[id]
	// if this is a BOF frame or a wild PREV frame we can ...
	if kf.typ == frames.BOF || (kf.typ == frames.PREV && kf.seg.pMax == -1 && kf.seg.pMin == 0) {
		// unless this isn't the last frame and the next frame is a non-wild PREV frame
		if id+1 < len(kfs) {
			next := kfs[id+1]
			if next.typ == frames.PREV && (next.seg.pMax > -1 || next.seg.pMin > 0) {
				return false
			}
		}
		return true
	}
	// if this is an EOF frame or SUCC frame we can ...
	if id > 0 {
		// so long as there isn't a previous frame that is a non-wild SUCC frame
		prev := kfs[id-1]
		if prev.typ == frames.SUCC && (prev.seg.pMax > -1 || prev.seg.pMin > 0) {
			return false
		}
	}
	return true
}

// test two key frames (current and previous) to see if they are connected and, if so, at what offsets
func (kf keyFrame) checkRelated(prevKf, nextKf keyFrame, thisOff, prevOff [][2]int64) ([][2]int64, bool) {
	switch kf.typ {
	case frames.BOF:
		return thisOff, true
	case frames.EOF, frames.SUCC:
		if prevKf.typ == frames.SUCC && !(prevKf.seg.pMax == -1 && prevKf.seg.pMin == 0) {
			ret := make([][2]int64, 0, len(thisOff))
			success := false
			for _, v := range thisOff {
				for _, v1 := range prevOff {
					dif := v[0] - v1[0] - v1[1]
					if dif > -1 {
						if dif < prevKf.seg.pMin || (prevKf.seg.pMax > -1 && dif > prevKf.seg.pMax) {
							continue
						} else {
							ret = append(ret, v)
							success = true
							// if this type is EOF, we only need one match
							if kf.typ == frames.EOF {
								return ret, success
							}
						}
					}
				}
			}
			return ret, success
		} else {
			return thisOff, true
		}
	default:
		if kf.seg.pMax == -1 && kf.seg.pMin == 0 {
			return thisOff, true
		}
		ret := make([][2]int64, 0, len(thisOff))
		success := false
		for _, v := range thisOff {
			for _, v1 := range prevOff {
				dif := v[0] - v1[0] - v1[1] // current offset, minus previous offset, minus previous length
				if dif > -1 {
					if dif < kf.seg.pMin || (kf.seg.pMax > -1 && dif > kf.seg.pMax) {
						continue
					} else {
						ret = append(ret, v)
						success = true
						// if the next type isn't a non-wild PREV, we only need one match
						if nextKf.typ != frames.PREV || (nextKf.seg.pMax == -1 && nextKf.seg.pMin == 0) {
							return ret, success
						}
					}
				}
			}
		}
		return ret, success
	}
}
