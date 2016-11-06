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

// Package bytematcher builds a matching engine from a set of signatures and performs concurrent matching against an input siegreader.Buffer.
package bytematcher

import (
	"fmt"
	"sync"

	wac "github.com/richardlehane/match/fwac"
	"github.com/richardlehane/siegfried/core"
	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/internal/siegreader"
)

// Matcher matches byte signatures against the siegreader.Buffer.
type Matcher struct {
	// the following fields are persisted
	keyFrames  [][]keyFrame
	tests      []*testTree
	bofFrames  *frameSet
	eofFrames  *frameSet
	bofSeq     *seqSet
	eofSeq     *seqSet
	knownBOF   int
	knownEOF   int
	maxBOF     int
	maxEOF     int
	priorities *priority.Set
	// remaining fields are not persisted
	mu     *sync.Mutex
	bAho   wac.Wac
	eAho   wac.Wac
	lowmem bool
}

// SignatureSet for a bytematcher is a slice of frames.Signature.
type SignatureSet []frames.Signature

// Load loads a Matcher.
func Load(ls *persist.LoadSaver) core.Matcher {
	if !ls.LoadBool() {
		return nil
	}
	return &Matcher{
		keyFrames:  loadKeyFrames(ls),
		tests:      loadTests(ls),
		bofFrames:  loadFrameSet(ls),
		eofFrames:  loadFrameSet(ls),
		bofSeq:     loadSeqSet(ls),
		eofSeq:     loadSeqSet(ls),
		knownBOF:   ls.LoadInt(),
		knownEOF:   ls.LoadInt(),
		maxBOF:     ls.LoadInt(),
		maxEOF:     ls.LoadInt(),
		priorities: priority.Load(ls),
		mu:         &sync.Mutex{},
	}
}

// Save persists a Matcher.
func Save(c core.Matcher, ls *persist.LoadSaver) {
	if c == nil {
		ls.SaveBool(false)
		return
	}
	b := c.(*Matcher)
	ls.SaveBool(true)
	saveKeyFrames(ls, b.keyFrames)
	saveTests(ls, b.tests)
	b.bofFrames.save(ls)
	b.eofFrames.save(ls)
	b.bofSeq.save(ls)
	b.eofSeq.save(ls)
	ls.SaveInt(b.knownBOF)
	ls.SaveInt(b.knownEOF)
	ls.SaveInt(b.maxBOF)
	ls.SaveInt(b.maxEOF)
	b.priorities.Save(ls)
}

type sigErrors []error

func (se sigErrors) Error() string {
	str := "bytematcher.Signatures errors:"
	for _, v := range se {
		str += v.Error()
		str += "\n"
	}
	return str
}

// Add a set of signatures to a bytematcher.
// The priority list should be of equal length to the signatures, or nil (if no priorities are to be set).
//
// Example:
//   m, n, err := Add(bm, []frames.Signature{frames.Signature{frames.NewFrame(frames.BOF, patterns.Sequence{'p','d','f'}, 0, 0)}}, nil)
func Add(c core.Matcher, ss core.SignatureSet, priorities priority.List) (core.Matcher, int, error) {
	var b *Matcher
	if c == nil {
		b = &Matcher{
			bofFrames:  &frameSet{},
			eofFrames:  &frameSet{},
			bofSeq:     &seqSet{},
			eofSeq:     &seqSet{},
			priorities: &priority.Set{},
			mu:         &sync.Mutex{},
		}
	} else {
		b = c.(*Matcher)
	}
	sigs, ok := ss.(SignatureSet)
	if !ok {
		return nil, -1, fmt.Errorf("Byte matcher: can't convert signature set to BM signature set")
	}
	var se sigErrors
	// process each of the sigs, adding them to b.Sigs and the various seq/frame/testTree sets
	var bof, eof int
	for _, sig := range sigs {
		if err := b.addSignature(sig); err == nil {
			// get the local max bof and eof by popping last keyframe and testing
			kf := b.keyFrames[len(b.keyFrames)-1]
			bof, eof = maxBOF(bof, kf), maxEOF(eof, kf)
		} else {
			se = append(se, err)
		}
	}
	if len(se) > 0 {
		return nil, -1, se
	}
	// set the maximum distances for this test tree so can properly size slices for matching
	for _, t := range b.tests {
		t.maxLeftDistance = maxLength(t.left)
		t.maxRightDistance = maxLength(t.right)
	}
	// add the priorities to the priority set
	b.priorities.Add(priorities, len(sigs), bof, eof)
	return b, len(b.keyFrames), nil
}

