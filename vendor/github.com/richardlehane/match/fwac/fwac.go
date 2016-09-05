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

// Package fwac is a modified Aho-Corasick multiple string search algorithm.
// This package differs from the wac package by removing interfaces from the hot loop. The wac package
// should be preferred unless performance is vital.
//
// This algorithm allows for sequences that are composed of sub-sequences
// that can contain an arbitrary number of wildcards. Sequences can also be
// given a maximum offset that defines the maximum byte position of the first sub-sequence.
//
// The results returned are for the matches on subsequences (NOT the full sequences).
// The index of those subsequences and the offset is returned.
// It is up to clients to verify that the complete sequence that they are interested in has matched.
// A "progress" result is sent from offset 1024 onwards. This is to update clients on scanning progress and has index -1,-1.
// This result is sent on powers of two (1024, 2048, 4096, etc.)

// Example usage:
//
//   seq := fwac.Seq{
//     MaxOffsets: []int64{5, -1},
//     Choices: []fwac.Choice{
//       fwac.Choice{[]byte{'b'},[]byte{'c'},[]byte{'d'}},
//       fwac.Choice{[]byte{'a','d'}},
//       fwac.Choice{[]byte{'r', 'x'}},
//       []fwac.Choice{[]byte{'a'}},
//     }
//   }
//   secondSeq := fwac.Seq{
//     MaxOffsets: []int64{0},
//     Choices: []wac.Choice{fwac.Choice{[]byte{'b'}}},
//   }
//   w := fwac.New([]fwac.Seq{seq, secondSeq})
//   for result := range w.Index(bytes.NewBuffer([]byte("abracadabra"))) {
// 	   fmt.Println(result.Index, "-", result.Offset)
//   }

package fwac

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type Wac interface {
	Index(io.ByteReader) chan Result
}

// Result contains the index and offset of matches.
type Result struct {
	Index  [2]int // a double index: index of the Seq and index of the Choice
	Offset int64
	Length int
}

// Choice represents the different byte slices that can occur at each position of the Seq
type Choice [][]byte

// Seq is an ordered set of slices of Choices, with maximum offsets for each choice
type Seq struct {
	MaxOffsets []int64 // maximum offsets for each choice. Can be -1 for wildcard.
	Choices    []Choice
}

func (s Seq) String() string {
	str := "{Offsets:"
	for n, v := range s.MaxOffsets {
		if n > 0 {
			str += ","
		}
		str += fmt.Sprintf(" %d", v)
	}
	str += "; Choices:"
	for n, v := range s.Choices {
		if n > 0 {
			str += ","
		}
		str += " ["
		strs := make([]string, len(v))
		for i := range v {
			strs[i] = string(v[i])
		}
		str += strings.Join(strs, " | ")
		str += "]"
	}
	return str + "}"
}

// New creates an Wild Aho-Corasick tree
func New(seqs []Seq) Wac {
	zero := &node{keys: make([]byte, 0, 1)}
	zero.addGotos(seqs, true) // TODO: cld use low memory here?
	root := zero.addFails(true)
	root.addGotos(seqs, false)
	root.addFails(false)
	return &fwac{
		zero: zero,
		root: root,
		p:    newPool(seqs),
	}
}

// New Low Mem creates a Wild Aho-Corasick tree with lower memory requirements (single tree, low mem transitions)
func NewLowMem(seqs []Seq) Wac {
	root := &nodelm{}
	root.addGotos(seqs, true)
	root.addFails(false)
	return &fwaclm{
		root: root,
		p:    newPool(seqs),
	}
}

// fwac is a wild Aho-Corasick tree
type fwac struct {
	zero *node
	root *node
	p    *pool // pool of preconditions
}

// fwaclm is a wild Aho-Corasick tree that takes less RAM
type fwaclm struct {
	root *nodelm
	p    *pool // pool of preconditions
}

// Nodes

// out function is shared
type out struct {
	max      int64 // maximum offset at which can occur
	seqIndex int   // index within all the Seqs in the Wac
	subIndex int   // index of the Choice within the Seq
	length   int   // length of byte slice
}

func contains(op []out, o out) bool {
	if op == nil {
		return false
	}
	for _, o1 := range op {
		if o == o1 {
			return true
		}
	}
	return false
}

func addOutput(op []out, o out, outMax int64, outMaxL int) ([]out, int64, int) {
	if op == nil {
		return []out{o}, o.max, o.length
	}
	if outMax > -1 && (o.max == -1 || o.max > outMax) {
		outMax = o.max
	}
	if o.length > outMaxL {
		outMaxL = o.length
	}
	return append(op, o), outMax, outMaxL
}

// regular node
type node struct {
	val     byte
	keys    []byte
	transit [256]*node // the goto function
	fail    *node      // the fail function
	output  []out      // the output function
	outMax  int64
	outMaxL int
}

