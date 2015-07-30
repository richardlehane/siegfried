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
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

// scorer is a mutable object that tallies incoming strikes (part matches) from the BOF/EOF byte and frame matchers
// it returns results and is responsible for signalling quit if an exit condition is met - either a) satisfied or b) the incoming channel closed
type scorer struct {
	bm       *Matcher
	buf      siegreader.Buffer
	quit     chan struct{}
	results  chan core.Result
	incoming chan strike

	strikeCache strikeCache
	kfHits      chan kfHit
	waitSet     *priority.WaitSet
	queue       *sync.WaitGroup
	once        *sync.Once
	stop        chan struct{}
	halt        chan bool

	tally *tally
}

func (b *Matcher) newScorer(buf siegreader.Buffer, q chan struct{}, r chan core.Result) chan strike {
	incoming := make(chan strike) // buffer this chan?
	s := &scorer{
		bm:       b,
		buf:      buf,
		quit:     q,
		results:  r,
		incoming: incoming,

		strikeCache: make(strikeCache),
		kfHits:      make(chan kfHit),
		waitSet:     b.priorities.WaitSet(),
		queue:       &sync.WaitGroup{},
		once:        &sync.Once{},
		stop:        make(chan struct{}),
		halt:        make(chan bool),

		tally: &tally{&sync.Mutex{}, make(map[[2]int][][2]int64), make(map[[2]int]int)},
	}
	go s.filterHits() // this goroutine is a gateway that takes all keyframe hits and reports results on the results channel
	go s.score()      // this is the main goroutine: it ranges on incoming, routing strikes
	return incoming
}

func (s *scorer) score() {
	for in := range s.incoming {
		select {
		case <-s.quit:
			return
		default:
		}
		if in.idxa == -1 { // check whether should keep waiting for progress strikes
			if !s.continueWait(in.offset, in.reverse) {
				break
			}
		} else {
			s.stash(in) // stash the strike
		}
	}
	s.shutdown(true) // shutdown at eof
}

// Strikes

// strike is a raw hit from either the WAC matchers or the BOF/EOF frame matchers
// progress strikes aren't hits: and have -1 for idxa, they just report how far we have scanned
type strike struct {
	idxa    int
	idxb    int   // a test tree index = idxa + idxb
	offset  int64 // offset of match
	length  int
	reverse bool
	frame   bool // is it a frameset match?
	final   bool // last in a sequence of strikes?
}

func (st strike) String() string {
	return fmt.Sprintf("{STRIKE Test: [%d:%d], Offset: %d, Length: %d, Reverse: %t, Frame: %t, Final: %t}", st.idxa, st.idxb, st.offset, st.length, st.reverse, st.frame, st.final)
}

// progress strikes are special results from the WAC matchers that periodically report on progress, these aren't hits
func progressStrike(off int64, rev bool) strike {
	return strike{
		idxa:    -1,
		idxb:    -1,
		offset:  off,
		reverse: rev,
		final:   true,
	}
}

// Cache Strikes

// strike cache holds strikes until it is necessary to evaluate them
type strikeCache map[int]*cacheItem

type cacheItem struct {
	finalised  bool
	potentials []keyFrameID // list of sigs this might match
	first      strike       // full details for the first match

	mu         *sync.Mutex
	successive [][2]int64 // just cache the offsets of successive matches
	strikeIdx  int        // -1 signals that the strike is in the first position, otherwise an index into the successive slice
	satisfying bool       // state when a cacheItem is currently trying a strike
}

func (s *scorer) newCacheItem(st strike) *cacheItem {
	return &cacheItem{
		finalised:  st.final,
		potentials: s.bm.tests[st.idxa+st.idxb].keyFrames(),
		first:      st,
		mu:         &sync.Mutex{},
		strikeIdx:  -1,
	}
}

// adds a new strike to an existing cache item, returns the current satisfying state
func (c *cacheItem) push(st strike) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.successive == nil {
		c.successive = make([][2]int64, 1, 10)
		c.successive[0][0], c.successive[0][1] = st.offset, int64(st.length)
		return c.satisfying
	}
	c.successive = append(c.successive, [2]int64{st.offset, int64(st.length)})
	return c.satisfying
}

