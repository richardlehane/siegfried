package frames

import "github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns"

// a Sequencer turns sequential frames into a set of plain byte sequences. The set represents possible choices.
type Sequencer func(Frame) [][]byte

func NewSequencer(rev bool) Sequencer {
	ret := make([][]byte, 0)
	return func(f Frame) [][]byte {
		var s []patterns.Sequence
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

func appendSeq(b [][]byte, s []patterns.Sequence) [][]byte {
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
