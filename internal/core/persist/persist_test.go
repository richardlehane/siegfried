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

package persist

import (
	"testing"
	"time"
)

func TestByte(t *testing.T) {
	saver := NewLoadSaver(nil)
	saver.SaveByte(5)
	saver.SaveBool(true)
	saver.SaveBool(false)
	loader := NewLoadSaver(saver.Bytes())
	i := loader.LoadByte()
	if i != 5 {
		t.Errorf("expecting %d, got %d", 5, i)
	}
	b := loader.LoadBool()
	if !b {
		t.Error("expecting true, got false")
	}
	b = loader.LoadBool()
	if b {
		t.Error("expecting false, got true")
	}
}

func TestBoolField(t *testing.T) {
	saver := NewLoadSaver(nil)
	saver.SaveBoolField(true, false, false, true, false, true, true, true)
	loader := NewLoadSaver(saver.Bytes())
	a, b, c, d, e, f, g, h := loader.LoadBoolField()
	if !a || b || c || !d || e || !f || !g || !h {
		t.Errorf("expecting 'true, false, false, true, false, true, true, true', got %v, %v, %v, %v, %v, %v, %v, %v", a, b, c, d, e, f, g, h)
	}
}

func TestTinyInt(t *testing.T) {
	saver := NewLoadSaver(nil)
	saver.SaveTinyInt(5)
	saver.SaveTinyInt(-1)
	saver.SaveTinyInt(127)
	saver.SaveTinyInt(0)
	saver.SaveTinyInt(-127)
	saver.SaveInts([]int{5, -1, 127, 0, -127})
	loader := NewLoadSaver(saver.Bytes())
	i := loader.LoadTinyInt()
	if i != 5 {
		t.Errorf("expecting %d, got %d", 5, i)
	}
	i = loader.LoadTinyInt()
	if i != -1 {
		t.Errorf("expecting %d, got %d", -1, i)
	}
	i = loader.LoadTinyInt()
	if i != 127 {
		t.Errorf("expecting %d, got %d", 127, i)
	}
	i = loader.LoadTinyInt()
	if i != 0 {
		t.Errorf("expecting %d, got %d", 0, i)
	}
	i = loader.LoadTinyInt()
	if i != -127 {
		t.Errorf("expecting %d, got %d", -127, i)
	}
	is := loader.LoadInts()
	if len(is) != 5 {
		t.Errorf("expecting 5 ints got %d", len(is))
	}
	switch {
	case is[0] != 5:
		t.Errorf("expecting 5, got %d", is[0])
	case is[1] != -1:
		t.Errorf("expecting -1, got %d", is[1])
	case is[2] != 127:
		t.Errorf("expecting 127, got %d", is[2])
	case is[3] != 0:
		t.Errorf("expecting 0, got %d", is[3])
	case is[4] != -127:
		t.Errorf("expecting -127, got %d", is[4])
	}
}

func TestSmallInt(t *testing.T) {
	saver := NewLoadSaver(nil)
	saver.SaveSmallInt(5)
	saver.SaveSmallInt(-1)
	saver.SaveSmallInt(32767)
	saver.SaveSmallInt(0)
	saver.SaveSmallInt(-32767)
	saver.SaveInts([]int{-1, 32767, 0, -32767})
	loader := NewLoadSaver(saver.Bytes())
	i := loader.LoadSmallInt()
	if i != 5 {
		t.Errorf("expecting %d, got %d", 5, i)
	}
	i = loader.LoadSmallInt()
	if i != -1 {
		t.Errorf("expecting %d, got %d", -1, i)
	}
	i = loader.LoadSmallInt()
	if i != 32767 {
		t.Errorf("expecting %d, got %d", 32767, i)
	}
	i = loader.LoadSmallInt()
	if i != 0 {
		t.Errorf("expecting %d, got %d", 0, i)
	}
	i = loader.LoadSmallInt()
	if i != -32767 {
		t.Errorf("expecting %d, got %d", -32767, i)
	}
	c := loader.LoadInts()
	if len(c) != 4 {
		t.Fatalf("expecting 4 results got %v", c)
	}
	if c[0] != -1 {
		t.Errorf("expecting -1, got %v", c[0])
	}
	if c[1] != 32767 {
		t.Errorf("expecting 32767, got %v", c[1])
	}
	if c[2] != 0 {
		t.Errorf("expecting 0, got %v", c[2])
	}
	if c[3] != -32767 {
		t.Errorf("expecting -32767, got %v", c[3])
	}
}

