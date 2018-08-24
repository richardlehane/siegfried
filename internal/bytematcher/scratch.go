package bytematcher

import "fmt"

//var reverse bool
//if len(partials[len(partials)-1]) < len(partials[0]) {
//	reverse = true
//}

func searchPartials(partials [][][2]int64, kfs []keyFrame) (bool, string) {
	idxs := make([]int, len(partials))
	var i, j, o = 0, 1, 0

	for {
		// if we've made it to the last item, we've made it through
		if j == len(idxs) {
			basis := make([][2]int64, len(idxs))
			for i, v := range idxs {
				basis[i] = partials[i][v]
			}
			return true, fmt.Sprintf("byte match at %v", basis)
		}
		for {
			ok, lock := checkRelatedKF(kfs[j], kfs[i], partials[j][idxs[j]], partials[i][idxs[i]])
			if ok {
				if lock {
					o = j
				}
				i++
				j++
				break
			}
			if idxs[j] < len(partials[j])-1 {
				idxs[j]++
				continue
			}
			if idxs[i] < len(partials[i])-1 {
				idxs[j] = 0
				idxs[i]++
				if i > o {
					i--
					j--
				}
				continue
			}
			return false, ""
		}
	}

	/* dross below
	var ev int
	return func(start, end int) [][2]int64 {
		if idxs == nil || start >= len(idxs) || ev >= end {
			return nil
		}
		for i, v := range idxs[start:] {
			ret[i+start] = partials[i+start][v]
		}
		for i := range partials[start:] {
			if idxs[i+start] == len(partials[i+start])-1 {
				if i+start == len(partials)-1 { // if we are at the end
					idxs = nil
					break
				}
				if i+start >= end { // if we are at the end value
					ev = end
				}
				idxs[i+start] = 0
				continue
			}
			idxs[i+start]++
			break
		}
		return ret
	}

	next := iteratePartials(h.partials)
	var start, end int
	var basis [][2]int64
	for basis = next(start, len(kfs)-1); basis != nil; basis = next(start, end) {
		prevKf := kfs[0]
		var ok, checkpoint bool
		for i, kf := range kfs[1:] {
			ok, checkpoint = checkRelatedKF(kf, prevKf, basis[i+1], basis[i])
			if !ok {
				if end < i+1 {
					end = i + 1
				}
				break
			}
			if checkpoint && start < i+1 {
				start = i + 1
			}
		}
		if ok {
			return true, fmt.Sprintf("byte match at %v", basis)
		}
	}*/
}
