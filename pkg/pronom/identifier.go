package pronom

import (
	"sync"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type PronomIdentifier struct {
	Bm    *bytematcher.Bytematcher
	Puids []string
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

func (pi *PronomIdentifier) Identify(b *siegreader.Buffer, c chan core.Identification, wg *sync.WaitGroup) {
	ids, _ := pi.Bm.Identify(b)
	for i := range ids {
		c <- PronomIdentification{pi.Puids[i], 0.9}
	}
	wg.Done()
}
