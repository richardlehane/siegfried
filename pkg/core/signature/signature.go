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

// Package signature defines how siegfried marshals and unmarshals signatures as binary data
package signature

// todo - look at BinaryMarshaler and BinaryUnmarshaler in "encoding"

// type PatternLoader func(*core.LoadSaver) patterns.Pattern
// And for save - just add a Save(*core.LoadSaver) method to Patterns interface
// LoadBytematcher(*core.LoadSaver) core.Matcher
// And for save - Save(*core.LoadSaver) method on core.Matcher

import (
	"encoding/binary"
	"errors"
)

type LoadSaver struct {
	buf []byte
	i   int
	Err error
}

func NewLoadSaver(b []byte) *LoadSaver {
	if len(b) == 0 {
		b = make([]byte, 16)
	}
	return &LoadSaver{
		b,
		0,
		nil,
	}
}

func (l *LoadSaver) Bytes() []byte {
	return l.buf[:l.i]
}

func (l *LoadSaver) read(i int) []byte {
	if l.Err != nil {
		return nil
	}
	if l.i+i > len(l.buf) {
		l.Err = errors.New("error loading signature file, overflowed")
		return nil
	}
	l.i += i
	return l.buf[l.i-i : l.i]
}

func (l *LoadSaver) write(b []byte) {
	if l.Err != nil {
		return
	}
	if len(b)+l.i > len(l.buf) {
		nbuf := make([]byte, len(l.buf)*2)
		copy(nbuf, l.buf[:l.i])
		l.buf = nbuf
	}
	copy(l.buf[l.i:len(b)+l.i], b)
	l.i += len(b)
}

func (l *LoadSaver) LoadByte() byte {
	le := l.read(1)
	if le == nil {
		return 0
	}
	return byte(le[0])
}

func (l *LoadSaver) SaveByte(b byte) {
	l.write([]byte{b})
}

func (l *LoadSaver) LoadBool() bool {
	b := l.LoadByte()
	if b == 0xFF {
		return true
	}
	return false
}

func (l *LoadSaver) SaveBool(b bool) {
	if b {
		l.SaveByte(0xFF)
	} else {
		l.SaveByte(0)
	}
}

func (l *LoadSaver) LoadTinyInt() int {
	i := int(l.LoadByte())
	if i > 128 {
		return -256 + i
	}
	return i
}

func (l *LoadSaver) SaveTinyInt(i int) {
	if i <= -128 || i >= 128 {
		l.Err = errors.New("int overflows byte")
		return
	}
	l.SaveByte(byte(i))
}

func (l *LoadSaver) LoadTinyUInt() int {
	return int(l.LoadByte())
}

func (l *LoadSaver) SaveTinyUInt(i int) {
	if i < 0 || i >= 256 {
		l.Err = errors.New("int overflows byte as a uint")
		return
	}
	l.SaveByte(byte(i))
}

func (l *LoadSaver) LoadSmallInt() int {
	le := l.read(2)
	if le == nil {
		return 0
	}
	i := int(binary.LittleEndian.Uint16(le))
	if i > 32768 {
		return -65536 + i
	}
	return i
}

func (l *LoadSaver) SaveSmallInt(i int) {
	if i <= -32768 || i >= 32768 {
		l.Err = errors.New("int overflows uint16")
		return
	}
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(i))
	l.write(buf)
}

func (l *LoadSaver) LoadInt() int {
	le := l.read(4)
	if le == nil {
		return 0
	}
	i := int64(binary.LittleEndian.Uint32(le))
	if i > 2147483648 {
		return int(-4294967296 + i)
	}
	return int(i)
}

func (l *LoadSaver) SaveInt(i int) {
	if int64(i) <= -2147483648 || int64(i) >= 2147483648 {
		l.Err = errors.New("int overflows uint32")
		return
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(i))
	l.write(buf)
}

func (l *LoadSaver) LoadBytes() []byte {
	i := l.LoadSmallInt()
	if i == 0 {
		return nil
	}
	return l.read(i)
}

func (l *LoadSaver) SaveBytes(b []byte) {
	l.SaveSmallInt(len(b))
	l.write(b)
}

func (l *LoadSaver) LoadString() string {
	return string(l.LoadBytes())
}

func (l *LoadSaver) SaveString(s string) {
	l.SaveBytes([]byte(s))
}

func (l *LoadSaver) LoadStrings() []string {
	ssl := l.LoadSmallInt()
	if ssl == 0 {
		return nil
	}
	ret := make([]string, ssl)
	for i := range ret {
		ret[i] = l.LoadString()
	}
	return ret
}

func (l *LoadSaver) SaveStrings(ss []string) {
	l.SaveSmallInt(len(ss))
	for _, s := range ss {
		l.SaveString(s)
	}
}
