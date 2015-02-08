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
type scorer struct {
	incoming       chan strike
	bm             *Matcher
	buf            siegreader.Buffer
	partialMatches map[[2]int][][2]int64 // map of a keyframe to a slice of offsets and lengths where it has matched
	partialStrikes map[[2]int]int        // represents complete/incomplete keyframe hits
	strikeCache    map[int]*cacheItem
	*tally
}

func (b *Matcher) newScorer(buf siegreader.Buffer, q chan struct{}, r chan core.Result) chan strike {
	incoming := make(chan strike) // buffer ? Use benchmarks to check
	s := &scorer{
		incoming:       incoming,
		bm:             b,
		buf:            buf,
		partialMatches: make(map[[2]int][][2]int64),
		partialStrikes: make(map[[2]int]int),
		strikeCache:    make(map[int]*cacheItem),
	}
	s.tally = newTally(r, q, s)
	go s.score()
	return incoming
}

func (s *scorer) score() {
	for in := range s.incoming {
		if in.idxa == -1 {
			if in.reverse {
				s.eofQueue.Wait()
				if !s.continueWait(in.offset, false) {
					break
				}
			} else {
				s.bofQueue.Wait()
				if !s.continueWait(in.offset, false) {
					break
				}
			}
		} else {
			s.processStrike(in)
		}
	}
	s.shutdown(true)
}

type strike struct {
	idxa    int
	idxb    int   // a test tree index = idxa + idxb
	offset  int64 // offset of match
	length  int
	reverse bool
	frame   bool // is it a frameset match?
	final   bool // last in a sequence of strikes?
}

var progressStrike = strike{
	idxa:  -1,
	idxb:  -1,
	final: true,
}

func (s strike) String() string {
	return fmt.Sprintf("{STRIKE Test: [%d:%d], Offset: %d, Length: %d, Reverse: %t, Frame: %t, Final: %t}", s.idxa, s.idxb, s.offset, s.length, s.reverse, s.frame, s.final)
}

type cacheItem struct {
	finalised bool
	strikes   []strike
}

