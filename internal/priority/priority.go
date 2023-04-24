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

	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/pkg/core"
)

// a priority map links subordinate results to a list of priority results
type Map map[string][]string

func (m Map) Difference(mb Map) Map {
	mc := make(Map)
	for k, v := range m {
		vb, ok := mb[k]
		if !ok {
			mc[k] = v
			continue
		}
		e := extras(v, vb)
		if len(e) > 0 {
			mc[k] = e
		}
	}
	return mc
}

func (m Map) Elements() [][2]string {
	fmts := make(map[string]bool)
	elements := make([][2]string, 0, len(m)*3)
	for k, v := range m {
		for _, sup := range v {
			elements = append(elements, [2]string{k, sup})
			fmts[sup] = true
		}
	}
	for k, v := range m {
		if len(v) == 0 && !fmts[k] {
			elements = append(elements, [2]string{k, ""})
		}
	}
	return elements
}

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

// make sure that a set of superiors doesn't include self
func trimSelf(ss []string, s string) []string {
	if !containsStr(ss, s) {
		return ss
	}
	ret := make([]string, 0, len(ss))
	for _, v := range ss {
		if v != s {
			ret = append(ret, v)
		}
	}
	return ret
}

// add a subordinate-superior relationship to the priority map
func (m Map) Add(subordinate string, superior string) {
	if subordinate == "" || superior == "" || subordinate == superior {
		return
	}
	_, ok := m[subordinate]
	if ok {
		m[subordinate] = addStr(m[subordinate], superior)
		return
	}
	m[subordinate] = []string{superior}
}

