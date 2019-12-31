// Copyright 2016 Richard Lehane. All rights reserved.
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

package mimeinfo

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"github.com/richardlehane/siegfried/internal/bytematcher/patterns"
	"github.com/richardlehane/siegfried/internal/persist"
)

func init() {
	patterns.Register(int8Loader, loadInt8)
	patterns.Register(big16Loader, loadBig16)
	patterns.Register(big32Loader, loadBig32)
	patterns.Register(little16Loader, loadLittle16)
	patterns.Register(little32Loader, loadLittle32)
	patterns.Register(host16Loader, loadHost16)
	patterns.Register(host32Loader, loadHost32)
	patterns.Register(ignoreCaseLoader, loadIgnoreCase)
	patterns.Register(maskLoader, loadMask)
}

const (
	int8Loader = iota + 16
	big16Loader
	big32Loader
	little16Loader
	little32Loader
	host16Loader
	host32Loader
	ignoreCaseLoader
	maskLoader
)

type Int8 byte

// Test bytes against the pattern.
func (n Int8) Test(b []byte) ([]int, int) {
	if len(b) < 1 {
		return nil, 0
	}
	if b[0] == byte(n) {
		return []int{1}, 1
	}
	return nil, 1
}

// Test bytes against the pattern in reverse.
func (n Int8) TestR(b []byte) ([]int, int) {
	if len(b) < 1 {
		return nil, 0
	}
	if b[len(b)-1] == byte(n) {
		return []int{1}, 1
	}
	return nil, 1
}

// Equals reports whether a pattern is identical to another pattern.
func (n Int8) Equals(pat patterns.Pattern) bool {
	n2, ok := pat.(Int8)
	if ok {
		return n == n2
	}
	return false
}

// Length returns a minimum and maximum length for the pattern.
func (n Int8) Length() (int, int) {
	return 1, 1
}

// NumSequences reports how many plain sequences are needed to represent this pattern.
func (n Int8) NumSequences() int {
	return 1
}

// Sequences converts the pattern into a slice of plain sequences.
func (n Int8) Sequences() []patterns.Sequence {
	return []patterns.Sequence{{byte(n)}}
}

func (n Int8) String() string {
	return "int8 " + hex.EncodeToString([]byte{byte(n)})
}

// Save persists the pattern.
func (n Int8) Save(ls *persist.LoadSaver) {
	ls.SaveByte(int8Loader)
	ls.SaveByte(byte(n))
}

func loadInt8(ls *persist.LoadSaver) patterns.Pattern {
	return Int8(ls.LoadByte())
}

type Big16 uint16

// Test bytes against the pattern.
func (n Big16) Test(b []byte) ([]int, int) {
	if len(b) < 2 {
		return nil, 0
	}
	if binary.BigEndian.Uint16(b[:2]) == uint16(n) {
		return []int{2}, 1
	}
	return nil, 1
}

// Test bytes against the pattern in reverse.
func (n Big16) TestR(b []byte) ([]int, int) {
	if len(b) < 2 {
		return nil, 0
	}
	if binary.BigEndian.Uint16(b[len(b)-2:]) == uint16(n) {
		return []int{2}, 1
	}
	return nil, 1
}

// Equals reports whether a pattern is identical to another pattern.
func (n Big16) Equals(pat patterns.Pattern) bool {
	n2, ok := pat.(Big16)
	if ok {
		return n == n2
	}
	return false
}

// Length returns a minimum and maximum length for the pattern.
func (n Big16) Length() (int, int) {
	return 2, 2
}

// NumSequences reports how many plain sequences are needed to represent this pattern.
func (n Big16) NumSequences() int {
	return 1
}

// Sequences converts the pattern into a slice of plain sequences.
func (n Big16) Sequences() []patterns.Sequence {
	seq := make(patterns.Sequence, 2)
	binary.BigEndian.PutUint16([]byte(seq), uint16(n))
	return []patterns.Sequence{seq}
}

func (n Big16) String() string {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(n))
	return "big16 " + hex.EncodeToString(buf)
}

// Save persists the pattern.
func (n Big16) Save(ls *persist.LoadSaver) {
	ls.SaveByte(big16Loader)
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(n))
	ls.SaveBytes(buf)
}

func loadBig16(ls *persist.LoadSaver) patterns.Pattern {
	return Big16(binary.BigEndian.Uint16(ls.LoadBytes()))
}

type Big32 uint32

