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

// Package ac is a minimal implemenation of the Aho-Corasick multiple string search algorithm.
//
// This implementation is tuned for fast matching speed. Building the Aho-Corasick tree is relatively slow and memory intensive.
// It only returns the index (within the byte slices that made the tree) and offset of matches.
// For a more fully featured and balanced implementation, use http://godoc.org/code.google.com/p/ahocorasick
//
// Example:
//   ac := ac.New([][]byte{[]byte("ab"), []byte("c"), []byte("def")})
//   for result := range ac.Index(bytes.NewBuffer([]byte("abracadabra"))) {
//     fmt.Println(result.Index, "-", result.Offset)
//   }
package ac

import "io"

type Ac struct {
	val   byte
	trans *trans // the goto function
	fail  *Ac    // the fail function
	out   out    // the output function
}

type trans struct {
	keys  []byte
	gotos *[256]*Ac // the goto function is a pointer to an array of 256 nodes, indexed by the byte val
}

type out [][2]int

func (t *trans) put(b byte, ac *Ac) {
	t.keys = append(t.keys, b)
	t.gotos[b] = ac
}

func (t *trans) get(b byte) (*Ac, bool) {
	node := t.gotos[b]
	if node == nil {
		return node, false
	}
	return node, true
}

func newTrans() *trans { return &trans{keys: make([]byte, 0, 50), gotos: new([256]*Ac)} }

func (o out) contains(i int) bool {
	for _, v := range o {
		if v[0] == i {
			return true
		}
	}
	return false
}

func newNode() *Ac { return &Ac{trans: newTrans(), out: make(out, 0, 10)} }

// New creates an Aho-Corasick tree from a slice of byte slices
func New(seqs [][]byte) *Ac {
	root := newNode()
	root.addGotos(seqs, false)
	root.addFails()
	return root
}

// New creates an Aho-Corasick tree that only has gotos, the fail functions are all set to root.
// Creates a smaller tree if you are only wanting to use the IndexFixed() function.
func NewFixed(seqs [][]byte) *Ac {
	root := newNode()
	root.addGotos(seqs, true)
	root.fail = root
	return root
}

func (root *Ac) addGotos(seqs [][]byte, fixed bool) {
	// iterate through byte sequences adding goto links to the link matrix
	for id, seq := range seqs {
		curr := root
		for _, seqByte := range seq {
			if trans, ok := curr.trans.get(seqByte); ok {
				curr = trans
			} else {
				node := newNode()
				node.val = seqByte
				if fixed {
					node.fail = root
				}
				curr.trans.put(seqByte, node)
				curr = node
			}
		}
		curr.out = append(curr.out, [2]int{id, len(seq)})
	}
}

func (root *Ac) addFails() {
	// root and its children fail to root
	root.fail = root
	for _, k := range root.trans.keys {
		root.trans.gotos[k].fail = root
	}
	// traverse tree in breadth first search adding fails
	queue := make([]*Ac, 0, 50)
	queue = append(queue, root)
	for len(queue) > 0 {
		pop := queue[0]
		for _, key := range pop.trans.keys {
			node := pop.trans.gotos[key]
			queue = append(queue, node)
			// starting from the node's parent, follow the fails back towards root,
			// and stop at the first fail that has a goto to the node's value
			fail := pop.fail
			_, ok := fail.trans.get(node.val)
			for fail != root && !ok {
				fail = fail.fail
				_, ok = fail.trans.get(node.val)
			}
			fnode, ok := fail.trans.get(node.val)
			if ok && fnode != node {
				node.fail = fnode
			} else {
				node.fail = root
			}
			// another traverse back to root following the fails. This time add any unique out functions to the node
			fail = node.fail
			for fail != root {
				for _, id := range fail.out {
					if !node.out.contains(id[0]) {
						node.out = append(node.out, id)
					}
				}
				fail = fail.fail
			}
		}
		queue = queue[1:]
	}
}

// Index returns a channel of results, these contain the indexes (in the list of sequences that made the tree)
// and offsets (in the input byte slice) of matching sequences.
func (ac *Ac) Index(input io.ByteReader) chan Result {
	output := make(chan Result, 20)
	go ac.match(input, output)
	return output
}

// Index returns a channel of results, these contain the indexes (in the list of sequences that made the tree)
// and offsets (in the input byte slice) of matching sequences.
// Has a quit channel that should be closed to signal quit.
func (ac *Ac) IndexQ(input io.ByteReader, quit chan struct{}) chan Result {
	output := make(chan Result, 20)
	go ac.matchQ(input, output, quit)
	return output
}

// IndexFixed returns a channel of indexes (in the list of sequences that made the tree) of matching sequences.
// Fail links are not followed, so all matches have an offset of 0.
func (ac *Ac) IndexFixed(input io.ByteReader) chan int {
	output := make(chan int)
	go ac.fixed(input, output)
	return output
}

// IndexFixed returns a channel of indexes (in the list of sequences that made the tree) of matching sequences.
// Fail links are not followed, so all matches have an offset of 0.
// Has a quit channel that should be closed to signal quit.
func (ac *Ac) IndexFixedQ(input io.ByteReader, quit chan struct{}) chan int {
	output := make(chan int)
	go ac.fixedQ(input, output, quit)
	return output
}

// Result contains the index (in the list of sequences that made the tree) and offset of matches.
type Result struct {
	Index  int
	Offset int
}

func (ac *Ac) match(input io.ByteReader, results chan Result) {
	var offset int
	root := ac
	curr := root

	for c, err := input.ReadByte(); err == nil; c, err = input.ReadByte() {
		offset++
		if trans, ok := curr.trans.get(c); ok {
			curr = trans
		} else {
			for curr != root {
				curr = curr.fail
				if trans, ok := curr.trans.get(c); ok {
					curr = trans
					break
				}
			}
		}
		for _, id := range curr.out {
			results <- Result{Index: id[0], Offset: offset - id[1]}
		}
	}
	close(results)
}

func (ac *Ac) matchQ(input io.ByteReader, results chan Result, quit chan struct{}) {
	var offset int
	root := ac
	curr := root

	for {
		select {
		case <-quit:
			close(results)
			return
		default:
		}
		c, err := input.ReadByte()
		if err != nil {
			break
		}
		offset++
		if trans, ok := curr.trans.get(c); ok {
			curr = trans
		} else {
			for curr != root {
				curr = curr.fail
				if trans, ok := curr.trans.get(c); ok {
					curr = trans
					break
				}
			}
		}
		for _, id := range curr.out {
			results <- Result{Index: id[0], Offset: offset - id[1]}
		}
	}
	close(results)
}

func (ac *Ac) fixed(input io.ByteReader, results chan int) {
	root := ac
	curr := root

	for c, err := input.ReadByte(); err == nil; c, err = input.ReadByte() {
		if trans, ok := curr.trans.get(c); ok {
			curr = trans
			for _, id := range curr.out {
				results <- id[0]
			}
		} else {
			break
		}
	}
	close(results)
}

func (ac *Ac) fixedQ(input io.ByteReader, results chan int, quit chan struct{}) {
	root := ac
	curr := root

	for {
		select {
		case <-quit:
			close(results)
			return
		default:
		}
		c, err := input.ReadByte()
		if err != nil {
			break
		}
		if trans, ok := curr.trans.get(c); ok {
			curr = trans
			for _, id := range curr.out {
				results <- id[0]
			}
		} else {
			break
		}
	}
	close(results)
}
