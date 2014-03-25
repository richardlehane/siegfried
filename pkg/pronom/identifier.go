package pronom

import (
	"sync"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type PronomIdentifier struct {
	Bm         *bytematcher.Bytematcher
	Puids      []string
	Priorities map[string][]int
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

func (pi *PronomIdentifier) Identify(b *siegreader.Buffer, n string, c chan core.Identification, wg *sync.WaitGroup) {
	if len(n) > 0 {

	}
	ids, limit := pi.Bm.Identify(b)
	for i := range ids {
		puid := pi.Puids[i]
		c <- PronomIdentification{puid, 0.9}
		l, ok := pi.Priorities[puid]
		if !ok {
			close(limit)
			break
		} else {
			limit <- l
		}
	}
	wg.Done()
}
