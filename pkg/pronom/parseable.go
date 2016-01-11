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
	"github.com/richardlehane/siegfried/pkg/core/parseable"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/pronom/mappings"
)

// turn generic FormatInfo into PRONOM formatInfo
func infos(m map[string]parseable.FormatInfo) map[string]formatInfo {
	i := make(map[string]formatInfo, len(m))
	for k, v := range m {
		i[k] = v.(formatInfo)
	}
	return i
}

// DoublesFilter removes the byte signatures where container signatures are also defined
type doublesFilter struct {
	ids []string
	parseable.Parseable
}

func (db *doublesFilter) Signatures() ([]frames.Signature, []string, error) {
	filter := parseable.Filter(db.ids, db.Parseable)
	return filter.Signatures()
}

// REPORTS
type reports struct {
	p  []string
	r  []*mappings.Report
	ip map[int]string
}

func (r *reports) IDs() []string {
	return r.p
}

func (r *reports) Infos() map[string]parseable.FormatInfo {
	infos := make(map[string]parseable.FormatInfo)
	for i, v := range r.r {
		infos[r.p[i]] = formatInfo{v.Name, strings.TrimSpace(v.Version), v.MIME()}
	}
	return infos
}

func (r *reports) Globs() ([][]string, []string) {
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

func (r *reports) MIMEs() ([][]string, []string) {
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

func (r *reports) XMLs() ([][][2]string, []string) {
	return nil, nil
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
}

func (d *droid) IDs() []string {
	puids := make([]string, len(d.FileFormats))
	for i, v := range d.FileFormats {
		puids[i] = v.Puid
	}
	return puids
}

func (d *droid) Infos() map[string]parseable.FormatInfo {
	infos := make(map[string]parseable.FormatInfo)
	for _, v := range d.FileFormats {
		infos[v.Puid] = formatInfo{v.Name, v.Version, v.MIMEType}
	}
	return infos
}

func (d *droid) Globs() ([][]string, []string) {
	p := d.IDs()
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

func (d *droid) MIMEs() ([][]string, []string) {
	p := d.IDs()
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

func (d *droid) XMLs() ([][][2]string, []string) {
	return nil, nil
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
		seqs[v.Id] = v.ByteSequences
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
