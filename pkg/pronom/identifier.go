// Copyright 2014 Richard Lehane. All rights reserved.
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

package pronom

import (
	"fmt"
	"sort"
	"strings"

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/persist"
)

func init() {
	core.RegisterIdentifier(core.Pronom, Load)
}

type Identifier struct {
	p          *pronom
	name       string
	details    string
	noPriority bool // was noPriority set when built?
	infos      map[string]formatInfo
	eStart     int
	ePuids     []string // slice of puids that corresponds to the extension matcher's int slice of signatures
	cStart     int
	cPuids     []string
	bStart     int
	bPuids     []string // slice of puids that corresponds to the bytematcher's int slice of signatures
}

type formatInfo struct {
	name     string
	version  string
	mimeType string
}

func (i *Identifier) Save(ls *persist.LoadSaver) {
	ls.SaveByte(core.Pronom)
	ls.SaveString(i.name)
	ls.SaveString(i.details)
	ls.SaveBool(i.noPriority)
	ls.SaveSmallInt(len(i.infos))
	for k, v := range i.infos {
		ls.SaveString(k)
		ls.SaveString(v.name)
		ls.SaveString(v.version)
		ls.SaveString(v.mimeType)
	}
	ls.SaveInt(i.eStart)
	ls.SaveStrings(i.ePuids)
	ls.SaveInt(i.cStart)
	ls.SaveStrings(i.cPuids)
	ls.SaveInt(i.bStart)
	ls.SaveStrings(i.bPuids)
}

func Load(ls *persist.LoadSaver) core.Identifier {
	i := &Identifier{}
	i.name = ls.LoadString()
	i.details = ls.LoadString()
	i.noPriority = ls.LoadBool()
	i.infos = make(map[string]formatInfo)
	le := ls.LoadSmallInt()
	for j := 0; j < le; j++ {
		i.infos[ls.LoadString()] = formatInfo{
			ls.LoadString(),
			ls.LoadString(),
			ls.LoadString(),
		}
	}
	i.eStart = ls.LoadInt()
	i.ePuids = ls.LoadStrings()
	i.cStart = ls.LoadInt()
	i.cPuids = ls.LoadStrings()
	i.bStart = ls.LoadInt()
	i.bPuids = ls.LoadStrings()
	return i
}

func New(opts ...config.Option) (*Identifier, error) {
	for _, v := range opts {
		v()
	}
	pronom, err := newPronom()
	if err != nil {
		return nil, err
	}
	return pronom.identifier(), nil
}

func (i *Identifier) Add(m core.Matcher) error {
	return i.p.add(m)
}

func (i *Identifier) Describe() [2]string {
	return [2]string{i.name, i.details}
}

func (i *Identifier) String() string {
	str := fmt.Sprintf("Name: %s\nDetails: %s\n", i.name, i.details)
	str += fmt.Sprintf("Number of extension persists: %d \n", len(i.ePuids))
	str += fmt.Sprintf("Number of container persists: %d \n", len(i.cPuids))
	str += fmt.Sprintf("Number of byte persists: %d \n", len(i.bPuids))
	return str
}

func (i *Identifier) Recognise(m core.MatcherType, idx int) (bool, string) {
	switch m {
	default:
		return false, ""
	case core.ExtensionMatcher:
		if idx >= i.eStart && idx < i.eStart+len(i.ePuids) {
			idx = idx - i.eStart
			return true, i.name + ": " + i.ePuids[idx]
		} else {
			return false, ""
		}
	case core.ContainerMatcher:
		if idx >= i.cStart && idx < i.cStart+len(i.cPuids) {
			idx = idx - i.cStart
			return true, i.name + ": " + i.cPuids[idx]
		} else {
			return false, ""
		}
	case core.ByteMatcher:
		if idx >= i.bStart && idx < i.bStart+len(i.bPuids) {
			return true, i.name + ": " + i.bPuids[idx]
		} else {
			return false, ""
		}
	}
}

func (i *Identifier) Recorder() core.Recorder {
	return &Recorder{i, make(pids, 0, 10), 1}
}

type Recorder struct {
	*Identifier
	ids    pids
	cscore int
}

