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
	"path/filepath"
	"strings"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

func (m Matcher) Identify(n string, b *siegreader.Buffer) chan core.Result {
	res := make(chan core.Result)
	// check trigger
	buf, err := b.Slice(0, 8)
	if err != nil {
		close(res)
		return res
	}
	for _, c := range m {
		if c.trigger(buf) {
			if q, i := c.defaultMatch(n); q {
				go func() {
					res <- defaultHit(i)
					close(res)
				}()
				return res
			}
			rdr, err := c.rdr(b)
			if err != nil {
				close(res)
				return res
			}
			go c.identify(rdr, res)
			return res
		}
	}
	// nothing ... move on
	close(res)
	return res
}

func (c *ContainerMatcher) defaultMatch(n string) (bool, int) {
	if c.Default == "" {
		return false, 0
	}
	ext := filepath.Ext(n)
	if len(ext) > 0 && strings.TrimPrefix(ext, ".") == c.Default {
		// the default is a negative value calculated from the CType
		return true, -1 - int(c.CType)
	}
	return false, 0
}

func (c *ContainerMatcher) identify(rdr Reader, res chan core.Result) {
	// safe to call on a nil matcher
	if c == nil {
		close(res)
		return
	}
	// reset
	c.waitSet = c.Priorities.WaitSet()
	if c.started {
		for i := range c.partsMatched {
			c.partsMatched[i] = c.partsMatched[i][:0]
			c.ruledOut[i] = false
		}
	} else {
		c.entryBuf = siegreader.New()
		c.partsMatched = make([][]hit, len(c.Parts))
		c.ruledOut = make([]bool, len(c.Parts))
		c.hits = make([]hit, 0, 20) // shared hits buffer to avoid allocs
		c.started = true
	}
	var err error
	for err = rdr.Next(); err == nil; err = rdr.Next() {
		ct, ok := c.NameCTest[rdr.Name()]
		if !ok {
			continue
		}
		// name has matched, lets test the CTests
		// ct.identify will generate a slice of hits which pass to
		// processHits which will return true if we can stop
		if c.processHits(ct.identify(c, rdr, rdr.Name()), ct, rdr.Name(), res) {
			break
		}
	}
	close(res)
}

func (ct *CTest) identify(c *ContainerMatcher, rdr Reader, name string) []hit {
	// reset hits
	c.hits = c.hits[:0]
	for _, h := range ct.Satisfied {
		if c.waitSet.Check(h) {
			c.hits = append(c.hits, hit{h, name, "name only"})
		}
	}
	if ct.Unsatisfied != nil {
		rdr.SetSource(c.entryBuf) // NOTE: an error is ignored here.
		for r := range ct.BM.Identify("", c.entryBuf) {
			h := ct.Unsatisfied[r.Index()]
			if c.waitSet.Check(h) && c.checkHits(h) {
				c.hits = append(c.hits, hit{h, name, r.Basis()})
			}
		}
		rdr.Close()
	}
	return c.hits
}

// process the hits from the ctest: adding hits to the parts matched, checking priorities
// return true if satisfied and can quit
func (c *ContainerMatcher) processHits(hits []hit, ct *CTest, name string, res chan core.Result) bool {
	// if there are no hits, rule out any sigs in the ctest
	if len(hits) == 0 {
		for _, v := range ct.Satisfied {
			c.ruledOut[v] = true
		}
		for _, v := range ct.Unsatisfied {
			c.ruledOut[v] = true
		}
		return false
	}
	for _, h := range hits {
		c.partsMatched[h.id] = append(c.partsMatched[h.id], h)
		if len(c.partsMatched[h.id]) == c.Parts[h.id] {
			if c.waitSet.Check(h.id) {
				idx, _ := c.Priorities.Index(h.id)
				res <- toResult(c.Sindexes[idx], c.partsMatched[h.id]) // send a Result here
				// set a priority list and return early if can
				if c.waitSet.Put(h.id) {
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
		if len(c.partsMatched[v]) == 0 || c.partsMatched[v][len(c.partsMatched[v])-1].name != name {
			c.ruledOut[v] = true
		}
	}
	for _, v := range ct.Unsatisfied {
		if len(c.partsMatched[v]) == 0 || c.partsMatched[v][len(c.partsMatched[v])-1].name != name {
			c.ruledOut[v] = true
		}
	}
	// if we haven't got a waitList yet, then we should return false
	waitingOn := c.waitSet.WaitingOn()
	if waitingOn == nil {
		return false
	}
	// loop over the wait list, seeing if they are all ruled out
	for _, v := range waitingOn {
		if !c.ruledOut[v] {
			return false
		}
	}
	return true
}

// eliminate duplicate hits - must do this since rely on number of matches for each sig as test for full match
func (c *ContainerMatcher) checkHits(i int) bool {
	for _, h := range c.hits {
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
