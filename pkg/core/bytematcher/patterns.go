package bytematcher

import (
	"bytes"
	"encoding/gob"
	"strconv"
)

func init() {
	gob.Register(Sequence{})
	gob.Register(Choice{})
}

// Patterns are the smallest building blocks of a format signature.
// Exact byte sequence matches are a type of pattern, as are byte ranges, non-sequence matches etc.
// You can define custom patterns (e.g. for W3C date type) by implementing this interface.
type Pattern interface {
	Test([]byte) (bool, int)  // Returns boolean for match. For a positive match, the integer value represents the length of the match. For a negative match, the integer represents an offset jump before a subsequent test. That offset should be 0 if the remaining byte slice is smaller than the pattern.
	TestR([]byte) (bool, int) // Same as Test but for testing in reverse (from the right-most position of the byte slice).
	Equals(Pattern) bool      // Test equality with another pattern
	Length() (int, int)       // Minimum and maximum lengths of the pattern
	NumSequences() int        // Number of simple sequences represented by a pattern. Return 0 if the pattern cannot be represented by a defined number of simple sequence (e.g. for an indirect offset pattern) or, if in your opinion, the number of sequences is unreasonably large.
	Sequences() []Sequence    // Convert the pattern to a slice of sequences. Return an empty slice if the pattern cannot be represented by a defined number of simple sequences.
	ValidBytes(int) []byte    // Valid (matching) bytes at a particular offset of the pattern
	String() string
}

// A sequence is a matching sequence of bytes.
type Sequence []byte

func (s Sequence) Test(b []byte) (bool, int) {
	if len(b) < len(s) {
		return false, 0
	}
	if bytes.Equal(s, b[:len(s)]) {
		return true, len(s)
	}
	return false, 1
}

func (s Sequence) TestR(b []byte) (bool, int) {
	if len(b) < len(s) {
		return false, 0
	}
	if bytes.Equal(s, b[len(b)-len(s):]) {
		return true, len(s)
	}
	return false, 1
}

func (s Sequence) Equals(pat Pattern) bool {
	seq2, ok := pat.(Sequence)
	if ok {
		return bytes.Equal(s, seq2)
	}
	return false
}

func (s Sequence) Length() (int, int) {
	return len(s), len(s)
}

func (s Sequence) NumSequences() int {
	return 1
}

func (s Sequence) Sequences() []Sequence {
	return []Sequence{s}
}

func (s Sequence) ValidBytes(i int) []byte {
	if i < len(s) {
		return []byte{s[i]}
	}
	return []byte{}
}

func (s Sequence) String() string {
	return "seq" + strconv.Itoa(len(s))
}

// The Reverse method is unique to this pattern. It is used for the EOF byte sequence set
func (s Sequence) Reverse() Sequence {
	p := make(Sequence, len(s))
	for i, j := 0, len(s)-1; j > -1; i, j = i+1, j-1 {
		p[i] = s[j]
	}
	return p
}

// A choice is a slice of patterns, any of which can test true for the pattern to succeed. Returns the longest matching pattern
type Choice []Pattern

func (c Choice) test(b []byte, f func(Pattern, []byte) (bool, int)) (bool, int) {
	var r, res bool
	var tl, fl, lgth int
	for _, pat := range c {
		res, lgth = f(pat, b)
		if res {
			r = true
			if lgth > tl {
				tl = lgth
			}
		} else if lgth > fl {
			fl = lgth
		}
	}
	if r {
		return r, tl
	}
	return r, fl
}

func (c Choice) Test(b []byte) (bool, int) {
	return c.test(b, Pattern.Test)
}

func (c Choice) TestR(b []byte) (bool, int) {
	return c.test(b, Pattern.TestR)
}

func (c Choice) Equals(pat Pattern) bool {
	c2, ok := pat.(Choice)
	if ok {
		if len(c) == len(c2) {
			for _, p := range c {
				ident := false
				for _, p2 := range c2 {
					if p.Equals(p2) {
						ident = true
					}
				}
				if !ident {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (c Choice) Length() (int, int) {
	var min, max int
	if len(c) > 0 {
		min, max = c[0].Length()
	}
	for _, pat := range c {
		min2, max2 := pat.Length()
		if min2 < min {
			min = min2
		}
		if max2 > max {
			max = max2
		}
	}
	return min, max
}

func (c Choice) NumSequences() int {
	var s int
	for _, pat := range c {
		num := pat.NumSequences()
		if num == 0 { // if any of the patterns can't be converted to sequences, don't return any
			return 0
		}
		s += num
	}
	return s
}

func (c Choice) Sequences() []Sequence {
	num := c.NumSequences()
	seqs := make([]Sequence, 0, num)
	for _, pat := range c {
		seqs = append(seqs, pat.Sequences()...)
	}
	return seqs
}

func (c Choice) ValidBytes(i int) []byte {
	res := make([]byte, 0)
	for _, pat := range c {
		byts := pat.ValidBytes(i)
		for _, b := range byts {
			if bytes.IndexByte(res, b) < 0 {
				res = append(res, b)
			}
		}
	}
	return res
}

func (c Choice) String() string {
	s := "c["
	for i, pat := range c {
		s += pat.String()
		if i < len(c)-1 {
			s += ","
		}
	}
	return s + "]"
}
