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

type FormatInfo struct {
	Name     string
	Version  string
	MIMEType string
}

type Identifier struct {
	p       *pronom
	Name    string
	Details string
	Infos   map[string]FormatInfo

	EStart int
	EPuids []string // slice of puids that corresponds to the extension matcher's int signatures
	CStart int
	CPuids []string
	BStart int
	BPuids []string         // slice of puids that corresponds to the bytematcher's int signatures
	PuidsB map[string][]int // map of puids to slices of bytematcher int signatures
}

func New(opts ...config.Option) (*Identifier, error) {
	for _, v := range opts {
		v()
	}
	pronom, err := NewPronom()
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
		i.Name, i.Details)
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
			r.cscore *= 1.1
			r.ids = add(r.ids, r.Name, "x-fmt/263", r.Infos["x-fmt/263"], res.Basis(), r.cscore) // not great to have this hardcoded
			return false
		}
		if res.Index() >= r.CStart && res.Index() < r.CStart+len(r.CPuids) {
			idx := res.Index() - r.CStart
			r.cscore *= 1.1
			r.ids = add(r.ids, r.Name, r.CPuids[idx], r.Infos[r.CPuids[idx]], res.Basis(), r.cscore)
			return true
		} else {
			return false
		}
	case core.ByteMatcher:
		if res.Index() >= r.BStart && res.Index() < r.BStart+len(r.BPuids) {
			idx := res.Index() - r.BStart
			r.cscore *= 1.1
			r.ids = add(r.ids, r.Name, r.BPuids[idx], r.Infos[r.BPuids[idx]], res.Basis(), r.cscore)
			return true
		} else {
			return false
		}
	}
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
		conf := r.ids[0].confidence
		// if we've only got extension matches, check if those matches are ruled out by lack of byte match
		// add warnings too
		if conf == 0.1 {
			nids := make([]Identification, 0, len(r.ids))
			for _, v := range r.ids {
				if _, ok := r.PuidsB[v.puid]; !ok {
					v.warning = "match on extension only"
					nids = append(nids, v)
				}
			}
			if len(nids) == 0 {
				poss := make([]string, len(r.ids))
				for i, v := range r.ids {
					poss[i] = v.puid
				}
				nids = []Identification{Identification{r.Name, "UNKNOWN", "", "", "", nil, fmt.Sprintf("no match; possibilities based on extension are %v", strings.Join(poss, ", ")), 0}}
			}
			r.ids = nids
		}
		res <- r.ids[0]
		if len(r.ids) > 1 {
			for i, v := range r.ids[1:] {
				if v.confidence == conf {
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
	identifier string
	puid       string
	name       string
	version    string
	mime       string
	basis      []string
	warning    string
	confidence float64
}

func (id Identification) String() string {
	return id.puid
}

func quoteText(s string) string {
	if len(s) == 0 {
		return s
	}
	return "\"" + s + "\""
}

func (id Identification) Yaml() string {
	var basis string
	if len(id.basis) > 0 {
		basis = quoteText(strings.Join(id.basis, "; "))
	}
	return fmt.Sprintf("  - id      : %v\n    puid    : %v\n    format  : %v\n    version : %v\n    mime    : %v\n    basis   : %v\n    warning : %v\n",
		id.identifier, id.puid, quoteText(id.name), quoteText(id.version), quoteText(id.mime), basis, quoteText(id.warning))
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
	if len(id.basis) > 0 {
		basis = strings.Join(id.basis, "; ")
	}
	b, err := json.Marshal(jsonid{id.puid, id.name, id.version, id.mime, basis, id.warning})
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

type pids []Identification

func (p pids) Len() int { return len(p) }

func (p pids) Less(i, j int) bool { return p[j].confidence < p[i].confidence }

func (p pids) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func add(p pids, id string, f string, info FormatInfo, basis string, c float64) pids {
	for i, v := range p {
		if v.puid == f {
			p[i].confidence += c
			p[i].basis = append(p[i].basis, basis)
			return p
		}
	}
	return append(p, Identification{id, f, info.Name, info.Version, info.MIMEType, []string{basis}, "", c})
}
