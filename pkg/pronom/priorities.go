package pronom

import (
	"github.com/richardlehane/siegfried/pkg/core/priority"
)

// returns a map of puids and the indexes of byte signatures that those puids should give priority to
func (p pronom) priorities() priority.Map {
	pMap := make(priority.Map)
	for _, f := range p.droid.FileFormats {
		superior := f.Puid
		for _, v := range f.Priorities {
			subordinate := p.ids[v]
			pMap.Add(subordinate, superior)
		}
	}
	pMap.Complete()
	return pMap
}