func TestInt(t *testing.T) {
	saver := NewLoadSaver(nil)
	saver.SaveInt(5)
	saver.SaveInt(-1)
	saver.SaveInt(2147483647)
	saver.SaveInt(0)
	saver.SaveInt(-2147483647)
	saver.SaveInts([]int{5, -2147483647, 2147483647, 0})
	loader := NewLoadSaver(saver.Bytes())
	i := loader.LoadInt()
	if i != 5 {
		t.Errorf("expecting %d, got %d", 5, i)
	}
	i = loader.LoadInt()
	if i != -1 {
		t.Errorf("expecting %d, got %d", -1, i)
	}
	i = loader.LoadInt()
	if i != 2147483647 {
		t.Errorf("expecting %d, got %d", 2147483647, i)
	}
	i = loader.LoadInt()
	if i != 0 {
		t.Errorf("expecting %d, got %d", 0, i)
	}
	i = loader.LoadInt()
	if i != -2147483647 {
		t.Errorf("expecting %d, got %d", -2147483647, i)
	}

	c := loader.LoadInts()
	if len(c) != 4 {
		t.Fatalf("expecting 4 results got %v", c)
	}
	if c[0] != 5 {
		t.Errorf("expecting 5, got %v", c[0])
	}
	if c[1] != -2147483647 {
		t.Errorf("expecting -1, got %v", c[1])
	}
	if c[2] != 2147483647 {
		t.Errorf("expecting 2147483647, got %v", c[2])
	}
	if c[3] != 0 {
		t.Errorf("expecting 0, got %v", c[3])
	}
}

func TestString(t *testing.T) {
	saver := NewLoadSaver(nil)
	saver.SaveString("apple")
	saver.SaveString("banana")
	loader := NewLoadSaver(saver.Bytes())
	s := loader.LoadString()
	if s != "apple" {
		t.Errorf("expecting %s, got %s", "apple", s)
	}
	s = loader.LoadString()
	if s != "banana" {
		t.Errorf("expecting %s, got %s", "banana", s)
	}
}

func TestStrings(t *testing.T) {
	saver := NewLoadSaver(nil)
	saver.SaveString("apple")
	saver.SaveStrings([]string{"banana", "orange"})
	saver.SaveStrings([]string{"banana", "grapefruit"})
	loader := NewLoadSaver(saver.Bytes())
	s := loader.LoadString()
	if s != "apple" {
		t.Errorf("expecting %s, got %s", "apple", s)
	}
	ss := loader.LoadStrings()
	if len(ss) != 2 {
		t.Errorf("expecting a slice of two strings got %v", ss)
	}
	if ss[0] != "banana" {
		t.Errorf("expecting %s, got %s", "banana", ss[0])
	}
	if ss[1] != "orange" {
		t.Errorf("expecting %s, got %s", "orange", ss[1])
	}
	s2 := loader.LoadStrings()
	if len(s2) != 2 {
		t.Errorf("expecting a slice of two strings got %v", s2)
	}
	if s2[0] != "banana" {
		t.Errorf("expecting %s, got %s", "banana", s2[0])
	}
	if s2[1] != "grapefruit" {
		t.Errorf("expecting %s, got %s", "grapefruit", s2[1])
	}
}

func TestTime(t *testing.T) {
	saver := NewLoadSaver(nil)
	now := time.Now()
	saver.SaveTime(now)
	loader := NewLoadSaver(saver.Bytes())
	then := loader.LoadTime()
	if now.String() != then.String() {
		t.Errorf("expecting %s to equal %s, errs %v & %v, raw: %v", now, then, loader.Err, saver.Err, saver.Bytes())
	}
}
