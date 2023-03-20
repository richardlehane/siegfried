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

	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

func init() {
	core.RegisterIdentifier(core.MIMEInfo, Load)
}

type Identifier struct {
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
	// add extensions
	for _, v := range config.Extend() {
		e, err := newMIMEInfo(v)
		if err != nil {
			return nil, fmt.Errorf("MIMEinfo: error loading extension file %s; got %s", v, err)
		}
		mi = identifier.Join(mi, e)
	}
	// apply config
	mi = identifier.ApplyConfig(mi)
	// get version
	// return identifier
	return &Identifier{
		infos: infos(mi.Infos()),
		Base:  identifier.New(mi, config.ZipMIME(), config.MIMEVersion()...),
	}, nil
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
	if r.Identifier.Active(m) {
		switch m {
		case core.NameMatcher:
			r.globActive = true
		case core.MIMEMatcher:
			r.mimeActive = true
		case core.TextMatcher:
			r.textActive = true
		}
	}
}

func (r *Recorder) Record(m core.MatcherType, res core.Result) bool {
	switch m {
	default:
		return false
	case core.NameMatcher:
		if hit, id := r.Hit(m, res.Index()); hit {
			r.ids = add(r.ids, r.Name(), id, r.infos[id], res.Basis(), m, rel(r.Place(core.NameMatcher, res.Index())))
			return true
		} else {
			return false
		}
	case core.MIMEMatcher, core.XMLMatcher:
		if hit, id := r.Hit(m, res.Index()); hit {
			r.ids = add(r.ids, r.Name(), id, r.infos[id], res.Basis(), m, 0)
			return true
		} else {
			return false
		}
	case core.ByteMatcher:
		if hit, id := r.Hit(m, res.Index()); hit {
			if r.satisfied {
				return true
			}
			basis := res.Basis()
			p, t := r.Place(core.ByteMatcher, res.Index())
			if t > 1 {
				basis = basis + fmt.Sprintf(" (signature %d/%d)", p, t)
			}
			r.ids = add(r.ids, r.Name(), id, r.infos[id], basis, m, p-1)
			return true
		} else {
			return false
		}
	case core.TextMatcher:
		if hit, _ := r.Hit(m, res.Index()); hit {
			if r.satisfied {
				return true
			}
			if len(r.IDs(m)) > 0 {
				r.ids = bulkAdd(r.ids, r.Name(), r.IDs(m), r.infos, res.Basis(), core.TextMatcher, 0)
			}
			return true
		} else {
			return false
		}
	}
}

func rel(prev, post int) int {
	return prev - 1
}

func (r *Recorder) Satisfied(mt core.MatcherType) (bool, core.Hint) {
	if r.NoPriority() {
		return false, core.Hint{}
	}
	sort.Sort(r.ids)
	if len(r.ids) > 0 && (r.ids[0].xmlMatch || (r.ids[0].magicScore > 0 && r.ids[0].ID != config.TextMIME())) {
		if mt == core.ByteMatcher {
			return true, core.Hint{Exclude: r.Start(mt), Pivot: nil}
		}
		return true, core.Hint{}
	}
	return false, core.Hint{}
}

func (r *Recorder) Report() []core.Identification {
	// no results
	if len(r.ids) == 0 {
		return []core.Identification{Identification{
			Namespace: r.Name(),
			ID:        "UNKNOWN",
			Warning:   "no match",
		}}
	}
	sort.Sort(r.ids)
	// exhaustive
	if r.Multi() == config.Exhaustive {
		ret := make([]core.Identification, len(r.ids))
		for i, v := range r.ids {
			ret[i] = r.updateWarning(v)
		}
		return ret
	}
	// if we've only got weak matches (match is filename/mime only) report only the first
	if !r.ids[0].xmlMatch && r.ids[0].magicScore == 0 {
		var nids []Identification
		if len(r.ids) == 1 || r.ids.Less(0, 1) { // // Less reports whether the element with index i (0) should sort before the element with index j
			if r.ids[0].ID != config.TextMIME() || r.ids[0].textMatch || !r.textActive {
				nids = []Identification{r.ids[0]}
			}
		}
		var conf string
		if len(nids) != 1 {
			lowConfidence := confidenceTrick()
			poss := make([]string, len(r.ids))
			for i, v := range r.ids {
				poss[i] = v.ID
				conf = lowConfidence(v)
			}
			return []core.Identification{Identification{
				Namespace: r.Name(),
				ID:        "UNKNOWN",
				Warning:   fmt.Sprintf("no match; possibilities based on %s are %v", conf, strings.Join(poss, ", ")),
			}}
		}
		r.ids = nids
	}
	// handle single result only
	if r.Multi() == config.Single && len(r.ids) > 1 && !r.ids.Less(0, 1) {
		poss := make([]string, 0, len(r.ids))
		for i, v := range r.ids {
			if i > 0 && r.ids.Less(i-1, i) {
				break
			}
			poss = append(poss, v.ID)
		}
		return []core.Identification{Identification{
			Namespace: r.Name(),
			ID:        "UNKNOWN",
			Warning:   fmt.Sprintf("multiple matches %v", strings.Join(poss, ", ")),
		}}
	}
	ret := make([]core.Identification, len(r.ids))
	for i, v := range r.ids {
		if i > 0 {
			switch r.Multi() {
			case config.Single:
				return ret[:i]
			case config.Conclusive:
				if r.ids.Less(i-1, i) {
					return ret[:i]
				}
			default:
				if !v.xmlMatch && v.magicScore == 0 { // if weak
					return ret[:i]
				}
			}
		}
		ret[i] = r.updateWarning(v)
	}
	return ret
}

func (r *Recorder) updateWarning(i Identification) Identification {
	// weak match
	if !i.xmlMatch && i.magicScore == 0 {
		lowConfidence := confidenceTrick()
		if len(i.Warning) > 0 {
			i.Warning += "; " + "match on " + lowConfidence(i) + " only"
		} else {
			i.Warning = "match on " + lowConfidence(i) + " only"
		}
		// if the match has no corresponding byte or xml signature...
		if r.HasSig(i.ID, core.XMLMatcher, core.ByteMatcher) {
			i.Warning += "; byte/xml signatures for this format did not match"
		}
	}
	// apply mismatches
	if r.globActive && i.globScore == 0 {
		for _, v := range r.IDs(core.NameMatcher) {
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

func (id Identification) Values() []string {
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
