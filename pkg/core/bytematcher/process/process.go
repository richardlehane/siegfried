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

package process

import (
	"fmt"

	"github.com/richardlehane/match/wac"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
)

type Options struct {
	Distance  int
	Range     int
	Choices   int
	VarLength int
}

type Process struct {
	KeyFrames [][]keyFrame
	Tests     []*testTree
	BOFFrames *frameSet
	EOFFrames *frameSet
	BOFSeq    *seqSet
	EOFSeq    *seqSet
	MaxBOF    int
	MaxEOF    int
	Options
}

func New() *Process {
	return &Process{
		BOFFrames: &frameSet{},
		EOFFrames: &frameSet{},
		BOFSeq:    &seqSet{},
		EOFSeq:    &seqSet{},
	}
}

func (p *Process) AddSignature(sig frames.Signature) error {
	segments := p.splitSegments(sig)
	// apply config no eof option
	if config.NoEOF() {
		var hasEof bool
		var x int
		for i, segment := range segments {
			c := characterise(segment)
			if c > prev {
				hasEof = true
				x = i
				break
			}
		}
		if hasEof {
			if x == 0 {
				p.KeyFrames = append(p.KeyFrames, []keyFrame{})
				return nil
			}
			segments = segments[:x] // Otherwise trim segments to the first SUCC/EOF segment
		}
	}
	kf := make([]keyFrame, len(segments))
	clstr := newCluster(p)
	for i, segment := range segments {
		var pos position
		c := characterise(segment)
		switch c {
		case bofZero:
			pos = bofLength(segment, p.Choices)
		case eofZero:
			pos = eofLength(segment, p.Choices)
		default:
			pos = varLength(segment, p.Choices)
		}
		if pos.length < 1 {
			switch c {
			case bofZero, bofWindow:
				kf[i] = p.addToFrameSet(segment, i, p.BOFFrames, 0, 1)
			case eofZero, eofWindow:
				kf[i] = p.addToFrameSet(segment, i, p.EOFFrames, len(segment)-1, len(segment))
			default:
				return fmt.Errorf("Variable offset segment encountered that can't be turned into a sequence: signature %d, segment %d", len(p.KeyFrames), i)
			}
		} else {
			switch c {
			case bofZero, bofWindow, bofWild:
				clstr = clstr.commit()
				kf[i] = clstr.add(segment, i, pos)
			case prev:
				kf[i] = clstr.add(segment, i, pos)
			case succ:
				if !clstr.rev {
					clstr = clstr.commit()
					clstr.rev = true
				}
				kf[i] = clstr.add(segment, i, pos)
			case eofZero, eofWindow, eofWild:
				if !clstr.rev {
					clstr = clstr.commit()
					clstr.rev = true
				}
				kf[i] = clstr.add(segment, i, pos)
				clstr = clstr.commit()
				clstr.rev = true
			}
		}
	}
	clstr.commit()
	updatePositions(kf)
	p.MaxBOF = maxBOF(p.MaxBOF, kf)
	p.MaxEOF = maxEOF(p.MaxEOF, kf)
	p.KeyFrames = append(p.KeyFrames, kf)
	return nil
}

type cluster struct {
	rev    bool
	kfs    []keyFrame
	p      *Process
	w      wac.Seq
	ks     []int
	lefts  [][]frames.Frame
	rights [][]frames.Frame
}

func newCluster(p *Process) *cluster {
	return &cluster{p: p}
}

func (c *cluster) add(seg frames.Signature, i int, pos position) keyFrame {
	sequences := frames.NewSequencer(c.rev)
	k, left, right := toKeyFrame(seg, pos)
	c.kfs = append(c.kfs, k)
	var seqs [][]byte
	// do it all backwards
	if c.rev {
		for j := pos.end - 1; j >= pos.start; j-- {
			seqs = sequences(seg[j])
		}
		c.w.Choices = append([]wac.Choice{wac.Choice(seqs)}, c.w.Choices...)
		c.ks = append([]int{i}, c.ks...)
		c.lefts = append([][]frames.Frame{left}, c.lefts...)
		c.rights = append([][]frames.Frame{right}, c.rights...)
	} else {
		for _, f := range seg[pos.start:pos.end] {
			seqs = sequences(f)
		}
		c.w.Choices = append(c.w.Choices, wac.Choice(seqs))
		c.ks = append(c.ks, i)
		c.lefts = append(c.lefts, left)
		c.rights = append(c.rights, right)
	}
	return k
}

func (c *cluster) commit() *cluster {
	// commit nothing if the cluster is empty
	if len(c.w.Choices) == 0 {
		return newCluster(c.p)
	}
	updatePositions(c.kfs)
	c.w.MaxOffsets = make([]int, len(c.kfs))
	if c.rev {
		for i := range c.w.MaxOffsets {
			c.w.MaxOffsets[i] = c.kfs[len(c.kfs)-1-i].Key.PMax
		}
	} else {
		for i, v := range c.kfs {
			c.w.MaxOffsets[i] = v.Key.PMax
		}
	}
	var ss *seqSet
	if c.rev {
		ss = c.p.EOFSeq
	} else {
		ss = c.p.BOFSeq
	}
	hi := ss.add(c.w, len(c.p.Tests))
	l := len(c.ks)
	if hi == len(c.p.Tests) {
		for i := 0; i < l; i++ {
			c.p.Tests = append(c.p.Tests, &testTree{})
		}
	}
	for i := 0; i < l; i++ {
		c.p.Tests[hi+i].add([2]int{len(c.p.KeyFrames), c.ks[i]}, c.lefts[i], c.rights[i])
	}
	return newCluster(c.p)
}

func (p *Process) addToFrameSet(segment frames.Signature, i int, fs *frameSet, start, end int) keyFrame {
	k, left, right := toKeyFrame(segment, position{0, start, end})
	hi := fs.add(segment[start], len(p.Tests))
	if hi == len(p.Tests) {
		p.Tests = append(p.Tests, &testTree{})
	}
	p.Tests[hi].add([2]int{len(p.KeyFrames), i}, left, right)
	return k
}
