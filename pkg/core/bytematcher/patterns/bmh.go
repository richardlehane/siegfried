package patterns

// THIS FILE IS NOT YET IMPLEMENTED
// IT IS AN OPTIMISATION THAT CAN BE DEFERRED

import (
	"bytes"
	"strconv"
)

// A BMH Sequence is an optimised version of the regular Sequence pattern
type BMHSequence struct {
	seq Sequence
	bmh [256]int
}

func NewBMHSequence(s Sequence) *BMHSequence {
	var badshift [256]int
	for i, _ := range badshift {
		badshift[i] = len(s)
	}
	last := len(s)
	for i := 0; i < last; i++ {
		badshift[s[i]] = last - i
	}
	return &BMHSequence{s, badshift}
}

func (s *BMHSequence) Test(b []byte) (bool, int) {
	if len(b) < len(s.seq) {
		return false, 0
	}
	scan := len(s.seq) - 1
	for ; b[scan] == s.seq[scan]; scan-- {
		if scan == 0 {
			return true, len(s.seq)
		}
	}
	return false, 1
}

func (s *BMHSequence) TestR(b []byte) (bool, int) {
	return false, 1
}

func (s *BMHSequence) TestByte(b byte, i int) bool {
	if i < len(s.seq) && s.seq[i] == b {
		return true
	}
	return false
}

func (s *BMHSequence) Equals(pat Pattern) bool {
	seq2, ok := pat.(*BMHSequence)
	if ok {
		return bytes.Equal(s.seq, seq2.seq)
	}
	return false
}

func (s *BMHSequence) Length() (int, int) {
	return len(s.seq), len(s.seq)
}

func (s *BMHSequence) NumSequences() int {
	return 1
}

func (s *BMHSequence) Sequences() []Sequence {
	return []Sequence{s.seq}
}

func (s *BMHSequence) ValidBytes(i int) []byte {
	if i < len(s.seq) {
		return []byte{s.seq[i]}
	}
	return []byte{}
}

func (s *BMHSequence) String() string {
	return "seq" + strconv.Itoa(len(s.seq))
}