// pops a strike from the item. Changes the satisfying state if returning the first strike. Returns strike and the satisfying state.
func (c *cacheItem) pop() (strike, bool) {
	ret := c.first
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.strikeIdx > -1 {
		// have we exhausted the cache?
		if c.strikeIdx > len(c.successive)-1 {
			c.satisfying = false // mark that no longer in a satisfying state - side effect ok as only satisfy loop calls pop
			return ret, false
		}
		ret.offset, ret.length = c.successive[c.strikeIdx][0], int(c.successive[c.strikeIdx][1])
	}
	c.strikeIdx++
	return ret, true
}

// 1. push the strike to the strike cache.
// 2. mark potentials: a slice marking which keyframes are potentially matched
// 3. filter those potentials (keyframes) against the waitset, and attempt to satisfy those that we are waiting on
func (s *scorer) stash(st strike) {
	stashed := s.strikeCache[st.idxa+st.idxb]
	if stashed == nil {
		stashed = s.newCacheItem(st)
		if st.final && st.idxb > 0 {
			if !s.strikeCache[st.idxa].finalised {
				// if not, do so now
				for i := st.idxb - 1; i >= 0; i-- {
					s.strikeCache[st.idxa+i].finalised = true
				}
			}
		}
		s.strikeCache[st.idxa+st.idxb] = stashed
	} else {
		if stashed.push(st) {
			return // return early if already satisfying
		}
	}
	s.markPotentials(stashed.potentials, st.idxa+st.idxb)
	if !stashed.finalised {
		return
	}
	pots := filterKF(stashed.potentials, s.waitSet)
	if len(pots) == 0 {
		return
	}
	s.satisfyPotentials(pots)
}

// range through the potentials, continuing for those keyframes that are potentially complete (all segments in the signature have strikes)
func (s *scorer) satisfyPotentials(pots []keyFrameID) {
	s.tally.mu.Lock() // during this phase - hold a lock on the tally
	for _, kf := range pots {
		if s.tally.completes(kf[0], len(s.bm.keyFrames[kf[0]])) {
			for i := 0; i < len(s.bm.keyFrames[kf[0]]); i++ {
				idx, ok := s.tally.potentialMatches[[2]int{kf[0], i}]
				if ok {
					s.satisfy(s.strikeCache[idx])
				}
			}
		}
	}
	s.tally.mu.Unlock()
}

func (s *scorer) satisfy(c *cacheItem) {
	c.mu.Lock()
	if c.satisfying {
		c.mu.Unlock()
		return
	}
	c.satisfying = true
	c.mu.Unlock()
	s.queue.Add(1)
	go func() {
		defer s.queue.Done()
		strike, ok := c.pop()
		if !ok {
			s.unmarkPotentials(c.potentials)
			return
		}
		if s.testStrike(strike) {
			return
		}
		for {
			strike, ok = c.pop()
			if !ok {
				s.unmarkPotentials(c.potentials)
				return
			}
			pots := filterKF(c.potentials, s.waitSet)
			if len(pots) == 0 {
				return
			}
			if !s.retainsPotential(pots) {
				c.mu.Lock()
				c.strikeIdx-- // backup
				c.satisfying = false
				c.mu.Unlock()
				return
			}
			if s.testStrike(strike) {
				return
			}
		}
	}()
}

// Tally

// this structure maintains the current state including actual (partial) and potential (stashed strikes) matches
type tally struct {
	mu               *sync.Mutex
	partialMatches   map[[2]int][][2]int64 // map of a keyframe to a slice of offsets and lengths where it has matched
	potentialMatches map[[2]int]int        // represents complete/incomplete keyframe hits
}

// completes returns true if a strike will complete a signature (all the other parts either match or potentially match)
func (t *tally) completes(a, l int) bool {
	for i := 0; i < l; i++ {
		_, partial := t.partialMatches[[2]int{a, i}]
		_, potential := t.potentialMatches[[2]int{a, i}]
		if !(partial || potential) {
			return false
		}
	}
	return true
}

func (s *scorer) unmarkPotentials(pots []keyFrameID) {
	s.tally.mu.Lock()
	for _, kf := range pots {
		delete(s.tally.potentialMatches, [2]int{kf[0], kf[1]})
	}
	s.tally.mu.Unlock()
}

func (s *scorer) markPotentials(pots []keyFrameID, idx int) {
	s.tally.mu.Lock()
	for _, kf := range pots {
		s.tally.potentialMatches[[2]int{kf[0], kf[1]}] = idx
	}
	s.tally.mu.Unlock()
}

