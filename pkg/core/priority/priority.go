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

// Package priority creates a subordinate-superiors map of identifications.
// These maps can be flattened into sorted lists for use by the bytematcher and containermatcher engines.
// Multiple priority lists can be added to priority sets. These contain the priorities of different identifiers within a bytematcher or containermatcher.
package priority

import (
	"fmt"
	"sort"
	"sync"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/process"
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
	if subordinate == "" || superior == "" {
		return
	}
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

// because keys can be duplicated in the slice given to List(), the list of superior indexes may be larger than the list of superior keys
func (m Map) expand(key string, iMap map[string][]int) []int {
	// use an empty, rather than nil slice for ret. This means a priority.List will never contain a nil slice.
	ret := make([]int, 0)
	superiors := m[key]
	for _, k := range superiors {
		ret = append(ret, iMap[k]...)
	}
	sort.Ints(ret)
	return ret
}

// Filter returns a new Priority Map that just contains formats in the provided slice
func (m Map) Filter(fmts []string) Map {
	ret := make(Map)
	for _, v := range fmts {
		l := m[v]
		n := []string{}
		for _, w := range l {
			for _, x := range fmts {
				if w == x {
					n = append(n, w)
					break
				}
			}
		}
		ret[v] = n
	}
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
		var indexes []int
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

// take a list of indexes, subtract the length of the previous priority list in a set (or 0) to get relative indexes,
// then map those against a priority list. Re-number according to indexes and return the common subset.
func (l List) Subset(indexes []int, prev int) List {
	if l == nil {
		return nil
	}
	submap := make(map[int]int)
	for i, v := range indexes {
		submap[v-prev] = i
	}
	subset := make(List, len(indexes))
	for i, v := range indexes {
		ns := make([]int, 0, len(l[v-prev]))
		for _, w := range l[v-prev] {
			if idx, ok := submap[w]; ok {
				ns = append(ns, idx)
			}
		}
		subset[i] = ns
	}
	return subset
}

func (l List) String() string {
	if l == nil {
		return "priority list: nil"
	}
	return fmt.Sprintf("priority list: %v", [][]int(l))
}

// A priority set holds a number of priority lists
// Fields exported so gobbable
type Set struct {
	Idx   []int
	Lists []List
}

// Add a priority list to a set. The length is the number of signatures the priority list applies to, not the length of the priority list.
// This length will only differ when no priorities are set for a given set of signatures.
func (s *Set) Add(l List, length int) {
	var last int
	if len(s.Idx) > 0 {
		last = s.Idx[len(s.Idx)-1]
	}
	s.Idx = append(s.Idx, length+last)
	s.Lists = append(s.Lists, l)
}

func (s *Set) list(i, j int) []int {
	if s.Lists[i] == nil {
		return nil
	} else {
		l := s.Lists[i][j]
		if l == nil {
			l = make([]int, 0)
		}
		return l
	}
}

// return the index of the s.Lists for the wait list, and return the previous tally
// previous tally is necessary for adding to the values in the priority list to give real priorities
func (s *Set) Index(i int) (int, int) {
	var prev int
	for idx, v := range s.Idx {
		if i < v {
			return idx, prev
		}
		prev = v
	}
	// should never get here. Signal error
	return -1, -1
}

// A wait set is a mutating structure that holds the set of indexes that should be waited for while matching underway
type WaitSet struct {
	*Set
	wait [][]int
	m    *sync.RWMutex
}

func (s *Set) WaitSet() *WaitSet {
	return &WaitSet{
		s,
		make([][]int, len(s.Lists)),
		&sync.RWMutex{},
	}
}

// Set the priority list & return a boolean indicating whether the WaitSet is satisfied such that matching can stop (i.e. no priority list is nil, and all are empty)
func (w *WaitSet) Put(i int) bool {
	idx, prev := w.Index(i)
	l := w.list(idx, i-prev)
	// no priorities for this set, return false immediately
	if l == nil {
		return false
	}
	w.m.Lock()
	defer w.m.Unlock()
	w.wait[idx] = l
	// if we have any priorities, then we aren't satisified
	if len(l) > 0 {
		return false
	}
	// if l is 0 and we have only one priority set, then we are satisfied
	if len(w.wait) == 1 {
		return true
	}
	// otherwise, let's check all the other priority sets
	for i, v := range w.wait {
		if i == idx {
			continue
		}
		if v == nil {
			return false
		}
		if len(v) > 0 {
			return false
		}
	}
	return true
}

// Check a signature index against the appropriate priority list. Should we continue trying to match this signature?
func (w *WaitSet) Check(i int) bool {
	idx, prev := w.Index(i)
	w.m.RLock()
	defer w.m.RUnlock()
	return w.check(i, idx, prev)
}

func (w *WaitSet) check(i, idx, prev int) bool {
	if w.wait[idx] == nil {
		return true
	}
	j := sort.SearchInts(w.wait[idx], i-prev)
	if j == len(w.wait[idx]) || w.wait[idx][j] != i-prev {
		return false
	}
	return true
}

// Filter a waitset with a list of potential matches, return only those that we are still waiting on. Return nil if none.
func (w *WaitSet) Filter(l []int) []int {
	ret := make([]int, 0, len(l))
	w.m.RLock()
	defer w.m.RUnlock()
	for _, v := range l {
		idx, prev := w.Index(v)
		if w.check(v, idx, prev) {
			ret = append(ret, v)
		}
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

// Filter a waitset with a list of potential matches, return only those still waiting on. Return nil if none.
// This a convenience method that takes key frame IDs rather than sig IDs.
func (w *WaitSet) FilterKfID(l []process.KeyFrameID) []process.KeyFrameID {
	ret := make([]process.KeyFrameID, 0, len(l))
	w.m.RLock()
	defer w.m.RUnlock()
	for _, v := range l {
		idx, prev := w.Index(v[0])
		if w.check(v[0], idx, prev) {
			ret = append(ret, v)
		}
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

// For periodic checking - what signatures are we currently waiting on?
// Accumulates values from all the priority lists within the set.
// Returns nil if *any* of the priority lists is nil.
func (w *WaitSet) WaitingOn() []int {
	w.m.RLock()
	defer w.m.RUnlock()
	var l int
	for _, v := range w.wait {
		if v == nil {
			return nil
		}
		l = l + len(v)
	}
	ret := make([]int, l)
	var prev, j int
	for i, v := range w.wait {
		for _, w := range v {
			ret[j] = w + prev
			j++
		}
		prev = w.Idx[i]
	}
	return ret
}