// Test bytes against the pattern.
func (n Big32) Test(b []byte) ([]int, int) {
	if len(b) < 4 {
		return nil, 0
	}
	if binary.BigEndian.Uint32(b[:4]) == uint32(n) {
		return []int{4}, 1
	}
	return nil, 1
}

// Test bytes against the pattern in reverse.
func (n Big32) TestR(b []byte) ([]int, int) {
	if len(b) < 4 {
		return nil, 0
	}
	if binary.BigEndian.Uint32(b[len(b)-4:]) == uint32(n) {
		return []int{4}, 1
	}
	return nil, 1
}

// Equals reports whether a pattern is identical to another pattern.
func (n Big32) Equals(pat patterns.Pattern) bool {
	n2, ok := pat.(Big32)
	if ok {
		return n == n2
	}
	return false
}

// Length returns a minimum and maximum length for the pattern.
func (n Big32) Length() (int, int) {
	return 4, 4
}

// NumSequences reports how many plain sequences are needed to represent this pattern.
func (n Big32) NumSequences() int {
	return 1
}

// Sequences converts the pattern into a slice of plain sequences.
func (n Big32) Sequences() []patterns.Sequence {
	seq := make(patterns.Sequence, 4)
	binary.BigEndian.PutUint32([]byte(seq), uint32(n))
	return []patterns.Sequence{seq}
}

func (n Big32) String() string {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(n))
	return "big32 " + hex.EncodeToString(buf)
}

// Save persists the pattern.
func (n Big32) Save(ls *persist.LoadSaver) {
	ls.SaveByte(big32Loader)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(n))
	ls.SaveBytes(buf)
}

func loadBig32(ls *persist.LoadSaver) patterns.Pattern {
	return Big32(binary.BigEndian.Uint32(ls.LoadBytes()))
}

type Little16 uint16

// Test bytes against the pattern.
func (n Little16) Test(b []byte) ([]int, int) {
	if len(b) < 2 {
		return nil, 0
	}
	if binary.LittleEndian.Uint16(b[:2]) == uint16(n) {
		return []int{2}, 1
	}
	return nil, 1
}

// Test bytes against the pattern in reverse.
func (n Little16) TestR(b []byte) ([]int, int) {
	if len(b) < 2 {
		return nil, 0
	}
	if binary.LittleEndian.Uint16(b[len(b)-2:]) == uint16(n) {
		return []int{2}, 1
	}
	return nil, 1
}

// Equals reports whether a pattern is identical to another pattern.
func (n Little16) Equals(pat patterns.Pattern) bool {
	n2, ok := pat.(Little16)
	if ok {
		return n == n2
	}
	return false
}

// Length returns a minimum and maximum length for the pattern.
func (n Little16) Length() (int, int) {
	return 2, 2
}

// NumSequences reports how many plain sequences are needed to represent this pattern.
func (n Little16) NumSequences() int {
	return 1
}

// Sequences converts the pattern into a slice of plain sequences.
func (n Little16) Sequences() []patterns.Sequence {
	seq := make(patterns.Sequence, 2)
	binary.LittleEndian.PutUint16([]byte(seq), uint16(n))
	return []patterns.Sequence{seq}
}

func (n Little16) String() string {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(n))
	return "little16 " + hex.EncodeToString(buf)
}

// Save persists the pattern.
func (n Little16) Save(ls *persist.LoadSaver) {
	ls.SaveByte(little16Loader)
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(n))
	ls.SaveBytes(buf)
}

func loadLittle16(ls *persist.LoadSaver) patterns.Pattern {
	return Little16(binary.LittleEndian.Uint16(ls.LoadBytes()))
}

type Little32 uint32

// Test bytes against the pattern.
func (n Little32) Test(b []byte) ([]int, int) {
	if len(b) < 4 {
		return nil, 0
	}
	if binary.LittleEndian.Uint32(b[:4]) == uint32(n) {
		return []int{4}, 1
	}
	return nil, 1
}

// Test bytes against the pattern in reverse.
func (n Little32) TestR(b []byte) ([]int, int) {
	if len(b) < 4 {
		return nil, 0
	}
	if binary.LittleEndian.Uint32(b[len(b)-4:]) == uint32(n) {
		return []int{4}, 1
	}
	return nil, 1
}

// Equals reports whether a pattern is identical to another pattern.
func (n Little32) Equals(pat patterns.Pattern) bool {
	n2, ok := pat.(Little32)
	if ok {
		return n == n2
	}
	return false
}

// Length returns a minimum and maximum length for the pattern.
func (n Little32) Length() (int, int) {
	return 4, 4
}

