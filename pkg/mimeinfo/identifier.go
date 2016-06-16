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

package mimeinfo

import (
	"fmt"
	"sort"
	"strings"

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/identifier"
	"github.com/richardlehane/siegfried/pkg/core/mimematcher"
	"github.com/richardlehane/siegfried/pkg/core/namematcher"
	"github.com/richardlehane/siegfried/pkg/core/parseable"
	"github.com/richardlehane/siegfried/pkg/core/persist"
	"github.com/richardlehane/siegfried/pkg/core/textmatcher"
	"github.com/richardlehane/siegfried/pkg/core/xmlmatcher"
)

func init() {
	core.RegisterIdentifier(core.MIMEInfo, Load)
}

type Identifier struct {
	p     parseable.Parseable
	infos map[string]formatInfo
	*identifier.Base
}

func (i *Identifier) Save(ls *persist.LoadSaver) {
	ls.SaveByte(core.MIMEInfo)
	ls.SaveSmallInt(len(i.infos))
	for k, v := range i.infos {
		ls.SaveString(k)
		ls.SaveString(v.comment)
		ls.SaveBool(v.text)
		ls.SaveInts(v.globWeights)
		ls.SaveInts(v.magicWeights)
	}
	i.Base.Save(ls)
}

func Load(ls *persist.LoadSaver) core.Identifier {
	i := &Identifier{}
	i.infos = make(map[string]formatInfo)
	le := ls.LoadSmallInt()
	for j := 0; j < le; j++ {
		i.infos[ls.LoadString()] = formatInfo{
			ls.LoadString(),
			ls.LoadBool(),
			ls.LoadInts(),
			ls.LoadInts(),
		}
	}
	i.Base = identifier.Load(ls)
	return i
}

func contains(ss []string, str string) bool {
	for _, s := range ss {
		if s == str {
			return true
		}
	}
	return false
}

func New(opts ...config.Option) (core.Identifier, error) {
	for _, v := range opts {
		v()
	}
	mi, err := newMIMEInfo(config.MIMEInfo())
	if err != nil {
		return nil, err
	}
	// if we are inspecting...
	if config.Inspect() {
		mi = parseable.Filter(config.Limit(mi.IDs()), mi)
		is := infos(mi.Infos())
		sigs, ids, err := mi.Signatures()
		if err != nil {
			return nil, fmt.Errorf("MIMEinfo: parsing signatures; got %s", err)
		}
		var id string
		for i, sig := range sigs {
			if ids[i] != id {
				id = ids[i]
				fmt.Printf("%s: \n", is[id].comment)
			}
			fmt.Println(sig)
		}
		return nil, nil
	}
	// apply limit or exclude
	if config.HasLimit() || config.HasExclude() {
		var ids []string
		if config.HasLimit() {
			ids = config.Limit(mi.IDs())
		} else if config.HasExclude() {
			ids = config.Exclude(mi.IDs())
		}
		mi = parseable.Filter(ids, mi)
	}
	// add extensions
	for _, v := range config.Extend() {
		e, err := newMIMEInfo(v)
		if err != nil {
			return nil, fmt.Errorf("MIMEinfo: error loading extension file %s; got %s", v, err)
		}
		mi = parseable.Join(mi, e)
	}
	id := &Identifier{
		p:     mi,
		infos: infos(mi.Infos()),
		Base:  identifier.New(contains(mi.IDs(), config.ZipMIME())),
	}
	return id, nil
}

func (i *Identifier) Fields() []string {
	return []string{"namespace", "id", "format", "mime", "basis", "warning"}
}

func (i *Identifier) Recorder() core.Recorder {
	return &Recorder{
		Identifier: i,
		ids:        make(ids, 0, 1),
	}
}

type Recorder struct {
	*Identifier
	ids        ids
	satisfied  bool
	globActive bool
	mimeActive bool
	textActive bool
}

func (r *Recorder) Active(m core.MatcherType) {
	switch m {
	case core.NameMatcher:
		if len(r.gids) > 0 {
			r.globActive = true
		}
	case core.MIMEMatcher:
		if len(r.mids) > 0 {
			r.mimeActive = true
		}
	case core.TextMatcher:
		if r.tstart > 0 {
			r.textActive = true
		}
	}
}