// cache strikes until we have all sequences in a cluster (group of frames with a single BOF or EOF anchor)
func (s *scorer) stash(st strike) (bool, []strike) {
	stashed, ok := s.strikeCache[st.idxa]
	if ok && stashed.finalised {
		return true, []strike{st}
	}
	if st.final {
		if !ok {
			return true, []strike{st}
		}
		stashed.finalised = true
		return true, append(stashed.strikes, st)
	}
	if !ok {
		s.strikeCache[st.idxa] = &cacheItem{false, []strike{st}}
	} else {
		s.strikeCache[st.idxa].strikes = append(s.strikeCache[st.idxa].strikes, st)
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

func (s *scorer) processStrike(st strike) {
	var queue *sync.WaitGroup
	if st.reverse {
		queue = s.eofQueue
	} else {
		queue = s.bofQueue
	}
	if st.frame {
		queue.Add(1)
		go s.tryStrike(st, queue)
		return
	}
	ok, strks := s.stash(st)
	if ok {
		for _, v := range strks {
			queue.Add(1)
			go s.tryStrike(v, queue)
		}
	}
}

// this will block until quit if EOF is inaccessible
func (s *scorer) calcOffset(st strike) int64 {
	if !st.reverse {
		return st.offset
	}
	return s.buf.Size() - st.offset - int64(st.length)
}

func (s *scorer) tryStrike(st strike, queue *sync.WaitGroup) {
	defer queue.Done()
	// the offsets we *record* are always BOF offsets - these can be interpreted as EOF offsets when necessary
	off := s.calcOffset(st)
	// if we've quitted, the calculated offset will be a negative int
	if off < 0 {
		return
	}

	// grab the relevant testTree
	t := s.bm.Tests[st.idxa+st.idxb]

	// immediately apply key frames for the completes
	for _, kf := range t.Complete {
		if s.bm.KeyFrames[kf[0]][kf[1]].Check(st.offset) && s.waitSet.Check(kf[0]) {
			s.kfHits <- kfHit{kf, off, st.length}
			if <-s.halt {
				return
			}
		}
	}

	// if there are no incompletes, we are done
	if len(t.Incomplete) < 1 {
		return
	}
	/*
		TODO: HANDLE INCOMPLETE CHECKS AS GOROUTINE
	*/
	// see what incompletes are worth pursuing
	var checkl, checkr bool
	for _, v := range t.Incomplete {
		if checkl && checkr {
			break
		}
		if s.bm.KeyFrames[v.Kf[0]][v.Kf[1]].Check(st.offset) && s.waitSet.Check(v.Kf[0]) {
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
	var lpos, rpos int64
	var llen, rlen int
	if st.reverse {
		lpos, llen = st.offset+int64(st.length), t.MaxLeftDistance
		rpos, rlen = st.offset-int64(t.MaxRightDistance), t.MaxRightDistance
		if rpos < 0 {
			rlen = rlen + int(rpos)
			rpos = 0
		}
	} else {
		lpos, llen = st.offset-int64(t.MaxLeftDistance), t.MaxLeftDistance
		rpos, rlen = st.offset+int64(st.length), t.MaxRightDistance
		if lpos < 0 {
			llen = llen + int(lpos)
			lpos = 0
		}
	}

	//  the partials slice has a mirror entry for each of the testTree incompletes
	partials := make([]partial, len(t.Incomplete))

	// test left (if there are valid left tests to try)
	if checkl {
		if st.reverse {
			lslc, _ = s.buf.EofSlice(lpos, llen)
		} else {
			lslc, _ = s.buf.Slice(lpos, llen)
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
	// test right (if there are valid right tests to try)
	if checkr {
		if st.reverse {
			rslc, _ = s.buf.EofSlice(rpos, rlen)
		} else {
			rslc, _ = s.buf.Slice(rpos, rlen)
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
			if s.bm.KeyFrames[kf[0]][kf[1]].Check(st.offset) && s.waitSet.Check(kf[0]) {
				if !p.l {
					p.ldistances = []int{0}
				}
				if !p.r {
					p.rdistances = []int{0}
				}
				for _, ldistance := range p.ldistances {
					for _, rdistance := range p.rdistances {
						moff := off - int64(ldistance)
						length := ldistance + st.length + rdistance
						s.kfHits <- kfHit{kf, moff, length}
						if <-s.halt {
							return
						}
					}
				}
			}
		}
	}
}

func (s *scorer) applyKeyFrame(kfID process.KeyFrameID, o int64, l int) (bool, string) {
	kf := s.bm.KeyFrames[kfID[0]]
	if len(kf) == 1 {
		return true, fmt.Sprintf("byte match at %d, %d", o, l)
	}
	if _, ok := s.partialMatches[kfID]; ok {
		s.partialMatches[kfID] = append(s.partialMatches[kfID], [2]int64{o, int64(l)})
	} else {
		s.partialMatches[kfID] = [][2]int64{[2]int64{o, int64(l)}}
	}
	return s.checkKeyFrames(kfID[0])
}

// check key frames checks the relationships between neighbouring frames
func (s *scorer) checkKeyFrames(i int) (bool, string) {
	kfs := s.bm.KeyFrames[i]
	for j := range kfs {
		_, ok := s.partialMatches[[2]int{i, j}]
		if !ok {
			return false, ""
		}
	}
	prevOff := s.partialMatches[[2]int{i, 0}]
	basis := make([][][2]int64, len(kfs))
	basis[0] = prevOff
	prevKf := kfs[0]
	var ok bool
	for j, kf := range kfs[1:] {
		thisOff := s.partialMatches[[2]int{i, j + 1}]
		prevOff, ok = kf.CheckRelated(prevKf, thisOff, prevOff)
		if !ok {
			return false, ""
		}
		basis[j+1] = prevOff
		prevKf = kf
	}
	return true, fmt.Sprintf("byte match at %v", basis)
}
