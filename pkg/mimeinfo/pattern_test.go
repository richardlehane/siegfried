// Copyright 2015 Richard Lehane. All rights reserved.
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
	"encoding/binary"
	"testing"

	"github.com/richardlehane/siegfried/internal/bytematcher/patterns"
	"github.com/richardlehane/siegfried/internal/persist"
)

var (
	i8  byte  = 8
	i16 int16 = -5000
	i32 int32 = 12345678

	b16, l16 = make([]byte, 2), make([]byte, 2)
	b32, l32 = make([]byte, 4), make([]byte, 4)
)

func init() {
	binary.BigEndian.PutUint16(b16, uint16(i16))
	binary.LittleEndian.PutUint16(l16, uint16(i16))
	binary.BigEndian.PutUint32(b32, uint32(i32))
	binary.LittleEndian.PutUint32(l32, uint32(i32))
}

func TestInt8(t *testing.T) {
	if !Int8(i8).Equals(Int8(i8)) {
		t.Error("Int8 fail: Equality")
	}
	if r, _ := Int8(i8).Test([]byte{7}); r > -1 {
		t.Error("Int8 fail: shouldn't match")
	}
	if r, _ := Int8(i8).Test([]byte{i8}); r < 0 || r != 1 {
		t.Error("Int8 fail: should match")
	}
	if r, _ := Int8(i8).TestR([]byte{i8}); r < 0 || r != 1 {
		t.Error("Int8 fail: should match reverse")
	}
	saver := persist.NewLoadSaver(nil)
	Int8(i8).Save(saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	_ = loader.LoadByte()
	p := loadInt8(loader)
	if !p.Equals(Int8(i8)) {
		t.Errorf("expecting %d, got %s", i8, p)
	}
}

func TestBig16(t *testing.T) {
	if !Big16(i16).Equals(Big16(i16)) {
		t.Error("Big16 fail: Equality")
	}
	if r, _ := Big16(i16).Test(l16); r > -1 {
		t.Error("Big16 fail: shouldn't match")
	}
	if r, _ := Big16(i16).Test(b16); r > 0 || r != 2 {
		t.Error("Big16 fail: should match")
	}
	if r, _ := Big16(i16).TestR(b16); r > 0 || r != 2 {
		t.Error("Big16 fail: should match reverse")
	}
	saver := persist.NewLoadSaver(nil)
	Big16(i16).Save(saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	_ = loader.LoadByte()
	p := loadBig16(loader)
	if !p.Equals(Big16(i16)) {
		t.Errorf("expecting %d, got %s", i16, p)
	}
}

func TestLittle16(t *testing.T) {
	if !Little16(i16).Equals(Little16(i16)) {
		t.Error("Little16 fail: Equality")
	}
	if r, _ := Little16(i16).Test(b16); r > -1 {
		t.Error("Little16 fail: shouldn't match")
	}
	if r, _ := Little16(i16).Test(l16); r > 0 || r != 2 {
		t.Error("Little16 fail: should match")
	}
	if r, _ := Little16(i16).TestR(l16); r > 0 || r != 2 {
		t.Error("Little16 fail: should match reverse")
	}
	saver := persist.NewLoadSaver(nil)
	Little16(i16).Save(saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	_ = loader.LoadByte()
	p := loadLittle16(loader)
	if !p.Equals(Little16(i16)) {
		t.Errorf("expecting %d, got %s", i16, p)
	}
}

func TestHost16(t *testing.T) {
	if !Host16(i16).Equals(Host16(i16)) {
		t.Error("Host16 fail: Equality")
	}
	if r, _ := Host16(i16).Test(b32); r > -1 {
		t.Error("Host16 fail: shouldn't match")
	}
	if r, _ := Host16(i16).Test(l16); r > 0 || r != 2 {
		t.Error("Host16 fail: should match")
	}
	if r, _ := Host16(i16).Test(b16); r > 0 || r != 2 {
		t.Error("Host16 fail: should match")
	}
	if r, _ := Host16(i16).TestR(l16); r > 0 || r != 2 {
		t.Error("Host16 fail: should match reverse")
	}
	if r, _ := Host16(i16).TestR(b16); r > 0 || r != 2 {
		t.Error("Host16 fail: should match reverse")
	}
	saver := persist.NewLoadSaver(nil)
	Host16(i16).Save(saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	_ = loader.LoadByte()
	p := loadHost16(loader)
	if !p.Equals(Host16(i16)) {
		t.Errorf("expecting %d, got %s", i16, p)
	}
}

func TestBig32(t *testing.T) {
	if !Big32(i32).Equals(Big32(i32)) {
		t.Error("Big32 fail: Equality")
	}
	if r, _ := Big32(i32).Test(l32); r > -1 {
		t.Error("Big32 fail: shouldn't match")
	}
	if r, _ := Big32(i32).Test(b32); r > 0 || r != 4 {
		t.Error("Big32 fail: should match")
	}
	if r, _ := Big32(i32).TestR(b32); r > 0 || r != 4 {
		t.Error("Big32 fail: should match")
	}
	saver := persist.NewLoadSaver(nil)
	Big32(i32).Save(saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	_ = loader.LoadByte()
	p := loadBig32(loader)
	if !p.Equals(Big32(i32)) {
		t.Errorf("expecting %d, got %s", i32, p)
	}
}

func TestLittle32(t *testing.T) {
	if !Little32(i32).Equals(Little32(i32)) {
		t.Error("Little32 fail: Equality")
	}
	if r, _ := Little32(i32).Test(b32); r > -1 {
		t.Error("Big32 fail: shouldn't match")
	}
	if r, _ := Little32(i32).Test(l32); r > 0 || r != 4 {
		t.Error("Little32 fail: should match")
	}
	if r, _ := Little32(i32).TestR(l32); r > 0 || r != 4 {
		t.Error("Little32 fail: should match")
	}
	saver := persist.NewLoadSaver(nil)
	Little32(i32).Save(saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	_ = loader.LoadByte()
	p := loadLittle32(loader)
	if !p.Equals(Little32(i32)) {
		t.Errorf("expecting %d, got %s", i32, p)
	}
}

func TestHost32(t *testing.T) {
	if !Host32(i32).Equals(Host32(i32)) {
		t.Error("Host32 fail: Equality")
	}
	if r, _ := Host32(i32).Test(b16); r > -1 {
		t.Error("Host32 fail: shouldn't match")
	}
	if r, _ := Host32(i32).Test(l32); r > 0 || r != 4 {
		t.Error("Host32 fail: should match")
	}
	if r, _ := Host32(i32).Test(b32); r > 0 || r != 4 {
		t.Error("Host32 fail: should match")
	}
	if r, _ := Host32(i32).TestR(l32); r > 0 || r != 4 {
		t.Error("Host32 fail: should match reverse")
	}
	if r, _ := Host32(i32).TestR(b32); r > 0 || r != 4 {
		t.Error("Host32 fail: should match reverse")
	}
	saver := persist.NewLoadSaver(nil)
	Host32(i32).Save(saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	_ = loader.LoadByte()
	p := loadHost32(loader)
	if !p.Equals(Host32(i32)) {
		t.Errorf("expecting %d, got %s", i32, p)
	}
}

func TestIgnoreCase(t *testing.T) {
	apple := []byte("AppLe")
	apple2 := []byte("apple")
	if !IgnoreCase(apple).Equals(IgnoreCase(apple2)) {
		t.Error("IgnoreCase fail: Equality")
	}
	if r, _ := IgnoreCase(apple).Test([]byte("banana")); r > -1 {
		t.Error("IgnoreCase fail: shouldn't match")
	}
	if r, _ := IgnoreCase(apple).Test(IgnoreCase(apple2)); r < 0 || r != 5 {
		t.Error("IgnoreCase fail: should match")
	}
	if r, _ := IgnoreCase(apple).TestR(IgnoreCase(apple2)); r < 0 || r != 5 {
		t.Error("IgnoreCase fail: should match reverse")
	}
	if i := IgnoreCase("!bYt*e").NumSequences(); i != 16 {
		t.Errorf("IgnoreCase fail: numsequences expected %d, got %d", 16, i)
	}
	saver := persist.NewLoadSaver(nil)
	IgnoreCase(apple).Save(saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	_ = loader.LoadByte()
	p := loadIgnoreCase(loader)
	if !p.Equals(IgnoreCase(apple)) {
		t.Errorf("expecting %v, got %v", IgnoreCase(apple), p)
	}
	if seqs := IgnoreCase([]byte("a!cd")).Sequences(); len(seqs) != 8 {
		t.Errorf("IgnoreCase sequences %v", seqs)
	}
}

func TestMask(t *testing.T) {
	apple := Mask{
		pat: patterns.Sequence{'a', 'p', 'p', 0, 0, 'l', 'e'},
		val: []byte{255, 255, 255, 0, 0, 255, 255},
	}
	apple2 := Mask{
		pat: patterns.Sequence{'a', 'p', 'p', 0, 0, 'l', 'e'},
		val: []byte{255, 255, 255, 0, 0, 255, 255},
	}
	if !apple.Equals(apple2) {
		t.Error("Mask fail: Equality")
	}
	if r, _ := apple.Test([]byte("apPyzle")); r > -1 {
		t.Error("Mask fail: shouldn't match")
	}
	if r, _ := apple.Test([]byte("appyzle")); r < 0 || r != 7 {
		t.Error("Mask fail: should match")
	}
	if r, _ := apple.TestR([]byte("appyzle")); r < 0 || r != 7 {
		t.Error("Mask fail: should match reverse")
	}
	saver := persist.NewLoadSaver(nil)
	apple.Save(saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	_ = loader.LoadByte()
	p := loadMask(loader)
	if !p.Equals(apple) {
		t.Errorf("expecting %s, got %s", apple, p)
	}
	seqsTest := Mask{
		pat: patterns.Sequence("ap"),
		val: []byte{0xFF, 0xFE},
	}
	if seqs := seqsTest.Sequences(); len(seqs) != 2 || seqs[1][1] != 'q' {
		t.Error(seqs)
	}
	pats, ints := unmask(apple)
	if len(ints) != 2 || ints[0] != 0 || ints[1] != 2 {
		t.Errorf("Unmask fail, got ints %v", ints)
	}
	if len(pats) != 2 || !pats[0].Equals(patterns.Sequence("app")) || !pats[1].Equals(patterns.Sequence("le")) {
		t.Errorf("Unmask fail, got pats %v", pats)
	}
	pats, ints = unmask(Mask{
		pat: patterns.Sequence{'A', 'C', '0', '0', '0', '0'},
		val: []byte{0xFF, 0xFF, 0xF0, 0xF0, 0xF0, 0xF0},
	})
	if len(ints) != 2 || ints[0] != 0 || ints[1] != 0 {
		t.Errorf("Unmask fail, got ints %v", ints)
	}
	if len(pats) != 2 || !pats[0].Equals(patterns.Sequence("AC")) || !pats[1].Equals(Mask{
		pat: patterns.Sequence{'0', '0', '0', '0'},
		val: []byte{0xF0, 0xF0, 0xF0, 0xF0},
	}) {
		t.Errorf("Unmask fail, got pats %v", pats)
	}
}