func (s *scorer) retainsPotential(pots []keyFrameID) bool {
	s.tally.mu.Lock()
	defer s.tally.mu.Unlock()
	for _, kf := range pots {
		if s.tally.completes(kf[0], len(s.bm.keyFrames[kf[0]])) {
			return true
		}
	}
	return false
}

func (s *scorer) applyKeyFrame(kfID keyFrameID, o int64, l int) (bool, string) {
	kf := s.bm.keyFrames[kfID[0]]
	if len(kf) == 1 {
		return true, fmt.Sprintf("byte match at %d, %d", o, l)
	}
	s.tally.mu.Lock()
	defer s.tally.mu.Unlock()
	if _, ok := s.tally.partialMatches[kfID]; ok {
		s.tally.partialMatches[kfID] = append(s.tally.partialMatches[kfID], [2]int64{o, int64(l)})
	} else {
		s.tally.partialMatches[kfID] = [][2]int64{[2]int64{o, int64(l)}}
	}
	return s.checkKeyFrames(kfID[0])
}

// check key frames checks the relationships between neighbouring frames
func (s *scorer) checkKeyFrames(i int) (bool, string) {
	kfs := s.bm.keyFrames[i]
	for j := range kfs {
		_, ok := s.tally.partialMatches[[2]int{i, j}]
		if !ok {
			return false, ""
		}
	}
	prevOff := s.tally.partialMatches[[2]int{i, 0}]
	basis := make([][][2]int64, len(kfs))
	basis[0] = prevOff
	prevKf := kfs[0]
	var ok bool
	for j, kf := range kfs[1:] {
		thisOff := s.tally.partialMatches[[2]int{i, j + 1}]
		prevOff, ok = kf.checkRelated(prevKf, thisOff, prevOff)
		if !ok {
			return false, ""
		}
		basis[j+1] = prevOff
		prevKf = kf
	}
	return true, fmt.Sprintf("byte match at %v", basis)
}

// 2. Test strikes

// a partial
type partial struct {
	l          bool
	r          bool
	ldistances []int
	rdistances []int
}

// this will block until quit if EOF is inaccessible
func (s *scorer) calcOffset(st strike) int64 {
	if !st.reverse {
		return st.offset
	}
	return s.buf.Size() - st.offset - int64(st.length)
}

