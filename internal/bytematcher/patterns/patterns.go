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

// Package patterns describes the Pattern interface.
// Standard patterns are also defined in this package: Sequence (as well as BMH and reverse BMH Sequence), Choice, List and Not.
package patterns

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/richardlehane/siegfried/internal/persist"
)

func init() {
	Register(sequenceLoader, loadSequence)
	Register(choiceLoader, loadChoice)
	Register(listLoader, loadList)
	Register(notLoader, loadNot)
	Register(bmhLoader, loadBMH)
	Register(rbmhLoader, loadRBMH)
	Register(maskLoader, loadMask)
	Register(anyMaskLoader, loadAnyMask)
}

// Stringify returns a string version of a byte slice.
// If all bytes are UTF8, an ASCII string is returned
// Otherwise a hex string is returned.
func Stringify(b []byte) string {
	if utf8.Valid(b) {
		return strconv.QuoteToASCII(string(b))
	}
	return hex.EncodeToString(b)
}

// Patterns are the smallest building blocks of a format signature.
// Exact byte sequence matches are a type of pattern, as are byte ranges, non-sequence matches etc.
// You can define custom patterns (e.g. for W3C date type) by implementing this interface.
type Pattern interface {
	Test([]byte) ([]int, int)  // For a positive match, returns slice of lengths of the match and bytes to advance for a subsequent test. For a negative match, returns nil or empty slice and the bytes to advance for subsequent test (or 0 if the length of the pattern is longer than the length of the slice).
	TestR([]byte) ([]int, int) // Same as Test but for testing in reverse (from the right-most position of the byte slice).
	Equals(Pattern) bool       // Test equality with another pattern
	Length() (int, int)        // Minimum and maximum lengths of the pattern
	NumSequences() int         // Number of simple sequences represented by a pattern. Return 0 if the pattern cannot be represented by a defined number of simple sequence (e.g. for an indirect offset pattern) or, if in your opinion, the number of sequences is unreasonably large.
	Sequences() []Sequence     // Convert the pattern to a slice of sequences. Return an empty slice if the pattern cannot be represented by a defined number of simple sequences.
	String() string
	Save(*persist.LoadSaver) // encode the pattern into bytes for saving in a persist file
}

// Loader loads a Pattern.
type Loader func(*persist.LoadSaver) Pattern

const (
	sequenceLoader byte = iota
	choiceLoader
	listLoader
	notLoader
	bmhLoader
	rbmhLoader
	maskLoader
	anyMaskLoader
)

var loaders = [32]Loader{}

// Register a new Loader (provide an id higher than 16).
func Register(id byte, l Loader) {
	loaders[int(id)] = l
}

// Load loads the Pattern, choosing the correct Loader by the leading id byte.
func Load(ls *persist.LoadSaver) Pattern {
	id := ls.LoadByte()
	l := loaders[int(id)]
	if l == nil {
		if ls.Err == nil {
			ls.Err = errors.New("bad pattern loader")
		}
		return nil
	}
	return l(ls)
}

// Index reports the offset of one pattern within another (or -1 if not contained)
func Index(a, b Pattern) int {
	if a.Equals(b) {
		return 0
	}
	seq1, ok := a.(Sequence)
	seq2, ok2 := b.(Sequence)
	if ok && ok2 {
		return bytes.Index(seq1, seq2)
	}
	return -1
}

// Sequence is a matching sequence of bytes.
type Sequence []byte

// Test bytes against the pattern.
func (s Sequence) Test(b []byte) ([]int, int) {
	if len(b) < len(s) {
		return nil, 0
	}
	if bytes.Equal(s, b[:len(s)]) {
		return []int{len(s)}, 1
	}
	return nil, 1
}

// Test bytes against the pattern in reverse.
func (s Sequence) TestR(b []byte) ([]int, int) {
	if len(b) < len(s) {
		return nil, 0
	}
	if bytes.Equal(s, b[len(b)-len(s):]) {
		return []int{len(s)}, 1
	}
	return nil, 1
}

