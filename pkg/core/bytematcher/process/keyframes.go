package process

import (
	"strconv"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type keyFramePos struct {
	PMin int // Minimum and maximum position
	PMax int
	LMin int // Minimum and maximum length
	LMax int
}

func calcLen(fs []frames.Frame) (int, int) {
	var min, max int
	for _, f := range fs {
		fmin, fmax := f.Length()
		min += fmin
		max += fmax
	}
	return min, max
}

// Each segment in a signature is represented by a single keyFrame. A slice of keyFrames represents a full signature. The keyFrames type is a slice of these slices.
// The keyFrame includes the range of offsets that need to match for a successful hit.
type keyFrame struct {
	Typ frames.OffType // defined in frames.go
	Seg keyFramePos    // positioning info for segment as a whole (min/max length and offset in relation to BOF/EOF/PREV/SUCC)
	Key keyFramePos    // positioning info for keyFrame portion of segment (min/max length and offset in relation to BOF/EOF)
}

func (kf keyFrame) String() string {
	return frames.OffString[kf.Typ] + " Min:" + strconv.Itoa(kf.Seg.PMin) + " Max:" + strconv.Itoa(kf.Seg.PMax)
}

// A double index: the first int is for the signature's position within the set of all signatures,
// the second int is for the keyFrames position within the segments of the signature.
type KeyFrameID [2]int

// Turn a signature segment into a keyFrame and left and right frame slices.
func toKeyFrame(seg frames.Signature, pos position) (keyFrame, []frames.Frame, []frames.Frame) {
	left, right := make([]frames.Frame, 0), make([]frames.Frame, 0)
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
		return keyFrame{typ, segPos, keyPos}, left, right
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
	return keyFrame{typ, segPos, keyPos}, left, right
}

func calcMinMax(min, max int, sp keyFramePos) (int, int) {
	min = min + sp.PMin + sp.LMin
	if max < 0 || sp.PMax < 0 {
		return min, -1
	}
	max = max + sp.PMax + sp.LMax
	return min, max
}

func updatePositions(ks []keyFrame) {
	var min, max int
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

// at a given offset, must a keyFrame have already appeared?
// rev = true for EOF matching
// Just uses BOF:
// - doesn't do anything for PREV/SUCC
// - doesn't do anything for EOF??
func (kf keyFrame) MustExist(o int, rev bool) bool {
	switch kf.Typ {
	case frames.BOF:
		// it could be that we are searching the EOF and a VAR BOF has not yet appeared
		if rev && kf.Seg.PMax > 0 {
			return false
		}
		if o > kf.Seg.PMin && (kf.Seg.PMax > -1 && o > kf.Seg.PMax) {
			return true
		}
		return false
	default:
		return false
	}
}

// REPLACE WITH MORE ACCURATE ONE THAT USES NEW KF POSITION INFO + STRIKE
// quick check performed before applying a keyFrame ID
func (kf keyFrame) Check(o, l int, buf *siegreader.Buffer) bool {
	switch kf.Typ {
	case frames.EOF:
		o = buf.Size() - l - o
		fallthrough
	case frames.BOF:
		if o < kf.Seg.PMin || (kf.Seg.PMax > -1 && o > kf.Seg.PMax) {
			return false
		}
		return true
	default:
		return true
	}
}

// test two key frames (current and previous) to see if they are connected and, if so, at what offsets
func (kf keyFrame) CheckRelated(prevKf keyFrame, thisOff, prevOff [][2]int) ([][2]int, bool) {
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