// testStrike checks a strike for a match. Return true if we can halt now (nothing better to wait for)
func (s *scorer) testStrike(st strike) bool {
	// the offsets we *record* are always BOF offsets - these can be interpreted as EOF offsets when necessary
	off := s.calcOffset(st)
	// if we've quitted, the calculated offset will be a negative int
	if off < 0 {
		return true
	}

	// grab the relevant testTree
	t := s.bm.tests[st.idxa+st.idxb]

	// immediately apply key frames for the completes
	for _, kf := range t.complete {
		if s.bm.keyFrames[kf[0]][kf[1]].check(st.offset) && s.waitSet.Check(kf[0]) {
			s.kfHits <- kfHit{kf, off, st.length}
			if <-s.halt {
				return true
			}
		}
	}

	// if there are no incompletes, we are done
	if len(t.incomplete) < 1 {
		return false
	}

	// see what incompletes are worth pursuing
	//TODO: HANDLE INCOMPLETE CHECKS AS GOROUTINE
	var checkl, checkr bool
	for _, v := range t.incomplete {
		if checkl && checkr {
			break
		}
		if s.bm.keyFrames[v.kf[0]][v.kf[1]].check(st.offset) && s.waitSet.Check(v.kf[0]) {
			if v.l {
				checkl = true
			}
			if v.r {
				checkr = true
			}
		}
	}
	if !checkl && !checkr {
		return false
	}

	// calculate the offset and lengths for the left and right test slices
	var lslc, rslc []byte
	var lpos, rpos int64
	var llen, rlen int
	if st.reverse {
		lpos, llen = st.offset+int64(st.length), t.maxLeftDistance
		rpos, rlen = st.offset-int64(t.maxRightDistance), t.maxRightDistance
		if rpos < 0 {
			rlen = rlen + int(rpos)
			rpos = 0
		}
	} else {
		lpos, llen = st.offset-int64(t.maxLeftDistance), t.maxLeftDistance
		rpos, rlen = st.offset+int64(st.length), t.maxRightDistance
		if lpos < 0 {
			llen = llen + int(lpos)
			lpos = 0
		}
	}

	//  the partials slice has a mirror entry for each of the testTree incompletes
	partials := make([]partial, len(t.incomplete))

	// test left (if there are valid left tests to try)
	if checkl {
		if st.reverse {
			lslc, _ = s.buf.EofSlice(lpos, llen)
		} else {
			lslc, _ = s.buf.Slice(lpos, llen)
		}
		left := matchTestNodes(t.left, lslc, true)
		for _, lp := range left {
			if partials[lp.followUp].l {
				partials[lp.followUp].ldistances = append(partials[lp.followUp].ldistances, lp.distances...)
			} else {
				partials[lp.followUp].l = true
				partials[lp.followUp].ldistances = lp.distances
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
		right := matchTestNodes(t.right, rslc, false)
		for _, rp := range right {
			if partials[rp.followUp].r {
				partials[rp.followUp].rdistances = append(partials[rp.followUp].rdistances, rp.distances...)
			} else {
				partials[rp.followUp].r = true
				partials[rp.followUp].rdistances = rp.distances
			}
		}
	}

	// now iterate through the partials, checking whether they fulfil any of the incompletes
	for i, p := range partials {
		if p.l == t.incomplete[i].l && p.r == t.incomplete[i].r {
			kf := t.incomplete[i].kf
			if s.bm.keyFrames[kf[0]][kf[1]].check(st.offset) && s.waitSet.Check(kf[0]) {
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
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func (s *scorer) shutdown(eof bool) {
	go s.once.Do(func() { s.finalise(eof) })
}

func (s *scorer) finalise(eof bool) {
	// if we've reached the end of the file, need to make sure any pending tests are completed
	if eof {
		s.queue.Wait()
	}
	close(s.quit) // signals siegreaders to end
	// drain any remaining matches
	for _ = range s.incoming {
	}
	// must wait for all pending tests to complete (so that the halt signal is sent back to them)
	if !eof {
		s.queue.Wait()
	}
	close(s.results)
	close(s.stop)
}

type kfHit struct {
	id     keyFrameID
	offset int64
	length int
}

func (s *scorer) filterHits() {
	var satisfied bool
	for {
		select {
		case <-s.stop:
			return
		case hit := <-s.kfHits:
			if satisfied {
				// the halt channel tells the testStrike goroutine
				// to continuing checking complete/incomplete tests for the strike
				s.halt <- true
				continue
			}
			// in case of a race
			if !s.waitSet.Check(hit.id[0]) {
				s.halt <- false
				continue
			}
			success, basis := s.applyKeyFrame(hit.id, hit.offset, hit.length)
			if success {
				if h := s.sendResult(hit.id[0], basis); h {
					s.halt <- true
					satisfied = true
					s.shutdown(false)
					continue
				}
			}
			s.halt <- false
		}
	}
}

type result struct {
	index int
	basis string
}

func (r result) Index() int {
	return r.index
}

func (r result) Basis() string {
	return r.basis
}

func (s *scorer) sendResult(idx int, basis string) bool {
	s.results <- result{idx, basis}
	return s.waitSet.Put(idx)
}

// check to see whether should still wait for signatures in the priority list, given the offset
func (s *scorer) continueWait(o int64, rev bool) bool {
	return true // test fix
	var fails int
	w := s.waitSet.WaitingOn()
	// must continue if any of the waitlists are nil
	if w == nil {
		return true
	}
	if len(w) == 0 {
		return false
	}
	s.tally.mu.Lock()
	defer s.tally.mu.Unlock()
	for _, v := range w {
		kf := s.bm.keyFrames[v]
		if rev {
			for i := len(kf) - 1; i >= 0 && kf[i].typ > frames.PREV; i-- {
				if kf[i].key.pMax == -1 || kf[i].key.pMax+int64(kf[i].key.lMax) > o {
					return true
				}
				if _, ok := s.tally.partialMatches[[2]int{v, i}]; ok {
					continue
				}
				if _, ok := s.tally.potentialMatches[[2]int{v, i}]; ok {
					continue
				}
				fails++
				break
			}
		} else {
			for i, f := range kf {
				if f.typ > frames.PREV {
					break
				}
				if f.key.pMax == -1 || f.key.pMax+int64(f.key.lMax) > o {
					return true
				}
				if _, ok := s.tally.partialMatches[[2]int{v, i}]; ok {
					continue
				}
				if _, ok := s.tally.potentialMatches[[2]int{v, i}]; ok {
					continue
				}
				fails++
				break
			}
		}
	}
	if fails == len(w) {
		return false
	}
	return true

}
