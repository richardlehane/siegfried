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

// Package Bytematcher builds a matching engine from a set of signatures and performs concurrent matching against an input siegreader.Buffer.
package bytematcher

import (
	"fmt"
	"sync"

	"github.com/richardlehane/match/wac"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/process"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
	"github.com/richardlehane/siegfried/pkg/core/signature"
)

// Bytematcher structure. Clients shouldn't need to get or set these fields directly, they are only exported so that this structure can be serialised and deserialised by encoding/gob.
type Matcher struct {
	*process.Process
	Priorities *priority.Set
	mu         *sync.Mutex
	bAho       *wac.Wac
	eAho       *wac.Wac
	lowmem     bool
}

func New() *Matcher {
	return &Matcher{
		process.New(),
		&priority.Set{},
		&sync.Mutex{},
		nil,
		nil,
		false,
	}
}

type SignatureSet []frames.Signature

func Load(ls *signature.LoadSaver) *Matcher {
	if !ls.LoadBool() {
		return nil
	}
	return &Matcher{
		Process:    process.Load(ls),
		Priorities: priority.Load(ls),
		mu:         &sync.Mutex{},
	}
}

func (b *Matcher) Save(ls *signature.LoadSaver) {
	if b == nil {
		ls.SaveBool(false)
		return
	}
	ls.SaveBool(true)
	b.Process.Save(ls)
	b.Priorities.Save(ls)
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
// The priorities should be a list of equal length to the signatures, or nil (if no priorities are to be set)
// Can give optional distance, range, choice, variable sequence length values to override the defaults of 8192, 2048, 64.
//   - the distance and range values dictate how signatures are segmented during processing
//   - the choices value controls how signature segments are converted into simple byte sequences during processing
//   - the varlen value controls what is the minimum length sequence acceptable for the variable Aho Corasick tree. The longer this length, the fewer false matches you will get during searching.
//
// Example:
//   err := Add([]Signature{Signature{NewFrame(BOF, Sequence{'p','d','f'}, nil, 0, 0)}})
func (b *Matcher) Add(ss core.SignatureSet, priorities priority.List) (int, error) {
	sigs, ok := ss.(SignatureSet)
	if !ok {
		return -1, fmt.Errorf("Byte matcher: can't convert signature set to BM signature set")
	}
	// set the options
	b.Distance, b.Range, b.Choices, b.VarLength = config.BMOptions()
	var se sigErrors
	// process each of the sigs, adding them to b.Sigs and the various seq/frame/testTree sets
	for _, sig := range sigs {
		err := b.AddSignature(sig)
		if err != nil {
			se = append(se, err)
		}
	}
	if len(se) > 0 {
		return -1, se
	}
	// set the maximum distances for this test tree so can properly size slices for matching
	for _, t := range b.Tests {
		t.MaxLeftDistance = process.MaxLength(t.Left)
		t.MaxRightDistance = process.MaxLength(t.Right)
	}
	// add the priorities to the priority set
	b.Priorities.Add(priorities, len(sigs))
	return len(b.KeyFrames), nil
}

type Result struct {
	index int
	basis string
}

func (r Result) Index() int {
	return r.index
}

func (r Result) Basis() string {
	return r.basis
}

// Identify matches a Bytematcher's signatures against the input siegreader.Buffer.
// Results are passed on the first returned int channel. These ints are the indexes of the matching signatures.
// The second and third int channels report on the Bytematcher's progress: returning offets from the beginning of the file and the end of the file.
//
// Example:
//   ret, bprog, eprog := bm.Identify(buf, q)
//   for v := range ret {
//     if v == 0 {
//       fmt.Print("Success! It is signature 0!")
//     }
//   }
func (b *Matcher) Identify(name string, sb siegreader.Buffer) (chan core.Result, error) {
	quit, ret := make(chan struct{}), make(chan core.Result)
	go b.identify(sb, quit, ret)
	return ret, nil
}

// Returns information about the Bytematcher including the number of BOF, VAR and EOF sequences, the number of BOF and EOF frames, and the total number of tests.
func (b *Matcher) String() string {
	str := fmt.Sprintf("BOF seqs: %v\n", len(b.BOFSeq.Set))
	str += fmt.Sprintf("EOF seqs: %v\n", len(b.EOFSeq.Set))
	str += fmt.Sprintf("BOF frames: %v\n", len(b.BOFFrames.Set))
	str += fmt.Sprintf("EOF frames: %v\n", len(b.EOFFrames.Set))
	str += fmt.Sprintf("Total Tests: %v\n", len(b.Tests))
	var c, ic, l, r, ml, mr int
	for _, t := range b.Tests {
		c += len(t.Complete)
		ic += len(t.Incomplete)
		l += len(t.Left)
		if ml < t.MaxLeftDistance {
			ml = t.MaxLeftDistance
		}
		r += len(t.Right)
		if mr < t.MaxRightDistance {
			mr = t.MaxRightDistance
		}
	}
	str += fmt.Sprintf("Complete Tests: %v\n", c)
	str += fmt.Sprintf("Incomplete Tests: %v\n", ic)
	str += fmt.Sprintf("Left Tests: %v\n", l)
	str += fmt.Sprintf("Right Tests: %v\n", r)
	str += fmt.Sprintf("Maximum Left Distance: %v\n", ml)
	str += fmt.Sprintf("Maximum Right Distance: %v\n", mr)
	str += fmt.Sprintf("Maximum BOF Distance: %v\n", b.MaxBOF)
	str += fmt.Sprintf("Maximum EOF Distance: %v\n", b.MaxEOF)
	str += fmt.Sprintf("Priorities: %v\n", b.Priorities)
	return str
}

func (b *Matcher) InspectTTI(i int) []int {
	if i < 0 || i >= len(b.Tests) {
		return nil
	}
	t := b.Tests[i]
	res := make([]int, len(t.Complete)+len(t.Incomplete))
	for i, v := range t.Complete {
		res[i] = v[0]
	}
	for i, v := range t.Incomplete {
		res[i+len(t.Complete)] = v.Kf[0]
	}
	return res
}

func (b *Matcher) SetLowMem() {
	b.lowmem = true
}
