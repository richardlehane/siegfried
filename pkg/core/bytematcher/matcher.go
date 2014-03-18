package bytematcher

import "sync"

// MUTABLE
type matcher struct {
	b                *Bytematcher
	r                chan int
	partialKeyframes map[[2]int][][2]int // map of a keyframe to a slice of offsets and lengths where it has matched
	wg               *sync.WaitGroup
}

type partial struct {
	l          bool
	r          bool
	ldistances []int
	rdistances []int
}

func NewMatcher(b *Bytematcher, r chan int, wg *sync.WaitGroup) *matcher {
	return &matcher{b, r, make(map[[2]int][][2]int), wg}
}

func (m *matcher) match(tti, o, l int, rev bool) {
	defer m.wg.Done()
	// the offsets we record are always BOF offsets - these can be interpreted as EOF offsets when necessary
	var off int
	if rev {
		off = m.b.buf.Size() - o - l
	} else {
		off = o
	}
	t := m.b.TestSet[tti]
	for _, kf := range t.Complete {
		complete := m.applyKeyFrame(kf, off, l)
		if complete {
			m.r <- kf[0]
		}
	}
	if len(t.Incomplete) < 0 {
		return
	}
	partials := make([]partial, len(t.Incomplete))
	var lslc, rslc []byte
	if rev {
		rpos, rlen := o-t.MaxRightDistance, t.MaxRightDistance
		if rpos < 0 {
			rlen = rlen + rpos
			rpos = 0
		}
		rslc = m.b.buf.MustSlice(rpos, rlen, true)
		lslc = m.b.buf.MustSlice(o+l, t.MaxLeftDistance, true)
	} else {
		lpos, llen := o-t.MaxLeftDistance, t.MaxLeftDistance
		if lpos < 0 {
			llen = llen + lpos
			lpos = 0
		}
		lslc = m.b.buf.MustSlice(lpos, llen, false)
		rslc = m.b.buf.MustSlice(o+l, t.MaxRightDistance, false)
	}
	left := matches(t.Left, lslc, true)
	for _, lp := range left {
		if partials[lp.followUp].l {
			partials[lp.followUp].ldistances = append(partials[lp.followUp].ldistances, lp.distances...)
		} else {
			partials[lp.followUp].l = true
			partials[lp.followUp].ldistances = lp.distances
		}
	}
	right := matches(t.Right, rslc, false)
	for _, rp := range right {
		if partials[rp.followUp].r {
			partials[rp.followUp].rdistances = append(partials[rp.followUp].rdistances, rp.distances...)
		} else {
			partials[rp.followUp].r = true
			partials[rp.followUp].rdistances = rp.distances
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

func (m *matcher) applyKeyFrame(kfID keyframeID, o, l int) bool {
	kf := m.b.Sigs[kfID[0]]
	if kf[kfID[1]].check(o, l, m.b.buf.Size()) {
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
	prevOff, ok := m.partialKeyframes[[2]int{i, 0}]
	if !ok {
		return false
	}
	prevKf := kfs[0]
	for j, kf := range kfs[1:] {
		thisOff, ok := m.partialKeyframes[[2]int{i, j + 1}]
		if !ok {
			return false
		}
		prevOff, ok = kf.checkRelated(prevKf, thisOff, prevOff)
		if !ok {
			return false
		}
		prevKf = kf
	}
	return true
}
