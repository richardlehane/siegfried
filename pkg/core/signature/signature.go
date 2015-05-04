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
	"time"
)

const MAXUINT23 = 256 * 256 * 128 // = approx 8mb address space

func getRef(b []byte) (int, bool) {
	return int(b[2]&^0x80)<<16 | int(b[1])<<8 | int(b[0]), b[2]&0x80 == 0x80
}

func (l *LoadSaver) makeRef(i int, ref bool) []byte {
	if i < 0 || i >= MAXUINT23 {
		l.Err = errors.New("cannot coerce integer to an unsigned 23bit")
		return nil
	}
	b := []byte{byte(i), byte(i >> 8), byte(i >> 16)}
	if ref {
		b[2] = b[2] | 0x80
	}
	return b
}

type LoadSaver struct {
	buf   []byte
	i     int
	uniqs map[string]int
	Err   error
}

func NewLoadSaver(b []byte) *LoadSaver {
	if len(b) == 0 {
		b = make([]byte, 16)
	}
	return &LoadSaver{
		b,
		0,
		make(map[string]int),
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

func (l *LoadSaver) getCollection() []byte {
	if l.Err != nil {
		return nil
	}
	byts := l.get(3)
	i, ref := getRef(byts)
	if ref {
		j, _ := getRef(l.buf[i : i+3])
		return l.buf[i+3 : i+3+j]
	}
	return l.get(i)
}

func (l *LoadSaver) putCollection(b []byte) {
	if l.Err != nil {
		return
	}
	i, ok := l.uniqs[string(append(l.makeRef(len(b), false), b...))]
	if !ok {
		l.put(l.makeRef(len(b), false))
		l.put(b)
		l.uniqs[string(append(l.makeRef(len(b), false), b...))] = l.i - len(b) - 3
	} else {
		l.put(l.makeRef(i, true))
	}
}

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

func (l *LoadSaver) LoadTinyInt() int {
	i := int(l.LoadByte())
	if i > 128 {
		return i - 256
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

func (l *LoadSaver) convertTinyInts(i []int) []byte {
	ret := make([]byte, len(i))
	for j := range ret {
		if i[j] <= -128 || i[j] >= 128 {
			l.Err = errors.New("int overflows byte: need a tiny int")
			return nil
		}
		ret[j] = byte(i[j])
	}
	return ret
}

func makeTinyInts(b []byte) []int {
	ret := make([]int, len(b))
	for i := range ret {
		n := int(b[i])
		if n > 128 {
			n -= 256
		}
		ret[i] = n
	}
	return ret
}

func (l *LoadSaver) LoadTinyInts() []int {
	return makeTinyInts(l.getCollection())
}

func (l *LoadSaver) SaveTinyInts(i []int) {
	l.putCollection(l.convertTinyInts(i))
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

func (l *LoadSaver) convertTinyUInts(i []int) []byte {
	ret := make([]byte, len(i))
	for j := range ret {
		if i[j] < 0 || i[j] >= 256 {
			l.Err = errors.New("int overflows byte: need a tiny uint")
			return nil
		}
		ret[j] = byte(i[j])
	}
	return ret
}

func makeTinyUInts(b []byte) []int {
	ret := make([]int, len(b))
	for i := range ret {
		ret[i] = int(b[i])
	}
	return ret
}

func (l *LoadSaver) LoadTinyUInts() []int {
	return makeTinyUInts(l.getCollection())
}

func (l *LoadSaver) SaveTinyUInts(i []int) {
	l.putCollection(l.convertTinyUInts(i))
}

func (l *LoadSaver) LoadSmallInt() int {
	le := l.get(2)
	if le == nil {
		return 0
	}
	i := int(binary.LittleEndian.Uint16(le))
	if i > 32768 {
		return i - 65536
	}
	return i
}

func (l *LoadSaver) SaveSmallInt(i int) {
	if i <= -32768 || i >= 32768 {
		l.Err = errors.New("int overflows int16")
		return
	}
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(i))
	l.put(buf)
}

func (l *LoadSaver) convertSmallInts(i []int) []byte {
	ret := make([]byte, len(i)*2)
	for j := range i {
		if i[j] <= -32768 || i[j] >= 32768 {
			l.Err = errors.New("int overflows int16")
			return nil
		}
		binary.LittleEndian.PutUint16(ret[j*2:], uint16(i[j]))
	}
	return ret
}

func makeSmallInts(b []byte) []int {
	ret := make([]int, len(b)/2)
	for i := range ret {
		ret[i] = int(binary.LittleEndian.Uint16(b[i*2:]))
		if ret[i] > 32768 {
			ret[i] -= 65536
		}
	}
	return ret
}

func (l *LoadSaver) LoadSmallInts() []int {
	return makeSmallInts(l.getCollection())
}

func (l *LoadSaver) SaveSmallInts(i []int) {
	l.putCollection(l.convertSmallInts(i))
}

func (l *LoadSaver) LoadInt() int {
	le := l.get(4)
	if le == nil {
		return 0
	}
	i := int64(binary.LittleEndian.Uint32(le))
	if i > 2147483648 {
		return int(i - 4294967296)
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
	l.put(buf)
}

func (l *LoadSaver) convertInts(i []int) []byte {
	ret := make([]byte, len(i)*4)
	for j := range i {
		if i[j] <= -2147483648 || i[j] >= 2147483648 {
			l.Err = errors.New("int overflows int32")
			return nil
		}
		binary.LittleEndian.PutUint32(ret[j*4:], uint32(i[j]))
	}
	return ret
}

func makeInts(b []byte) []int {
	ret := make([]int, len(b)/4)
	for i := range ret {
		n := int64(binary.LittleEndian.Uint32(b[i*4:]))
		if n > 2147483648 {
			n -= 4294967296
		}
		ret[i] = int(n)
	}
	return ret
}

func (l *LoadSaver) LoadInts() []int {
	return makeInts(l.getCollection())
}

func (l *LoadSaver) SaveInts(i []int) {
	l.putCollection(l.convertInts(i))
}

func makeBigInts(b []byte) []int64 {
	ret := make([]int64, len(b)/4)
	for i := range ret {
		ret[i] = int64(binary.LittleEndian.Uint32(b[i*4:]))
		if ret[i] > 2147483648 {
			ret[i] -= -4294967296
		}
	}
	return ret
}

func (l *LoadSaver) LoadBigInts() []int64 {
	return makeBigInts(l.getCollection())
}

func (l *LoadSaver) SaveBigInts(i []int64) {
	n := make([]int, len(i))
	for j := range i {
		n[j] = int(i[j])
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
