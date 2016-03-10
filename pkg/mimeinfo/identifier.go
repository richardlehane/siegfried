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
	p          parseable.Parseable
	name       string
	details    string
	noPriority bool
	zipDefault bool
	infos      map[string]formatInfo
	gstart     int
	gids       []string
	mstart     int
	mids       []string
	xstart     int
	xids       []string
	bstart     int
	bids       []string
	tstart     int
}

func (i *Identifier) Save(ls *persist.LoadSaver) {
	ls.SaveByte(core.MIMEInfo)
	ls.SaveString(i.name)
	ls.SaveString(i.details)
	ls.SaveBool(i.noPriority)
	ls.SaveBool(i.zipDefault)
	ls.SaveSmallInt(len(i.infos))
	for k, v := range i.infos {
		ls.SaveString(k)
		ls.SaveString(v.comment)
		ls.SaveInts(v.globWeights)
		ls.SaveInts(v.magicWeights)
	}
	ls.SaveInt(i.gstart)
	ls.SaveStrings(i.gids)
	ls.SaveInt(i.mstart)
	ls.SaveStrings(i.mids)
	ls.SaveInt(i.xstart)
	ls.SaveStrings(i.xids)
	ls.SaveInt(i.bstart)
	ls.SaveStrings(i.bids)
	ls.SaveSmallInt(i.tstart)
}

func Load(ls *persist.LoadSaver) core.Identifier {
	i := &Identifier{}
	i.name = ls.LoadString()
	i.details = ls.LoadString()
	i.noPriority = ls.LoadBool()
	i.zipDefault = ls.LoadBool()
	i.infos = make(map[string]formatInfo)
	le := ls.LoadSmallInt()
	for j := 0; j < le; j++ {
		i.infos[ls.LoadString()] = formatInfo{
			ls.LoadString(),
			ls.LoadInts(),
			ls.LoadInts(),
		}
	}
	i.gstart = ls.LoadInt()
	i.gids = ls.LoadStrings()
	i.mstart = ls.LoadInt()
	i.mids = ls.LoadStrings()
	i.xstart = ls.LoadInt()
	i.xids = ls.LoadStrings()
	i.bstart = ls.LoadInt()
	i.bids = ls.LoadStrings()
	i.tstart = ls.LoadSmallInt()
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
	mi, err := newMIMEInfo()
	if err != nil {
		return nil, err
	}
	id := &Identifier{
		p:       mi,
		name:    config.Name(),
		details: config.Details(),
		infos:   infos(mi.Infos()),
	}
	if contains(mi.IDs(), config.ZipMIME()) {
		id.zipDefault = true
	}
	return id, nil
}

func (i *Identifier) Add(m core.Matcher, t core.MatcherType) (core.Matcher, error) {
	var l int
	var err error
	switch t {
	default:
		return nil, fmt.Errorf("MIMEInfo: unknown matcher type %d", t)
	case core.NameMatcher:
		if !config.NoName() {
			var globs []string
			globs, i.gids = i.p.Globs()
			m, l, err = namematcher.Add(m, namematcher.SignatureSet(globs), nil)
			if err != nil {
				return nil, err
			}
			i.gstart = l - len(i.gids)
		}
	case core.MIMEMatcher:
		if !config.NoMIME() {
			var mimes []string
			mimes, i.mids = i.p.MIMEs()
			m, l, err = mimematcher.Add(m, mimematcher.SignatureSet(mimes), nil)
			if err != nil {
				return nil, err
			}
			i.mstart = l - len(i.mids)
		}
	case core.XMLMatcher:
		if !config.NoXML() {
			var xmls [][2]string
			xmls, i.xids = i.p.XMLs()
			m, l, err = xmlmatcher.Add(m, xmlmatcher.SignatureSet(xmls), nil)
			if err != nil {
				return nil, err
			}
			i.xstart = l - len(i.xids)
		}
	case core.ByteMatcher:
		var sigs []frames.Signature
		var err error
		sigs, i.bids, err = i.p.Signatures()
		if err != nil {
			return nil, err
		}
		m, l, err = bytematcher.Add(m, bytematcher.SignatureSet(sigs), nil)
		if err != nil {
			return nil, err
		}
		i.bstart = l - len(i.bids)
	case core.TextMatcher:
		if !config.NoText() && contains(i.p.IDs(), config.TextMIME()) {
			m, l, _ = textmatcher.Add(m, textmatcher.SignatureSet{}, nil)
			i.tstart = l
		}
	case core.ContainerMatcher:
	}
	return m, nil
}

func (i *Identifier) Name() string {
	return i.name
}

func (i *Identifier) Details() string {
	return i.details
}

func (i *Identifier) Fields() []string {
	return []string{"namespace", "id", "format", "mime", "basis", "warning"}
}