func (r *Recorder) Record(m core.MatcherType, res core.Result) bool {
	switch m {
	default:
		return false
	case core.ExtensionMatcher:
		if res.Index() >= r.eStart && res.Index() < r.eStart+len(r.ePuids) {
			idx := res.Index() - r.eStart
			r.ids = add(r.ids, r.name, r.ePuids[idx], r.infos[r.ePuids[idx]], res.Basis(), r.cscore)
			return true
		} else {
			return false
		}
	case core.ContainerMatcher:
		// add zip default
		if res.Index() < 0 {
			if !r.noPriority {
				r.cscore *= 2
			}
			r.ids = add(r.ids, r.name, "x-fmt/263", r.infos["x-fmt/263"], res.Basis(), r.cscore) // not great to have this hardcoded
			return false
		}
		if res.Index() >= r.cStart && res.Index() < r.cStart+len(r.cPuids) {
			idx := res.Index() - r.cStart
			if !r.noPriority {
				r.cscore *= 2
			}
			basis := res.Basis()
			p, t := place(idx, r.cPuids)
			if t > 1 {
				basis = basis + fmt.Sprintf(" (signature %d/%d)", p, t)
			}
			r.ids = add(r.ids, r.name, r.cPuids[idx], r.infos[r.cPuids[idx]], basis, r.cscore)
			return true
		} else {
			return false
		}
	case core.ByteMatcher:
		if res.Index() >= r.bStart && res.Index() < r.bStart+len(r.bPuids) {
			idx := res.Index() - r.bStart
			if !r.noPriority {
				r.cscore *= 2
			}
			basis := res.Basis()
			p, t := place(idx, r.bPuids)
			if t > 1 {
				basis = basis + fmt.Sprintf(" (signature %d/%d)", p, t)
			}
			r.ids = add(r.ids, r.name, r.bPuids[idx], r.infos[r.bPuids[idx]], basis, r.cscore)
			return true
		} else {
			return false
		}
	}
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

func (r *Recorder) Satisfied() bool {
	if r.cscore == 1 {
		return false
	}
	return true
}

func (r *Recorder) Report(res chan core.Identification) {
	if len(r.ids) > 0 {
		sort.Sort(r.ids)
		conf := r.ids[0].confidence
		// if we've only got extension matches, check if those matches are ruled out by lack of byte match
		// add warnings too
		if r.noPriority {
			for i := range r.ids {
				r.ids[i].Warning = "no priority set for this identifier"
			}
		} else if conf == 1 {
			nids := make([]Identification, 0, len(r.ids))
			for _, v := range r.ids {
				if ok := r.hasSig(v.Puid); !ok {
					v.Warning = "match on extension only"
					nids = append(nids, v)
				}
			}
			if len(nids) == 0 {
				poss := make([]string, len(r.ids))
				for i, v := range r.ids {
					poss[i] = v.Puid
				}
				nids = []Identification{Identification{r.name, "UNKNOWN", "", "", "", nil, fmt.Sprintf("no match; possibilities based on extension are %v", strings.Join(poss, ", ")), 0, 0}}
			}
			r.ids = nids
		}
		res <- r.checkExt(r.ids[0])
		if len(r.ids) > 1 {
			for i, v := range r.ids[1:] {
				if v.confidence == conf {
					res <- r.checkExt(r.ids[i+1])
				} else {
					break
				}
			}
		}
	} else {
		res <- Identification{r.name, "UNKNOWN", "", "", "", nil, "no match", 0, 0}
	}
}

func (r *Recorder) checkExt(i Identification) Identification {
	if i.confidence%2 == 0 {
		for _, v := range r.ePuids {
			if i.Puid == v {
				if len(i.Warning) > 0 {
					i.Warning += "; extension mismatch"
				} else {
					i.Warning = "extension mismatch"
				}
				return i
			}
		}
	}
	return i
}

func (r *Recorder) hasSig(puid string) bool {
	for _, v := range r.cPuids {
		if puid == v {
			return true
		}
	}
	for _, v := range r.bPuids {
		if puid == v {
			return true
		}
	}
	return false
}

type Identification struct {
	Identifier string
	Puid       string
	Name       string
	Version    string
	Mime       string
	Basis      []string
	Warning    string
	archive    config.Archive
	confidence int
}

func (id Identification) String() string {
	return id.Puid
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
	return fmt.Sprintf("  - id      : %v\n    puid    : %v\n    format  : %v\n    version : %v\n    mime    : %v\n    basis   : %v\n    warning : %v\n",
		id.Identifier, id.Puid, quoteText(id.Name), quoteText(id.Version), quoteText(id.Mime), basis, quoteText(id.Warning))
}

func (id Identification) JSON() string {
	var basis string
	if len(id.Basis) > 0 {
		basis = strings.Join(id.Basis, "; ")
	}
	return fmt.Sprintf("{\"id\":\"%s\",\"puid\":\"%s\",\"format\":\"%s\",\"version\":\"%s\",\"mime\":\"%s\",\"basis\":\"%s\",\"warning\":\"%s\"}",
		id.Identifier, id.Puid, id.Name, id.Version, id.Mime, basis, id.Warning)
}

func (id Identification) CSV() []string {
	var basis string
	if len(id.Basis) > 0 {
		basis = strings.Join(id.Basis, "; ")
	}
	return []string{
		id.Identifier,
		id.Puid,
		id.Name,
		id.Version,
		id.Mime,
		basis,
		id.Warning,
	}
}

func (id Identification) Archive() config.Archive {
	return id.archive
}

type pids []Identification

func (p pids) Len() int { return len(p) }

func (p pids) Less(i, j int) bool { return p[j].confidence < p[i].confidence }

func (p pids) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func add(p pids, id string, f string, info formatInfo, basis string, c int) pids {
	for i, v := range p {
		if v.Puid == f {
			p[i].confidence += c
			p[i].Basis = append(p[i].Basis, basis)
			return p
		}
	}
	return append(p, Identification{id, f, info.name, info.version, info.mimeType, []string{basis}, "", config.IsArchive(f), c})
}
