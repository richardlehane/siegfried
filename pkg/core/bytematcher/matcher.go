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
	"sync"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/process"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

// MUTABLE
type matcher struct {
	incoming       chan strike
	bm             *Matcher
	buf            *siegreader.Buffer
	bofProgress    chan int
	eofProgress    chan int
	gate           chan struct{}
	partialMatches map[[2]int][][2]int // map of a keyframe to a slice of offsets and lengths where it has matched
	strikeCache    map[int]*cacheItem
	*tally
}

func (b *Matcher) newMatcher(buf *siegreader.Buffer, q chan struct{}, r chan core.Result, bprog, eprog chan int, gate chan struct{}) chan strike {
	incoming := make(chan strike) // buffer ? Use benchmarks to check
	m := &matcher{
		incoming:       incoming,
		bm:             b,
		buf:            buf,
		bofProgress:    bprog,
		eofProgress:    eprog,
		gate:           gate,
		partialMatches: make(map[[2]int][][2]int),
		strikeCache:    make(map[int]*cacheItem),
	}
	m.tally = newTally(r, q, m)
	go m.match()
	return incoming
}

func (m *matcher) match() {
	for {
		select {
		case in, ok := <-m.incoming:
			// this happens when all of our matchers reach EOF
			if !ok {
				m.shutdown(true)
				return
			}
			m.processStrike(in)
		case p := <-m.bofProgress:
			if p == 12*1024 {
				close(m.gate)
			}
			if p%4096 == 0 {
				m.bofQueue.Wait()
				// see if need to continue here
			}
		case _ = <-m.eofProgress:
		}
	}
}

type strike struct {
	idxa    int
	idxb    int // a test tree index = idxa + idxb
	offset  int // offset of match
	length  int
	reverse bool
	frame   bool // is it a frameset match?
	final   bool // last in a sequence of strikes?
}

func (s strike) String() string {
	return fmt.Sprintf("{STRIKE Test: [%d:%d], Offset: %d, Length: %d, Reverse: %t, Frame: %t, Final: %t}", s.idxa, s.idxb, s.offset, s.length, s.reverse, s.frame, s.final)
}

type cacheItem struct {
	finalised bool
	strikes   []strike
}

// cache strikes until we have all sequences in a cluster (group of frames with a single BOF or EOF anchor)
func (m *matcher) stash(s strike) (bool, []strike) {
	stashed, ok := m.strikeCache[s.idxa]
	if ok && stashed.finalised {
		return true, []strike{s}
	}
	if s.final {
		if !ok {
			return true, []strike{s}
		}
		stashed.finalised = true
		return true, append(stashed.strikes, s)
	}
	if !ok {
		m.strikeCache[s.idxa] = &cacheItem{false, []strike{s}}
	} else {
		m.strikeCache[s.idxa].strikes = append(m.strikeCache[s.idxa].strikes, s)
	}
	return false, nil
}

// a partial
type partial struct {
	l          bool
	r          bool
	ldistances []int
	rdistances []int
}

func (m *matcher) processStrike(s strike) {
	var queue *sync.WaitGroup
	if s.reverse {
		queue = m.eofQueue
	} else {
		queue = m.bofQueue
	}
	if s.frame {
		queue.Add(1)
		go m.tryStrike(s, queue)
		return
	}
	st, strks := m.stash(s)
	if st {
		for _, v := range strks {
			queue.Add(1)
			go m.tryStrike(v, queue)
		}
	}
}

// this will block until quit if EOF is inaccessible
func (m *matcher) calcOffset(s strike) int {
	if !s.reverse {
		return s.offset
	}
	return m.buf.Size() - s.offset - s.length
}

