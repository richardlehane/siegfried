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

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/pronom/mappings"
)

// a parseable is something we can parse (either a DROID signature file or PRONOM report file)
// to derive extension and bytematcher signatures
type parseable interface {
	puids() []string
	infos() map[string]formatInfo
	extensions() ([][]string, []string)
	mimes() ([][]string, []string)
	signatures() ([]frames.Signature, []string, error)
	priorities() priority.Map
}

// JOINT
// joint allows two parseables to be logically joined (we want to merge droid signatures with pronom report signatures)
type joint struct {
	a, b parseable
}

func join(a, b parseable) *joint {
	return &joint{a, b}
}

func (j *joint) puids() []string {
	return append(j.a.puids(), j.b.puids()...)
}

func (j *joint) infos() map[string]formatInfo {
	infos := j.a.infos()
	for k, v := range j.b.infos() {
		infos[k] = v
	}
	return infos
}

func (j *joint) extensions() ([][]string, []string) {
	e, p := j.a.extensions()
	f, q := j.b.extensions()
	return append(e, f...), append(p, q...)
}

func (j *joint) mimes() ([][]string, []string) {
	e, p := j.a.mimes()
	f, q := j.b.mimes()
	return append(e, f...), append(p, q...)
}

func (j *joint) signatures() ([]frames.Signature, []string, error) {
	s, p, err := j.a.signatures()
	if err != nil {
		return nil, nil, err
	}
	t, q, err := j.b.signatures()
	if err != nil {
		return nil, nil, err
	}
	return append(s, t...), append(p, q...), nil
}

func (j *joint) priorities() priority.Map {
	ps := j.a.priorities()
	for k, v := range j.b.priorities() {
		for _, w := range v {
			ps.Add(k, w)
		}
	}
	return ps
}

// FILTERS
// a filter allows us to apply limit and exclude filters to a parseable (in both cases - provide the list of puids we want to show)
type filter struct {
	ids []string
	p   parseable
}

func applyFilter(puids []string, p parseable) *filter {
	return &filter{puids, p}
}

func (f *filter) puids() []string {
	ret := make([]string, 0, len(f.ids))
	for _, v := range f.p.puids() {
		for _, w := range f.ids {
			if v == w {
				ret = append(ret, v)
				break
			}
		}
	}
	return ret
}

func (f *filter) infos() map[string]formatInfo {
	ret, infos := make(map[string]formatInfo), f.p.infos()
	for _, v := range f.puids() {
		ret[v] = infos[v]
	}
	return ret
}

func (f *filter) extensions() ([][]string, []string) {
	ret, retp := make([][]string, 0, len(f.puids())), make([]string, 0, len(f.puids()))
	e, p := f.p.extensions()
	for i, v := range p {
		for _, w := range f.puids() {
			if v == w {
				ret, retp = append(ret, e[i]), append(retp, v)
				break
			}
		}
	}
	return ret, retp
}

func (f *filter) mimes() ([][]string, []string) {
	ret, retp := make([][]string, 0, len(f.puids())), make([]string, 0, len(f.puids()))
	e, p := f.p.mimes()
	for i, v := range p {
		for _, w := range f.puids() {
			if v == w {
				ret, retp = append(ret, e[i]), append(retp, v)
				break
			}
		}
	}
	return ret, retp
}

func (f *filter) signatures() ([]frames.Signature, []string, error) {
	s, p, err := f.p.signatures()
	if err != nil {
		return nil, nil, err
	}
	ret, retp := make([]frames.Signature, 0, len(f.puids())), make([]string, 0, len(f.puids()))
	for i, v := range p {
		for _, w := range f.puids() {
			if v == w {
				ret, retp = append(ret, s[i]), append(retp, v)
				break
			}
		}
	}
	return ret, retp, nil
}

func (f *filter) priorities() priority.Map {
	m := f.p.priorities()
	return m.Filter(f.puids())
}

// a limited filter that just removes the byte signatures where container signatures are also defined
type doublesFilter struct {
	ids []string
	parseable
}

func (db *doublesFilter) signatures() ([]frames.Signature, []string, error) {
	filter := applyFilter(db.ids, db.parseable)
	return filter.signatures()
}

// a mirror parseable mirrors the PREV wild segments within signatures as SUCC/EOF wild segments so they match at EOF as well as BOF
type mirror struct{ parseable }

