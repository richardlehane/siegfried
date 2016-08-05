// Copyright 2013 Richard Lehane. All rights reserved.
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

// Package rk is a minimal implemenation of the Rabin-Karp multiple string search algorithm.
//
// Useful for matching fixed length subsequences.
//
// Example:
//   rk := rk.New([][]byte{[]byte("abc"), []byte("cde"), []byte("def")})
//   for result := range rk.Index(bytes.NewBuffer([]byte("abracadabra"))) {
//     fmt.Println(result.Index, "-", result.Offset)
//   }
package rk

import (
	"bytes"
	"fmt"
	"io"
)

const windowSize = 64

type window struct {
	buf [windowSize]byte
	i   int
}

func (w *window) push(b byte) {
	w.buf[w.i] = b
	w.i++
	if w.i == windowSize {
		w.i = 0
	}
}

func (w *window) head(l int) byte {
	c := w.i - l
	if c >= 0 {
		return w.buf[c]
	}
	return w.buf[windowSize+c]
}

func (w *window) equals(seq []byte) bool {
	n := len(seq)
	if w.i >= n {
		return bytes.Equal(w.buf[w.i-n:w.i], seq)
	}
	if w.i == 0 {
		return bytes.Equal(w.buf[windowSize-n:], seq)
	}
	top := n - w.i
	return bytes.Equal(w.buf[windowSize-top:], seq[:top]) && bytes.Equal(w.buf[:w.i], seq[top:])
}

type seq struct {
	val   []byte
	index int
}

type Rk struct {
	l    int
	pow  uint32
	seqs map[uint32]*seq
}

// New creates an Rabin-Karp matcher from a slice of byte slices. N.B. The byte slices must be of equal length and not contain duplicates.
func New(s [][]byte) (*Rk, error) {
	rk := &Rk{
		len(s[0]),
		pow(len(s[0])),
		make(map[uint32]*seq),
	}
	if rk.l > windowSize {
		return nil, fmt.Errorf("Rabin-Karp: input sequence length is too long, max is ", windowSize)
	}
	for i, v := range s {
		if len(v) != rk.l {
			return nil, fmt.Errorf("Rabin-Karp: all input sequences must be of equal length. Bad argument: ", v)
		}
		hash := hash(v)
		if seq1, ok := rk.seqs[hash]; ok {
			if bytes.Equal(v, seq1.val) {
				return nil, fmt.Errorf("Rabin-Karp: all input sequences must be unique. Bad arguments: ", v, seq1.val)
			}
			return nil, fmt.Errorf("Rabin-Karp: hash collision! Collided: ", v, seq1.val)
		}
		rk.seqs[hash] = &seq{v, i}
	}
	return rk, nil
}

// primeRK, pow and hash are copied from implementation in golang strings package
const primeRK = 16777619

func pow(l int) uint32 {
	var pow, sq uint32 = 1, primeRK
	for i := l; i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return pow
}

func hash(seq []byte) uint32 {
	var hash uint32
	for _, b := range seq {
		hash = hash*primeRK + uint32(b)
	}
	return hash
}

// Index returns a channel of results, these contain the indexes (in the list of sequences given to New())
// and offsets (in the input byte slice) of matching sequences.
func (rk Rk) Index(input io.ByteReader) chan Result {
	output := make(chan Result, 20)
	go rk.match(input, output)
	return output
}

// Result contains the index (in the list of sequences given to New()) and offset of matches.
type Result struct {
	Index  int
	Offset int
}

func (rk Rk) match(input io.ByteReader, results chan Result) {
	var h uint32
	var offset int
	win := new(window)

	for i := 0; i < rk.l; i++ {
		offset++
		b, err := input.ReadByte()
		if err != nil {
			close(results)
			return
		}
		win.push(b)
		h = h*primeRK + uint32(b)
	}
	if seq, ok := rk.seqs[h]; ok {
		if win.equals(seq.val) {
			results <- Result{Index: seq.index, Offset: offset - rk.l}
		}
	}

	for c, err := input.ReadByte(); err == nil; c, err = input.ReadByte() {
		offset++
		h *= primeRK
		h += uint32(c)
		h -= rk.pow * uint32(win.head(rk.l))

		win.push(c)

		if seq, ok := rk.seqs[h]; ok {
			if win.equals(seq.val) {
				results <- Result{Index: seq.index, Offset: offset - rk.l}
			}
		}

	}
	close(results)
}