// NumSequences reports how many plain sequences are needed to represent this pattern.
func (n Little32) NumSequences() int {
	return 1
}

// Sequences converts the pattern into a slice of plain sequences.
func (n Little32) Sequences() []patterns.Sequence {
	seq := make(patterns.Sequence, 4)
	binary.LittleEndian.PutUint32([]byte(seq), uint32(n))
	return []patterns.Sequence{seq}
}

func (n Little32) String() string {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(n))
	return "little32 " + hex.EncodeToString(buf)
}

// Save persists the pattern.
func (n Little32) Save(ls *persist.LoadSaver) {
	ls.SaveByte(little32Loader)
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(n))
	ls.SaveBytes(buf)
}

func loadLittle32(ls *persist.LoadSaver) patterns.Pattern {
	return Little32(binary.LittleEndian.Uint32(ls.LoadBytes()))
}

type Host16 uint16

// Test bytes against the pattern.
func (n Host16) Test(b []byte) ([]int, int) {
	if len(b) < 2 {
		return nil, 0
	}
	if binary.LittleEndian.Uint16(b[:2]) == uint16(n) {
		return []int{2}, 1
	}
	if binary.BigEndian.Uint16(b[:2]) == uint16(n) {
		return []int{2}, 1
	}
	return nil, 1
}

// Test bytes against the pattern in reverse.
func (n Host16) TestR(b []byte) ([]int, int) {
	if len(b) < 2 {
		return nil, 0
	}
	if binary.LittleEndian.Uint16(b[len(b)-2:]) == uint16(n) {
		return []int{2}, 1
	}
	if binary.BigEndian.Uint16(b[len(b)-2:]) == uint16(n) {
		return []int{2}, 1
	}
	return nil, 1
}

// Equals reports whether a pattern is identical to another pattern.
func (n Host16) Equals(pat patterns.Pattern) bool {
	n2, ok := pat.(Host16)
	if ok {
		return n == n2
	}
	return false
}

// Length returns a minimum and maximum length for the pattern.
func (n Host16) Length() (int, int) {
	return 2, 2
}

// NumSequences reports how many plain sequences are needed to represent this pattern.
func (n Host16) NumSequences() int {
	return 2
}

// Sequences converts the pattern into a slice of plain sequences.
func (n Host16) Sequences() []patterns.Sequence {
	seq, seq2 := make(patterns.Sequence, 2), make(patterns.Sequence, 2)
	binary.LittleEndian.PutUint16([]byte(seq), uint16(n))
	binary.BigEndian.PutUint16([]byte(seq2), uint16(n))
	return []patterns.Sequence{seq, seq2}
}

func (n Host16) String() string {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(n))
	return "host16 " + hex.EncodeToString(buf)
}

// Save persists the pattern.
func (n Host16) Save(ls *persist.LoadSaver) {
	ls.SaveByte(host16Loader)
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(n))
	ls.SaveBytes(buf)
}

func loadHost16(ls *persist.LoadSaver) patterns.Pattern {
	return Host16(binary.LittleEndian.Uint16(ls.LoadBytes()))
}

type Host32 uint32

// Test bytes against the pattern.
func (n Host32) Test(b []byte) ([]int, int) {
	if len(b) < 4 {
		return nil, 0
	}
	if binary.LittleEndian.Uint32(b[:4]) == uint32(n) {
		return []int{4}, 1
	}
	if binary.BigEndian.Uint32(b[:4]) == uint32(n) {
		return []int{4}, 1
	}
	return nil, 1
}

// Test bytes against the pattern in reverse.
func (n Host32) TestR(b []byte) ([]int, int) {
	if len(b) < 4 {
		return nil, 0
	}
	if binary.LittleEndian.Uint32(b[len(b)-4:]) == uint32(n) {
		return []int{4}, 1
	}
	if binary.BigEndian.Uint32(b[len(b)-4:]) == uint32(n) {
		return []int{4}, 1
	}
	return nil, 1
}

// Equals reports whether a pattern is identical to another pattern.
func (n Host32) Equals(pat patterns.Pattern) bool {
	n2, ok := pat.(Host32)
	if ok {
		return n == n2
	}
	return false
}

// Length returns a minimum and maximum length for the pattern.
func (n Host32) Length() (int, int) {
	return 4, 4
}

// NumSequences reports how many plain sequences are needed to represent this pattern.
func (n Host32) NumSequences() int {
	return 2
}

