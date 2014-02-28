package bytematcher

// Signatures are split into segments (which are themselves signatures).
// This separation happens on wildcards or when the distance between frames is deemed too great.
// E.g. a signature of [BOF 0: "ABCD"][PREV 0-20: "EFG"][PREV Wild: "HI"][EOF 0: "XYZ"]
// has three segments:
// 1. [BOF 0: "ABCD"][PREV 0-20: "EFG"]
// 2. [PREV Wild: "HI"]
// 3. [EOF 0: "XYZ"]

// The maxDistance and maxRange args control the allowable distance and range between frames
// (i.e. a fixed offset of 5000 distant might be acceptable, where a range of 1-2000 might not be).
func segment(sig Signature, maxDistance, maxRange int) []Signature {
	if len(sig) < 2 {
		return []Signature{sig}
	}
	sigs := make([]Signature, 0, 1)
	newSig := Signature{sig[0]}
	for i, frame := range sig[1:] {
		if frame.Linked(sig[i], maxDistance, maxRange) {
			newSig = append(newSig, frame)
		} else {
			sigs = append(sigs, newSig)
			newSig = Signature{frame}
		}
	}
	return append(sigs, newSig)
}

type sigType int

const (
	bofZero sigType = iota // fixed offset, zero length from BOF
	eofZero
	bofWindow // offset is a window or fixed value greater than zero from BOF
	eofWindow
	bofWild
	eofWild
)

// Simple characterisation of a signature segment: is it relative to the BOF, or the EOF.
// Is it at a zero offset, a fixed/window offset, or wild (variable) offset?
func characterise(seg Signature, max int) sigType {
	switch seg[len(seg)-1].Orientation() {
	case SUCC:
		return eofWild
	case EOF:
		off := seg[len(seg)-1].Max()
		switch {
		case off == 0:
			return eofZero
		case off < 0, off >= max:
			return eofWild
		default:
			return eofWindow
		}
	}
	switch seg[0].Orientation() {
	case PREV:
		return bofWild
	case BOF:
		off := seg[0].Max()
		switch {
		case off == 0:
			return bofZero
		case off < 0, off >= max:
			return bofWild
		}
	}
	return bofWindow
}

// Is this a purely variable signature, with no segment anchored to the BOF or EOF?
func variable(segs []Signature, max int) bool {
	var vary bool = true
	for _, seg := range segs {
		if c := characterise(seg, max); c < bofWild {
			vary = false
		}
	}
	return vary
}

func bofLength(seg Signature, max int) (int, int, int) {
	var cur, length, start, end int
	num := seg[0].NumSequences()
	if num > 0 && num <= max {
		length, _ = seg[0].Length()
		start, end = 0, 1
		cur = num
	}
	if len(seg) > 1 {
		for i, f := range seg[1:] {
			if f.Linked(seg[i], 0, 0) {
				num = f.NumSequences()
				if num > 0 && num <= max {
					if length > 0 && cur*num <= max {
						l, _ := f.Length()
						length += l
						end = i + 2
						cur = cur * num
						continue
					}
				}
			}
			break
		}
	}
	return length, start, end
}

func varLength(seg Signature, max int) (int, int, int) {
	var cur, length, start, end, tallyL, tallyS, tallyE int
	num := seg[0].NumSequences()
	if num > 0 && num <= max && NonZero(seg[0]) {
		length, _ = seg[0].Length()
		tallyL, tallyS, tallyE = length, 0, 1
		cur = num
	}
	if len(seg) > 1 {
		for i, f := range seg[1:] {
			if f.Linked(seg[i], 0, 0) {
				num = f.NumSequences()
				if num > 0 && num <= max {
					if length > 0 && cur*num <= max {
						l, _ := f.Length()
						length += l
						end = i + 2
						cur = cur * num
					} else {
						length, _ = f.Length()
						start = i + 1
						end = i + 2
						cur = num
					}
				} else {
					length = 0
				}
			} else {
				num = f.NumSequences()
				if num > 0 && num <= max && NonZero(seg[0]) {
					length, _ = f.Length()
					start = i + 1
					end = i + 2
					cur = num
				} else {
					length = 0
				}
			}
			if length > tallyL {
				tallyL, tallyS, tallyE = length, start, end
			}
		}
	}
	return tallyL, tallyS, tallyE
}

func eofLength(seg Signature, max int) (int, int, int) {
	var cur, length, start, end int
	num := seg[len(seg)-1].NumSequences()
	if num > 0 && num <= max {
		length, _ = seg[len(seg)-1].Length()
		start, end = len(seg)-1, len(seg)
		cur = num
	}
	if len(seg) > 1 {
		for i := len(seg) - 2; i > -1; i-- {
			f := seg[i]
			if seg[i+1].Linked(f, 0, 0) {
				num = f.NumSequences()
				if num > 0 && num <= max {
					if length > 0 && cur*num <= max {
						l, _ := f.Length()
						length += l
						start = i
						cur = cur * num
						continue
					}
				}
			}
			break
		}
	}
	return length, start, end
}

type sequencer func(Frame) [][]byte

func newSequencer(rev bool) sequencer {
	ret := make([][]byte, 0)
	return func(f Frame) [][]byte {
		var s []Sequence
		if rev {
			s = f.Sequences()
			for i, _ := range s {
				s[i] = s[i].Reverse()
			}
		} else {
			s = f.Sequences()
		}
		ret = appendSeq(ret, s)
		return ret
	}
}

func appendSeq(b [][]byte, s []Sequence) [][]byte {
	var c [][]byte
	if len(b) == 0 {
		c = make([][]byte, len(s))
		for i, seq := range s {
			c[i] = make([]byte, len(seq))
			copy(c[i], []byte(seq))
		}
	} else {
		c = make([][]byte, len(b)*len(s))
		iter := 0
		for _, seq := range s {
			for _, orig := range b {
				c[iter] = make([]byte, len(orig)+len(seq))
				copy(c[iter], orig)
				copy(c[iter][len(orig):], []byte(seq))
				iter++
			}
		}
	}
	return c
}
