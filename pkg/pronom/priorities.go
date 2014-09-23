package pronom

import "github.com/richardlehane/siegfried/pkg/core/priority"

// returns a map of puids and the indexes of byte signatures that those puids should give priority to
func (p pronom) priorities(puids []string) priority.Map {
	pMap := make(priority.Map)
	var iter int
	for _, f := range p.droid.FileFormats {
		for _ = range f.Signatures {
			for _, v := range f.Priorities {
				subordinate := p.ids[v]
				superior := puids[iter]
				pMap.Add(subordinate, superior)
			}
			for _, r := range f.Relations {
				// only interested in subtypes
				if r.Type != "Is subtype of" {
					continue
				}
				subordinate := p.ids[r.ID]
				superior := puids[iter]
				pMap.Add(subordinate, superior)
			}
			iter++
		}
	}
	pMap.Complete()
	return pMap
}
