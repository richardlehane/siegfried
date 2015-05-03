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

package signature

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

func TestTinyInt(t *testing.T) {
	saver := NewLoadSaver(nil)
	saver.SaveTinyInt(5)
	saver.SaveTinyInt(-1)
	saver.SaveTinyInt(127)
	saver.SaveTinyInt(0)
	saver.SaveTinyInt(-127)
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
}

func TestSmallInt(t *testing.T) {
	saver := NewLoadSaver(nil)
	saver.SaveSmallInt(5)
	saver.SaveSmallInt(-1)
	saver.SaveSmallInt(32767)
	saver.SaveSmallInt(0)
	saver.SaveSmallInt(-32767)
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
}

func TestInt(t *testing.T) {
	saver := NewLoadSaver(nil)
	saver.SaveInt(5)
	saver.SaveInt(-1)
	saver.SaveInt(2147483647)
	saver.SaveInt(0)
	saver.SaveInt(-2147483647)
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
