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

// Package wac is a modified Aho-Corasick multiple string search algorithm.
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
//   seq := wac.Seq{
//     MaxOffsets: []int64{5, -1},
//     Choices: []wac.Choice{
//       wac.Choice{[]byte{'b'},[]byte{'c'},[]byte{'d'}},
//       wac.Choice{[]byte{'a','d'}},
//       wac.Choice{[]byte{'r', 'x'}},
//       []wac.Choice{[]byte{'a'}},
//     }
//   }
//   secondSeq := wac.Seq{
//     MaxOffsets: []int64{0},
//     Choices: []wac.Choice{wac.Choice{[]byte{'b'}}},
//   }
//   w := wac.New([]wac.Seq{seq, secondSeq})
//   for result := range w.Index(bytes.NewBuffer([]byte("abracadabra"))) {
// 	   fmt.Println(result.Index, "-", result.Offset)
//   }

package wac

import (
	"fmt"
	"io"
	"strings"
)

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
func New(seqs []Seq) *Wac {
	wac := new(Wac)
	zero := newNode(newTrans)
	zero.addGotos(seqs, true, newTrans) // TODO: cld use low memory here?
	root := zero.addFails(true, newTrans)
	root.addGotos(seqs, false, newTrans)
	root.addFails(false, nil)
	wac.zero, wac.root = zero, root
	wac.p = newPool(seqs)
	return wac
}

// New Low Mem creates a Wild Aho-Corasick tree with lower memory requirements (single tree, low mem transitions)
func NewLowMem(seqs []Seq) *Wac {
	wac := new(Wac)
	root := newNode(newTransLM)
	root.addGotos(seqs, true, newTransLM)
	root.addFails(false, nil)
	wac.zero, wac.root = root, root
	wac.p = newPool(seqs)
	return wac
}

// Wac is a wild Aho-Corasick tree
type Wac struct {
	zero *node
	root *node
	p    *pool // pool of preconditions
}

type node struct {
	val     byte
	transit transition // the goto function
	fail    *node      // the fail function
	output  []out      // the output function
	outMax  int64
	outMaxL int
}

func newNode(fn transitionFunc) *node { return &node{transit: fn()} }

type out struct {
	max      int64 // maximum offset at which can occur
	seqIndex int   // index within all the Seqs in the Wac
	subIndex int   // index of the Choice within the Seq
	length   int   // length of byte slice
}

func (n *node) contains(o out) bool {
	if n.output == nil {
		return false
	}
	for _, o1 := range n.output {
		if o == o1 {
			return true
		}
	}
	return false
}

func (n *node) addOutput(o out) {
	if n.output == nil {
		n.output = []out{o}
		n.outMax = o.max
		n.outMaxL = o.length
		return
	}
	if n.outMax > -1 && (o.max == -1 || o.max > n.outMax) {
		n.outMax = o.max
	}
	if o.length > n.outMaxL {
		n.outMaxL = o.length
	}
	n.output = append(n.output, o)
}

func (start *node) addGotos(seqs []Seq, zero bool, fn transitionFunc) {
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
					curr = curr.transit.put(byt, fn)
				}
				max := seq.MaxOffsets[i]
				curr.addOutput(out{max, id, i, len(byts)})
			}
		}
	}
}

func (start *node) addFails(zero bool, tfn transitionFunc) *node {
	// root and its children fail to root
	start.fail = start
	start.transit.finalise()
	for i := 0; ; i++ {
		n := start.transit.iter(i)
		if n == nil {
			break
		}
		n.fail = start
		n.transit.finalise()
	}
	// traverse tree in breadth first search adding fails
	queue := make([]*node, 0, 50)
	queue = append(queue, start)
	for len(queue) > 0 {
		pop := queue[0]
		for i := 0; ; i++ {
			node := pop.transit.iter(i)
			if node == nil {
				break
			}
			node.transit.finalise()
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
				if fail.output != nil {
					for _, o := range fail.output {
						if !node.contains(o) {
							node.addOutput(o)
						}
					}
				}
				fail = fail.fail
			}
		}
		queue = queue[1:]
	}
	// for the zero tree, rewrite the fail links so they now point to the root of the main tree
	if zero {
		root := newNode(tfn)
		start.fail = root
		for i := 0; ; i++ {
			n := start.transit.iter(i)
			if n == nil {
				break
			}
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
func (wac *Wac) Index(input io.ByteReader) chan Result {
	output := make(chan Result)
	go wac.match(input, output)
	return output
}

// Result contains the index and offset of matches.
type Result struct {
	Index  [2]int // a double index: index of the Seq and index of the Choice
	Offset int64
	Length int
}

func (wac *Wac) match(input io.ByteReader, results chan Result) {
	var offset int64
	var progressResult = Result{Index: [2]int{-1, -1}}
	precons := wac.p.get()
	curr := wac.zero
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
		if offset >= 1024 && offset&(^offset+1) == offset {
			progressResult.Offset = offset
			results <- progressResult
		}
	}
	wac.p.put(precons)
	close(results)
}
