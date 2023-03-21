// Copyright 2015 Richard Lehane. All rights reserved.
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
	"strings"

	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/pronom/internal/mappings"
)

type formatInfo struct {
	name     string
	version  string
	mimeType string
	class    string
}

func (f formatInfo) String() string {
	return f.name
}

// turn generic FormatInfo into PRONOM formatInfo. TODO: use real generics
func infos(m map[string]identifier.FormatInfo) map[string]formatInfo {
	i := make(map[string]formatInfo, len(m))
	for k, v := range m {
		i[k] = v.(formatInfo)
	}
	return i
}

// DoublesFilter removes the byte signatures where container signatures are also defined
type doublesFilter struct {
	ids []string
	identifier.Parseable
}

func (db doublesFilter) Signatures() ([]frames.Signature, []string, error) {
	filter := identifier.Filter(db.ids, db.Parseable)
	return filter.Signatures()
}

// REPORTS
type reports struct {
	p  []string
	r  []*mappings.Report
	ip map[int]string
	identifier.Blank
}

func word(w string) string {
	w = strings.TrimSpace(w)
	w = strings.ToLower(w)
	ws := strings.Split(w, " ")
	w = ws[0]
	if len(ws) > 1 {
		for _, s := range ws[1:] {
			s = strings.TrimSuffix(strings.TrimPrefix(s, "("), ")")
			s = strings.Replace(s, "-", "", 1)
			w += strings.Title(s)
		}
	}
	return w
}

func normalise(ws string) []string {
	ss := strings.Split(ws, ",")
	for i, s := range ss {
		ss[i] = word(s)
	}
	if len(ss) == 1 && ss[0] == "" {
		return nil
	}
	return ss
}

func (r *reports) FamilyTypes() (map[string][]string, map[string][]string) {
	retf, rett := make(map[string][]string), make(map[string][]string)
	for i, v := range r.r {
		f, t := normalise(v.Families), normalise(v.Types)
		this := v.Label(r.p[i])
		for _, fs := range f {
			retf[fs] = append(retf[fs], this)
		}
		for _, ts := range t {
			rett[ts] = append(rett[ts], this)
		}
	}
	return retf, rett
}

func (r *reports) Labels() []string {
	ret := make([]string, len(r.p))
	for i, v := range r.r {
		ret[i] = v.Label(r.p[i])
	}
	return ret
}

func (r *reports) IDs() []string {
	return r.p
}

func (r *reports) Infos() map[string]identifier.FormatInfo {
	infos := make(map[string]identifier.FormatInfo)
	for i, v := range r.r {
		infos[r.p[i]] = formatInfo{
			name:     strings.TrimSpace(v.Name),
			version:  strings.TrimSpace(v.Version),
			mimeType: strings.TrimSpace(v.MIME()),
			class:    strings.TrimSpace(v.Types),
		}
	}
	return infos
}

func globify(s []string) []string {
	ret := make([]string, 0, len(s))
	for _, v := range s {
		if len(v) > 0 {
			ret = append(ret, "*."+v)
		}
	}
	return ret
}

func (r *reports) Globs() ([]string, []string) {
	exts := make([]string, 0, len(r.r))
	puids := make([]string, 0, len(r.p))
	for i, v := range r.r {
		for _, e := range globify(v.Extensions) {
			exts = append(exts, e)
			puids = append(puids, r.p[i])
		}
	}
	return exts, puids
}

func (r *reports) MIMEs() ([]string, []string) {
	mimes, puids := make([]string, 0, len(r.r)), make([]string, 0, len(r.p))
	for i, v := range r.r {
		if len(v.MIME()) > 0 {
			mimes, puids = append(mimes, v.MIME()), append(puids, r.p[i])
		}
	}
	return mimes, puids
}

func (r *reports) XMLs() ([][2]string, []string) {
	return nil, nil
}

func (r *reports) Texts() []string {
	return []string{config.TextPuid()}
}

func (r *reports) idsPuids() map[int]string {
	if r.ip != nil {
		return r.ip
	}
	idsPuids := make(map[int]string)
	for i, v := range r.r {
		idsPuids[v.Id] = r.p[i]
	}
	return idsPuids
}

func (r *reports) Priorities() priority.Map {
	idsPuids := r.idsPuids()
	pMap := make(priority.Map)
	for i, v := range r.r {
		this := r.p[i]
		for _, sub := range v.Subordinates() {
			pMap.Add(idsPuids[sub], this)
		}
		for _, sup := range v.Superiors() {
			pMap.Add(this, idsPuids[sup])
		}
	}
	pMap.Complete()
	return pMap
}

func (r *reports) Signatures() ([]frames.Signature, []string, error) {
	sigs, puids := make([]frames.Signature, 0, len(r.r)*2), make([]string, 0, len(r.r)*2)
	for i, rep := range r.r {
		puid := r.p[i]
		for _, v := range rep.Signatures {
			s, err := processPRONOM(puid, v)
			if err != nil {
				return nil, nil, err
			}
			sigs = append(sigs, s)
			puids = append(puids, puid)
		}
	}
	return sigs, puids, nil
}

// DROID
type droid struct {
	*mappings.Droid
	identifier.Blank
}

