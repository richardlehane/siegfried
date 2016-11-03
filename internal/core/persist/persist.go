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

// Package persist marshals and unmarshals siegfried signatures as binary data
package persist

import (
	"encoding/binary"
	"errors"
	"time"
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

func (l *LoadSaver) get(i int) []byte {
	if l.Err != nil || i == 0 {
		return nil
	}
	if l.i+i > len(l.buf) {
		l.Err = errors.New("error loading signature file, overflowed")
		return nil
	}
	l.i += i
	return l.buf[l.i-i : l.i]
}

func (l *LoadSaver) put(b []byte) {
	if l.Err != nil || len(b) == 0 {
		return
	}
	if len(b)+l.i > len(l.buf) {
		nbuf := make([]byte, (len(b)+l.i)*2)
		copy(nbuf, l.buf[:l.i])
		l.buf = nbuf
	}
	copy(l.buf[l.i:len(b)+l.i], b)
	l.i += len(b)
}

const (
	_int8 byte = iota
	_uint8
	_int16
	_uint16
	_int32
	_uint32
)

const (
	min8, max8   = -128, 128
	maxu8        = 256
	min16, max16 = -32768, 32768
	maxu16       = 65536
	min32, max32 = -2147483648, 2147483648
	maxu32       = 4294967296
	maxu23       = 256 * 256 * 128 // used by collection refs = approx 8mb address space
)

func (l *LoadSaver) LoadByte() byte {
	le := l.get(1)
	if le == nil {
		return 0
	}
	return le[0]
}

func (l *LoadSaver) SaveByte(b byte) {
	l.put([]byte{b})
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

const (
	_a = 1 << iota
	_b
	_c
	_d
	_e
	_f
	_g
	_h
)

func (l *LoadSaver) LoadBoolField() (a bool, b bool, c bool, d bool, e bool, f bool, g bool, h bool) {
	byt := l.LoadByte()
	if byt&_a == _a {
		a = true
	}
	if byt&_b == _b {
		b = true
	}
	if byt&_c == _c {
		c = true
	}
	if byt&_d == _d {
		d = true
	}
	if byt&_e == _e {
		e = true
	}
	if byt&_f == _f {
		f = true
	}
	if byt&_g == _g {
		g = true
	}
	if byt&_h == _h {
		h = true
	}
	return
}

func (l *LoadSaver) SaveBoolField(a bool, b bool, c bool, d bool, e bool, f bool, g bool, h bool) {
	var byt byte
	if a {
		byt |= _a
	}
	if b {
		byt |= _b
	}
	if c {
		byt |= _c
	}
	if d {
		byt |= _d
	}
	if e {
		byt |= _e
	}
	if f {
		byt |= _f
	}
	if g {
		byt |= _g
	}
	if h {
		byt |= _h
	}
	l.SaveByte(byt)
}

func (l *LoadSaver) LoadTinyInt() int {
	i := int(l.LoadByte())
	if i > max8 {
		return i - maxu8
	}
	return i
}

func (l *LoadSaver) SaveTinyInt(i int) {
	if i <= min8 || i >= max8 {
		l.Err = errors.New("int overflows byte")
		return
	}
	l.SaveByte(byte(i))
}

func (l *LoadSaver) LoadTinyUInt() int {
	return int(l.LoadByte())
}

func (l *LoadSaver) SaveTinyUInt(i int) {
	if i < 0 || i >= maxu8 {
		l.Err = errors.New("int overflows byte as a uint")
		return
	}
	l.SaveByte(byte(i))
}

func (l *LoadSaver) LoadSmallInt() int {
	le := l.get(2)
	if le == nil {
		return 0
	}
	i := int(binary.LittleEndian.Uint16(le))
	if i > max16 {
		return i - maxu16
	}
	return i
}

func (l *LoadSaver) SaveSmallInt(i int) {
	if i <= min16 || i >= max16 {
		l.Err = errors.New("int overflows int16")
		return
	}
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(i))
	l.put(buf)
}

func (l *LoadSaver) LoadInt() int {
	le := l.get(4)
	if le == nil {
		return 0
	}
	i := int64(binary.LittleEndian.Uint32(le))
	if i > max32 {
		return int(i - maxu32)
	}
	return int(i)
}

func (l *LoadSaver) SaveInt(i int) {
	if int64(i) <= min32 || int64(i) >= max32 {
		l.Err = errors.New("int overflows uint32")
		return
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(i))
	l.put(buf)
}

func (l *LoadSaver) getCollection() []byte {
	if l.Err != nil {
		return nil
	}
	le := l.LoadSmallInt()
	return l.get(le)
}

func (l *LoadSaver) putCollection(b []byte) {
	if l.Err != nil {
		return
	}
	l.SaveSmallInt(len(b))
	l.put(b)
}

func characterise(is []int) (byte, error) {
	f := func(i, max int, sign bool) (int, bool) {
		if i < 0 {
			sign = true
			i *= -1
		}
		if i > max {
			return i, sign
		}
		return max, sign
	}
	var m int
	var s bool
	for _, v := range is {
		m, s = f(v, m, s)
	}
	switch {
	case m < max8:
		return _int8, nil
	case m < maxu8 && !s:
		return _uint8, nil
	case m < max16:
		return _int16, nil
	case m < maxu16 && !s:
		return _uint16, nil
	case int64(m) < max32:
		return _int32, nil
	case int64(m) < maxu32 && !s:
		return _uint32, nil
	default:
		return 0, errors.New("integer overflow when building signature - need 64 bit int types!")
	}
}

func (l *LoadSaver) convertInts(is []int) []byte {
	if len(is) == 0 {
		return nil
	}
	typ, err := characterise(is)
	if err != nil {
		l.Err = err
		return nil
	}
	var ret []byte
	switch typ {
	case _int8, _uint8:
		ret = make([]byte, len(is))
		for i := range ret {
			ret[i] = byte(is[i])
		}
	case _int16, _uint16:
		ret = make([]byte, len(is)*2)
		for i := range is {
			binary.LittleEndian.PutUint16(ret[i*2:], uint16(is[i]))
		}
	case _int32, _uint32:
		ret = make([]byte, len(is)*4)
		for i := range is {
			binary.LittleEndian.PutUint32(ret[i*4:], uint32(is[i]))
		}
	}
	return append([]byte{typ}, ret...)
}

func makeInts(b []byte) []int {
	if len(b) == 0 {
		return nil
	}
	var ret []int
	typ := b[0]
	b = b[1:]
	switch typ {
	case _int8:
		ret = make([]int, len(b))
		for i := range ret {
			ret[i] = int(b[i])
			if ret[i] > max8 {
				ret[i] -= maxu8
			}
		}
	case _uint8:
		ret = make([]int, len(b))
		for i := range ret {
			ret[i] = int(b[i])
		}
	case _int16:
		ret = make([]int, len(b)/2)
		for i := range ret {
			ret[i] = int(binary.LittleEndian.Uint16(b[i*2:]))
			if ret[i] > max16 {
				ret[i] -= maxu16
			}
		}
	case _uint16:
		ret = make([]int, len(b)/2)
		for i := range ret {
			ret[i] = int(binary.LittleEndian.Uint16(b[i*2:]))
		}
	case _int32:
		ret = make([]int, len(b)/4)
		for i := range ret {
			n := int64(binary.LittleEndian.Uint32(b[i*4:]))
			if n > max32 {
				n -= maxu32
			}
			ret[i] = int(n)
		}
	case _uint32:
		ret = make([]int, len(b)/4)
		for i := range ret {
			ret[i] = int(binary.LittleEndian.Uint32(b[i*4:]))
		}
	}
	return ret
}

func (l *LoadSaver) LoadInts() []int {
	return makeInts(l.getCollection())
}

func (l *LoadSaver) SaveInts(i []int) {
	l.putCollection(l.convertInts(i))
}

func (l *LoadSaver) LoadBigInts() []int64 {
	is := makeInts(l.getCollection())
	if is == nil {
		return nil
	}
	ret := make([]int64, len(is))
	for i := range is {
		ret[i] = int64(is[i])
	}
	return ret
}

func (l *LoadSaver) SaveBigInts(is []int64) {
	n := make([]int, len(is))
	for i := range is {
		n[i] = int(is[i])
	}
	l.SaveInts(n)
}

func (l *LoadSaver) LoadBytes() []byte {
	return l.getCollection()
}

func (l *LoadSaver) SaveBytes(b []byte) {
	l.putCollection(b)
}

func (l *LoadSaver) LoadString() string {
	return string(l.getCollection())
}

func (l *LoadSaver) SaveString(s string) {
	l.putCollection([]byte(s))
}

func (l *LoadSaver) LoadStrings() []string {
	le := l.LoadSmallInt()
	if le == 0 {
		return nil
	}
	ret := make([]string, le)
	for i := range ret {
		ret[i] = string(l.getCollection())
	}
	return ret
}

func (l *LoadSaver) SaveStrings(ss []string) {
	l.SaveSmallInt(len(ss))
	for _, s := range ss {
		l.putCollection([]byte(s))
	}
}

func (l *LoadSaver) SaveTime(t time.Time) {
	byts, err := t.MarshalBinary()
	if err != nil {
		l.Err = err
		return
	}
	l.put(byts)
}

func (l *LoadSaver) LoadTime() time.Time {
	buf := l.get(15)
	t := &time.Time{}
	l.Err = t.UnmarshalBinary(buf)
	return *t
}

func (l *LoadSaver) SaveFourCC(cc [4]byte) {
	l.put(cc[:])
}

func (l *LoadSaver) LoadFourCC() [4]byte {
	buf := l.get(4)
	var ret [4]byte
	copy(ret[:], buf)
	return ret
}