// Identify matches a Matcher's signatures against the input siegreader.Buffer.
// Results are passed on the returned channel.
//
// Example:
//   ret := bm.Identify("", buf)
//   for v := range ret {
//     if v.Index() == 0 {
//       fmt.Print("Success! It is signature 0!")
//     }
//   }
func (b *Matcher) Identify(name string, sb *siegreader.Buffer, exclude ...int) (chan core.Result, error) {
	quit, ret := make(chan struct{}), make(chan core.Result)
	go b.identify(sb, quit, ret, exclude...)
	return ret, nil
}

// String returns information about the Bytematcher including the number of BOF, VAR and EOF sequences, the number of BOF and EOF frames, and the total number of tests.
func (b *Matcher) String() string {
	str := fmt.Sprintf("BOF seqs: %v\n", len(b.bofSeq.set))
	str += fmt.Sprintf("EOF seqs: %v\n", len(b.eofSeq.set))
	str += fmt.Sprintf("BOF frames: %v\n", len(b.bofFrames.set))
	str += fmt.Sprintf("EOF frames: %v\n", len(b.eofFrames.set))
	str += fmt.Sprintf("Total Tests: %v\n", len(b.tests))
	var c, ic, l, r, ml, mr int
	for _, t := range b.tests {
		c += len(t.complete)
		ic += len(t.incomplete)
		l += len(t.left)
		if ml < t.maxLeftDistance {
			ml = t.maxLeftDistance
		}
		r += len(t.right)
		if mr < t.maxRightDistance {
			mr = t.maxRightDistance
		}
	}
	str += fmt.Sprintf("Complete Tests: %v\n", c)
	str += fmt.Sprintf("Incomplete Tests: %v\n", ic)
	str += fmt.Sprintf("Left Tests: %v\n", l)
	str += fmt.Sprintf("Right Tests: %v\n", r)
	str += fmt.Sprintf("Maximum Left Distance: %v\n", ml)
	str += fmt.Sprintf("Maximum Right Distance: %v\n", mr)
	str += fmt.Sprintf("Known BOF distance: %v\n", b.knownBOF)
	str += fmt.Sprintf("Known EOF distance: %v\n", b.knownEOF)
	str += fmt.Sprintf("Maximum BOF Distance: %v\n", b.maxBOF)
	str += fmt.Sprintf("Maximum EOF Distance: %v\n", b.maxEOF)
	str += fmt.Sprintf("priorities: %v\n", b.priorities)
	return str
}

// InspectTestTree reports which signatures are linked to a given index in the test tree.
// This is used by the -log debug and -log slow options for sf.
func (b *Matcher) InspectTestTree(i int) []int {
	if i < 0 || i >= len(b.tests) {
		return nil
	}
	t := b.tests[i]
	res := make([]int, len(t.complete)+len(t.incomplete))
	for i, v := range t.complete {
		res[i] = v[0]
	}
	for i, v := range t.incomplete {
		res[i+len(t.complete)] = v.kf[0]
	}
	return res
}

// SetLowMem instructs the Aho Corasick search tree to be built with a low memory opt (runs slightly slower than regular).
func (b *Matcher) SetLowMem() {
	b.lowmem = true
}
