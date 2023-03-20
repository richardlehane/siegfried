// Copyright 2015 Richard Lehane. All rights reserved.
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

	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/internal/siegreader"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

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
}

func (st strike) String() string {
	strikeOrientation := "BOF"
	if st.reverse {
		strikeOrientation = "EOF"
	}
	strikeType := "sequence"
	if st.frame {
		strikeType = "frametest"
	}
	return fmt.Sprintf("{%s %s hit - index: %d [%d], offset: %d, length: %d}", strikeOrientation, strikeType, st.idxa+st.idxb, st.idxb, st.offset, st.length)
}

// strikes are cached in a map of strikeItems indexed by strikes' idxa + idxb fields
type strikeItem struct {
	first      strike
	idx        int // allows us to 'pop' strikes off the strikeItem and records where we are in the successive slice
	successive [][2]int64
}

// have we exhausted the strikeItem i.e. popped off all the available strikes?
func (s *strikeItem) hasPotential() bool {
	return s.idx+1 <= len(s.successive)
}

func (s *strikeItem) numPotentials() int {
	return len(s.successive) - s.idx
}

func (s *strikeItem) pop() strike {
	s.idx++
	if s.idx > 0 {
		s.first.offset, s.first.length = s.successive[s.idx-1][0], int(s.successive[s.idx-1][1])
	}
	return s.first
}

// potential hits (signature matches) are marked in a map of hitItems indexed by keyframeID[0]
type hitItem struct {
	potentialIdxs []int        // indexes to the strike cache
	partials      [][][2]int64 // for each keyframe in a signature, a slice of offsets and lengths of matches
	matched       bool         // if we've already matched, mark so don't return
}

// search a set of partials for a complete match
func searchPartials(partials [][][2]int64, kfs []keyFrame) (bool, string) {
	res := make([][][2]int64, len(partials))
	idxs := make([][]int, len(partials))
	prevOff := partials[0]
	var idx []int
	ok := false
	res[0] = prevOff
	for i, kf := range kfs[1:] {
		var nextKf keyFrame
		if i+2 < len(kfs) {
			nextKf = kfs[i+2]
		}
		prevOff, idx, ok = checkRelated(kf, kfs[i], nextKf, partials[i+1], prevOff)
		if !ok {
			return false, ""
		}
		res[i+1] = prevOff
		idxs[i+1] = idx
	}
	basis := make([][2]int64, len(partials))
	basis[len(basis)-1] = res[len(res)-1][0]
	j := idxs[len(idxs)-1][0]
	for i := len(idxs) - 1; i > 0; i-- {
		basis[i-1] = res[i-1][j]
		if i > 1 {
			j = idxs[i-1][j]
		}
	}
	return true, fmt.Sprintf("byte match at %v", basis)
}

// returns the next strike for testing and true if should continue/false if done
func (h *hitItem) nextPotential(s map[int]*strikeItem) (strike, bool) {
	if h == nil || !h.potentiallyComplete(-1, s) {
		return strike{}, false
	}
	var minIdx, min int
	for i, v := range h.potentialIdxs {
		// first try sending only when we don't have any corresponding partial matches
		if h.partials[i] == nil {
			return s[v-1].pop(), true
		}
		// otherwise, if all are potential, start with the fewest potentials first (so as to exclude)
		if v > 0 && s[v-1].hasPotential() && (min == 0 || s[v-1].numPotentials() < min) {
			minIdx, min = v-1, s[v-1].numPotentials()
		}
	}
	// in case we are all partials, no potentials
	if min == 0 {
		return strike{}, false
	}
	return s[minIdx].pop(), true
}

// is a hit item potentially complete? - i.e. has at least one potential strike,
// and either partial matches or strikes for all segments
func (h *hitItem) potentiallyComplete(idx int, s map[int]*strikeItem) bool {
	if h.matched { // if matched, we don't want to resatisfy it
		return false
	}
	for i, v := range h.potentialIdxs {
		if i == idx {
			continue
		}
		if (v == 0 || !s[v-1].hasPotential()) && h.partials[i] == nil {
			return false
		}
	}
	return true
}

// return list of all hits, however fragmentary
func all(m map[int]*hitItem) []int {
	ret := make([]int, len(m))
	i := 0
	for k := range m {
		ret[i] = k
		i++
	}
	return ret
}

// kfHits are returned by the testStrike function defined in the scorer method below. They give offsets and lengths for hits on signatures' keyframes.
type kfHit struct {
	id     keyFrameID
	offset int64
	length int
}

