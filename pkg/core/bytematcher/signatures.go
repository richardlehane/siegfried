package bytematcher

// A signature is a slice of frames
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

func ShortestSeq(max int) func([]Signature) (int, int) {
	shortest := 100
	return func(segs []Signature) (int, int) {
		thisShortest := 100
		for _, seg := range segs {
			if seg[0].Orientation() == BOF || seg[len(seg)-1].Orientation() == EOF {
				continue
			}
			l, _, _ := varLength(seg, max)
			if l < shortest {
				shortest = l
			}
			if l < thisShortest {
				thisShortest = l
			}
		}
		return thisShortest, shortest
	}
}

func intersects(a absoluteFrame, b absoluteFrame) bool {
	return false
}

type absoluteFrame struct {
	OffType
	min int
	max int
	Frame
}

func (s Signature) absolute(i, min, max int, off OffType) (bool, absoluteFrame, int, int) {
	var af absoluteFrame
	f := s[i]
	if off == BOF {
		if f.Max() < 0 || f.Orientation() > PREV {
			return false, af, min, max
		}
	} else {
		if f.Max() < 0 || f.Orientation() < SUCC {
			return false, af, min, max
		}
	}
	switch f.Orientation() {
	case off:
		af = absoluteFrame{off, f.Min(), f.Max(), f}
		l1, l2 := f.Length()
		min, max = f.Min()+l1, f.Max()+l2
	default:
		min, max = f.Min()+min, f.Max()+max
		af = absoluteFrame{off, min, max, f}
		l1, l2 := f.Length()
		min, max = min+l1, max+l2
	}
	return true, af, min, max
}

func (s Signature) absolutes() ([]absoluteFrame, []absoluteFrame) {
	bof, eof := make([]absoluteFrame, 0), make([]absoluteFrame, 0)
	var cont bool
	var min, max int
	var af absoluteFrame
	for i := 0; i < len(s); i++ {
		cont, af, min, max = s.absolute(i, min, max, BOF)
		if cont {
			bof = append(bof, af)
		} else {
			break
		}
	}
	min, max = 0, 0
	for i := len(s) - 1; i > -1; i-- {
		cont, af, min, max = s.absolute(i, min, max, EOF)
		if cont {
			eof = append(eof, af)
		} else {
			break
		}
	}
	return bof, eof
}

// Test whether the successful match of one signature precludes the matching of a second signature
func (s Signature) Conflicts(b Signature) bool {
	bof, eof := s.absolutes()
	bof1, eof1 := b.absolutes()
	for _, f := range bof {
		for _, f1 := range bof1 {
			if f.Equals(f1.Frame) {
				return true
			}
		}
	}
	for _, f := range eof {
		for _, f1 := range eof1 {
			if f.Equals(f1.Frame) {
				return true
			}
		}
	}
	return false
}

// if signature A has matched, does this exclude the possibility of signature B matching?
func exclusive(a Signature, b Signature) bool {
	return false
}