func (m *mirror) signatures() ([]frames.Signature, []string, error) {
	sigs, puids, err := m.parseable.signatures()
	if err != nil {
		return sigs, puids, err
	}
	rsigs := make([]frames.Signature, 0, len(sigs)+100)
	rpuids := make([]string, 0, len(sigs)+100)
	for i, v := range sigs {
		rsigs = append(rsigs, v)
		rpuids = append(rpuids, puids[i])
		mirror := v.Mirror()
		if mirror != nil {
			rsigs = append(rsigs, mirror)
			rpuids = append(rpuids, puids[i])
		}
	}
	return rsigs, rpuids, nil
}

// REPORTS
type reports struct {
	p  []string
	r  []*mappings.Report
	ip map[int]string
}

func (r *reports) puids() []string {
	return r.p
}

func (r *reports) infos() map[string]formatInfo {
	infos := make(map[string]formatInfo)
	for i, v := range r.r {
		infos[r.p[i]] = formatInfo{v.Name, strings.TrimSpace(v.Version), v.MIME()}
	}
	return infos
}

func (r *reports) extensions() ([][]string, []string) {
	exts := make([][]string, 0, len(r.r))
	puids := make([]string, 0, len(r.p))
	for i, v := range r.r {
		if len(v.Extensions) > 0 {
			exts = append(exts, v.Extensions)
			puids = append(puids, r.p[i])
		}
	}
	return exts, puids
}

func (r *reports) mimes() ([][]string, []string) {
	mimes := make([][]string, 0, len(r.r))
	puids := make([]string, 0, len(r.p))
	for i, v := range r.r {
		if len(v.MIME()) > 0 {
			mimes = append(mimes, []string{v.MIME()})
			puids = append(puids, r.p[i])
		}
	}
	return mimes, puids
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

func (r *reports) priorities() priority.Map {
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
	return pMap
}

func (r *reports) signatures() ([]frames.Signature, []string, error) {
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
}

func (d *droid) puids() []string {
	puids := make([]string, len(d.FileFormats))
	for i, v := range d.FileFormats {
		puids[i] = v.Puid
	}
	return puids
}

func (d *droid) infos() map[string]formatInfo {
	infos := make(map[string]formatInfo)
	for _, v := range d.FileFormats {
		infos[v.Puid] = formatInfo{v.Name, v.Version, v.MIMEType}
	}
	return infos
}

func (d *droid) extensions() ([][]string, []string) {
	p := d.puids()
	exts := make([][]string, 0, len(d.FileFormats))
	puids := make([]string, 0, len(p))
	for i, v := range d.FileFormats {
		if len(v.Extensions) > 0 {
			exts = append(exts, v.Extensions)
			puids = append(puids, p[i])
		}
	}
	return exts, puids
}

func (d *droid) mimes() ([][]string, []string) {
	p := d.puids()
	mimes := make([][]string, 0, len(d.FileFormats))
	puids := make([]string, 0, len(p))
	for i, v := range d.FileFormats {
		if len(v.MIMEType) > 0 {
			mimes = append(mimes, []string{v.MIMEType})
			puids = append(puids, p[i])
		}
	}
	return mimes, puids
}

func (d *droid) idsPuids() map[int]string {
	idsPuids := make(map[int]string)
	for _, v := range d.FileFormats {
		idsPuids[v.Id] = v.Puid
	}
	return idsPuids
}

func (d *droid) puidsInternalIds() map[string][]int {
	puidsIIds := make(map[string][]int)
	for _, v := range d.FileFormats {
		if len(v.Signatures) > 0 {
			sigs := make([]int, len(v.Signatures))
			for j, w := range v.Signatures {
				sigs[j] = w
			}
			puidsIIds[v.Puid] = sigs
		}
	}
	return puidsIIds
}

func (d *droid) priorities() priority.Map {
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

func (d *droid) signatures() ([]frames.Signature, []string, error) {
	if len(d.Signatures) == 0 {
		return nil, nil, nil
	}
	sigs, puids := make([]frames.Signature, 0, len(d.Signatures)), make([]string, 0, len(d.Signatures))
	// first a map of internal sig ids to bytesequences
	seqs := make(map[int][]mappings.ByteSeq)
	for _, v := range d.Signatures {
		seqs[v.Id] = v.ByteSequences
	}
	m := d.puidsInternalIds()
	var err error
	for _, v := range d.puids() {
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