// partials are used within the testStrike function defined in the scorer method below.
// they mirror the testTree incompletes slice to record distances for hits to left and right of the matching segment
type partial struct {
	ldistances []int
	rdistances []int
}

// result is the bytematcher implementation of the Result interface.
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

func (b *Matcher) scorer(buf *siegreader.Buffer, waitSet *priority.WaitSet, q chan struct{}, r chan<- core.Result) (chan<- strike, <-chan []keyFrameID) {
	incoming := make(chan strike)
	resume := make(chan []keyFrameID)
	hits := make(map[int]*hitItem)
	strikes := make(map[int]*strikeItem)

	var bof int64
	var eof int64

	var quitting bool
	quit := func() {
		close(q)
		close(resume)
		quitting = true
	}

	newHit := func(i int) *hitItem {
		l := len(b.keyFrames[i])
		hit := &hitItem{
			potentialIdxs: make([]int, l),
			partials:      make([][][2]int64, l),
		}
		hits[i] = hit
		return hit
	}

	// Used for dwac
	dynamicSeqs := func(w []int) []keyFrameID {
		dynSeqs := make([]keyFrameID, 0, 20)
		for _, v := range w {
			kf := b.keyFrames[v]
			for i, f := range kf {
				var waitfor, excludable bool
				if f.key.pMax == -1 {
					waitfor = true
				} else if hit, ok := hits[v]; ok {
					if hit.partials[i] != nil {
						waitfor = true
					} else if hit.potentialIdxs[i] > 0 && strikes[hit.potentialIdxs[i]-1].hasPotential() {
						waitfor, excludable = true, true
					}
				}
				// if we've got to the end of the signature, and have determined this is a live one - return immediately & continue scan
				if waitfor {
					if i == len(kf)-1 {
						if config.Slow() {
							fmt.Fprintf(config.Out(), "waiting on: %d, potentially excludable: %t\n", v, excludable)
						}
						for ii, ff := range kf {
							if ff.key.pMax == -1 {
								dynSeqs = append(dynSeqs, keyFrameID{v, ii})
							}
						}
					}
					continue // we've got one, now get more!
				}
				break // no reason to keep checking this signature
			}
		}
		return dynSeqs
	}

	testStrike := func(st strike) []kfHit {
		// the offsets we *record* are always BOF offsets - these can be interpreted as EOF offsets when necessary
		off := st.offset
		if st.reverse {
			off = buf.Size() - st.offset - int64(st.length)
		}
		// grab the relevant testTree
		t := b.tests[st.idxa+st.idxb]
		res := make([]kfHit, 0, 10)
		// immediately apply key frames for the completes
		for _, kf := range t.complete {
			if b.keyFrames[kf[0]][kf[1]].check(st.offset) && waitSet.Check(kf[0]) {
				res = append(res, kfHit{kf, off, st.length})
			}
		}
		// if there are no incompletes, we are done
		if len(t.incomplete) < 1 {
			return res
		}
		// see what incompletes are worth pursuing
		var checkl, checkr bool
		for _, v := range t.incomplete {
			if checkl && checkr {
				break
			}
			if b.keyFrames[v.kf[0]][v.kf[1]].check(st.offset) && waitSet.Check(v.kf[0]) {
				if v.l {
					checkl = true
				}
				if v.r {
					checkr = true
				}
			}
		}
		if !checkl && !checkr {
			return res
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
				lslc, _ = buf.EofSlice(lpos, llen)
			} else {
				lslc, _ = buf.Slice(lpos, llen)
			}
			left := matchTestNodes(t.left, lslc, true)
			for _, lp := range left {
				if partials[lp.followUp].ldistances == nil {
					partials[lp.followUp].ldistances = lp.distances
				} else {
					partials[lp.followUp].ldistances = append(partials[lp.followUp].ldistances, lp.distances...)
				}
			}
		}
		// test right (if there are valid right tests to try)
		if checkr {
			if st.reverse {
				rslc, _ = buf.EofSlice(rpos, rlen)
			} else {
				rslc, _ = buf.Slice(rpos, rlen)
			}
			right := matchTestNodes(t.right, rslc, false)
			for _, rp := range right {
				if partials[rp.followUp].rdistances == nil {
					partials[rp.followUp].rdistances = rp.distances

				} else {
					partials[rp.followUp].rdistances = append(partials[rp.followUp].rdistances, rp.distances...)
				}
			}
		}
		// now iterate through the partials, checking whether they fulfil any of the incompletes
		for i, p := range partials {
			if (len(p.ldistances) > 0) == t.incomplete[i].l && (len(p.rdistances) > 0) == t.incomplete[i].r {
				kf := t.incomplete[i].kf
				if b.keyFrames[kf[0]][kf[1]].check(st.offset) && waitSet.Check(kf[0]) {
					if p.ldistances == nil {
						p.ldistances = []int{0}
					}
					if p.rdistances == nil {
						p.rdistances = []int{0}
					}
					// oneEnough is defined in keyframes.go and checks whether segments of a signature are anchored to other segments
					if oneEnough(kf[1], b.keyFrames[kf[0]]) {
						res = append(res, kfHit{kf, off - int64(p.ldistances[0]), p.ldistances[0] + st.length + p.rdistances[0]})
						continue
					}
					for _, ldistance := range p.ldistances {
						for _, rdistance := range p.rdistances {
							res = append(res, kfHit{kf, off - int64(ldistance), ldistance + st.length + rdistance})
						}
					}
				}
			}
		}
		return res
	}

	applyKeyFrame := func(hit kfHit) (bool, string) {
		kfs := b.keyFrames[hit.id[0]]
		if len(kfs) == 1 {
			return true, fmt.Sprintf("byte match at %d, %d", hit.offset, hit.length)
		}
		h, ok := hits[hit.id[0]]
		if !ok {
			h = newHit(hit.id[0])
		}
		if h.partials[hit.id[1]] == nil {
			h.partials[hit.id[1]] = [][2]int64{{hit.offset, int64(hit.length)}}
		} else {
			h.partials[hit.id[1]] = append(h.partials[hit.id[1]], [2]int64{hit.offset, int64(hit.length)})
		}
		for _, p := range h.partials {
			if p == nil {
				return false, ""
			}
		}
		return searchPartials(h.partials, kfs)
	}

	go func() {
		for in := range incoming {
			// if we've got a positive result, drain any remaining strikes from the matchers
			if quitting {
				continue
			}
			// HANDLE RESUME
			if in.idxa == -1 {
				w := waitSet.WaitingOn() // todo: this uses bof/eof which are less relevant I don't store progress
				if w == nil {
					w = all(hits)
				}
				if in.reverse {
					resume <- append(dynamicSeqs(w), b.unknownEOF...)
					continue
				}
				resume <- append(dynamicSeqs(w), b.unknownBOF...)
				continue
			}
			// HANDLE MATCH STRIKES
			var hasPotential bool
			potentials := filterKF(b.tests[in.idxa+in.idxb].keyFrames(), waitSet)
			for _, pot := range potentials {
				// if any of the signatures are single keyframe we can satisfy immediately and skip cache
				if len(b.keyFrames[pot[0]]) == 1 {
					hasPotential = true
					break
				}
				if hit, ok := hits[pot[0]]; ok && hit.potentiallyComplete(pot[1], strikes) {
					hasPotential = true
					break
				}
			}
			if !hasPotential {
				// cache the strike
				s, ok := strikes[in.idxa+in.idxb]
				if !ok {
					s = &strikeItem{in, -1, nil}
					strikes[in.idxa+in.idxb] = s
				} else {
					if s.successive == nil {
						s.successive = make([][2]int64, 0, 10)
					}
					s.successive = append(s.successive, [2]int64{in.offset, int64(in.length)})
				}
				// range over the potentials, linking to the strike
				for _, pot := range potentials {
					if b.keyFrames[pot[0]][pot[1]].check(in.offset) {
						hit, ok := hits[pot[0]]
						if !ok {
							hit = newHit(pot[0])
						}
						hit.potentialIdxs[pot[1]] = in.idxa + in.idxb + 1
					}
				}
				goto end
			}
			// satisfy the strike
			for {
				ks := testStrike(in)
				for _, k := range ks {
					if match, basis := applyKeyFrame(k); match {
						if waitSet.Check(k.id[0]) {
							r <- result{k.id[0], basis}
							if waitSet.PutAt(k.id[0], bof, eof) {
								quit()
								goto end
							}
						}
						if h, ok := hits[k.id[0]]; ok {
							h.matched = true
						}
					}
				}
				// given waitset, check if any potential matches remain to wait for
				potentials = filterKF(potentials, waitSet)
				var ok bool
				for _, pot := range potentials {
					in, ok = hits[pot[0]].nextPotential(strikes)
					if ok {
						break
					}
				}
				if !ok {
					break
				}
			}
		end: // keep looping until incoming is closed
		}
		close(r)
	}()
	return incoming, resume
}
