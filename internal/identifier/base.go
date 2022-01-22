// Copyright 2016 Richard Lehane. All rights reserved.
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

package identifier

import (
	"fmt"
	"strings"
	"sync"

	"github.com/richardlehane/siegfried/internal/bytematcher"
	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/containermatcher"
	"github.com/richardlehane/siegfried/internal/mimematcher"
	"github.com/richardlehane/siegfried/internal/namematcher"
	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/internal/riffmatcher"
	"github.com/richardlehane/siegfried/internal/textmatcher"
	"github.com/richardlehane/siegfried/internal/xmlmatcher"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

// A base identifier that can be embedded in other identifier
type Base struct {
	p                                        Parseable
	name                                     string
	details                                  string
	multi                                    config.Multi
	zipDefault                               bool
	gids, mids, cids, xids, bids, rids, tids *indexes
}

type indexes struct {
	start  int
	ids    []string
	once   sync.Once
	lookup map[string][]int
}

func (ii *indexes) find(ks []string) []int {
	ii.once.Do(func() {
		ii.lookup = make(map[string][]int)
		for i, v := range ii.ids {
			ii.lookup[v] = append(ii.lookup[v], ii.start+i)
		}
	})
	ret := make([]int, 0, len(ks)*2)
	for _, k := range ks {
		ret = append(ret, ii.lookup[k]...)
	}
	return ret
}

func (ii *indexes) hit(i int) (bool, string) {
	if i >= ii.start && i < ii.start+len(ii.ids) {
		return true, ii.ids[i-ii.start]
	}
	return false, ""
}

func (ii *indexes) first(i int) (bool, string) {
	if i == ii.start && len(ii.ids) > 0 {
		return true, ii.ids[0]
	}
	return false, ""
}

func (ii *indexes) save(ls *persist.LoadSaver) {
	ls.SaveInt(ii.start)
	ls.SaveStrings(ii.ids)
}

func (ii *indexes) place(i int) (int, int) {
	if i >= ii.start && i < ii.start+len(ii.ids) {
		idx, id := i-ii.start, ii.ids[i-ii.start]
		var prev, post int
		for j := idx - 1; j > -1 && ii.ids[j] == id; j-- {
			prev++
		}
		for j := idx + 1; j < len(ii.ids) && ii.ids[j] == id; j++ {
			post++
		}
		return prev + 1, prev + post + 1
	}
	return -1, -1
}

func loadIndexes(ls *persist.LoadSaver) *indexes {
	return &indexes{
		start: ls.LoadInt(),
		ids:   ls.LoadStrings(),
	}
}

func New(p Parseable, zip string, extra ...string) *Base {
	return &Base{
		p:          p,
		name:       config.Name(),
		details:    config.Details(extra...),
		multi:      config.GetMulti(),
		zipDefault: contains(p.IDs(), zip),
		gids:       &indexes{}, mids: &indexes{}, cids: &indexes{}, xids: &indexes{}, bids: &indexes{}, rids: &indexes{}, tids: &indexes{},
	}
}

func (b *Base) Save(ls *persist.LoadSaver) {
	ls.SaveString(b.name)
	ls.SaveString(b.details)
	ls.SaveTinyInt(int(b.multi))
	ls.SaveBool(b.zipDefault)
	b.gids.save(ls)
	b.mids.save(ls)
	b.cids.save(ls)
	b.xids.save(ls)
	b.bids.save(ls)
	b.rids.save(ls)
	b.tids.save(ls)
}

func Load(ls *persist.LoadSaver) *Base {
	return &Base{
		name:       ls.LoadString(),
		details:    ls.LoadString(),
		multi:      config.Multi(ls.LoadTinyInt()),
		zipDefault: ls.LoadBool(),
		gids:       loadIndexes(ls),
		mids:       loadIndexes(ls),
		cids:       loadIndexes(ls),
		xids:       loadIndexes(ls),
		bids:       loadIndexes(ls),
		rids:       loadIndexes(ls),
		tids:       loadIndexes(ls),
	}
}

func (b *Base) Name() string {
	return b.name
}

func (b *Base) Details() string {
	return b.details
}

func (b *Base) String() string {
	str := fmt.Sprintf("Name: %s\nDetails: %s\n", b.name, b.details)
	str += fmt.Sprintf("Number of filename signatures: %d \n", len(b.gids.ids))
	str += fmt.Sprintf("Number of MIME signatures: %d \n", len(b.mids.ids))
	str += fmt.Sprintf("Number of container signatures: %d \n", len(b.cids.ids))
	str += fmt.Sprintf("Number of XML signatures: %d \n", len(b.xids.ids))
	str += fmt.Sprintf("Number of byte signatures: %d \n", len(b.bids.ids))
	str += fmt.Sprintf("Number of RIFF signatures: %d \n", len(b.rids.ids))
	str += fmt.Sprintf("Number of text signatures: %d \n", len(b.tids.ids))
	return str
}

func (b *Base) Inspect(ids ...string) (string, error) {
	return inspect(b.p, ids...)
}

func graphP(p priority.Map, infos map[string]FormatInfo) string {
	elements := p.Elements()
	lines := make([]string, len(elements))
	for i, v := range elements {
		if v[1] == "" {
			lines[i] = fmt.Sprintf("\"%s (%s)\"", infos[v[0]].String(), v[0])
			continue
		}
		lines[i] = fmt.Sprintf("\"%s (%s)\" -> \"%s (%s)\"", infos[v[0]].String(), v[0], infos[v[1]].String(), v[1])
	}
	return "digraph {\n  " + strings.Join(lines, "\n  ") + "\n}"
}

const (
	Priorities int = iota
	Missing
	Implicit
)

func (b *Base) GraphP(i int) string {
	p := b.p.Priorities()
	if p == nil && i < Implicit {
		return "no priorities set"
	}
	switch i {
	case Missing:
		p = implicit(b.p.Signatures()).Difference(p)
	case Implicit:
		p = implicit(b.p.Signatures())
	}
	return graphP(p, b.p.Infos())
}

func implicit(sigs []frames.Signature, ids []string, e error) priority.Map {
	pm := make(priority.Map)
	if e != nil {
		return pm
	}
	for i, v := range sigs {
		for j, w := range sigs {
			if i == j || ids[i] == ids[j] {
				continue
			}
			if w.Contains(v) {
				pm.Add(ids[i], ids[j])
			}
		}
	}
	return pm
}

func (b *Base) NoPriority() bool {
	return b.multi >= config.Comprehensive
}

func (b *Base) Multi() config.Multi {
	return b.multi
}

func (b *Base) ZipDefault() bool {
	return b.zipDefault
}

func (b *Base) Hit(m core.MatcherType, idx int) (bool, string) {
	switch m {
	default:
		return false, ""
	case core.NameMatcher:
		return b.gids.hit(idx)
	case core.MIMEMatcher:
		return b.mids.hit(idx)
	case core.ContainerMatcher:
		return b.cids.hit(idx)
	case core.XMLMatcher:
		return b.xids.hit(idx)
	case core.ByteMatcher:
		return b.bids.hit(idx)
	case core.RIFFMatcher:
		return b.rids.hit(idx)
	case core.TextMatcher:
		return b.tids.first(idx) // textmatcher is unique as only returns a single hit per identifier
	}
}

func (b *Base) Place(m core.MatcherType, idx int) (int, int) {
	switch m {
	default:
		return -1, -1
	case core.NameMatcher:
		return b.gids.place(idx)
	case core.MIMEMatcher:
		return b.mids.place(idx)
	case core.ContainerMatcher:
		return b.cids.place(idx)
	case core.XMLMatcher:
		return b.xids.place(idx)
	case core.ByteMatcher:
		return b.bids.place(idx)
	case core.RIFFMatcher:
		return b.rids.place(idx)
	case core.TextMatcher:
		return b.tids.place(idx)
	}
}

func (b *Base) Lookup(m core.MatcherType, keys []string) []int {
	switch m {
	default:
		return nil
	case core.NameMatcher:
		return b.gids.find(keys)
	case core.MIMEMatcher:
		return b.mids.find(keys)
	case core.ContainerMatcher:
		return b.cids.find(keys)
	case core.XMLMatcher:
		return b.xids.find(keys)
	case core.ByteMatcher:
		return b.bids.find(keys)
	case core.RIFFMatcher:
		return b.rids.find(keys)
	case core.TextMatcher:
		return b.tids.find(keys)
	}
}

func (b *Base) Recognise(m core.MatcherType, idx int) (bool, string) {
	h, id := b.Hit(m, idx)
	if h {
		return true, b.name + ": " + id
	}
	return false, ""
}

func (b *Base) Add(m core.Matcher, t core.MatcherType) (core.Matcher, error) {
	var l int
	var err error
	switch t {
	default:
		return nil, fmt.Errorf("identifier: unknown matcher type %d", t)
	case core.NameMatcher:
		var globs []string
		globs, b.gids.ids = b.p.Globs()
		m, l, err = namematcher.Add(m, namematcher.SignatureSet(globs), nil)
		if err != nil {
			return nil, err
		}
		b.gids.start = l - len(b.gids.ids)
	case core.ContainerMatcher:
		znames, zsigs, zids, err := b.p.Zips()
		if err != nil {
			return nil, err
		}
		m, _, err = containermatcher.Add(
			m,
			containermatcher.SignatureSet{
				Typ:       containermatcher.Zip,
				NameParts: znames,
				SigParts:  zsigs,
			},
			b.p.Priorities().List(zids),
		)
		if err != nil {
			return nil, err
		}
		mnames, msigs, mids, err := b.p.MSCFBs()
		if err != nil {
			return nil, err
		}
		m, l, err = containermatcher.Add(
			m,
			containermatcher.SignatureSet{
				Typ:       containermatcher.Mscfb,
				NameParts: mnames,
				SigParts:  msigs,
			},
			b.p.Priorities().List(mids),
		)
		if err != nil {
			return nil, err
		}
		b.cids.ids = append(zids, mids...)
		b.cids.start = l - len(b.cids.ids)
	case core.MIMEMatcher:
		var mimes []string
		mimes, b.mids.ids = b.p.MIMEs()
		m, l, err = mimematcher.Add(m, mimematcher.SignatureSet(mimes), nil)
		if err != nil {
			return nil, err
		}
		b.mids.start = l - len(b.mids.ids)
	case core.XMLMatcher:
		var xmls [][2]string
		xmls, b.xids.ids = b.p.XMLs()
		m, l, err = xmlmatcher.Add(m, xmlmatcher.SignatureSet(xmls), nil)
		if err != nil {
			return nil, err
		}
		b.xids.start = l - len(b.xids.ids)
	case core.ByteMatcher:
		var sigs []frames.Signature
		var err error
		sigs, b.bids.ids, err = b.p.Signatures()
		if err != nil {
			return nil, err
		}
		m, l, err = bytematcher.Add(m, bytematcher.SignatureSet(sigs), b.p.Priorities().List(b.bids.ids))
		if err != nil {
			return nil, err
		}
		b.bids.start = l - len(b.bids.ids)
	case core.RIFFMatcher:
		var riffs [][4]byte
		riffs, b.rids.ids = b.p.RIFFs()
		m, l, err = riffmatcher.Add(m, riffmatcher.SignatureSet(riffs), b.p.Priorities().List(b.rids.ids))
		if err != nil {
			return nil, err
		}
		b.rids.start = l - len(b.rids.ids)
	case core.TextMatcher:
		b.tids.ids = b.p.Texts()
		if len(b.tids.ids) > 0 {
			m, l, _ = textmatcher.Add(m, textmatcher.SignatureSet{}, nil)
			b.tids.start = l
		}
	}
	return m, nil
}

func (b *Base) Active(m core.MatcherType) bool {
	switch m {
	default:
		return false
	case core.NameMatcher:
		return len(b.gids.ids) > 0
	case core.MIMEMatcher:
		return len(b.mids.ids) > 0
	case core.ContainerMatcher:
		return len(b.cids.ids) > 0
	case core.XMLMatcher:
		return len(b.xids.ids) > 0
	case core.ByteMatcher:
		return len(b.bids.ids) > 0
	case core.RIFFMatcher:
		return len(b.rids.ids) > 0
	case core.TextMatcher:
		return len(b.tids.ids) > 0
	}
}

func (b *Base) Start(m core.MatcherType) int {
	switch m {
	default:
		return 0
	case core.NameMatcher:
		return b.gids.start
	case core.MIMEMatcher:
		return b.mids.start
	case core.ContainerMatcher:
		return b.cids.start
	case core.XMLMatcher:
		return b.xids.start
	case core.ByteMatcher:
		return b.bids.start
	case core.RIFFMatcher:
		return b.rids.start
	case core.TextMatcher:
		return b.tids.start
	}
}

func (b *Base) IDs(m core.MatcherType) []string {
	switch m {
	default:
		return nil
	case core.NameMatcher:
		return b.gids.ids
	case core.MIMEMatcher:
		return b.mids.ids
	case core.ContainerMatcher:
		return b.cids.ids
	case core.XMLMatcher:
		return b.xids.ids
	case core.ByteMatcher:
		return b.bids.ids
	case core.RIFFMatcher:
		return b.rids.ids
	case core.TextMatcher:
		return b.tids.ids
	}
}

func (b *Base) HasSig(id string, ms ...core.MatcherType) bool {
	for _, m := range ms {
		for _, i := range b.IDs(m) {
			if id == i {
				return true
			}
		}
	}
	return false
}

func contains(strs []string, s string) bool {
	for _, v := range strs {
		if s == v {
			return true
		}
	}
	return false
}
