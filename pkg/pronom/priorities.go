package pronom

import "sort"

func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func containsInt(is []int, i int) bool {
	for _, v := range is {
		if v == i {
			return true
		}
	}
	return false
}

func appendUniq(is []int, i int) []int {
	if containsInt(is, i) {
		return is
	}
	return append(is, i)
}

func extras(a []int, b []int) []int {
	ret := make([]int, 0)
	for _, v := range a {
		var exists bool
		for _, v1 := range b {
			if v == v1 {
				exists = true
				break
			}
		}
		if !exists {
			ret = append(ret, v)
		}
	}
	return ret
}

func priorityWalk(k string, ps map[string][]int, puids []string) []int {
	tried := make([]string, 0)
	ret := make([]int, 0)
	var walkFn func(string)
	walkFn = func(p string) {
		vals, ok := ps[p]
		if !ok {
			return
		}
		for _, v := range vals {
			puid := puids[v]
			if containsStr(tried, puid) {
				continue
			}
			tried = append(tried, puid)
			priorityPriorities := ps[puid]
			ret = append(ret, extras(priorityPriorities, vals)...)
			walkFn(puid)
		}
	}
	walkFn(k)
	return ret
}

// returns a map of puids and the indexes of byte signatures that those puids should give priority to
func (p pronom) priorities() map[string][]int {
	var iter int
	priorities := make(map[string][]int)
	for _, f := range p.droid.FileFormats {
		for _ = range f.Signatures {
			for _, v := range f.Priorities {
				puid := p.ids[v]
				_, ok := priorities[puid]
				if ok {
					priorities[puid] = appendUniq(priorities[puid], iter)
				} else {
					priorities[puid] = []int{iter}
				}
			}
			for _, r := range f.Relations {
				// only interested in subtypes
				if r.Type != "Is subtype of" {
					continue
				}
				puid := p.ids[r.ID]
				_, ok := priorities[puid]
				if ok {
					priorities[puid] = appendUniq(priorities[puid], iter)
				} else {
					priorities[puid] = []int{iter}
				}
			}
			iter++
		}
	}

	// now check the priority tree to make sure that it is consistent,
	// i.e. that for any format with a superior fmt, then anything superior
	// to that superior fmt is also marked as superior to the base fmt, all the way down the tree
	puids, _ := p.GetPuids()
	for k, _ := range priorities {
		extras := priorityWalk(k, priorities, puids)
		if len(extras) > 0 {
			priorities[k] = append(priorities[k], extras...)
		}
	}

	for k := range priorities {
		sort.Ints(priorities[k])
	}
	return priorities
}

/* This adds substantially to benchmarks e.g. XML... needs more work*/
// Deal with non-explicit priorities?
// This is where there is no HasPriority relation but we should still wait anyway as we have an extension match
// Rule:-
// - for each extension match
// 	 - if the ID in question is not a priority of that extension match
//   - add BMIds to the wait list for that extension
func (pi *PronomIdentifier) addExtensionPriorities(id int, w []int, ems []int) []int {
	for _, v := range ems {
		ps := pi.Priorities[pi.EPuids[v]]
		var junior bool
		for _, psv := range ps {
			if psv == id {
				junior = true
			}
		}
		if !junior {
			w = append(w, pi.PuidsB[pi.EPuids[v]]...)
		}
	}
	sort.Ints(w)
	return w
}
