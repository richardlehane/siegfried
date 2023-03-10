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

package bytematcher

import (
	"fmt"

	"github.com/richardlehane/match/dwac"
	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/config"
)

func (b *Matcher) addSignature(sig frames.Signature) error {
	// todo: add cost to the Segment - or merge segments based on cost?
	segments := sig.Segment(config.Distance(), config.Range(), config.Cost(), config.Repetition())
	// apply config no eof option
	if config.NoEOF() {
		var hasEof bool
		var x int
		for i, segment := range segments {
			c := segment.Characterise()
			if c > frames.Prev {
				hasEof = true
				x = i
				break
			}
		}
		if hasEof {
			if x == 0 {
				b.keyFrames = append(b.keyFrames, []keyFrame{})
				return nil
			}
			segments = segments[:x] // Otherwise trim segments to the first SUCC/EOF segment
		}
	}
	kf := make([]keyFrame, len(segments))
	clstr := newCluster(b)
	for i, segment := range segments {
		var pos frames.Position
		c := segment.Characterise()
		switch c {
		case frames.Unknown:
			return fmt.Errorf("zero length segment: signature %d, %v, segment %d", len(b.keyFrames), sig, i)
		case frames.BOFZero:
			pos = frames.BOFLength(segment, config.Choices())
		case frames.EOFZero:
			pos = frames.EOFLength(segment, config.Choices())
		default:
			pos = frames.VarLength(segment, config.Choices())
		}
		if pos.Length < 1 {
			switch c {
			case frames.BOFZero, frames.BOFWindow:
				kf[i] = b.addToFrameSet(segment, i, b.bofFrames, 0, 1)
			case frames.EOFZero, frames.EOFWindow:
				kf[i] = b.addToFrameSet(segment, i, b.eofFrames, len(segment)-1, len(segment))
			default:
				return fmt.Errorf("variable offset segment encountered that can't be turned into a sequence: signature %d, segment %d", len(b.keyFrames), i)
			}
		} else {
			switch c {
			case frames.BOFZero, frames.BOFWild:
				clstr = clstr.commit()
				kf[i] = clstr.add(segment, i, pos)
			case frames.BOFWindow:
				if i > 0 {
					kfB, _, _ := toKeyFrame(segment, pos)
					if crossOver(kf[i-1], kfB) {
						clstr = clstr.commit()
					}
				} else {
					clstr = clstr.commit()
				}
				kf[i] = clstr.add(segment, i, pos)
			case frames.Prev:
				kf[i] = clstr.add(segment, i, pos)
			case frames.Succ:
				if !clstr.rev {
					clstr = clstr.commit()
					clstr.rev = true
				}
				kf[i] = clstr.add(segment, i, pos)
			case frames.EOFZero, frames.EOFWindow, frames.EOFWild:
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
	unknownBOF, unknownEOF := unknownBOFandEOF(len(b.keyFrames), kf)
	if len(unknownBOF) > 0 {
		b.unknownBOF = append(b.unknownBOF, unknownBOF...)
	}
	if len(unknownEOF) > 0 {
		b.unknownBOF = append(b.unknownEOF, unknownEOF...)
	}
	b.maxBOF = maxBOF(b.maxBOF, kf)
	b.maxEOF = maxEOF(b.maxEOF, kf)
	b.keyFrames = append(b.keyFrames, kf)
	return nil
}

type cluster struct {
	rev    bool
	kfs    []keyFrame
	b      *Matcher
	w      dwac.Seq
	ks     []int
	lefts  [][]frames.Frame
	rights [][]frames.Frame
}

func newCluster(b *Matcher) *cluster {
	return &cluster{b: b}
}

func (c *cluster) add(seg frames.Signature, i int, pos frames.Position) keyFrame {
	sequences := frames.NewSequencer(c.rev)
	k, left, right := toKeyFrame(seg, pos)
	c.kfs = append(c.kfs, k)
	var seqs [][]byte
	// do it all backwards
	if c.rev {
		for j := pos.End - 1; j >= pos.Start; j-- {
			seqs = sequences(seg[j])
		}
		c.w.Choices = append([]dwac.Choice{dwac.Choice(seqs)}, c.w.Choices...)
		c.ks = append([]int{i}, c.ks...)
		c.lefts = append([][]frames.Frame{left}, c.lefts...)
		c.rights = append([][]frames.Frame{right}, c.rights...)
	} else {
		for _, f := range seg[pos.Start:pos.End] {
			seqs = sequences(f)
		}
		c.w.Choices = append(c.w.Choices, dwac.Choice(seqs))
		c.ks = append(c.ks, i)
		c.lefts = append(c.lefts, left)
		c.rights = append(c.rights, right)
	}
	return k
}

func (c *cluster) commit() *cluster {
	// commit nothing if the cluster is empty
	if len(c.w.Choices) == 0 {
		return newCluster(c.b)
	}
	updatePositions(c.kfs)
	c.w.MaxOffsets = make([]int64, len(c.kfs))
	if c.rev {
		for i := range c.w.MaxOffsets {
			c.w.MaxOffsets[i] = c.kfs[len(c.kfs)-1-i].key.pMax
		}
	} else {
		for i, v := range c.kfs {
			c.w.MaxOffsets[i] = v.key.pMax
		}
	}
	var ss *seqSet
	if c.rev {
		ss = c.b.eofSeq
	} else {
		ss = c.b.bofSeq
	}
	hi := ss.add(c.w, len(c.b.tests))
	l := len(c.ks)
	if hi == len(c.b.tests) {
		for i := 0; i < l; i++ {
			c.b.tests = append(c.b.tests, &testTree{})
		}
	}
	for i := 0; i < l; i++ {
		c.b.tests[hi+i].add([2]int{len(c.b.keyFrames), c.ks[i]}, c.lefts[i], c.rights[i])
	}
	return newCluster(c.b)
}

func (b *Matcher) addToFrameSet(segment frames.Signature, i int, fs *frameSet, start, end int) keyFrame {
	k, left, right := toKeyFrame(segment, frames.Position{Length: 0, Start: start, End: end})
	hi := fs.add(segment[start], len(b.tests))
	if hi == len(b.tests) {
		b.tests = append(b.tests, &testTree{})
	}
	b.tests[hi].add([2]int{len(b.keyFrames), i}, left, right)
	return k
}
