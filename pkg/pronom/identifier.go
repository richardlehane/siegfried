package pronom

import (
	"fmt"
	"sort"
	"sync"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/core/namematcher"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type PronomIdentifier struct {
	SigVersion SigVersion
	BPuids     []string         // slice of puids that corresponds to the bytematcher's int signatures
	PuidsB     map[string][]int // map of puids to slices of bytematcher int signatures
	EPuids     []string         // slice of puids that corresponds to the extension matcher's int signatures
	Priorities map[string][]int // map of priorities - puids to bytematcher int signatures
	bm         bytematcher.Matcher
	em         namematcher.Matcher
	ids        pids
}

type PronomIdentification struct {
	puid       string
	confidence float64
}

func (pid PronomIdentification) String() string {
	return pid.puid
}

func (pid PronomIdentification) Confidence() float64 {
	return pid.confidence
}

func (pid PronomIdentification) Basis() string {
	return "because I said so" //obviously this needs to be changed!
}

func (pi *PronomIdentifier) Version() string {
	return fmt.Sprintf("Signature version: %d; based on droid sig: %s; and container sig: %s", pi.SigVersion.Gob, pi.SigVersion.Droid, pi.SigVersion.Containers)
}

func (pi *PronomIdentifier) Update(i int) bool {
	return i > pi.SigVersion.Gob
}

func (pi *PronomIdentifier) Identify(b *siegreader.Buffer, n string, c chan core.Identification, wg *sync.WaitGroup) {
	pi.ids = pi.ids[:0]
	var ems []int
	/*
		if len(n) > 0 {
			ems = pi.em.Identify(n)
			for _, v := range ems {
				pi.ids = add(pi.ids, pi.EPuids[v], 0.1)
			}
		}
	*/
	var cscore float64 = 0.1
	pi.bm.Start()
	ids, wait := pi.bm.Identify(b)

	for i := range ids {
		cscore *= 1.1
		puid := pi.BPuids[i]
		pi.ids = add(pi.ids, puid, cscore)
		wait <- pi.priorities(i, ems)
	}

	if len(pi.ids) > 0 {
		sort.Sort(pi.ids)
		conf := pi.ids[0].confidence
		c <- pi.ids[0]
		if len(pi.ids) > 1 {
			for i, v := range pi.ids[1:] {
				if v.confidence == conf {
					c <- pi.ids[i+1]
				} else {
					break
				}
			}
		}
	}
	wg.Done()
}

type pids []PronomIdentification

func (p pids) Len() int { return len(p) }

func (p pids) Less(i, j int) bool { return p[j].confidence < p[i].confidence }

func (p pids) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func add(p pids, f string, c float64) pids {
	for i, v := range p {
		if v.puid == f {
			p[i].confidence += c
			return p
		}
	}
	return append(p, PronomIdentification{f, c})
}

// Deal with non-explicity priorities
// This is where there is no HasPriority relation but we should still wait anyway as we have an extension match
// Rule:-
// - for each extension match
// 	 - if the ID in question is not a priority of that extension match
//   - add BMIds to the wait list for that extension
func (pi *PronomIdentifier) priorities(id int, ems []int) []int {
	w, ok := pi.Priorities[pi.BPuids[id]]
	if !ok {
		w = []int{}
	}
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

func (pi *PronomIdentifier) String() string {
	return pi.bm.String()
}