func (start *node) addGotos(seqs []Seq, zero bool) {
	// iterate through byte sequences adding goto links to the link matrix
	for id, seq := range seqs {
		for i, choice := range seq.Choices {
			// skip the first choice set if this isn't the zero tree and it is at 0 offset
			if !zero && i == 0 && seq.MaxOffsets[0] == 0 {
				continue
			}
			for _, byts := range choice {
				curr := start
				for _, byt := range byts {
					if curr.transit[byt] == nil {
						curr.transit[byt] = &node{
							val:  byt,
							keys: make([]byte, 0, 1),
						}
						curr.keys = append(curr.keys, byt)
					}
					curr = curr.transit[byt]
				}
				curr.output, curr.outMax, curr.outMaxL = addOutput(
					curr.output,
					out{seq.MaxOffsets[i], id, i, len(byts)},
					curr.outMax,
					curr.outMaxL)
			}
		}
	}
}

func (start *node) addFails(zero bool) *node {
	// root and its children fail to root
	start.fail = start
	for _, byt := range start.keys {
		start.transit[byt].fail = start
	}
	// traverse tree in breadth first search adding fails
	queue := make([]*node, 0, 50)
	queue = append(queue, start)
	for len(queue) > 0 {
		pop := queue[0]
		for _, byt := range pop.keys {
			node := pop.transit[byt]
			queue = append(queue, node)
			// starting from the node's parent, follow the fails back towards root,
			// and stop at the first fail that has a goto to the node's value
			fail := pop.fail
			ok := fail.transit[node.val]
			for fail != start && ok == nil {
				fail = fail.fail
				ok = fail.transit[node.val]
			}
			fnode := fail.transit[node.val]
			if fnode != nil && fnode != node {
				node.fail = fnode
			} else {
				node.fail = start
			}
			// another traverse back to root following the fails. This time add any unique out functions to the node
			fail = node.fail
			for fail != start {
				for _, o := range fail.output {
					if !contains(node.output, o) {
						node.output, node.outMax, node.outMaxL = addOutput(node.output, o, node.outMax, node.outMaxL)
					}
				}
				fail = fail.fail
			}
		}
		queue = queue[1:]
	}
	// for the zero tree, rewrite the fail links so they now point to the root of the main tree
	if zero {
		root := &node{keys: make([]byte, 0, 1)}
		start.fail = root
		for _, byt := range start.keys {
			start.transit[byt].fail = root
		}
		return root
	}
	return start
}

type nodelm struct {
	val     byte
	transit transLM // the goto function
	fail    *nodelm // the fail function
	output  []out   // the output function
	outMax  int64
	outMaxL int
}

// The low memory transition uses a slice of nodes with binary search. It is modelled on: https://code.google.com/p/ahocorasick/source/browse/aho.go
type link struct {
	b byte
	n *nodelm
}

type transLM []*link

func (t transLM) Len() int {
	return len(t)
}
func (t transLM) Less(i, j int) bool {
	return t[i].b < t[j].b
}
func (t transLM) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t transLM) get(b byte) *nodelm {
	top, bottom := len(t), 0
	for top > bottom {
		i := (top-bottom)/2 + bottom
		b2 := t[i].b
		if b2 > b {
			top = i
		} else if b2 < b {
			bottom = i + 1
		} else {
			return t[i].n
		}
	}
	return nil
}

func (start *nodelm) addGotos(seqs []Seq, zero bool) {
	// iterate through byte sequences adding goto links to the link matrix
	for id, seq := range seqs {
		for i, choice := range seq.Choices {
			// skip the first choice set if this isn't the zero tree and it is at 0 offset
			if !zero && i == 0 && seq.MaxOffsets[0] == 0 {
				continue
			}
			for _, byts := range choice {
				curr := start
				for _, byt := range byts {
					var n *nodelm
					for _, l := range curr.transit {
						if l.b == byt {
							n = l.n
							break
						}
					}
					if n == nil {
						n = &nodelm{
							val: byt,
						}
						curr.transit = append(curr.transit, &link{byt, n})
					}
					curr = n
				}
				curr.output, curr.outMax, curr.outMaxL = addOutput(
					curr.output,
					out{seq.MaxOffsets[i], id, i, len(byts)},
					curr.outMax,
					curr.outMaxL)
			}
		}
	}
}