func (d *droid) IDs() []string {
	puids := make([]string, len(d.FileFormats))
	for i, v := range d.FileFormats {
		puids[i] = v.Puid
	}
	return puids
}

func (d *droid) Infos() map[string]identifier.FormatInfo {
	infos := make(map[string]identifier.FormatInfo)
	for _, v := range d.FileFormats {
		infos[v.Puid] = formatInfo{
			name:     strings.TrimSpace(v.Name),
			version:  strings.TrimSpace(v.Version),
			mimeType: strings.TrimSpace(v.MIMEType),
		}
	}
	return infos
}

func (d *droid) Globs() ([]string, []string) {
	p := d.IDs()
	exts, puids := make([]string, 0, len(d.FileFormats)), make([]string, 0, len(p))
	for i, v := range d.FileFormats {
		if len(v.Extensions) > 0 {
			for _, e := range globify(v.Extensions) {
				exts = append(exts, e)
				puids = append(puids, p[i])
			}
		}
	}
	return exts, puids
}

func (d *droid) MIMEs() ([]string, []string) {
	p := d.IDs()
	mimes, puids := make([]string, 0, len(d.FileFormats)), make([]string, 0, len(p))
	for i, v := range d.FileFormats {
		if len(v.MIMEType) > 0 {
			mimes, puids = append(mimes, v.MIMEType), append(puids, p[i])
		}
	}
	return mimes, puids
}

func (d *droid) XMLs() ([][2]string, []string) {
	return nil, nil
}

func (d *droid) Texts() []string {
	return []string{config.TextPuid()}
}

func (d *droid) idsPuids() map[int]string {
	idsPuids := make(map[int]string)
	for _, v := range d.FileFormats {
		idsPuids[v.ID] = v.Puid
	}
	return idsPuids
}

func (d *droid) puidsInternalIds() map[string][]int {
	puidsIIds := make(map[string][]int)
	for _, v := range d.FileFormats {
		if len(v.Signatures) > 0 {
			sigs := make([]int, len(v.Signatures))
			copy(sigs, v.Signatures)
			puidsIIds[v.Puid] = sigs
		}
	}
	return puidsIIds
}

func (d *droid) Priorities() priority.Map {
	idsPuids := d.idsPuids()
	pMap := make(priority.Map)
	for _, v := range d.FileFormats {
		superior := v.Puid
		for _, w := range v.Priorities {
			subordinate := idsPuids[w]
			pMap.Add(subordinate, superior)
		}
	}
	pMap.Complete()
	return pMap
}

func (d *droid) Signatures() ([]frames.Signature, []string, error) {
	if len(d.Droid.Signatures) == 0 {
		return nil, nil, nil
	}
	sigs, puids := make([]frames.Signature, 0, len(d.Droid.Signatures)), make([]string, 0, len(d.Droid.Signatures))
	// first a map of internal sig ids to bytesequences
	seqs := make(map[int][]mappings.ByteSeq)
	for _, v := range d.Droid.Signatures {
		seqs[v.ID] = v.ByteSequences
	}
	m := d.puidsInternalIds()
	var err error
	for _, v := range d.IDs() {
		for _, w := range m[v] {
			sig, err := processDROID(v, seqs[w])
			if err != nil {
				return nil, nil, err
			}
			sigs = append(sigs, sig)
			puids = append(puids, v)
		}
	}
	return sigs, puids, err
}

// Containers
type container struct {
	*mappings.Container
	identifier.Blank
}

func (c *container) IDs() []string {
	return c.Puids()
}

func (c *container) containerSigs(t string) ([][]string, [][]frames.Signature, []string, error) {
	// store all the puids in a map
	cpuids := make(map[int]string)
	for _, fm := range c.FormatMappings {
		cpuids[fm.Id] = fm.Puid
	}
	cp := len(c.ContainerSignatures)
	names := make([][]string, 0, cp)
	sigs := make([][]frames.Signature, 0, cp)
	puids := make([]string, 0, cp)
	for _, c := range c.ContainerSignatures {
		if c.ContainerType != t {
			continue
		}
		puid := cpuids[c.Id]
		ns, ss := make([]string, 0, len(c.Files)), make([]frames.Signature, 0, len(c.Files))
		for _, f := range c.Files {
			sig, err := processDROID(puid, f.Signature.ByteSequences)
			if err != nil {
				return nil, nil, nil, err
			}
			// write over a File if it exists: address bug x-fmt/45 (# issues 89)
			var replace bool
			for i, nm := range ns {
				if nm == f.Path {
					if sig != nil {
						ss[i] = sig
					}
					replace = true
				}
			}
			if !replace {
				ns = append(ns, f.Path)
				ss = append(ss, sig)
			}
		}
		names = append(names, ns)
		sigs = append(sigs, ss)
		puids = append(puids, cpuids[c.Id])
	}
	return names, sigs, puids, nil
}

func (c *container) Zips() ([][]string, [][]frames.Signature, []string, error) {
	return c.containerSigs("ZIP")
}
func (c *container) MSCFBs() ([][]string, [][]frames.Signature, []string, error) {
	return c.containerSigs("OLE2")
}
