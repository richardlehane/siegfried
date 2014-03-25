package bytematcher

import (
	"sort"
	"sync"

	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

// MUTABLE
type matcher struct {
	b                *ByteMatcher
	buf              *siegreader.Buffer
	r                chan int
	partialKeyframes map[[2]int][][2]int // map of a keyframe to a slice of offsets and lengths where it has matched
	limit            []int
	limitm           *sync.RWMutex
	limitc           chan []int
	incoming         chan strike
	quit             chan struct{}
}

type strike struct {
	tti     int
	offset  int
	length  int
	reverse bool
	frame   bool
}

type partial struct {
	l          bool
	r          bool
	ldistances []int
	rdistances []int
}

func NewMatcher(b *ByteMatcher, buf *siegreader.Buffer, r chan int, l chan []int) *matcher {
	return &matcher{b, buf, r, make(map[[2]int][][2]int), nil, &sync.RWMutex{}, l, make(chan strike), make(chan struct{})}
}

func (m *matcher) setLimit(l []int) {
	m.limitm.Lock()
	m.limit = l
	m.limitm.Unlock()
}

func (m *matcher) checkLimit(i int) bool {
	m.limitm.RLock()
	defer m.limitm.RUnlock()
	if m.limit == nil {
		return true
	}
	idx := sort.SearchInts(m.limit, i)
	if idx == len(m.limit) || m.limit[idx] != i {
		return false
	}
	return true
}

func (m *matcher) match() {
	for {
		select {
		case l, ok := <-m.limitc:
			if !ok {
				close(m.quit)
				m.limitc = nil
			}
			m.setLimit(l)
		case in, ok := <-m.incoming:
			if !ok {
				close(m.r)
				return
			}
			m.tryStrike(in.tti, in.offset, in.length, in.reverse, in.frame)
		}
	}
}

func (m *matcher) tryStrike(tti, o, l int, rev, frame bool) {
	// the offsets we record are always BOF offsets - these can be interpreted as EOF offsets when necessary
	off := m.calcOffset(o, l, rev)

	// grab the relevant testTree
	t := m.b.TestSet[tti]

	// immediately apply key frames for the completes
	for _, kf := range t.Complete {
		success := m.applyKeyFrame(kf, off, l)
		if success {
			m.r <- kf[0]
		}
	}

	// if there are no incompletes, we are done
	if len(t.Incomplete) < 1 {
		return
	}

	// see what incompletes are worth pursuing
	checkl, checkr := m.checkLR(t, off, rev, frame)
	if !checkl && !checkr {
		return
	}
	partials := make([]partial, len(t.Incomplete))
	var lslc, rslc []byte
	var lpos, llen, rpos, rlen int
	if rev {
		lpos, llen = o+l, t.MaxLeftDistance
		rpos, rlen = o-t.MaxRightDistance, t.MaxRightDistance
		if rpos < 0 {
			rlen = rlen + rpos
			rpos = 0
		}
	} else {
		lpos, llen = o-t.MaxLeftDistance, t.MaxLeftDistance
		rpos, rlen = o+l, t.MaxRightDistance
		if lpos < 0 {
			llen = llen + lpos
			lpos = 0
		}
	}
	if checkl {
		lslc = m.buf.MustSlice(lpos, llen, rev)
		left := matchTestNodes(t.Left, lslc, true)
		for _, lp := range left {
			if partials[lp.followUp].l {
				partials[lp.followUp].ldistances = append(partials[lp.followUp].ldistances, lp.distances...)
			} else {
				partials[lp.followUp].l = true
				partials[lp.followUp].ldistances = lp.distances
			}
		}
	}
	if checkr {
		rslc = m.buf.MustSlice(rpos, rlen, rev)
		right := matchTestNodes(t.Right, rslc, false)
		for _, rp := range right {
			if partials[rp.followUp].r {
				partials[rp.followUp].rdistances = append(partials[rp.followUp].rdistances, rp.distances...)
			} else {
				partials[rp.followUp].r = true
				partials[rp.followUp].rdistances = rp.distances
			}
		}
	}
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
					length := ldistance + l + rdistance
					complete := m.applyKeyFrame(kf, moff, length)
					if complete {
						m.r <- kf[0]
					}
				}
			}
		}
	}
}

func (m *matcher) calcOffset(o, l int, r bool) int {
	if !r {
		return o
	}
	return m.buf.Size() - o - l
}

func (m *matcher) checkLR(t *testTree, o int, rev, frame bool) (bool, bool) {
	var checkl, checkr bool
	for _, v := range t.Incomplete {
		if checkl && checkr {
			break
		}
		if m.checkLimit(v.Kf[0]) && m.aliveKeyFrame(v.Kf, o, rev, frame) {
			if v.L {
				checkl = true
			}
			if v.R {
				checkr = true
			}
		}
	}
	return checkl, checkr
}

func (m *matcher) aliveKeyFrame(kfID keyframeID, o int, rev, frame bool) bool {
	// if we are doing BOF frame matching, we must assume that there might be sequences yet to be matched
	if frame {
		return true
	}
	j := kfID[1]
	for i := 0; i < j; i++ {
		kf := m.b.Sigs[kfID[0]][i]
		if !kf.mustExist(o, rev) {
			return true
		}
		_, ok := m.partialKeyframes[[2]int{kfID[0], i}]
		if !ok {
			return false
		}
	}
	return true
}

func (m *matcher) applyKeyFrame(kfID keyframeID, o, l int) bool {
	kf := m.b.Sigs[kfID[0]]
	if kf[kfID[1]].check(o, l, m.buf) {
		if len(kf) == 1 {
			return true
		}
	} else {
		return false
	}
	if _, ok := m.partialKeyframes[kfID]; ok {
		m.partialKeyframes[kfID] = append(m.partialKeyframes[kfID], [2]int{o, l})
	} else {
		m.partialKeyframes[kfID] = [][2]int{[2]int{o, l}}
	}
	return m.checkKeyFrames(kfID[0])
}

func (m *matcher) checkKeyFrames(i int) bool {
	kfs := m.b.Sigs[i]
	for j := range kfs {
		_, ok := m.partialKeyframes[[2]int{i, j}]
		if !ok {
			return false
		}
	}
	prevOff := m.partialKeyframes[[2]int{i, 0}]
	prevKf := kfs[0]
	var ok bool
	for j, kf := range kfs[1:] {
		thisOff := m.partialKeyframes[[2]int{i, j + 1}]
		prevOff, ok = kf.checkRelated(prevKf, thisOff, prevOff)
		if !ok {
			return false
		}
		prevKf = kf
	}
	return true
}
