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
	"github.com/richardlehane/siegfried/pkg/core/signature"
)

func init() {
	core.RegisterIdentifier(core.Pronom, Load)
}

type Identifier struct {
	p          *pronom
	Name       string
	Details    string
	NoPriority bool // was noPriority set when built?
	Infos      map[string]FormatInfo

	EStart int
	EPuids []string // slice of puids that corresponds to the extension matcher's int signatures
	CStart int
	CPuids []string
	BStart int
	BPuids []string // slice of puids that corresponds to the bytematcher's int signatures
}

type FormatInfo struct {
	Name     string
	Version  string
	MIMEType string
}

func (i *Identifier) Save(ls *signature.LoadSaver) {
	ls.SaveByte(core.Pronom)
	ls.SaveString(i.Name)
	ls.SaveString(i.Details)
	ls.SaveBool(i.NoPriority)
	ls.SaveSmallInt(len(i.Infos))
	for k, v := range i.Infos {
		ls.SaveString(k)
		ls.SaveString(v.Name)
		ls.SaveString(v.Version)
		ls.SaveString(v.MIMEType)
	}
	ls.SaveInt(i.EStart)
	ls.SaveStrings(i.EPuids)
	ls.SaveInt(i.CStart)
	ls.SaveStrings(i.CPuids)
	ls.SaveInt(i.BStart)
	ls.SaveStrings(i.BPuids)
}

func Load(ls *signature.LoadSaver) core.Identifier {
	i := &Identifier{}
	i.Name = ls.LoadString()
	i.Details = ls.LoadString()
	i.NoPriority = ls.LoadBool()
	i.Infos = make(map[string]FormatInfo)
	le := ls.LoadSmallInt()
	for j := 0; j < le; j++ {
		i.Infos[ls.LoadString()] = FormatInfo{
			ls.LoadString(),
			ls.LoadString(),
			ls.LoadString(),
		}
	}
	i.EStart = ls.LoadInt()
	i.EPuids = ls.LoadStrings()
	i.CStart = ls.LoadInt()
	i.CPuids = ls.LoadStrings()
	i.BStart = ls.LoadInt()
	i.BPuids = ls.LoadStrings()
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
	return [2]string{i.Name, i.Details}
}

func (i *Identifier) String() string {
	str := fmt.Sprintf("Name: %s\nDetails: %s\n", i.Name, i.Details)
	str += fmt.Sprintf("Number of extension signatures: %d \n", len(i.EPuids))
	str += fmt.Sprintf("Number of container signatures: %d \n", len(i.CPuids))
	str += fmt.Sprintf("Number of byte signatures: %d \n", len(i.BPuids))
	return str
}

func (i *Identifier) Recognise(m core.MatcherType, idx int) (bool, string) {
	switch m {
	default:
		return false, ""
	case core.ExtensionMatcher:
		if idx >= i.EStart && idx < i.EStart+len(i.EPuids) {
			idx = idx - i.EStart
			return true, i.Name + ": " + i.EPuids[idx]
		} else {
			return false, ""
		}
	case core.ContainerMatcher:
		if idx >= i.CStart && idx < i.CStart+len(i.CPuids) {
			idx = idx - i.CStart
			return true, i.Name + ": " + i.CPuids[idx]
		} else {
			return false, ""
		}
	case core.ByteMatcher:
		if idx >= i.BStart && idx < i.BStart+len(i.BPuids) {
			return true, i.Name + ": " + i.BPuids[idx]
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
		if res.Index() >= r.EStart && res.Index() < r.EStart+len(r.EPuids) {
			idx := res.Index() - r.EStart
			r.ids = add(r.ids, r.Name, r.EPuids[idx], r.Infos[r.EPuids[idx]], res.Basis(), r.cscore)
			return true
		} else {
			return false
		}
	case core.ContainerMatcher:
		// add zip default
		if res.Index() < 0 {
			if !r.NoPriority {
				r.cscore *= 2
			}
			r.ids = add(r.ids, r.Name, "x-fmt/263", r.Infos["x-fmt/263"], res.Basis(), r.cscore) // not great to have this hardcoded
			return false
		}
		if res.Index() >= r.CStart && res.Index() < r.CStart+len(r.CPuids) {
			idx := res.Index() - r.CStart
			if !r.NoPriority {
				r.cscore *= 2
			}
			basis := res.Basis()
			p, t := place(idx, r.CPuids)
			if t > 1 {
				basis = basis + fmt.Sprintf(" (signature %d/%d)", p, t)
			}
			r.ids = add(r.ids, r.Name, r.CPuids[idx], r.Infos[r.CPuids[idx]], basis, r.cscore)
			return true
		} else {
			return false
		}
	case core.ByteMatcher:
		if res.Index() >= r.BStart && res.Index() < r.BStart+len(r.BPuids) {
			idx := res.Index() - r.BStart
			if !r.NoPriority {
				r.cscore *= 2
			}
			basis := res.Basis()
			p, t := place(idx, r.BPuids)
			if t > 1 {
				basis = basis + fmt.Sprintf(" (signature %d/%d)", p, t)
			}
			r.ids = add(r.ids, r.Name, r.BPuids[idx], r.Infos[r.BPuids[idx]], res.Basis(), r.cscore)
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
		if r.NoPriority {
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
				nids = []Identification{Identification{r.Name, "UNKNOWN", "", "", "", nil, fmt.Sprintf("no match; possibilities based on extension are %v", strings.Join(poss, ", ")), 0, 0}}
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
		res <- Identification{r.Name, "UNKNOWN", "", "", "", nil, "no match", 0, 0}
	}
}

func (r *Recorder) checkExt(i Identification) Identification {
	if i.confidence%2 == 0 {
		for _, v := range r.EPuids {
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
	for _, v := range r.CPuids {
		if puid == v {
			return true
		}
	}
	for _, v := range r.BPuids {
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

func (id Identification) Yaml() string {
	var basis string
	if len(id.Basis) > 0 {
		basis = quoteText(strings.Join(id.Basis, "; "))
	}
	return fmt.Sprintf("  - id      : %v\n    puid    : %v\n    format  : %v\n    version : %v\n    mime    : %v\n    basis   : %v\n    warning : %v\n",
		id.Identifier, id.Puid, quoteText(id.Name), quoteText(id.Version), quoteText(id.Mime), basis, quoteText(id.Warning))
}

func (id Identification) Json() string {
	var basis string
	if len(id.Basis) > 0 {
		basis = strings.Join(id.Basis, "; ")
	}
	return fmt.Sprintf("{\"id\":\"%s\",\"puid\":\"%s\",\"format\":\"%s\",\"version\":\"%s\",\"mime\":\"%s\",\"basis\":\"%s\",\"warning\":\"%s\"}",
		id.Identifier, id.Puid, id.Name, id.Version, id.Mime, basis, id.Warning)
}

func (id Identification) Csv() []string {
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

func add(p pids, id string, f string, info FormatInfo, basis string, c int) pids {
	for i, v := range p {
		if v.Puid == f {
			p[i].confidence += c
			p[i].Basis = append(p[i].Basis, basis)
			return p
		}
	}
	return append(p, Identification{id, f, info.Name, info.Version, info.MIMEType, []string{basis}, "", config.IsArchive(f), c})
}
