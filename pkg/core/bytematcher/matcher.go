package bytematcher

import "sync"

// MUTABLE
type matcher struct {
	b                *Bytematcher
	r                chan int
	n                []byte
	partialKeyframes map[[2]int][][2]int // map of a keyframe to a slice of offsets and lengths where it has matched
	wg               *sync.WaitGroup
}

type partial struct {
	l          bool
	r          bool
	ldistances []int
	rdistances []int
}

func NewMatcher(b *Bytematcher, r chan int, n []byte, wg *sync.WaitGroup) *matcher {
	return &matcher{b, r, n, make(map[[2]int][][2]int), wg}
}

func (m *matcher) match(tti, o, l int, rev bool) {
	defer m.wg.Done()
	// the offsets we record are always BOF offsets - these can be interpreted as EOF offsets when necessary
	if rev {
		o = len(m.n) - o - l
	}
	t := m.b.TestSet[tti]
	for _, kf := range t.Complete {
		complete := m.applyKeyFrame(kf, o, l)
		if complete {
			m.r <- kf[0]
		}
	}
	if len(t.Incomplete) < 0 {
		return
	}
	partials := make([]partial, len(t.Incomplete))
	left := matches(t.Left, m.n[:o], true)
	for _, lp := range left {
		if partials[lp.followUp].l {
			partials[lp.followUp].ldistances = append(partials[lp.followUp].ldistances, lp.distances...)
		} else {
			partials[lp.followUp].l = true
			partials[lp.followUp].ldistances = lp.distances
		}
	}
	if o+l < len(m.n) {
		right := matches(t.Right, m.n[o+l:], false)
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
					off := o - ldistance
					length := ldistance + l + rdistance
					complete := m.applyKeyFrame(kf, off, length)
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
	if kf[kfID[1]].check(o, l, len(m.n)) {
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
