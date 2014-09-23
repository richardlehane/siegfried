package containermatcher

import (
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

const (
	mscfbTrigger = 0xE11AB1A1E011CFD0
	zipTrigger   = 0x04034B50
)

func (c *ContainerMatcher) Identify(b *siegreader.Buffer) (core.Result, bool) {
	// reset
	c.partsMatched = make([]int, len(c.Parts))
	c.nameMatches = c.nameMatches[:0]
	// check trigger
	buf, err := b.Slice(0, 8)
	var res core.Result
	if err != nil {
		return core.Result{}, false
	}
	var rdr Reader
	if buf.uint64() == mscfbTrigger {
		rdr = newMscfb(b)
	} else if buf[:4].uint32() == zipTrigger {
		rdr = newZip(b)
	} else {
		return res, false
	}
	var success bool
Loop:
	for err := rdr.Next(); err != nil; err = rdr.Next() {
		ct, ok := c.NameCTest[rdr.Name()]
		if !ok {
			continue
		}
		matches := ct.identify(rdr)
		c.nameMatches = append(c.nameMatches, rdr.Name())
		for _, m := range matches {
			c.partsMatched[m.Index]++
			if c.partsMatched[m.Index] == c.Parts[m.Index] && len(c.Priorities[m.Index]) == 0 {
				res = m
				success = true
				break Loop
			}
		}
		// ugly
		// loop again, checking if priorities exhausted
		for _, m := range matches {
			if c.partsMatched[m.Index] == c.Parts[m.Index] {

			}
		}
	}
	rdr.Quit()
	return res, success
}