func (m *matcher) tryStrike(s strike, queue *sync.WaitGroup) {
	defer queue.Done()
	// the offsets we *record* are always BOF offsets - these can be interpreted as EOF offsets when necessary
	off := m.calcOffset(s)
	// if we've quitted, the calculated offset will be a negative int
	if off < 0 {
		return
	}

	// grab the relevant testTree
	t := m.bm.Tests[s.idxa+s.idxb]

	// immediately apply key frames for the completes
	for _, kf := range t.Complete {
		if m.bm.KeyFrames[kf[0]][kf[1]].Check(s.offset) && m.waitSet.Check(kf[0]) {
			m.kfHits <- kfHit{kf, off, s.length}
			if <-m.halt {
				return
			}
		}
	}

	// if there are no incompletes, we are done
	if len(t.Incomplete) < 1 {
		return
	}

	// see what incompletes are worth pursuing
	var checkl, checkr bool
	for _, v := range t.Incomplete {
		if checkl && checkr {
			break
		}
		if m.bm.KeyFrames[v.Kf[0]][v.Kf[1]].Check(s.offset) && m.waitSet.Check(v.Kf[0]) {
			if v.L {
				checkl = true
			}
			if v.R {
				checkr = true
			}
		}
	}
	if !checkl && !checkr {
		return
	}

	// calculate the offset and lengths for the left and right test slices
	var lslc, rslc []byte
	var lpos, llen, rpos, rlen int
	if s.reverse {
		lpos, llen = s.offset+s.length, t.MaxLeftDistance
		rpos, rlen = s.offset-t.MaxRightDistance, t.MaxRightDistance
		if rpos < 0 {
			rlen = rlen + rpos
			rpos = 0
		}
	} else {
		lpos, llen = s.offset-t.MaxLeftDistance, t.MaxLeftDistance
		rpos, rlen = s.offset+s.length, t.MaxRightDistance
		if lpos < 0 {
			llen = llen + lpos
			lpos = 0
		}
	}

	//  the partials slice has a mirror entry for each of the testTree incompletes
	partials := make([]partial, len(t.Incomplete))

	// test left (if there are valid left tests to try)
	if checkl {
		lslc, _ = m.buf.SafeSlice(lpos, llen, s.reverse)
		// if we've quit already, we'll return a nil slice
		if lslc == nil {
			return
		}
		left := process.MatchTestNodes(t.Left, lslc, true)
		for _, lp := range left {
			if partials[lp.FollowUp].l {
				partials[lp.FollowUp].ldistances = append(partials[lp.FollowUp].ldistances, lp.Distances...)
			} else {
				partials[lp.FollowUp].l = true
				partials[lp.FollowUp].ldistances = lp.Distances
			}
		}
	}
	// test right (if there are valid left tests to try)
	if checkr {
		rslc, _ = m.buf.SafeSlice(rpos, rlen, s.reverse)
		// if we've quit already, we'll return a nil slice
		if rslc == nil {
			return
		}
		right := process.MatchTestNodes(t.Right, rslc, false)
		for _, rp := range right {
			if partials[rp.FollowUp].r {
				partials[rp.FollowUp].rdistances = append(partials[rp.FollowUp].rdistances, rp.Distances...)
			} else {
				partials[rp.FollowUp].r = true
				partials[rp.FollowUp].rdistances = rp.Distances
			}
		}
	}

	// now iterate through the partials, checking whether they fulfil any of the incompletes
	for i, p := range partials {
		if p.l == t.Incomplete[i].L && p.r == t.Incomplete[i].R {
			kf := t.Incomplete[i].Kf
			if m.bm.KeyFrames[kf[0]][kf[1]].Check(s.offset) && m.waitSet.Check(kf[0]) {
				if !p.l {
					p.ldistances = []int{0}
				}
				if !p.r {
					p.rdistances = []int{0}
				}
				for _, ldistance := range p.ldistances {
					for _, rdistance := range p.rdistances {
						toff := s.offset - ldistance
						if s.reverse {
							toff = s.offset - rdistance
						}
						if m.bm.KeyFrames[kf[0]][kf[1]].CheckSeg(toff) {
							moff := off - ldistance
							length := ldistance + s.length + rdistance
							m.kfHits <- kfHit{kf, moff, length}
							if <-m.halt {
								return
							}
						}
					}
				}
			}
		}
	}
}

func (m *matcher) applyKeyFrame(kfID process.KeyFrameID, o, l int) (bool, string) {
	kf := m.bm.KeyFrames[kfID[0]]
	if len(kf) == 1 {
		return true, fmt.Sprintf("byte match at %d, %d", o, l)
	}
	if _, ok := m.partialMatches[kfID]; ok {
		m.partialMatches[kfID] = append(m.partialMatches[kfID], [2]int{o, l})
	} else {
		m.partialMatches[kfID] = [][2]int{[2]int{o, l}}
	}
	return m.checkKeyFrames(kfID[0])
}

// check key frames checks the relationships between neighbouring frames
func (m *matcher) checkKeyFrames(i int) (bool, string) {
	kfs := m.bm.KeyFrames[i]
	for j := range kfs {
		_, ok := m.partialMatches[[2]int{i, j}]
		if !ok {
			return false, ""
		}
	}
	prevOff := m.partialMatches[[2]int{i, 0}]
	basis := make([][][2]int, len(kfs))
	basis[0] = prevOff
	prevKf := kfs[0]
	var ok bool
	for j, kf := range kfs[1:] {
		thisOff := m.partialMatches[[2]int{i, j + 1}]
		prevOff, ok = kf.CheckRelated(prevKf, thisOff, prevOff)
		if !ok {
			return false, ""
		}
		basis[j+1] = prevOff
		prevKf = kf
	}
	return true, fmt.Sprintf("byte match at %v", basis)
}
