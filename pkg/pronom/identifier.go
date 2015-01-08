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
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

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
	BPuids []string         // slice of puids that corresponds to the bytematcher's int signatures
	PuidsB map[string][]int // map of puids to slices of bytematcher int signatures
}

type FormatInfo struct {
	Name     string
	Version  string
	MIMEType string
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

func (i *Identifier) Yaml() string {
	return fmt.Sprintf("  - name    : %v\n    details : %v\n",
		i.Name, quoteText(i.Details))
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

func (i *Identifier) Save(w io.Writer) (int, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(i)
	if err != nil {
		return 0, err
	}
	sz := buf.Len()
	_, err = buf.WriteTo(w)
	if err != nil {
		return 0, err
	}
	return sz, nil
}

func Load(r io.Reader) (*Identifier, error) {
	i := &Identifier{}
	dec := gob.NewDecoder(r)
	err := dec.Decode(i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (i *Identifier) Recorder() core.Recorder {
	return &Recorder{i, make(pids, 0, 10), 0.1}
}

type Recorder struct {
	*Identifier
	ids    pids
	cscore float64
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
				r.cscore *= 1.1
			}
			r.ids = add(r.ids, r.Name, "x-fmt/263", r.Infos["x-fmt/263"], res.Basis(), r.cscore) // not great to have this hardcoded
			return false
		}
		if res.Index() >= r.CStart && res.Index() < r.CStart+len(r.CPuids) {
			idx := res.Index() - r.CStart
			if !r.NoPriority {
				r.cscore *= 1.1
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
				r.cscore *= 1.1
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
	if r.cscore == 0.1 {
		return false
	}
	return true
}

func (r *Recorder) Report(res chan core.Identification) {
	if len(r.ids) > 0 {
		sort.Sort(r.ids)
		conf := r.ids[0].Confidence
		// if we've only got extension matches, check if those matches are ruled out by lack of byte match
		// add warnings too
		if r.NoPriority {
			for i := range r.ids {
				r.ids[i].Warning = "no priority set for this identifier"
			}
		} else if conf == 0.1 {
			nids := make([]Identification, 0, len(r.ids))
			for _, v := range r.ids {
				if _, ok := r.PuidsB[v.Puid]; !ok {
					v.Warning = "match on extension only"
					nids = append(nids, v)
				}
			}
			if len(nids) == 0 {
				poss := make([]string, len(r.ids))
				for i, v := range r.ids {
					poss[i] = v.Puid
				}
				nids = []Identification{Identification{r.Name, "UNKNOWN", "", "", "", nil, fmt.Sprintf("no match; possibilities based on extension are %v", strings.Join(poss, ", ")), 0}}
			}
			r.ids = nids
		}
		res <- r.ids[0]
		if len(r.ids) > 1 {
			for i, v := range r.ids[1:] {
				if v.Confidence == conf {
					res <- r.ids[i+1]
				} else {
					break
				}
			}
		}
	} else {
		res <- Identification{r.Name, "UNKNOWN", "", "", "", nil, "no match", 0}
	}
}

type Identification struct {
	Identifier string
	Puid       string
	Name       string
	Version    string
	Mime       string
	Basis      []string
	Warning    string
	Confidence float64
}

func (id Identification) String() string {
	return id.Puid
}

func quoteText(s string) string {
	if len(s) == 0 {
		return s
	}
	return "\"" + s + "\""
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
	type jsonid struct {
		Puid    string `json:"puid"`
		Name    string `json:"name"`
		Version string `json:"version"`
		Mime    string `json:"mime"`
		Basis   string `json:"basis"`
		Warning string `json:"warning"`
	}
	var basis string
	if len(id.Basis) > 0 {
		basis = strings.Join(id.Basis, "; ")
	}
	b, err := json.Marshal(jsonid{id.Puid, id.Name, id.Version, id.Mime, basis, id.Warning})
	if err != nil {
		return `{
			"puid": "",
			"name": "",
			"version": "",
			"mime": "",
			"basis": "",
			"warning": "json encoding error"
			}`
	}
	return string(b)
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

type pids []Identification

func (p pids) Len() int { return len(p) }

func (p pids) Less(i, j int) bool { return p[j].Confidence < p[i].Confidence }

func (p pids) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func add(p pids, id string, f string, info FormatInfo, basis string, c float64) pids {
	for i, v := range p {
		if v.Puid == f {
			p[i].Confidence += c
			p[i].Basis = append(p[i].Basis, basis)
			return p
		}
	}
	return append(p, Identification{id, f, info.Name, info.Version, info.MIMEType, []string{basis}, "", c})
}