// Sequences converts the pattern into a slice of plain sequences.
func (n Host32) Sequences() []patterns.Sequence {
	seq, seq2 := make(patterns.Sequence, 4), make(patterns.Sequence, 4)
	binary.LittleEndian.PutUint32([]byte(seq), uint32(n))
	binary.BigEndian.PutUint32([]byte(seq2), uint32(n))
	return []patterns.Sequence{seq, seq2}
}

func (n Host32) String() string {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(n))
	return "host32 " + hex.EncodeToString(buf)
}

// Save persists the pattern.
func (n Host32) Save(ls *persist.LoadSaver) {
	ls.SaveByte(host32Loader)
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(n))
	ls.SaveBytes(buf)
}

func loadHost32(ls *persist.LoadSaver) patterns.Pattern {
	return Host32(binary.LittleEndian.Uint32(ls.LoadBytes()))
}

type IgnoreCase []byte

func (c IgnoreCase) Test(b []byte) ([]int, int) {
	if len(b) < len(c) {
		return nil, 0
	}
	for i, v := range c {
		if v != b[i] {
			if 'a' <= v && v <= 'z' && b[i] == v-'a'-'A' {
				continue
			}
			if 'A' <= v && v <= 'Z' && b[i] == v+'a'-'A' {
				continue
			}
			return nil, 1
		}
	}
	return []int{len(c)}, 1
}

func (c IgnoreCase) TestR(b []byte) ([]int, int) {
	if len(b) < len(c) {
		return nil, 0
	}
	for i, v := range c {
		if v != b[len(b)-len(c)+i] {
			if 'a' <= v && v <= 'z' && b[len(b)-len(c)+i] == v-'a'-'A' {
				continue
			}
			if 'A' <= v && v <= 'Z' && b[len(b)-len(c)+i] == v+'a'-'A' {
				continue
			}
			return nil, 1
		}
	}
	return []int{len(c)}, 1
}

// Equals reports whether a pattern is identical to another pattern.
func (c IgnoreCase) Equals(pat patterns.Pattern) bool {
	c2, ok := pat.(IgnoreCase)
	if ok && bytes.Equal(bytes.ToLower(c), bytes.ToLower(c2)) {
		return true
	}
	return false
}

// Length returns a minimum and maximum length for the pattern.
func (c IgnoreCase) Length() (int, int) {
	return len(c), len(c)
}

// NumSequences reports how many plain sequences are needed to represent this pattern.
func (c IgnoreCase) NumSequences() int {
	i := 1
	for _, v := range c {
		if 'A' <= v && v <= 'z' {
			i *= 2
		}
	}
	return i
}

// Sequences converts the pattern into a slice of plain sequences.
func (c IgnoreCase) Sequences() []patterns.Sequence {
	var ret []patterns.Sequence
	for _, v := range c {
		switch {
		case 'a' <= v && v <= 'z':
			ret = sequences(ret, v, v-('a'-'A'))
		case 'A' <= v && v <= 'Z':
			ret = sequences(ret, v, v+('a'-'A'))
		default:
			ret = sequences(ret, v)
		}
	}
	return ret
}

func (c IgnoreCase) String() string {
	return "ignore case " + string(c)
}

// Save persists the pattern.
func (c IgnoreCase) Save(ls *persist.LoadSaver) {
	ls.SaveByte(ignoreCaseLoader)
	ls.SaveBytes(c)
}

func loadIgnoreCase(ls *persist.LoadSaver) patterns.Pattern {
	return IgnoreCase(ls.LoadBytes())
}

func sequences(pats []patterns.Sequence, opts ...byte) []patterns.Sequence {
	if len(pats) == 0 {
		pats = []patterns.Sequence{{}}
	}
	ret := make([]patterns.Sequence, len(opts)*len(pats))
	var i int
	for _, b := range opts {
		for _, p := range pats {
			seq := make(patterns.Sequence, len(p)+1)
			copy(seq, p)
			seq[len(p)] = b
			ret[i] = seq
			i++
		}
	}
	return ret
}

type Mask struct {
	pat patterns.Pattern
	val []byte // masks for numerical types can be any number; masks for strings must be in base16 and start with 0x
}

func (m Mask) Test(b []byte) ([]int, int) {
	if len(b) < len(m.val) {
		return nil, 0
	}
	t := make([]byte, len(m.val))
	for i := range t {
		t[i] = b[i] & m.val[i]
	}
	return m.pat.Test(t)
}

