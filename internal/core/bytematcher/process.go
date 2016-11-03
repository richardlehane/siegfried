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

	wac "github.com/richardlehane/match/fwac"
	"github.com/richardlehane/siegfried/internal/config"
	"github.com/richardlehane/siegfried/internal/core/bytematcher/frames"
)

func (b *Matcher) addSignature(sig frames.Signature) error {
	segments := sig.Segment(config.Distance(), config.Range())
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
				b.keyFrames = append(b.keyFrames, []keyFrame{})
				return nil
			}
			segments = segments[:x] // Otherwise trim segments to the first SUCC/EOF segment
		}
	}
	kf := make([]keyFrame, len(segments))
	clstr := newCluster(b)
	for i, segment := range segments {
		var pos position
		c := characterise(segment)
		switch c {
		case unknown:
			return fmt.Errorf("Zero length segment: signature %d, %v, segment %d", len(b.keyFrames), sig, i)
		case bofZero:
			pos = bofLength(segment, config.Choices())
		case eofZero:
			pos = eofLength(segment, config.Choices())
		default:
			pos = varLength(segment, config.Choices())
		}
		if pos.length < 1 {
			switch c {
			case bofZero, bofWindow:
				kf[i] = b.addToFrameSet(segment, i, b.bofFrames, 0, 1)
			case eofZero, eofWindow:
				kf[i] = b.addToFrameSet(segment, i, b.eofFrames, len(segment)-1, len(segment))
			default:
				return fmt.Errorf("Variable offset segment encountered that can't be turned into a sequence: signature %d, segment %d", len(b.keyFrames), i)
			}
		} else {
			switch c {
			case bofZero, bofWild:
				clstr = clstr.commit()
				kf[i] = clstr.add(segment, i, pos)
			case bofWindow:
				if i > 0 {
					kfB, _, _ := toKeyFrame(segment, pos)
					if crossOver(kf[i-1], kfB) {
						clstr = clstr.commit()
					}
				} else {
					clstr = clstr.commit()
				}
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
	b.knownBOF, b.knownEOF = firstBOFandEOF(b.knownBOF, b.knownEOF, kf)
	b.maxBOF = maxBOF(b.maxBOF, kf)
	b.maxEOF = maxEOF(b.maxEOF, kf)
	b.keyFrames = append(b.keyFrames, kf)
	return nil
}

type cluster struct {
	rev    bool
	kfs    []keyFrame
	b      *Matcher
	w      wac.Seq
	ks     []int
	lefts  [][]frames.Frame
	rights [][]frames.Frame
}

func newCluster(b *Matcher) *cluster {
	return &cluster{b: b}
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
	k, left, right := toKeyFrame(segment, position{0, start, end})
	hi := fs.add(segment[start], len(b.tests))
	if hi == len(b.tests) {
		b.tests = append(b.tests, &testTree{})
	}
	b.tests[hi].add([2]int{len(b.keyFrames), i}, left, right)
	return k
}