// Equals reports whether a pattern is identical to another pattern.
func (s Sequence) Equals(pat Pattern) bool {
	seq2, ok := pat.(Sequence)
	if ok {
		return bytes.Equal(s, seq2)
	}
	return false
}

// Length returns a minimum and maximum length for the pattern.
func (s Sequence) Length() (int, int) {
	return len(s), len(s)
}

// NumSequences reports how many plain sequences are needed to represent this pattern.
func (s Sequence) NumSequences() int {
	return 1
}

// Sequences converts the pattern into a slice of plain sequences.
func (s Sequence) Sequences() []Sequence {
	return []Sequence{s}
}

func (s Sequence) String() string {
	return "seq " + Stringify(s)
}

// The Reverse method is unique to this pattern. It is used for the EOF byte sequence set
func (s Sequence) Reverse() Sequence {
	p := make(Sequence, len(s))
	for i, j := 0, len(s)-1; j > -1; i, j = i+1, j-1 {
		p[i] = s[j]
	}
	return p
}

// Save persists the pattern.
func (s Sequence) Save(ls *persist.LoadSaver) {
	ls.SaveByte(sequenceLoader)
	ls.SaveBytes(s)
}

func loadSequence(ls *persist.LoadSaver) Pattern {
	return Sequence(ls.LoadBytes())
}

// Choice is a slice of patterns, any of which can test successfully for the pattern to succeed. For advance, returns shortest
type Choice []Pattern

func (c Choice) test(b []byte, f func(Pattern, []byte) ([]int, int)) ([]int, int) {
	var r, res []int
	var tl, fl, adv int // trueLen and falseLen
	for _, pat := range c {
		res, adv = f(pat, b)
		if len(res) > 0 {
			r = append(r, res...)
			if tl == 0 || (adv > 0 && adv < tl) {
				tl = adv
			}
		} else if fl == 0 || (adv > 0 && adv < fl) {
			fl = adv
		}
	}
	if len(r) > 0 {
		return r, tl
	}
	return nil, fl
}

// Test bytes against the pattern.
func (c Choice) Test(b []byte) ([]int, int) {
	return c.test(b, Pattern.Test)
}

// Test bytes against the pattern in reverse.
func (c Choice) TestR(b []byte) ([]int, int) {
	return c.test(b, Pattern.TestR)
}

// Equals reports whether a pattern is identical to another pattern.
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

// Length returns a minimum and maximum length for the pattern.
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

// NumSequences reports how many plain sequences are needed to represent this pattern.
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

