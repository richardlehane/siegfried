package bytematcher

import (
	"encoding/gob"
	"strconv"
)

func init() {
	gob.Register(Fixed{})
	gob.Register(Window{})
	gob.Register(Wild{})
	gob.Register(WildMin{})
}

type Matchling struct {
	Offset  int
	Typ     OffType
	Reverse bool
}

var FAIL = Matchling{-1, -1, false}

func (m Matchling) Pass() bool {
	if m.Offset < 0 {
		return false
	}
	return true
}

type Frame interface {
	Match([]byte) (bool, []int)
	MatchR([]byte) (bool, []int)
	Equals(Frame) bool
	String() string
	Linked(Frame, int, int) bool // Is a frame linked to a preceding frame (by a preceding or succeding relationship) with an offset and range that is less than the supplied ints?
	Min() int                    // minimum offset
	Max() int                    // maximum offset. Return -1 for no limit (wildcard, *)
	Pat() Pattern

	// The following methods are inherited from the enclosed OffType
	Orientation() OffType
	SwitchOff() OffType

	// The following methods are inherited from the enclosed pattern
	Length() (int, int) // min and max lengths of the enclosed pattern
	NumSequences() int  // // the number of simple sequences that the enclosed pattern can be represented by. Return 0 if the pattern cannot be represented by a defined number of simple sequence (e.g. for an indirect offset pattern) or, if in your opinion, the number of sequences is unreasonably large.
	Sequences() []Sequence
	ValidBytes(i int) []byte
}

type OffType int

const (
	BOF  OffType = iota // beginning of file offset                 // end of file offset
	PREV                // offset from previous frame
	SUCC
	EOF // offset from successive frame
)

func (o OffType) Orientation() OffType {
	return o
}

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

var OffString = map[OffType]string{
	BOF:  "B",
	PREV: "P",
	SUCC: "S",
	EOF:  "E",
}

// Generates Fixed, Window, Wild and WildMin frames.
// The offsets argument controls what type of frame is created.
// - for a Wild frame, give no offsets or give a max offset of < 0 and a min of < 1
// - for a WildMin frame, give one offset, or give a max offset of < 0 and a min of > 0
// - for a Fixed frame, give two offsets that are both >= 0 and that are equal to each other
// - for a Window frame, give two offsets that are both >= 0 and that are not equal to each other.
func NewFrame(typ OffType, pat Pattern, offsets ...int) Frame {
	switch len(offsets) {
	case 0:
		return Frame(Wild{typ, pat})
	case 1:
		if offsets[0] > 0 {
			return Frame(WildMin{typ, offsets[0], pat})
		} else {
			return Frame(Wild{typ, pat})
		}
	}
	if offsets[1] < 0 {
		if offsets[0] > 0 {
			return Frame(WildMin{typ, offsets[0], pat})
		} else {
			return Frame(Wild{typ, pat})
		}
	}
	if offsets[0] < 0 {
		offsets[0] = 0
	}
	if offsets[0] == offsets[1] {
		return Frame(Fixed{typ, offsets[0], pat})
	}
	return Frame(Window{typ, offsets[0], offsets[1], pat})
}

func SwitchFrame(f Frame, p Pattern) Frame {
	return NewFrame(f.SwitchOff(), p, f.Min(), f.Max())
}

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

// e.g. 0
type Fixed struct {
	OffType
	Off int
	Pattern
}

func (f Fixed) Match(b []byte) (bool, []int) {
	if f.Off >= len(b) {
		return false, nil
	}
	if success, length := f.Test(b[f.Off:]); success {
		return true, []int{f.Off + length}
	}
	return false, nil
}

func (f Fixed) MatchR(b []byte) (bool, []int) {
	if f.Off >= len(b) {
		return false, nil
	}
	if success, length := f.TestR(b[:len(b)-f.Off]); success {
		return true, []int{f.Off + length}
	}
	return false, nil
}

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

func (f Fixed) Min() int {
	return f.Off
}

func (f Fixed) Max() int {
	return f.Off
}

func (f Fixed) Linked(prev Frame, maxDistance, maxRange int) bool {
	switch f.OffType {
	case PREV:
		if f.Off > maxDistance {
			return false
		}
		return true
	case SUCC, EOF:
		if prev.Orientation() != SUCC || prev.Max() < 0 || prev.Max() > maxDistance {
			return false
		}
		return true
	default:
		return false
	}
}

func (f Fixed) Pat() Pattern {
	return f.Pattern
}

// e.g. 1-1500
type Window struct {
	OffType
	MinOff int
	MaxOff int
	Pattern
}

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

func (w Window) Min() int {
	return w.MinOff
}

func (w Window) Max() int {
	return w.MaxOff
}

func (w Window) Linked(prev Frame, maxDistance, maxRange int) bool {
	switch w.OffType {
	case PREV:
		if w.MaxOff > maxDistance || w.MaxOff-w.MinOff > maxRange {
			return false
		}
		return true
	case SUCC, EOF:
		if prev.Orientation() != SUCC || prev.Max() < 0 || prev.Max() > maxDistance || prev.Max()-prev.Min() > maxRange {
			return false
		}
		return true
	default:
		return false
	}
}

func (w Window) Pat() Pattern {
	return w.Pattern
}

// *
type Wild struct {
	OffType
	Pattern
}

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

func (w Wild) Min() int {
	return 0
}

func (w Wild) Max() int {
	return -1
}

func (w Wild) Linked(prev Frame, maxDistance, maxRange int) bool {
	switch w.OffType {
	case SUCC, EOF:
		if prev.Orientation() != SUCC || prev.Max() < 0 || prev.Max() > maxDistance || prev.Max()-prev.Min() > maxRange {
			return false
		}
		return true
	default:
		return false
	}
}

func (w Wild) Pat() Pattern {
	return w.Pattern
}

// 200-*
type WildMin struct {
	OffType
	MinOff int
	Pattern
}

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

func (w WildMin) Min() int {
	return w.MinOff
}

func (w WildMin) Max() int {
	return -1
}

func (w WildMin) Linked(prev Frame, maxDistance, maxRange int) bool {
	switch w.OffType {
	case SUCC, EOF:
		if prev.Orientation() != SUCC || prev.Max() < 0 || prev.Max() > maxDistance || prev.Max()-prev.Min() > maxRange {
			return false
		}
		return true
	default:
		return false
	}
}

func (w WildMin) Pat() Pattern {
	return w.Pattern
}
