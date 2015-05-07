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
	"strings"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

func (m Matcher) Identify(n string, b siegreader.Buffer) (chan core.Result, error) {
	res := make(chan core.Result)
	// check trigger
	buf, err := b.Slice(0, 8)
	if err != nil {
		close(res)
		return res, nil
	}
	for _, c := range m {
		if c.trigger(buf) {
			rdr, err := c.rdr(b)
			if err != nil {
				close(res)
				return res, err
			}
			go c.identify(rdr, res)
			return res, nil
		}
	}
	// nothing ... move on
	close(res)
	return res, nil
}

type identifier struct {
	partsMatched [][]hit // hits for parts
	ruledOut     []bool  // mark additional signatures as negatively matched
	waitSet      *priority.WaitSet
	hits         []hit // shared buffer of hits used when matching
}

func (c *ContainerMatcher) newIdentifier(numParts int) *identifier {
	return &identifier{
		make([][]hit, numParts),
		make([]bool, numParts),
		c.Priorities.WaitSet(),
		make([]hit, 0, 1),
	}
}

func (c *ContainerMatcher) identify(rdr Reader, res chan core.Result) {
	// safe to call on a nil matcher (i.e. container matching switched off)
	if c == nil {
		close(res)
		return
	}
	id := c.newIdentifier(len(c.Parts))
	var err error
	var hit bool
	for err = rdr.Next(); err == nil; err = rdr.Next() {
		ct, ok := c.NameCTest[rdr.Name()]
		if !ok {
			continue
		}
		// name has matched, lets test the CTests
		// ct.identify will generate a slice of hits which pass to
		// processHits which will return true if we can stop
		if c.processHits(ct.identify(c, id, rdr, rdr.Name()), id, ct, rdr.Name(), res) {
			hit = true
			break
		}
	}
	// if we have no hits and a default value for this matcher, send it
	if !hit && c.Default {
		// the default is a negative value calculated from the CType
		res <- defaultHit(-1 - int(c.CType))
	}
	close(res)
}

func (ct *CTest) identify(c *ContainerMatcher, id *identifier, rdr Reader, name string) []hit {
	// reset hits
	id.hits = id.hits[:0]
	for _, h := range ct.Satisfied {
		if id.waitSet.Check(h) {
			id.hits = append(id.hits, hit{h, name, "name only"})
		}
	}
	if ct.Unsatisfied != nil {
		buf, _ := rdr.SetSource(c.entryBufs) // NOTE: an error is ignored here.
		bmc, _ := ct.BM.Identify("", buf)
		for r := range bmc {
			h := ct.Unsatisfied[r.Index()]
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
func (c *ContainerMatcher) processHits(hits []hit, id *identifier, ct *CTest, name string, res chan core.Result) bool {
	// if there are no hits, rule out any sigs in the ctest
	if len(hits) == 0 {
		for _, v := range ct.Satisfied {
			id.ruledOut[v] = true
		}
		for _, v := range ct.Unsatisfied {
			id.ruledOut[v] = true
		}
		return false
	}
	for _, h := range hits {
		id.partsMatched[h.id] = append(id.partsMatched[h.id], h)
		if len(id.partsMatched[h.id]) == c.Parts[h.id] {
			if id.waitSet.Check(h.id) {
				idx, _ := c.Priorities.Index(h.id)
				res <- toResult(c.Sindexes[idx], id.partsMatched[h.id]) // send a Result here
				// set a priority list and return early if can
				if id.waitSet.Put(h.id) {
					return true
				}
			}
		}
	}
	// if nothing ruled out by this test, then we must continue
	if len(hits) == len(ct.Satisfied)+len(ct.Unsatisfied) {
		return false
	}
	// we can rule some possible matches out...
	for _, v := range ct.Satisfied {
		if len(id.partsMatched[v]) == 0 || id.partsMatched[v][len(id.partsMatched[v])-1].name != name {
			id.ruledOut[v] = true
		}
	}
	for _, v := range ct.Unsatisfied {
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