func (r *Recorder) Record(m core.MatcherType, res core.Result) bool {
	switch m {
	default:
		return false
	case core.NameMatcher:
		if res.Index() >= r.gstart && res.Index() < r.gstart+len(r.gids) {
			idx := res.Index() - r.gstart
			r.ids = add(r.ids, r.name, r.gids[idx], r.infos[r.gids[idx]], res.Basis(), core.NameMatcher, rel(idx, r.gids))
			return true
		} else {
			return false
		}
	case core.MIMEMatcher:
		if res.Index() >= r.mstart && res.Index() < r.mstart+len(r.mids) {
			idx := res.Index() - r.mstart
			r.ids = add(r.ids, r.name, r.mids[idx], r.infos[r.mids[idx]], res.Basis(), core.MIMEMatcher, 0)
			return true
		} else {
			return false
		}
	case core.XMLMatcher:
		if res.Index() >= r.xstart && res.Index() < r.xstart+len(r.xids) {
			idx := res.Index() - r.xstart
			r.ids = add(r.ids, r.name, r.xids[idx], r.infos[r.xids[idx]], res.Basis(), core.XMLMatcher, 0)
			return true
		} else {
			return false
		}
	case core.ByteMatcher:
		if res.Index() >= r.bstart && res.Index() < r.bstart+len(r.bids) {
			if r.satisfied {
				return true
			}
			idx := res.Index() - r.bstart
			basis := res.Basis()
			p, t := place(idx, r.bids)
			if t > 1 {
				basis = basis + fmt.Sprintf(" (signature %d/%d)", p, t)
			}
			r.ids = add(r.ids, r.name, r.bids[idx], r.infos[r.bids[idx]], basis, core.ByteMatcher, p-1)
			return true
		} else {
			return false
		}
	case core.TextMatcher:
		if res.Index() == r.tstart {
			if r.satisfied {
				return true
			}
			if _, ok := r.infos[config.TextMIME()]; ok {
				r.ids = add(r.ids, r.name, config.TextMIME(), r.infos[config.TextMIME()], res.Basis(), core.TextMatcher, 0)
			}
			if len(r.tids) > 0 {
				r.ids = bulkAdd(r.ids, r.name, r.tids, r.infos, res.Basis(), core.TextMatcher, 0)
			}
			return true
		} else {
			return false
		}
	}
}

func rel(idx int, ids []string) int {
	prev, _ := place(idx, ids)
	return prev - 1
}

func place(idx int, ids []string) (int, int) {
	puid := ids[idx]
	var prev, post int
	for i := idx - 1; i > -1 && ids[i] == puid; i-- {
		prev++
	}
	for i := idx + 1; i < len(ids) && ids[i] == puid; i++ {
		post++
	}
	return prev + 1, prev + post + 1
}

func (r *Recorder) Satisfied(mt core.MatcherType) (bool, int) {
	sort.Sort(r.ids)
	if len(r.ids) > 0 && (r.ids[0].xmlMatch || (r.ids[0].magicScore > 0 && r.ids[0].ID != config.TextMIME())) {
		if mt == core.ByteMatcher {
			return true, r.bstart
		}
		return true, 0
	}
	return false, 0
}

