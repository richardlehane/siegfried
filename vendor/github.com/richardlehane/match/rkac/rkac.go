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

// Package rkac is an implemenation of the Rabin-Karp multiple string search algorithm.
// The sequences are stored in a tree similar to an Aho-Corasick tree. This enables variable
// length inputs.
//
// Example:
//   rk := rkac.New([][]byte{[]byte("abra"), []byte("cd"), []byte("bra")})
//   for result := range rkac.Index(bytes.NewBuffer([]byte("abracadabra"))) {
//     fmt.Println(result.Index, "-", result.Offset)
//   }

// To do: add a reverse mode
package rkac

import (
	"fmt"
	"io"
)

const windowSize = 256

type window struct {
	buf      [windowSize]byte
	marks    [windowSize][]int
	i        int
	numMarks int
}

func newWindow() *window {
	win := new(window)
	for i, _ := range win.marks {
		win.marks[i] = make([]int, 0, 5)
	}
	return win
}

func (w *window) mark(wl, l int) {
	idx := l - wl - 1 + w.i
	if idx >= windowSize {
		idx = idx - windowSize
	}
	w.marks[idx] = append(w.marks[idx], l)
	w.numMarks++
}

func (w *window) push(b byte) ([]int, bool) {
	var ok bool
	w.buf[w.i] = b
	marks := w.marks[w.i]
	if len(marks) > 0 {
		ok = true
	}
	w.i++
	if w.i == windowSize {
		w.i = 0
	}
	return marks, ok
}

func (w *window) clear() {
	idx := w.i
	if idx == 0 {
		idx = windowSize - 1
	} else {
		idx--
	}
	w.marks[idx] = w.marks[idx][:0]
}

func (w *window) flush() []int {
	flushes := make([]int, 0, 10)
	iter, idx := 1, w.i
	for w.numMarks > 0 {
		marks := w.marks[idx]
		for _, l := range marks {
			flushes = append(flushes, l-iter)
			w.numMarks--
		}
		iter++
		idx++
		if idx == windowSize {
			idx = 0
		}
	}
	return flushes
}

func (w *window) read(l int) byte {
	c := w.i - l
	if c >= 0 {
		return w.buf[c]
	}
	return w.buf[windowSize+c]
}

type ac struct {
	val   byte
	trans *trans // the goto function
	out   out    // the output function
}

type trans struct {
	gotos *[256]*ac // the goto function is a pointer to an array of 256 nodes, indexed by the byte val
}

type out [][2]int

func (t *trans) put(b byte, ac *ac) {
	t.gotos[b] = ac
}

func (t *trans) get(b byte) (*ac, bool) {
	node := t.gotos[b]
	if node == nil {
		return node, false
	}
	return node, true
}

func newTrans() *trans { return &trans{new([256]*ac)} }

func (o out) contains(i int) bool {
	for _, v := range o {
		if v[0] == i {
			return true
		}
	}
	return false
}

func newNode() *ac { return &ac{trans: newTrans(), out: make(out, 0, 10)} }

func (root *ac) addGotos(seq []byte, id int) {
	// iterate through byte sequences adding goto links to the link matrix
	curr := root
	for _, seqByte := range seq {
		if trans, ok := curr.trans.get(seqByte); ok {
			curr = trans
		} else {
			node := newNode()
			node.val = seqByte
			curr.trans.put(seqByte, node)
			curr = node
		}
	}
	curr.out = append(curr.out, [2]int{id, len(seq)})
}

type Rkac struct {
	min  int
	max  int
	ac   *ac
	pow  uint32
	seqs map[uint32]int
}

// New creates a hybrid Rabin-Karp/ Aho Corasick matcher from a slice of byte slices.
func New(s [][]byte) (*Rkac, error) {
	if len(s) <= 0 {
		return nil, fmt.Errorf("Rkac: no byte slices provided")
	}
	rk := new(Rkac)
	rk.min, rk.max = len(s[0]), len(s[0])
	for _, seq := range s {
		if rk.min > len(seq) {
			rk.min = len(seq)
		}
		if rk.max < len(seq) {
			rk.max = len(seq)
		}
	}
	if rk.max > windowSize {
		return nil, fmt.Errorf("Rkac: input sequence max length is too long, expected %v, got %v", windowSize, rk.max)
	}
	rk.pow = pow(rk.min)
	rk.ac = newNode()
	rk.seqs = make(map[uint32]int)
	for i, seq := range s {
		rk.ac.addGotos(seq, i)
		hash := hash(seq[:rk.min])
		if l, ok := rk.seqs[hash]; ok {
			if l < len(seq) {
				rk.seqs[hash] = len(seq)
			}
		} else {
			rk.seqs[hash] = len(seq)
		}
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
func (rk *Rkac) Index(input io.ByteReader) chan Result {
	output := make(chan Result, 20)
	go rk.match(input, output)
	return output
}

// Result contains the index (in the list of sequences given to New()) and offset of matches.
type Result struct {
	Index  int
	Offset int
}

func (rk *Rkac) match(input io.ByteReader, results chan Result) {
	var h uint32
	var offset int

	win := newWindow()

	for i := 0; i < rk.min; i++ {
		offset++
		b, err := input.ReadByte()
		if err != nil {
			close(results)
			return
		}
		win.push(b)
		h = h*primeRK + uint32(b)
	}
	if l, ok := rk.seqs[h]; ok {
		if l == rk.min {
			rk.index(win, results, l, offset-l)
		} else {
			win.mark(rk.min, l)

		}
	}

	for c, err := input.ReadByte(); err == nil; c, err = input.ReadByte() {
		offset++
		h *= primeRK
		h += uint32(c)
		h -= rk.pow * uint32(win.read(rk.min))

		if ls, ok := win.push(c); ok {
			for _, l := range ls {
				rk.index(win, results, l, offset-l)
				win.numMarks--
			}
			win.clear()
		}
		if l, ok := rk.seqs[h]; ok {
			if l == rk.min {
				rk.index(win, results, l, offset-l)
			} else {
				win.mark(rk.min, l)

			}
		}

	}
	ls := win.flush()
	for _, l := range ls {
		rk.index(win, results, l, offset-l)
	}

	close(results)
}

func (rk *Rkac) index(win *window, results chan Result, idx, offset int) {
	curr := rk.ac
	for ; idx >= 0; idx-- {
		c := win.read(idx)
		if trans, ok := curr.trans.get(c); ok {
			curr = trans
			for _, id := range curr.out {
				results <- Result{Index: id[0], Offset: offset}
			}
		} else {
			break
		}
	}
}