func (start *nodelm) addFails(zero bool) *nodelm {
	// root and its children fail to root
	start.fail = start
	sort.Sort(start.transit)
	for _, l := range start.transit {
		n := l.n
		n.fail = start
		sort.Sort(n.transit)
	}
	// traverse tree in breadth first search adding fails
	queue := make([]*nodelm, 0, 50)
	queue = append(queue, start)
	for len(queue) > 0 {
		pop := queue[0]
		for _, l := range pop.transit {
			node := l.n
			sort.Sort(node.transit)
			queue = append(queue, node)
			// starting from the node's parent, follow the fails back towards root,
			// and stop at the first fail that has a goto to the node's value
			fail := pop.fail
			ok := fail.transit.get(node.val)
			for fail != start && ok == nil {
				fail = fail.fail
				ok = fail.transit.get(node.val)
			}
			fnode := fail.transit.get(node.val)
			if fnode != nil && fnode != node {
				node.fail = fnode
			} else {
				node.fail = start
			}
			// another traverse back to root following the fails. This time add any unique out functions to the node
			fail = node.fail
			for fail != start {
				for _, o := range fail.output {
					if !contains(node.output, o) {
						node.output, node.outMax, node.outMaxL = addOutput(node.output, o, node.outMax, node.outMaxL)
					}
				}
				fail = fail.fail
			}
		}
		queue = queue[1:]
	}
	// for the zero tree, rewrite the fail links so they now point to the root of the main tree
	if zero {
		root := &nodelm{}
		start.fail = root
		for _, l := range start.transit {
			n := l.n
			n.fail = root
		}
		return root
	}
	return start
}

// preconditions ensure that subsequent (>0) Choices in a Seq are only sent when previous Choices have already matched
// previous matches are stored as offsets to prevent overlapping matches resulting in false positives
type precons [][]int64

func newPrecons(t []int) precons {
	p := make([][]int64, len(t))
	for i, v := range t {
		p[i] = make([]int64, v)
	}
	return p
}

func clear(p precons) precons {
	for i := range p {
		for j := range p[i] {
			p[i][j] = 0
		}
	}
	return p
}

// Index returns a channel of results, these contain the indexes (a double index: index of the Seq and index of the Choice)
// and offsets (in the input byte slice) of matching sequences.
func (wac *fwac) Index(input io.ByteReader) chan Result {
	output := make(chan Result)
	go wac.match(input, output)
	return output
}

func (wac *fwac) match(input io.ByteReader, results chan Result) {
	var offset int64
	var progressResult = Result{Index: [2]int{-1, -1}}
	precons := wac.p.get()
	curr := wac.zero
	for c, err := input.ReadByte(); err == nil; c, err = input.ReadByte() {
		offset++
		if trans := curr.transit[c]; trans != nil {
			curr = trans
		} else {
			for curr != wac.root {
				curr = curr.fail
				if trans := curr.transit[c]; trans != nil {
					curr = trans
					break
				}

			}
		}
		if curr.output != nil && (curr.outMax == -1 || curr.outMax >= offset-int64(curr.outMaxL)) {
			for _, o := range curr.output {
				if o.max == -1 || o.max >= offset-int64(o.length) {
					if o.subIndex == 0 || (precons[o.seqIndex][o.subIndex-1] != 0 && offset-int64(o.length) >= precons[o.seqIndex][o.subIndex-1]) {
						if precons[o.seqIndex][o.subIndex] == 0 {
							precons[o.seqIndex][o.subIndex] = offset
						}
						results <- Result{Index: [2]int{o.seqIndex, o.subIndex}, Offset: offset - int64(o.length), Length: o.length}
					}
				}
			}
		}
		if offset&(^offset+1) == offset && offset >= 1024 { // send powers of 2 greater than 512
			progressResult.Offset = offset
			results <- progressResult
		}
	}
	wac.p.put(precons)
	close(results)
}

// Index returns a channel of results, these contain the indexes (a double index: index of the Seq and index of the Choice)
// and offsets (in the input byte slice) of matching sequences.
func (wac *fwaclm) Index(input io.ByteReader) chan Result {
	output := make(chan Result)
	go wac.match(input, output)
	return output
}

func (wac *fwaclm) match(input io.ByteReader, results chan Result) {
	var offset int64
	var progressResult = Result{Index: [2]int{-1, -1}}
	precons := wac.p.get()
	curr := wac.root
	for c, err := input.ReadByte(); err == nil; c, err = input.ReadByte() {
		offset++
		if trans := curr.transit.get(c); trans != nil {
			curr = trans
		} else {
			for curr != wac.root {
				curr = curr.fail
				if trans := curr.transit.get(c); trans != nil {
					curr = trans
					break
				}

			}
		}
		if curr.output != nil && (curr.outMax == -1 || curr.outMax >= offset-int64(curr.outMaxL)) {
			for _, o := range curr.output {
				if o.max == -1 || o.max >= offset-int64(o.length) {
					if o.subIndex == 0 || (precons[o.seqIndex][o.subIndex-1] != 0 && offset-int64(o.length) >= precons[o.seqIndex][o.subIndex-1]) {
						if precons[o.seqIndex][o.subIndex] == 0 {
							precons[o.seqIndex][o.subIndex] = offset
						}
						results <- Result{Index: [2]int{o.seqIndex, o.subIndex}, Offset: offset - int64(o.length), Length: o.length}
					}
				}
			}
		}
		if offset&(^offset+1) == offset && offset >= 1024 { // send powers of 2 greater than 512
			progressResult.Offset = offset
			results <- progressResult
		}
	}
	wac.p.put(precons)
	close(results)
}