func (r *Recorder) Report(res chan core.Identification) {
	if len(r.ids) == 0 {
		res <- Identification{
			Namespace: r.name,
			ID:        "UNKNOWN",
			Warning:   "no match",
		}
		return
	}
	sort.Sort(r.ids)
	// if match is filename/mime only
	// Less reports whether index i should sort before index j
	if !r.ids[0].xmlMatch && r.ids[0].magicScore == 0 && !r.noPriority {
		lowConfidence := confidenceTrick()
		var nids []Identification
		if len(r.ids) == 1 || r.ids.Less(0, 1) {
			v := r.ids[0]
			if v.ID != config.TextMIME() || v.textMatch || !r.textActive {
				if len(v.Warning) > 0 {
					v.Warning += "; " + "match on " + lowConfidence(v) + " only"
				} else {
					v.Warning = "match on " + lowConfidence(v) + " only"
				}
				// if the match has no corresponding byte or xml signature...
				if r.hasSig(v.ID) {
					v.Warning += "; byte/xml signatures for this format did not match"
				}
				nids = []Identification{v}
			}
		}
		var conf string
		if len(nids) != 1 {
			poss := make([]string, len(r.ids))
			for i, v := range r.ids {
				poss[i] = v.ID
				conf = lowConfidence(v)
			}
			nids = []Identification{Identification{
				Namespace: r.name,
				ID:        "UNKNOWN",
				Warning:   fmt.Sprintf("no match; possibilities based on %s are %v", conf, strings.Join(poss, ", ")),
			},
			}
		}
		r.ids = nids
	}
	res <- r.checkActive(r.ids[0])
	if len(r.ids) > 1 {
		for i, _ := range r.ids[1:] {
			if !r.noPriority && r.ids.Less(i, i+1) {
				break
			}
			res <- r.checkActive(r.ids[i+1])
		}
	}
}

func (r *Recorder) checkActive(i Identification) Identification {
	if r.globActive && i.globScore == 0 {
		for _, v := range r.gids {
			if i.ID == v {
				if len(i.Warning) > 0 {
					i.Warning += "; filename mismatch"
				} else {
					i.Warning = "filename mismatch"
				}
				break
			}
		}
	}
	if r.mimeActive && !i.mimeMatch {
		if len(i.Warning) > 0 {
			i.Warning += "; MIME mismatch"
		} else {
			i.Warning = "MIME mismatch"
		}
	}
	return i
}

func confidenceTrick() func(i Identification) string {
	var ls = make([]string, 0, 1)
	return func(i Identification) string {
		if i.globScore > 0 && !contains(ls, "filename") {
			ls = append(ls, "filename")
		}
		if i.mimeMatch && !contains(ls, "MIME") {
			ls = append(ls, "MIME")
		}
		if i.textMatch && !contains(ls, "text") {
			ls = append(ls, "text")
		}
		switch len(ls) {
		case 0:
			return ""
		case 1:
			return ls[0]
		case 2:
			return ls[0] + " and " + ls[1]
		default:
			return strings.Join(ls[:len(ls)-1], ", ") + " and " + ls[len(ls)-1]
		}
	}
}

func (r *Recorder) hasSig(id string) bool {
	for _, v := range r.xids {
		if id == v {
			return true
		}
	}
	for _, v := range r.bids {
		if id == v {
			return true
		}
	}
	return false
}

type Identification struct {
	Namespace string
	ID        string
	Name      string
	Basis     []string
	Warning   string
	archive   config.Archive

	xmlMatch    bool
	magicScore  int
	globScore   int
	mimeMatch   bool
	textMatch   bool
	textDefault bool
}

func (id Identification) String() string {
	return id.ID
}

func (id Identification) Known() bool {
	return id.ID != "UNKNOWN"
}

func (id Identification) Warn() string {
	return id.Warning
}

func quoteText(s string) string {
	if len(s) == 0 {
		return s
	}
	return "'" + s + "'"
}

func (id Identification) YAML() string {
	var basis string
	if len(id.Basis) > 0 {
		basis = quoteText(strings.Join(id.Basis, "; "))
	}
	return fmt.Sprintf("  - ns      : %v\n    id      : %v\n    format  : %v\n    mime    : %v\n    basis   : %v\n    warning : %v\n",
		id.Namespace, quoteText(id.ID), quoteText(id.Name), quoteText(id.ID), basis, quoteText(id.Warning))
}

func (id Identification) JSON() string {
	var basis string
	if len(id.Basis) > 0 {
		basis = strings.Join(id.Basis, "; ")
	}
	return fmt.Sprintf("{\"ns\":\"%s\",\"id\":\"%s\",\"format\":\"%s\",\"mime\":\"%s\",\"basis\":\"%s\",\"warning\":\"%s\"}",
		id.Namespace, id.ID, id.Name, id.ID, basis, id.Warning)
}

