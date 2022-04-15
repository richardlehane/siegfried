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

package containermatcher

import (
	"fmt"
	"path/filepath"

	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/internal/siegreader"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

func (m Matcher) Identify(n string, b *siegreader.Buffer, hints ...core.Hint) (chan core.Result, error) {
	res := make(chan core.Result)
	// check trigger
	buf, err := b.Slice(0, 8)
	if err != nil {
		close(res)
		return res, nil
	}
	divhints := m.divideHints(hints)
	for i, c := range m {
		if c.trigger(buf) {
			rdr, err := c.rdr(b)
			if err != nil {
				close(res)
				return res, err
			}
			go c.identify(n, rdr, res, divhints[i]...)
			return res, nil
		}
	}
	// nothing ... move on
	close(res)
	return res, nil
}

// ranges allows referencing a container hit back to a specific container matcher (used by divideHints)
// returns running number / matcher index / identifier index
func (m Matcher) ranges() [][3]int {
	var l int
	for _, c := range m {
		l += len(c.startIndexes)
	}
	ret := make([][3]int, l)
	var prev, this, idx, jdx int
	for i := range ret {
		prev = this
		this = m[idx].startIndexes[jdx]
		ret[i] = [3]int{prev + this, idx, jdx}
		idx++
		if idx >= len(m) {
			jdx++
			idx = 0
		}
	}
	return ret
}

func findID(id int, rng [][3]int) (int, int) {
	var idx int
	for idx = range rng {
		if rng[idx][0] > id {
			if idx > 0 {
				idx--
			}
			break
		}
	}
	return rng[idx][1], rng[idx][2]
}

func (m Matcher) divideHints(hints []core.Hint) [][]core.Hint {
	ret := make([][]core.Hint, len(m))
	rng := m.ranges()
	for _, h := range hints {
		if len(h.Pivot) == 0 {
			continue
		}
		first := make([]bool, len(m))
		for _, p := range h.Pivot {
			midx, iidx := findID(p, rng)
			if !first[midx] {
				first[midx] = true
				_, excl := m[midx].priorities.Index(p - m[midx].startIndexes[iidx])
				ret[midx] = append(ret[midx], core.Hint{excl, nil})
			}
			ret[midx][len(ret[midx])-1].Pivot = append(ret[midx][len(ret[midx])-1].Pivot, p-m[midx].startIndexes[iidx])
		}
	}
	return ret
}

type identifier struct {
	partsMatched [][]hit // hits for parts
	ruledOut     []bool  // mark additional signatures as negatively matched
	waitSet      *priority.WaitSet
	hits         []hit // shared buffer of hits used when matching
	result       bool
}

func (c *ContainerMatcher) newIdentifier(numParts int, hints ...core.Hint) *identifier {
	return &identifier{
		make([][]hit, numParts),
		make([]bool, numParts),
		c.priorities.WaitSet(hints...),
		make([]hit, 0, 1),
		false,
	}
}

func (c *ContainerMatcher) identify(n string, rdr Reader, res chan core.Result, hints ...core.Hint) {
	// safe to call on a nil matcher (i.e. container matching switched off)
	if c == nil {
		close(res)
		return
	}
	id := c.newIdentifier(len(c.parts), hints...)
	var err error
	for err = rdr.Next(); err == nil; err = rdr.Next() {
		ct, ok := c.nameCTest[rdr.Name()]
		if !ok {
			continue
		}
		if config.Debug() {
			fmt.Fprintf(config.Out(), "{Name match - %s (container %d))}\n", rdr.Name(), c.conType)
		}
		// name has matched, let's test the CTests
		// ct.identify will generate a slice of hits which pass to
		// processHits which will return true if we can stop
		if c.processHits(ct.identify(c, id, rdr, rdr.Name()), id, ct, rdr.Name(), res) {
			break
		}
	}
	// send a default hit if no result and extension matches
	if c.extension != "" && !id.result && filepath.Ext(n) == "."+c.extension {
		res <- defaultHit(-1 - int(c.conType))
	}
	close(res)
}

func (ct *cTest) identify(c *ContainerMatcher, id *identifier, rdr Reader, name string) []hit {
	// reset hits
	id.hits = id.hits[:0]
	for _, h := range ct.satisfied {
		if id.waitSet.Check(h) {
			id.hits = append(id.hits, hit{h, name, "name only"})
		}
	}
	if ct.unsatisfied != nil && !rdr.IsDir() {
		buf, err := rdr.SetSource(c.entryBufs)
		if buf == nil {
			rdr.Close()
			if config.Debug() {
				fmt.Fprintf(config.Out(), "{Container error - %s (container %d)); error: %v}\n", rdr.Name(), c.conType, err)
			}
			return id.hits
		}
		bmc, _ := ct.bm.Identify("", buf)
		for r := range bmc {
			h := ct.unsatisfied[r.Index()]
			if id.waitSet.Check(h) && id.checkHits(h) {
				id.hits = append(id.hits, hit{h, name, r.Basis()})
			}
		}
		rdr.Close()
		c.entryBufs.Put(buf)
	}
	return id.hits
}

// process the hits from the ctest: adding hits to the parts matched, checking priorities
// return true if satisfied and can quit
func (c *ContainerMatcher) processHits(hits []hit, id *identifier, ct *cTest, name string, res chan core.Result) bool {
	// if there are no hits, rule out any sigs in the ctest
	if len(hits) == 0 {
		for _, v := range ct.satisfied {
			id.ruledOut[v] = true
		}
		for _, v := range ct.unsatisfied {
			id.ruledOut[v] = true
		}
		return false
	}
	for _, h := range hits {
		id.partsMatched[h.id] = append(id.partsMatched[h.id], h)
		if len(id.partsMatched[h.id]) == c.parts[h.id] {
			if id.waitSet.Check(h.id) {
				idx, _ := c.priorities.Index(h.id)
				res <- toResult(c.startIndexes[idx], id.partsMatched[h.id]) // send a Result here
				id.result = true                                            // mark id as having a result (for zip default)
				// set a priority list and return early if can
				if id.waitSet.Put(h.id) {
					return true
				}
			}
		}
	}
	// if nothing ruled out by this test, then we must continue
	if len(hits) == len(ct.satisfied)+len(ct.unsatisfied) {
		return false
	}
	// we can rule some possible matches out...
	for _, v := range ct.satisfied {
		if len(id.partsMatched[v]) == 0 || id.partsMatched[v][len(id.partsMatched[v])-1].name != name {
			id.ruledOut[v] = true
		}
	}
	for _, v := range ct.unsatisfied {
		if len(id.partsMatched[v]) == 0 || id.partsMatched[v][len(id.partsMatched[v])-1].name != name {
			id.ruledOut[v] = true
		}
	}
	// if we haven't got a waitList yet, then we should return false
	waitingOn := id.waitSet.WaitingOn()
	if waitingOn == nil {
		return false
	}
	// loop over the wait list, seeing if they are all ruled out
	for _, v := range waitingOn {
		if !id.ruledOut[v] {
			return false
		}
	}
	return true
}

// eliminate duplicate hits - must do this since rely on number of matches for each sig as test for full match
func (id *identifier) checkHits(i int) bool {
	for _, h := range id.hits {
		if i == h.id {
			return false
		}
	}
	return true
}

func toResult(i int, h []hit) result {
	if len(h) == 0 {
		return result(h)
	}
	h[0].id += i
	return result(h)
}

type result []hit

func (r result) Index() int {
	if len(r) == 0 {
		return -1
	}
	return r[0].id
}

func (r result) Basis() string {
	var basis string
	for i, v := range r {
		if i < 1 {
			basis += "container "
		} else {
			basis += "; "
		}
		basis += "name " + v.name
		if len(v.basis) > 0 {
			basis += " with " + v.basis
		}
	}
	return basis
}

type hit struct {
	id    int
	name  string
	basis string
}

type defaultHit int

func (d defaultHit) Index() int {
	return int(d)
}

func (d defaultHit) Basis() string {
	return "container match with trigger and default extension"
}
