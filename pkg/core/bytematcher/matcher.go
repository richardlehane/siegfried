package bytematcher

import (
	"sort"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/process"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

// MUTABLE
type matcher struct {
	b              *ByteMatcher
	buf            *siegreader.Buffer
	r              chan int
	partialMatches map[[2]int][][2]int // map of a keyframe to a slice of offsets and lengths where it has matched
	incoming       chan strike
	quit           chan struct{}
	bofProgress    chan int
	bofOff         int
	eofProgress    chan int
	eofOff         int
	waitc          chan []int
	wait           []int
	strikeCache    map[int]*cacheItem
}

func (b *ByteMatcher) newMatcher(buf *siegreader.Buffer, q chan struct{}, r, bprog, eprog chan int, wait chan []int) *matcher {
	return &matcher{
		b,
		buf,
		r,
		make(map[[2]int][][2]int),
		make(chan strike, 100),
		q,
		bprog,
		0,
		eprog,
		0,
		wait,
		nil,
		make(map[int]*cacheItem),
	}
}

func (m *matcher) match() {
	for {
		select {
		case in, ok := <-m.incoming:
			// this happens when all of our matchers reach EOF
			if !ok {
				m.finalise()
				return
			}
			if halt := m.processStrike(in); halt {
				return
			}
		case bp := <-m.bofProgress:
			m.bofOff = bp
			if m.wait != nil {
				// check if anything we are waiting on is still alive
			}
		case ep := <-m.eofProgress:
			m.eofOff = ep
			if m.wait != nil {
				// check if anything we are waiting on is still alive
			}
		}

	}
}

func (m *matcher) sendResult(res int) bool {
	m.r <- res
	w := <-m.waitc // every result sent must result in a new priority list being returned & we need to drain this or it will block
	// nothing more to wait for
	if len(w) == 0 {
		return m.finalise()
	}
	m.wait = w
	return false
}

func (m *matcher) finalise() bool {
	close(m.quit)
	/*
		try to break deadlock
		for _ = range m.incoming {
		}
	*/
	close(m.r)
	return true
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

type strike struct {
	idxa    int
	idxb    int // a test tree index = idxa + idxb
	offset  int // offset of match
	length  int
	reverse bool
	frame   bool // is it a frameset match?
	final   bool // last in a sequence of strikes?
}

// a partial
type partial struct {
	l          bool
	r          bool
	ldistances []int
	rdistances []int
}

func (m *matcher) processStrike(s strike) bool {
	if s.frame {
		return m.tryStrike(s)
	}
	st, strks := m.stash(s)
	if st {
		for _, v := range strks {
			if halt := m.tryStrike(v); halt {
				return true
			}
		}
	}
	return false
}

func (m *matcher) tryStrike(s strike) bool {
	// the offsets we *record* are always BOF offsets - these can be interpreted as EOF offsets when necessary
	off := m.calcOffset(s)

	// grab the relevant testTree
	t := m.b.Tests[s.idxa+s.idxb]

	// immediately apply key frames for the completes
	for _, kf := range t.Complete {
		success := m.applyKeyFrame(kf, off, s.length)
		if success {
			if halt := m.sendResult(kf[0]); halt {
				return true
			}
		}
	}

	// if there are no incompletes, we are done
	if len(t.Incomplete) < 1 {
		return false
	}

	// see what incompletes are worth pursuing
	var checkl, checkr bool
	for _, v := range t.Incomplete {
		if checkl && checkr {
			break
		}
		if m.checkWait(v.Kf[0]) && m.aliveKeyFrame(v.Kf, s) {
			if v.L {
				checkl = true
			}
			if v.R {
				checkr = true
			}
		}
	}
	if !checkl && !checkr {
		return false
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
		lslc = m.buf.MustSlice(lpos, llen, s.reverse)
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
		rslc = m.buf.MustSlice(rpos, rlen, s.reverse)
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
			if !p.l {
				p.ldistances = []int{0}
			}
			if !p.r {
				p.rdistances = []int{0}
			}
			for _, ldistance := range p.ldistances {
				for _, rdistance := range p.rdistances {
					moff := off - ldistance
					length := ldistance + s.length + rdistance
					complete := m.applyKeyFrame(kf, moff, length)
					if complete {
						if halt := m.sendResult(kf[0]); halt {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func (m *matcher) calcOffset(s strike) int {
	if !s.reverse {
		return s.offset
	}
	return m.buf.Size() - s.offset - s.length
}

// check a signature ID against the priority list
func (m *matcher) checkWait(i int) bool {
	if m.wait == nil {
		return true
	}
	idx := sort.SearchInts(m.wait, i)
	if idx == len(m.wait) || m.wait[idx] != i {
		return false
	}
	return true
}

func (m *matcher) aliveKeyFrame(kfID process.KeyFrameID, s strike) bool {
	j := kfID[1]
	for i := 0; i < j; i++ {
		kf := m.b.KeyFrames[kfID[0]][i]
		if !kf.MustExist(s.offset, s.reverse) {
			return true
		}
		_, ok := m.partialMatches[[2]int{kfID[0], i}]
		if !ok {
			return false
		}
	}
	return true
}

func (m *matcher) applyKeyFrame(kfID process.KeyFrameID, o, l int) bool {
	kf := m.b.KeyFrames[kfID[0]]
	if kf[kfID[1]].Check(o, l, m.buf) {
		if len(kf) == 1 {
			return true
		}
	} else {
		return false
	}
	if _, ok := m.partialMatches[kfID]; ok {
		m.partialMatches[kfID] = append(m.partialMatches[kfID], [2]int{o, l})
	} else {
		m.partialMatches[kfID] = [][2]int{[2]int{o, l}}
	}
	return m.checkKeyFrames(kfID[0])
}

// check key frames checks the relationships between neighbouring frames
func (m *matcher) checkKeyFrames(i int) bool {
	kfs := m.b.KeyFrames[i]
	for j := range kfs {
		_, ok := m.partialMatches[[2]int{i, j}]
		if !ok {
			return false
		}
	}
	prevOff := m.partialMatches[[2]int{i, 0}]
	prevKf := kfs[0]
	var ok bool
	for j, kf := range kfs[1:] {
		thisOff := m.partialMatches[[2]int{i, j + 1}]
		prevOff, ok = kf.CheckRelated(prevKf, thisOff, prevOff)
		if !ok {
			return false
		}
		prevKf = kf
	}
	return true
}