func (i *Identifier) String() string {
	str := fmt.Sprintf("Name: %s\nDetails: %s\n", i.name, i.details)
	str += fmt.Sprintf("Number of file infos: %d \n", len(i.infos))
	str += fmt.Sprintf("Number of filename signatures: %d \n", len(i.gids))
	str += fmt.Sprintf("Number of MIME signatures: %d \n", len(i.mids))
	str += fmt.Sprintf("Number of XML signatures: %d \n", len(i.xids))
	str += fmt.Sprintf("Number of byte signatures: %d \n", len(i.bids))
	return str
}

func (i *Identifier) Recognise(m core.MatcherType, idx int) (bool, string) {
	switch m {
	default:
		return false, ""
	case core.NameMatcher:
		if idx >= i.gstart && idx < i.gstart+len(i.gids) {
			idx = idx - i.gstart
			return true, i.name + ": " + i.gids[idx]
		}
		return false, ""
	case core.MIMEMatcher:
		if idx >= i.mstart && idx < i.mstart+len(i.mids) {
			idx = idx - i.mstart
			return true, i.name + ": " + i.mids[idx]
		}
		return false, ""
	case core.XMLMatcher:
		if idx >= i.xstart && idx < i.xstart+len(i.xids) {
			idx = idx - i.xstart
			return true, i.name + ": " + i.xids[idx]
		}
		return false, ""
	case core.ContainerMatcher:
		return false, ""
	case core.ByteMatcher:
		if idx >= i.bstart && idx < i.bstart+len(i.bids) {
			return true, i.name + ": " + i.bids[idx]
		}
		return false, ""
	case core.TextMatcher:
		if idx == i.tstart {
			return true, i.name + ": " + config.TextPuid()
		}
		return false, ""
	}
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
			r.ids = add(r.ids, r.name, config.TextMIME(), r.infos[config.TextMIME()], res.Basis(), core.TextMatcher, 0)
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
	if len(r.ids) > 0 && (r.ids[0].xmlMatch || r.ids[0].magicScore > 0) {
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
	if !r.ids[0].xmlMatch && r.ids[0].magicScore == 0 {
		nids := make([]Identification, 0, 1)
		for idx, v := range r.ids {
			// if we have plain text result that is based on ext or mime only,
			// and not on a text match, and if text matcher is on for this identifier,
			// then don't report a text match
			if v.ID == config.TextMIME() && !v.textMatch && r.textActive {
				continue
			}
			// rule out any lesser confident matches
			if idx < len(r.ids)-1 && r.ids.Less(idx, idx+1) {
				break
			}
			// if the match has no corresponding byte or xml signature...
			if ok := r.hasSig(v.ID); !ok {
				// break immediately if more than one match
				if len(nids) > 0 {
					nids = nids[:0]
					break
				}
				if len(v.Warning) > 0 {
					v.Warning += "; " + "match on " + lowConfidence("", v) + " only"
				} else {
					v.Warning = "match on " + lowConfidence("", v) + " only"
				}
				nids = append(nids, v)
			}
		}
		var conf string
		if len(nids) != 1 {
			poss := make([]string, len(r.ids))
			for i, v := range r.ids {
				poss[i] = v.ID
				conf = lowConfidence(conf, v)
			}
			nids = []Identification{Identification{
				Namespace: r.name,
				ID:        "UNKNOWN",
				Warning:   fmt.Sprintf("no match; possibilities based on %v are %v", conf, strings.Join(poss, ", ")),
			},
			}
		}
		r.ids = nids
	}
	res <- r.checkActive(r.ids[0])
	if len(r.ids) > 1 {
		for i, _ := range r.ids[1:] {
			if !r.ids.Less(i, i+1) { // || (r.noPriority && v.confidence >= incScore) - noPriority semantics for MI??
				res <- r.checkActive(r.ids[i+1])
			} else {
				break
			}
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

func lowConfidence(str string, i Identification) string {
	if i.globScore > 0 && !strings.Contains(str, "filename") {
		if str == "" {
			str = "filename"
		} else {
			str += "; filename"
		}
	}
	if i.mimeMatch && !strings.Contains(str, "MIME") {
		if str == "" {
			str = "MIME"
		} else {
			str += "; MIME"
		}
	}
	if i.textMatch && !strings.Contains(str, "text") {
		if str == "" {
			str = "text"
		} else {
			str += "; text"
		}
	}
	return str
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

	xmlMatch   bool
	magicScore int
	globScore  int
	mimeMatch  bool
	textMatch  bool
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

func tieBreak(m1, m2 bool, gs1, gs2 int) bool {
	switch {
	case m1 && !m2:
		return true
	case m2 && !m1:
		return false
	}
	return gs2 < gs1
}

func multisignal(m bool, ms, gs int) bool {
	switch {
	case m && ms > 0:
		return true
	case ms > 0 && gs > 0:
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
		return tieBreak(m[i].mimeMatch, m[j].mimeMatch, m[i].globScore, m[j].globScore)
	}
	msi, msj := multisignal(m[i].mimeMatch, m[i].magicScore, m[i].globScore), multisignal(m[j].mimeMatch, m[j].magicScore, m[j].globScore)
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
	return tieBreak(m[i].mimeMatch, m[j].mimeMatch, m[i].globScore, m[j].globScore)
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
	}
	return id
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