func (id Identification) CSV() []string {
	var basis string
	if len(id.Basis) > 0 {
		basis = strings.Join(id.Basis, "; ")
	}
	return []string{
		id.Namespace,
		id.ID,
		id.Name,
		id.ID,
		basis,
		id.Warning,
	}
}

func (id Identification) Archive() config.Archive {
	return id.archive
}

type ids []Identification

func (m ids) Len() int { return len(m) }

func tieBreak(m1, m2, t1, t2, td1, td2 bool, gs1, gs2 int) bool {
	switch {
	case m1 && !m2:
		return true
	case m2 && !m1:
		return false
	}
	if gs1 == gs2 {
		if t1 && !t2 {
			return true
		}
		if t2 && !t1 {
			return false
		}
		if td1 && !td2 {
			return true
		}
	}
	return gs2 < gs1
}

func multisignal(m, t bool, ms, gs int) bool {
	switch {
	case m && ms > 0:
		return true
	case ms > 0 && gs > 0:
		return true
	case m && t:
		return true
	case t && gs > 0:
		return true
	}
	return false
}

func (m ids) Less(i, j int) bool {
	switch {
	case m[i].xmlMatch && !m[j].xmlMatch:
		return true
	case !m[i].xmlMatch && m[j].xmlMatch:
		return false
	case m[i].xmlMatch && m[j].xmlMatch:
		return tieBreak(m[i].mimeMatch, m[j].mimeMatch, m[i].textMatch, m[j].textMatch, m[i].textDefault, m[j].textDefault, m[i].globScore, m[j].globScore)
	}
	msi, msj := multisignal(m[i].mimeMatch, m[i].textMatch, m[i].magicScore, m[i].globScore), multisignal(m[j].mimeMatch, m[j].textMatch, m[j].magicScore, m[j].globScore)
	switch {
	case msi && !msj:
		return true
	case !msi && msj:
		return false
	}
	switch {
	case m[i].magicScore > m[j].magicScore:
		return true
	case m[i].magicScore < m[j].magicScore:
		return false
	}
	return tieBreak(m[i].mimeMatch, m[j].mimeMatch, m[i].textMatch, m[j].textMatch, m[i].textDefault, m[j].textDefault, m[i].globScore, m[j].globScore)
}

func (m ids) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

func applyScore(id Identification, info formatInfo, t core.MatcherType, rel int) Identification {
	switch t {
	case core.NameMatcher:
		score := info.globWeights[rel]
		if score > id.globScore {
			id.globScore = score
		}
	case core.MIMEMatcher:
		id.mimeMatch = true
	case core.XMLMatcher:
		id.xmlMatch = true
	case core.ByteMatcher:
		score := info.magicWeights[rel]
		if score > id.magicScore {
			id.magicScore = score
		}
	case core.TextMatcher:
		id.textMatch = true
		if id.ID == config.TextMIME() {
			id.textDefault = true
		}
	}
	return id
}

func bulkAdd(m ids, ns string, bids []string, infs map[string]formatInfo, basis string, t core.MatcherType, rel int) ids {
	nids := make(ids, len(m), len(m)+len(bids))
	for _, bid := range bids {
		var has bool
		for i, v := range m {
			if v.ID == bid {
				m[i].Basis = append(m[i].Basis, basis)
				m[i] = applyScore(m[i], infs[bid], t, rel)
				has = true
				break
			}
		}
		if !has {
			md := Identification{
				Namespace: ns,
				ID:        bid,
				Name:      infs[bid].comment,
				Basis:     []string{basis},
				Warning:   "",
				archive:   config.IsArchive(bid),
			}
			nids = append(nids, applyScore(md, infs[bid], t, rel))
		}
	}
	copy(nids, m)
	return nids
}

func add(m ids, ns string, id string, info formatInfo, basis string, t core.MatcherType, rel int) ids {
	for i, v := range m {
		if v.ID == id {
			m[i].Basis = append(m[i].Basis, basis)
			m[i] = applyScore(m[i], info, t, rel)
			return m
		}
	}
	md := Identification{
		Namespace: ns,
		ID:        id,
		Name:      info.comment,
		Basis:     []string{basis},
		Warning:   "",
		archive:   config.IsArchive(id),
	}
	return append(m, applyScore(md, info, t, rel))
}