// Sequences converts the pattern into a slice of plain sequences.
func (c Choice) Sequences() []Sequence {
	num := c.NumSequences()
	seqs := make([]Sequence, 0, num)
	for _, pat := range c {
		seqs = append(seqs, pat.Sequences()...)
	}
	return seqs
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

// Save persists the pattern.
func (c Choice) Save(ls *persist.LoadSaver) {
	ls.SaveByte(choiceLoader)
	ls.SaveSmallInt(len(c))
	for _, pat := range c {
		pat.Save(ls)
	}
}

func loadChoice(ls *persist.LoadSaver) Pattern {
	l := ls.LoadSmallInt()
	choices := make(Choice, l)
	for i := range choices {
		choices[i] = Load(ls)
	}
	return choices
}

// List is a slice of patterns, all of which must test true sequentially in order for the pattern to succeed.
type List []Pattern

// Test bytes against the pattern.
func (l List) Test(b []byte) ([]int, int) {
	if len(l) < 1 {
		return nil, 0
	}
	totals := []int{0}
	for _, pat := range l {
		nts := make([]int, 0, len(totals))
		for _, t := range totals {
			les, _ := pat.Test(b[t:])
			for _, le := range les {
				nts = append(nts, t+le)
			}
		}
		if len(nts) < 1 {
			return nil, 1
		}
		totals = nts
	}
	return totals, 1
}

// Test bytes against the pattern in reverse.
func (l List) TestR(b []byte) ([]int, int) {
	if len(l) < 1 {
		return nil, 0
	}
	totals := []int{0}
	for i := len(l) - 1; i >= 0; i-- {
		nts := make([]int, 0, len(totals))
		for _, t := range totals {
			les, _ := l[i].TestR(b[:len(b)-t])
			for _, le := range les {
				nts = append(nts, t+le)
			}
		}
		if len(nts) < 1 {
			return nil, 1
		}
		totals = nts
	}
	return totals, 1
}

// Equals reports whether a pattern is identical to another pattern.
func (l List) Equals(pat Pattern) bool {
	l2, ok := pat.(List)
	if ok {
		if len(l) == len(l2) {
			for i, p := range l {
				if !p.Equals(l2[i]) {
					return false
				}
			}
		}
	}
	return true
}

// Length returns a minimum and maximum length for the pattern.
func (l List) Length() (int, int) {
	var min, max int
	for _, pat := range l {
		pmin, pmax := pat.Length()
		min += pmin
		max += pmax
	}
	return min, max
}

// NumSequences reports how many plain sequences are needed to represent this pattern.
func (l List) NumSequences() int {
	s := 1
	for _, pat := range l {
		num := pat.NumSequences()
		if num == 0 { // if any of the patterns can't be converted to sequences, don't return any
			return 0
		}
		s *= num
	}
	return s
}

// Sequences converts the pattern into a slice of plain sequences.
func (l List) Sequences() []Sequence {
	total := l.NumSequences()
	seqs := make([]Sequence, total)
	for _, pat := range l {
		num := pat.NumSequences()
		times := total / num
		idx := 0
		for _, seq := range pat.Sequences() {
			for i := 0; i < times; i++ {
				seqs[idx] = append(seqs[idx], seq...)
				idx++
			}
		}
	}
	return seqs
}

func (l List) String() string {
	s := "l["
	for i, pat := range l {
		s += pat.String()
		if i < len(l)-1 {
			s += ","
		}
	}
	return s + "]"
}

// Save persists the pattern.
func (l List) Save(ls *persist.LoadSaver) {
	ls.SaveByte(listLoader)
	ls.SaveSmallInt(len(l))
	for _, pat := range l {
		pat.Save(ls)
	}
}

func loadList(ls *persist.LoadSaver) Pattern {
	le := ls.LoadSmallInt()
	list := make(List, le)
	for i := range list {
		list[i] = Load(ls)
	}
	return list
}

// Not contains a pattern and reports the opposite of that pattern's result when testing.
type Not struct{ Pattern }

// Test bytes against the pattern.
func (n Not) Test(b []byte) ([]int, int) {
	min, _ := n.Pattern.Length()
	if len(b) < min {
		return nil, 0
	}
	ok, _ := n.Pattern.Test(b)
	if len(ok) < 1 {
		return []int{min}, 1
	}
	return nil, 1
}

// Test bytes against the pattern in reverse.
func (n Not) TestR(b []byte) ([]int, int) {
	min, _ := n.Pattern.Length()
	if len(b) < min {
		return nil, 0
	}
	ok, _ := n.Pattern.TestR(b)
	if len(ok) < 1 {
		return []int{min}, 1
	}
	return nil, 1
}

// Equals reports whether a pattern is identical to another pattern.
func (n Not) Equals(pat Pattern) bool {
	n2, ok := pat.(Not)
	if ok {
		return n.Pattern.Equals(n2.Pattern)
	}
	return false
}

// Length returns a minimum and maximum length for the pattern.
func (n Not) Length() (int, int) {
	min, _ := n.Pattern.Length()
	return min, min
}

// NumSequences reports how many plain sequences are needed to represent this pattern.
func (n Not) NumSequences() int {
	_, max := n.Pattern.Length()
	if max > 1 {
		return 0
	}
	num := n.Pattern.NumSequences()
	if num == 0 {
		return 0
	}
	return 256 - num
}

// Sequences converts the pattern into a slice of plain sequences.
func (n Not) Sequences() []Sequence {
	num := n.NumSequences()
	if num < 1 {
		return nil
	}
	seqs := make([]Sequence, 0, num)
	pseqs := n.Pattern.Sequences()
	allBytes := make([]Sequence, 256)
	for i := 0; i < 256; i++ {
		allBytes[i] = Sequence{byte(i)}
	}
	for _, v := range allBytes {
		eq := false
		for _, w := range pseqs {
			if v.Equals(w) {
				eq = true
				break
			}
		}
		if eq {
			continue
		}
		seqs = append(seqs, v)
	}
	return seqs
}

func (n Not) String() string {
	return "not[" + n.Pattern.String() + "]"
}

// Save persists the pattern.
func (n Not) Save(ls *persist.LoadSaver) {
	ls.SaveByte(notLoader)
	n.Pattern.Save(ls)
}

func loadNot(ls *persist.LoadSaver) Pattern {
	return Not{Load(ls)}
}

type Mask byte

func (m Mask) Test(b []byte) ([]int, int) {
	if len(b) == 0 {
		return nil, 0
	}
	if byte(m)&b[0] == byte(m) {
		return []int{1}, 1
	}
	return nil, 1
}

func (m Mask) TestR(b []byte) ([]int, int) {
	if len(b) == 0 {
		return nil, 0
	}
	if byte(m)&b[len(b)-1] == byte(m) {
		return []int{1}, 1
	}
	return nil, 1
}

func (m Mask) Equals(pat Pattern) bool {
	msk, ok := pat.(Mask)
	if ok {
		if m == msk {
			return true
		}
	}
	return false
}

func (m Mask) Length() (int, int) {
	return 1, 1
}

func countBits(b byte) int {
	var count uint
	for b > 0 {
		b &= b - 1
		count++
	}
	return 256 / (1 << count)
}

func allBytes() []byte {
	all := make([]byte, 256)
	for i := range all {
		all[i] = byte(i)
	}
	return all
}

func (m Mask) NumSequences() int {
	return countBits(byte(m))
}

func (m Mask) Sequences() []Sequence {
	seqs := make([]Sequence, 0, m.NumSequences())
	for _, b := range allBytes() {
		if byte(m)&b == byte(m) {
			seqs = append(seqs, Sequence{b})
		}
	}
	return seqs
}

func (m Mask) String() string {
	return fmt.Sprintf("m %#x", byte(m))
}

func (m Mask) Save(ls *persist.LoadSaver) {
	ls.SaveByte(maskLoader)
	ls.SaveByte(byte(m))
}

func loadMask(ls *persist.LoadSaver) Pattern {
	return Mask(ls.LoadByte())
}

type AnyMask byte

func (am AnyMask) Test(b []byte) ([]int, int) {
	if len(b) == 0 {
		return nil, 0
	}
	if byte(am)&b[0] != 0 {
		return []int{1}, 1
	}
	return nil, 1
}

func (am AnyMask) TestR(b []byte) ([]int, int) {
	if len(b) == 0 {
		return nil, 0
	}
	if byte(am)&b[len(b)-1] != 0 {
		return []int{1}, 1
	}
	return nil, 1
}

func (am AnyMask) Equals(pat Pattern) bool {
	amsk, ok := pat.(AnyMask)
	if ok {
		if am == amsk {
			return true
		}
	}
	return false
}

func (am AnyMask) Length() (int, int) {
	return 1, 1
}

func (am AnyMask) NumSequences() int {
	return 256 - countBits(byte(am))
}

func (am AnyMask) Sequences() []Sequence {
	seqs := make([]Sequence, 0, am.NumSequences())
	for _, b := range allBytes() {
		if byte(am)&b != 0 {
			seqs = append(seqs, Sequence{b})
		}
	}
	return seqs
}

func (am AnyMask) String() string {
	return fmt.Sprintf("am %#x", byte(am))
}

func (am AnyMask) Save(ls *persist.LoadSaver) {
	ls.SaveByte(anyMaskLoader)
	ls.SaveByte(byte(am))
}

func loadAnyMask(ls *persist.LoadSaver) Pattern {
	return AnyMask(ls.LoadByte())
}
