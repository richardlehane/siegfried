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

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
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
	return fmt.Sprintf("{%s %s hit - test index: %d [%d], offset: %d, length: %d}", strikeOrientation, strikeType, st.idxa+st.idxb, st.idxb, st.offset, st.length)
}

// progress strikes are special results from the WAC matchers that periodically report on progress, these aren't hits
func progressStrike(off int64, rev bool) strike {
	return strike{
		idxa:    -1,
		idxb:    -1,
		offset:  off,
		reverse: rev,
	}
}

// strikes are cached in a map of strike items
type strikeItem struct {
	first      strike
	idx        int
	successive [][2]int64
}

func (s *strikeItem) hasPotential() bool {
	return s.idx+1 <= len(s.successive)
}

func (s *strikeItem) pop() strike {
	s.idx++
	if s.idx > 0 {
		s.first.offset, s.first.length = s.successive[s.idx-1][0], int(s.successive[s.idx-1][1])
	}
	return s.first
}

// potential hits are marked in a map of hitItems
type hitItem struct {
	potentialIdxs []int        // indexes to the strike cache
	partials      [][][2]int64 // for each keyframe in a signature, a slice of offsets and lengths of matches
	matched       bool         // if we've already matched, mark so don't return
}

// return next strike to test, true if continue/false if done
func (h *hitItem) nextPotential(s map[int]*strikeItem) (strike, bool) {
	if h == nil || !h.potentiallyComplete(-1, s) {
		return strike{}, false
	}
	for i, v := range h.potentialIdxs {
		// first try sending only when we don't have any corresponding partial matches
		if v > 0 && h.partials[i] == nil && s[v-1].hasPotential() {
			return s[v-1].pop(), true
		}
	}
	// now start retrying other potentials, starting from the left
	for _, v := range h.potentialIdxs {
		if v > 0 && s[v-1].hasPotential() {
			return s[v-1].pop(), true
		}
	}
	return strike{}, false
}

// is a hit item potentially complete - i.e. has at least one potential strike,
// and either partial matches or strikes for all segments
func (h *hitItem) potentiallyComplete(idx int, s map[int]*strikeItem) bool {
	if h.matched {
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

type kfHit struct {
	id     keyFrameID
	offset int64
	length int
}

type partial struct {
	l          bool
	r          bool
	ldistances []int
	rdistances []int
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

func (b *Matcher) scorer(buf siegreader.Buffer, q chan struct{}, r chan<- core.Result) chan<- strike {
	incoming := make(chan strike)
	waitSet := b.priorities.WaitSet()
	hits := make(map[int]*hitItem)
	strikes := make(map[int]*strikeItem)

	var bof int64
	var eof int64

	var quitting bool
	quit := func() {
		close(q)
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

	// given the current bof and eof, is there anything worth waiting for?
	continueWaiting := func(w []int) bool {
		// now for each of the possible signatures we are either waiting on or have partial/potential matches for, check whether there are live contenders
		for _, v := range w {
			kf := b.keyFrames[v]
			for i, f := range kf {
				off := bof
				if f.typ > frames.PREV {
					off = eof
				}
				var waitfor bool
				if off > 0 && (f.key.pMax == -1 || f.key.pMax+int64(f.key.lMax) > off) {
					waitfor = true
				} else if hit, ok := hits[v]; ok && (hit.partials[i] != nil || (hit.potentialIdxs[i] > 0 && strikes[hit.potentialIdxs[i]-1].hasPotential())) {
					waitfor = true
				}
				// if we've got to the end of the signature, and have determined this is a live one - return immediately & continue scan
				if waitfor {
					if i == len(kf)-1 {
						return true
					}
					continue
				}
				break
			}
		}
		return false
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
				rslc, _ = buf.EofSlice(rpos, rlen)
			} else {
				rslc, _ = buf.Slice(rpos, rlen)
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
				if b.keyFrames[kf[0]][kf[1]].check(st.offset) && waitSet.Check(kf[0]) {
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
							res = append(res, kfHit{kf, moff, length})
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
			h.partials[hit.id[1]] = [][2]int64{[2]int64{hit.offset, int64(hit.length)}}
		} else {
			h.partials[hit.id[1]] = append(h.partials[hit.id[1]], [2]int64{hit.offset, int64(hit.length)})
		}
		for _, p := range h.partials {
			if p == nil {
				return false, ""
			}
		}
		prevOff := h.partials[0]
		basis := make([][][2]int64, len(kfs))
		basis[0] = prevOff
		prevKf := kfs[0]
		ok = false
		for i, kf := range kfs[1:] {
			thisOff := h.partials[i+1]
			prevOff, ok = kf.checkRelated(prevKf, thisOff, prevOff)
			if !ok {
				return false, ""
			}
			basis[i+1] = prevOff
			prevKf = kf
		}
		return true, fmt.Sprintf("byte match at %v", basis)
	}

	go func() {
		for in := range incoming {
			// if we've got a postive result, drain any remaining strikes from the matchers
			if quitting {
				continue
			}
			// if the strike reports progress, check if we should be continuing to wait
			if in.idxa == -1 {
				// update with the latest offset
				if in.reverse {
					eof = in.offset
				} else {
					bof = in.offset
				}
				w := waitSet.WaitingOn()
				// if any of the waitlists are nil, we will continue - unless we are past the known bof and known eof (points at which we *should* have got at least partial matches), in which case we will check if any partial/potential matches are live
				if w == nil {
					// keep going if we don't have a maximum known bof, or if our current bof/eof are less than the maximum known bof/eof
					if b.knownBOF < 0 || int64(b.knownBOF) > bof || int64(b.knownEOF) > eof {
						continue
					}
					// if we don't have a waitlist, and we are past the known bof and known eof, grab all the partials and potentials to check if any are live
					w = all(hits)
				}
				// exhausted all contenders, we can stop scanning
				if !continueWaiting(w) {
					quit()
				}
				continue
			}
			// now cache or satisfy the strike
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
							if waitSet.Put(k.id[0]) {
								quit()
								goto end
							}
						}
						if h, ok := hits[k.id[0]]; ok {
							h.matched = true
						}
					}
				}
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
	return incoming
}
