// Package priority creates a subordinate-superiors map of identifications.
// These maps can be flattened into sorted lists for use by the bytematcher and containermatcher engines.
package priority

import (
	"fmt"
	"sort"
)

// a priority map links subordinate results to a list of priority restuls
type Map map[string][]string

func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func addStr(ss []string, s string) []string {
	if containsStr(ss, s) {
		return ss
	}
	return append(ss, s)
}

// add a subordinate-superior relationship to the priority map
func (m Map) Add(subordinate string, superior string) {
	_, ok := m[subordinate]
	if ok {
		m[subordinate] = addStr(m[subordinate], superior)
		return
	}
	m[subordinate] = []string{superior}
}

func extras(a []string, b []string) []string {
	ret := make([]string, 0)
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

func (m Map) priorityWalk(k string) []string {
	tried := make([]string, 0)
	ret := make([]string, 0)
	var walkFn func(string)
	walkFn = func(id string) {
		vals, ok := m[id]
		if !ok {
			return
		}
		for _, v := range vals {
			// avoid cycles
			if containsStr(tried, v) {
				continue
			}
			tried = append(tried, v)
			priorityPriorities := m[v]
			ret = append(ret, extras(priorityPriorities, vals)...)
			walkFn(v)
		}
	}
	walkFn(k)
	return ret
}

// After adding all priorities, walk the priority map to make sure that it is consistent,
// i.e. that for any format with a superior fmt, then anything superior
// to that superior fmt is also marked as superior to the base fmt, all the way down the tree
func (m Map) Complete() {
	for k, _ := range m {
		extraPriorities := m.priorityWalk(k)
		m[k] = append(m[k], extraPriorities...)
	}
}

func (m Map) expand(key string, iMap map[string][]int) []int {
	ret := make([]int, 0)
	superiors := m[key]
	for _, k := range superiors {
		ret = append(ret, iMap[k]...)
	}
	sort.Ints(ret)
	return ret
}

// return a priority list using the indexes from the supplied slice of keys (keys can be duplicated in that slice)
func (m Map) List(keys []string) List {
	// build a map of keys to their indexes in the supplied slice
	iMap := make(map[string][]int)
	for _, k := range keys {
		// continue on if the key has already been added
		_, ok := iMap[k]
		if ok {
			continue
		}
		indexes := make([]int, 0)
		for i, v := range keys {
			if v == k {
				indexes = append(indexes, i)
			}
		}
		iMap[k] = indexes
	}
	l := make(List, len(keys))
	for i, k := range keys {
		l[i] = m.expand(k, iMap)
	}
	return l
}

type List [][]int

func (l List) Subset(indexes []int) List {
	if l == nil {
		return nil
	}
	subset := make(List, len(indexes))
	for i, v := range indexes {
		subset[i] = l[v]
	}
	return subset
}

func (l List) String() string {
	if l == nil {
		return "0 priorities defined"
	}
	var total int
	for _, v := range l {
		total += len(v)
	}
	return fmt.Sprintf("%d priorities defined", total)
}