// create a list of all strings that appear in 'a' but not 'b', 'c', 'd', ...
func extras(a []string, bs ...[]string) []string {
	ret := make([]string, 0, len(a))
	for _, v := range a {
		var exists bool
	outer:
		for _, b := range bs {
			for _, v1 := range b {
				if v == v1 {
					exists = true
					break outer
				}
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
			ret = append(ret, extras(priorityPriorities, vals, ret)...)
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
	for k := range m {
		extraPriorities := m.priorityWalk(k)
		extraPriorities = trimSelf(extraPriorities, k)
		m[k] = append(m[k], extras(extraPriorities, m[k])...)
		sort.Strings(m[k])
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

// is this a superior result? (it has no superiors among the set of intial hits)
func superior(sups, hits []string) bool {
	for _, sup := range sups {
		for _, hit := range hits {
			if sup == hit {
				return false
			}
		}
	}
	return true
}

// Apply checks a list of hits against a priority map and returns a subset of that list for any hits
// that don't have superiors also in that list
func (m Map) Apply(hits []string) []string {
	ret := make([]string, 0, len(hits))
	for _, hit := range hits {
		if superior(m[hit], hits) {
			ret = append(ret, hit)
		}
	}
	return ret
}

// return a priority list using the indexes from the supplied slice of keys (keys can be duplicated in that slice)
func (m Map) List(keys []string) List {
	if m == nil {
		return nil
	}
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
type Set struct {
	idx        []int
	lists      []List
	maxOffsets [][2]int
}

func (s *Set) Save(ls *persist.LoadSaver) {
	ls.SaveInts(s.idx)
	ls.SaveSmallInt(len(s.lists))
	for _, v := range s.lists {
		ls.SaveSmallInt(len(v))
		for _, w := range v {
			ls.SaveInts(w)
		}
	}
	ls.SaveSmallInt(len(s.maxOffsets))
	for _, v := range s.maxOffsets {
		ls.SaveInt(v[0])
		ls.SaveInt(v[1])
	}
}

func Load(ls *persist.LoadSaver) *Set {
	set := &Set{}
	set.idx = ls.LoadInts()
	set.lists = make([]List, ls.LoadSmallInt())
	for i := range set.lists {
		le := ls.LoadSmallInt()
		if le == 0 {
			continue
		}
		set.lists[i] = make(List, le)
		for j := range set.lists[i] {
			set.lists[i][j] = ls.LoadInts()
		}
	}
	set.maxOffsets = make([][2]int, ls.LoadSmallInt())
	for i := range set.maxOffsets {
		set.maxOffsets[i] = [2]int{ls.LoadInt(), ls.LoadInt()}
	}
	return set
}

// Add a priority list to a set. The length is the number of signatures the priority list applies to, not the length of the priority list.
// This length will only differ when no priorities are set for a given set of signatures.
func (s *Set) Add(l List, length, bof, eof int) {
	var last int
	if len(s.idx) > 0 {
		last = s.idx[len(s.idx)-1]
	}
	s.idx = append(s.idx, length+last)
	s.lists = append(s.lists, l)
	s.maxOffsets = append(s.maxOffsets, [2]int{bof, eof})
}

func (s *Set) list(i, j int) []int {
	if s.lists[i] == nil {
		return nil
	} else {
		l := s.lists[i][j]
		if l == nil {
			l = []int{}
		}
		return l
	}
}

// at given BOF and EOF offsets, should we still wait on a given priority set?
func (s *Set) await(idx int, bof, eof int64) bool {
	if s.maxOffsets[idx][0] < 0 || (s.maxOffsets[idx][0] > 0 && int64(s.maxOffsets[idx][0]) >= bof) {
		return true
	}
	if s.maxOffsets[idx][1] < 0 || (s.maxOffsets[idx][1] > 0 && int64(s.maxOffsets[idx][1]) >= eof) {
		return true
	}
	return false
}

// Index return the index of the s.lists for the wait list, and return the previous tally
// previous tally is necessary for adding to the values in the priority list to give real priorities
func (s *Set) Index(i int) (int, int) {
	var prev int
	for idx, v := range s.idx {
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
	wait  [][]int // a nil list means we're not waiting on anything yet; an empty list means nothing to wait for i.e. satisifed
	this  []int   // record last hit so can avoid pivotting to weaker matches
	pivot [][]int // a pivot list is a list of indexes that we could potentially pivot to. E.g. for a .pdf file that has mp3 signatures, but is actually a PDF
	m     *sync.RWMutex
}

// WaitSet creates a new WaitSet given a list of hints
func (s *Set) WaitSet(hints ...core.Hint) *WaitSet {
	ws := &WaitSet{
		s,
		make([][]int, len(s.lists)),
		make([]int, len(s.lists)),
		make([][]int, len(s.lists)),
		&sync.RWMutex{},
	}
	for _, h := range hints {
		idx, _ := s.Index(h.Exclude)
		if h.Pivot == nil { // if h.Pivot is nil (as opposed to empty slice), it is a signal that that matcher is satisfied
			ws.wait[idx] = []int{}
		} else {
			ws.pivot[idx] = h.Pivot
		}
	}
	return ws
}

// MaxOffsets returns max/min offset info in order to override the max/min offsets set on the bytematcher when
// any identifiers have been excluded.
func (w *WaitSet) MaxOffsets() (int, int) {
	var bof, eof int
	for i, v := range w.wait {
		if v == nil {
			if bof >= 0 && (w.maxOffsets[i][0] < 0 || bof < w.maxOffsets[i][0]) {
				bof = w.maxOffsets[i][0]
			}
			if eof >= 0 && (w.maxOffsets[i][1] < 0 || eof < w.maxOffsets[i][1]) {
				eof = w.maxOffsets[i][1]
			}
		}
	}
	return bof, eof
}

func inPivot(i int, ii []int) bool {
	for _, v := range ii {
		if i == v {
			return true
		}
	}
	return false
}

func mightPivot(i int, ii []int) bool {
	return len(ii) > 0 && !inPivot(i, ii)
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
	// set the wait list
	w.wait[idx] = l
	// set this
	w.this[idx] = i - prev
	mp := mightPivot(i, w.pivot[idx])
	if !mp {
		w.pivot[idx] = nil // ditch the pivot list if it is just confirming a match or empty
	}
	// if we have any priorities, then we aren't satisified
	if len(l) > 0 || mp {
		return false
	}
	// if l is 0, and we have only one priority set, and we're not going to pivot, then we are satisfied
	if len(w.wait) == 1 && !mp {
		return true
	}
	// otherwise, let's check all the other priority sets for wait sets or pivot lists
	for i, v := range w.wait {
		if i == idx {
			continue
		}
		if v == nil || len(v) > 0 || len(w.pivot[i]) > 0 {
			return false
		}
	}
	return true
}

// Set the priority list & return a boolean indicating whether the WaitSet is satisfied such that matching can stop (i.e. no priority list is nil, and all are empty)
func (w *WaitSet) PutAt(i int, bof, eof int64) bool {
	idx, prev := w.Index(i)
	l := w.list(idx, i-prev)
	// no priorities for this set, return false immediately
	if l == nil && w.await(idx, bof, eof) {
		return false
	}
	w.m.Lock()
	defer w.m.Unlock()
	// set the wait list
	w.wait[idx] = l
	// set this
	w.this[idx] = i - prev
	mp := mightPivot(i, w.pivot[idx])
	if !mp {
		w.pivot[idx] = nil // ditch the pivot list if it is just confirming a match or empty
	}
	// if we have any priorities, then we aren't satisified
	if (len(l) > 0 || mp) && w.await(idx, bof, eof) {
		return false
	}
	// if l is 0, and we have only one priority set, and we're not going to pivot, then we are satisfied
	if len(w.wait) == 1 && !mp {
		return true
	}
	// otherwise, let's check all the other priority sets
	for i, v := range w.wait {
		if i == idx {
			continue
		}
		if w.await(i, bof, eof) {
			if v == nil || len(v) > 0 || len(w.pivot[i]) > 0 {
				return false
			}
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
		if inPivot(i, w.pivot[idx]) {
			l := w.list(idx, i-prev)
			k := sort.SearchInts(l, w.this[idx])
			if k < len(l) && l[k] == w.this[idx] {
				return false
			}
			return true
		}
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

type Filterable interface {
	Next() int
	Mark(bool)
}

func (w *WaitSet) ApplyFilter(f Filterable) {
	w.m.RLock()
	defer w.m.RUnlock()
	for i := f.Next(); i > -1; i = f.Next() {
		idx, prev := w.Index(i)
		f.Mark(w.check(i, idx, prev))
	}
}

// For periodic checking - what signatures are we currently waiting on?
// Accumulates values from all the priority lists within the set.
// Returns nil if *any* of the priority lists is nil.
func (w *WaitSet) WaitingOn() []int {
	w.m.RLock()
	defer w.m.RUnlock()
	var l int
	for i, v := range w.wait {
		if v == nil {
			return nil
		}
		l = l + len(v) + len(w.pivot[i])
	}
	ret := make([]int, l)
	var prev, j int
	for i, v := range w.wait {
		for _, x := range v {
			ret[j] = x + prev
			j++
		}
		copy(ret[j:], w.pivot[i])
		j += len(w.pivot[i])
		prev = w.idx[i]
	}
	return ret
}