func (m Mask) TestR(b []byte) ([]int, int) {
	if len(b) < len(m.val) {
		return nil, 0
	}
	t := make([]byte, len(m.val))
	for i := range t {
		t[i] = b[len(b)-len(t)+i] & m.val[i]
	}
	return m.pat.TestR(t)
}

// Equals reports whether a pattern is identical to another pattern.
func (m Mask) Equals(pat patterns.Pattern) bool {
	m2, ok := pat.(Mask)
	if ok && m.pat.Equals(m2.pat) && bytes.Equal(m.val, m2.val) {
		return true
	}
	return false
}

// Length returns a minimum and maximum length for the pattern.
func (m Mask) Length() (int, int) {
	return m.pat.Length()
}

func validMasks(a, b byte) []byte {
	var ret []byte
	var byt byte
	for ; ; byt++ {
		if a&byt == b {
			ret = append(ret, byt)
		}
		if byt == 255 {
			break
		}
	}
	return ret
}

// NumSequences reports how many plain sequences are needed to represent this pattern.
func (m Mask) NumSequences() int {
	if n := m.pat.NumSequences(); n != 1 {
		return 0
	}
	seq := m.pat.Sequences()[0]
	if len(m.val) != len(seq) {
		return 0
	}
	var ret int
	for i, b := range m.val {
		ret *= len(validMasks(b, seq[i]))
	}
	return ret
}

// Sequences converts the pattern into a slice of plain sequences.
func (m Mask) Sequences() []patterns.Sequence {
	if n := m.pat.NumSequences(); n != 1 {
		return nil
	}
	seq := m.pat.Sequences()[0]
	if len(m.val) != len(seq) {
		return nil
	}
	var ret []patterns.Sequence
	for i, b := range m.val {
		ret = sequences(ret, validMasks(b, seq[i])...)
	}
	return ret
}

func (m Mask) String() string {
	return "mask " + hex.EncodeToString(m.val) + " (" + m.pat.String() + ")"
}

// Save persists the pattern.
func (m Mask) Save(ls *persist.LoadSaver) {
	ls.SaveByte(maskLoader)
	m.pat.Save(ls)
	ls.SaveBytes(m.val)
}

func loadMask(ls *persist.LoadSaver) patterns.Pattern {
	return Mask{
		pat: patterns.Load(ls),
		val: ls.LoadBytes(),
	}
}

func repairMask(m Mask) (Mask, bool) {
	seq := m.pat.Sequences()[0]
	for _, b := range m.val {
		if b != 0xFF && b != 0x00 {
			return m, false
		}
	}
	for i, b := range seq {
		if b == '.' {
			break
		}
		if i == len(seq)-1 {
			return m, false
		}
	}
	nv := make([]byte, len(seq))
	for i, v := range seq {
		if v == '.' {
			nv[i] = 0x00
		} else {
			nv[i] = 0xFF
		}
	}
	return Mask{seq, nv}, true
}

// Unmask turns 0xFF00 masks into a slice of patterns and slice of distances between those patterns.
func unmask(m Mask) ([]patterns.Pattern, []int) {
	if m.pat.NumSequences() != 1 {
		return []patterns.Pattern{m}, []int{0}
	}
	seq := m.pat.Sequences()[0]
	if len(seq) != len(m.val) {
		var ok bool
		m, ok = repairMask(m)
		if !ok {
			return []patterns.Pattern{m}, []int{0}
		}
	}
	pret, iret := []patterns.Pattern{}, []int{}
	var slc, skip int
	for idx, byt := range m.val {
		switch byt {
		case 0xFF:
			slc++
		case 0x00:
			if slc > 0 {
				pat := make(patterns.Sequence, slc)
				copy(pat, seq[idx-slc:idx])
				pret = append(pret, pat)
				iret = append(iret, skip)
				slc, skip = 0, 0
			}
			skip++
		default:
			if slc > 0 {
				pat := make(patterns.Sequence, slc)
				copy(pat, seq[idx-slc:idx])
				pret = append(pret, pat)
				iret = append(iret, skip)
				slc, skip = 0, 0
			}
			pat := make(patterns.Sequence, len(m.val)-idx)
			copy(pat, seq[idx:])
			pret = append(pret, Mask{pat: pat, val: m.val[idx:]})
			iret = append(iret, skip)
			return pret, iret
		}
	}
	if slc > 0 {
		pat := make(patterns.Sequence, slc)
		copy(pat, seq[len(m.val)-slc:])
		pret = append(pret, pat)
		iret = append(iret, skip)
	}
	return pret, iret
}
