package bytematcher

import (
	"strconv"
)

// Each segment in a signature is represented by a single keyframe. A slice of keyframes therefore represents a full signature. The Keyframes type is a slice of these slices.
// The keyframe includes the range of offsets that need to match for a successful hit.
type keyFrame struct {
	Typ OffType // defined in frames.go
	Min int
	Max int
}

func (kf keyFrame) String() string {
	return OffString[kf.Typ] + " Min:" + strconv.Itoa(kf.Min) + " Max:" + strconv.Itoa(kf.Max)
}

// A double index: the second int is for the keyframe's position within its signature, the first is for the signature's position within the set of all signatures.
type keyframeID [2]int

// Turn a signature segment into a keyframe and left and right frame slices.
func keyframe(seg Signature, x, y int) (keyFrame, []Frame, []Frame) {
	l, r := make([]Frame, 0), make([]Frame, 0)
	var typ OffType
	var min, max int
	if seg[0].Orientation() < SUCC {
		typ, min, max = seg[0].Orientation(), seg[0].Min(), seg[0].Max()
		for i, f := range seg[:x+1] {
			if x > i {
				l = append([]Frame{SwitchFrame(seg[i+1], f.Pat())}, l...)
			}
		}
		if y < len(seg) {
			r = seg[y:]
		}
		return keyFrame{typ, min, max}, l, r
	}
	typ, min, max = seg[len(seg)-1].Orientation(), seg[len(seg)-1].Min(), seg[len(seg)-1].Max()
	if y < len(seg) {
		for i, f := range seg[y:] {
			r = append(r, SwitchFrame(seg[y+i-1], f.Pat()))
		}
	}
	for _, f := range seg[:x] {
		l = append([]Frame{f}, l...)
	}
	return keyFrame{typ, min, max}, l, r
}

// quick check performed before applying a keyFrame ID
func (kf keyFrame) check(o, l, buflen int) bool {
	switch kf.Typ {
	case EOF:
		o = buflen - l - o
		fallthrough
	case BOF:
		if o < kf.Min || (kf.Max > -1 && o > kf.Max) {
			return false
		}
		return true
	default:
		return true
	}
}

// test two key frames (current and previous) to see if they are connected and, if so, at what offsets
func (kf keyFrame) checkRelated(prevKf keyFrame, thisOff, prevOff [][2]int) ([][2]int, bool) {
	switch kf.Typ {
	case BOF:
		return thisOff, true
	case EOF, SUCC:
		if prevKf.Typ == SUCC {
			ret := make([][2]int, 0, len(thisOff))
			success := false
			for _, v := range thisOff {
				for _, v1 := range prevOff {
					dif := v[0] - v1[0] - v1[1]
					if dif > -1 {
						if dif < prevKf.Min || (prevKf.Max > -1 && dif > prevKf.Max) {
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
					if dif < kf.Min || (kf.Max > -1 && dif > kf.Max) {
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
