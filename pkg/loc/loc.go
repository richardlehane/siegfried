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

package loc

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/loc/internal/mappings"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

type fdds struct {
	f []mappings.FDD
	p identifier.Parseable
	identifier.Blank
}

func newLOC(path string) (identifier.Parseable, error) {
	rc, err := zip.OpenReader(path)
	if err != nil {
		return nil, errors.New("reading " + path + "; " + err.Error())
	}
	defer rc.Close()

	fs := make([]mappings.FDD, 0, len(rc.File))
	for _, f := range rc.File {
		dir, nm := filepath.Split(f.Name)
		if dir == "fddXML/" && nm != "" && filepath.Ext(nm) == ".xml" && !strings.Contains(nm, "test") {
			res := mappings.FDD{}
			rdr, err := f.Open()
			if err != nil {
				return nil, err
			}
			buf, err := ioutil.ReadAll(rdr)
			rdr.Close()
			if err != nil {
				return nil, err
			}
			err = xml.Unmarshal(buf, &res)
			if err != nil {
				return nil, err
			}
			fs = append(fs, res)
		}
	}

	var p identifier.Parseable = identifier.Blank{}
	if !config.NoPRONOM() {
		p, err = pronom.NewPronom()
		if err != nil {
			return nil, err
		}
	}
	return fdds{fs, p, identifier.Blank{}}, nil
}

const dateFmt = "2006-01-02"

func (f fdds) Updated() time.Time {
	t, _ := time.Parse(dateFmt, "2000-01-01")
	for _, v := range f.f {
		for _, u := range v.Updates {
			tt, err := time.Parse(dateFmt, u)
			if err == nil && tt.After(t) {
				t = tt
			}
		}
	}
	return t
}

func (f fdds) IDs() []string {
	ids := make([]string, len(f.f))
	for i, v := range f.f {
		ids[i] = v.ID
	}
	return ids
}

type formatInfo struct {
	name     string
	longName string
	mimeType string
}

func (f formatInfo) String() string {
	return f.name
}

// turn generic FormatInfo into fdd formatInfo
func infos(m map[string]identifier.FormatInfo) map[string]formatInfo {
	i := make(map[string]formatInfo, len(m))
	for k, v := range m {
		i[k] = v.(formatInfo)
	}
	return i
}

func (f fdds) Infos() map[string]identifier.FormatInfo {
	fmap := make(map[string]identifier.FormatInfo, len(f.f))
	for _, v := range f.f {
		var mime string
		if len(v.MIMEs) > 0 {
			mime = v.MIMEs[0]
		}
		fi := formatInfo{
			name:     v.Name,
			longName: v.LongName,
			mimeType: mime,
		}
		fmap[v.ID] = fi
	}
	return fmap
}

func (f fdds) Globs() ([]string, []string) {
	globs, ids := make([]string, 0, len(f.f)), make([]string, 0, len(f.f))
	for _, v := range f.f {
		for _, w := range v.Extensions {
			globs, ids = append(globs, "*."+w), append(ids, v.ID)
		}
	}
	return globs, ids
}

func (f fdds) MIMEs() ([]string, []string) {
	mimes, ids := make([]string, 0, len(f.f)), make([]string, 0, len(f.f))
	for _, v := range f.f {
		for _, w := range v.MIMEs {
			mimes, ids = append(mimes, w), append(ids, v.ID)
		}
	}
	return mimes, ids
}

func (f fdds) Signatures() ([]frames.Signature, []string, error) {
	var errs []error
	var puidsIDs map[string][]string
	if len(f.p.IDs()) > 0 {
		puidsIDs = make(map[string][]string)
	}
	sigs, ids := make([]frames.Signature, 0, len(f.f)), make([]string, 0, len(f.f))
	for _, v := range f.f {
		ss, e := magics(v.Magics)
		if e != nil {
			errs = append(errs, e)
		}
		for _, s := range ss {
			sigs = append(sigs, s)
			ids = append(ids, v.ID)
		}
		if puidsIDs != nil {
			for _, puid := range v.PUIDs() {
				puidsIDs[puid] = append(puidsIDs[puid], v.ID)
			}
		}
	}
	if puidsIDs != nil {
		puids := make([]string, 0, len(puidsIDs))
		for p := range puidsIDs {
			puids = append(puids, p)
		}
		np := identifier.Filter(puids, f.p)
		ns, ps, e := np.Signatures()
		if e != nil {
			errs = append(errs, e)
		}
		for i, v := range ps {
			for _, id := range puidsIDs[v] {
				sigs = append(sigs, ns[i])
				ids = append(ids, id)
			}
		}
	}
	var err error
	if len(errs) > 0 {
		errStrs := make([]string, len(errs))
		for i, e := range errs {
			errStrs[i] = e.Error()
		}
		err = errors.New(strings.Join(errStrs, "; "))
	}
	return sigs, ids, err
}

func (f fdds) containers(typ string) ([][]string, [][]frames.Signature, []string, error) {
	if _, ok := f.p.(identifier.Blank); ok {
		return nil, nil, nil, nil
	}
	puidsIDs := make(map[string][]string)
	for _, v := range f.f {
		for _, puid := range v.PUIDs() {
			puidsIDs[puid] = append(puidsIDs[puid], v.ID)
		}
	}
	puids := make([]string, 0, len(puidsIDs))
	for p := range puidsIDs {
		puids = append(puids, p)
	}

	np := identifier.Filter(puids, f.p)

	names, sigs, ids := make([][]string, 0, len(f.f)), make([][]frames.Signature, 0, len(f.f)), make([]string, 0, len(f.f))
	var (
		ns  [][]string
		ss  [][]frames.Signature
		is  []string
		err error
	)
	switch typ {
	default:
		err = errors.New("Unknown container type " + typ)
	case "ZIP":
		ns, ss, is, err = np.Zips()
	case "OLE2":
		ns, ss, is, err = np.MSCFBs()
	}

	if err != nil {
		return nil, nil, nil, err
	}
	for i, puid := range is {
		for _, id := range puidsIDs[puid] {
			names = append(names, ns[i])
			sigs = append(sigs, ss[i])
			ids = append(ids, id)
		}
	}
	return names, sigs, ids, nil
}

func (f fdds) Zips() ([][]string, [][]frames.Signature, []string, error) {
	return f.containers("ZIP")
}

func (f fdds) MSCFBs() ([][]string, [][]frames.Signature, []string, error) {
	return f.containers("OLE2")
}

func (f fdds) RIFFs() ([][4]byte, []string) {
	riffs, ids := make([][4]byte, 0, len(f.f)), make([]string, 0, len(f.f))
	for _, v := range f.f {
		for _, w := range v.Others {
			if w.Tag == "Microsoft FOURCC" {
				for _, x := range w.Values {
					if len(x) == 4 {
						val := [4]byte{}
						copy(val[:], x[:])
						riffs, ids = append(riffs, val), append(ids, v.ID)
					}
				}
			}
		}
	}
	return riffs, ids
}

func (f fdds) Priorities() priority.Map {
	p := make(priority.Map)
	for _, v := range f.f {
		for _, r := range v.Relations {
			switch r.Typ {
			case "Subtype of", "Modification of", "Version of", "Extension of", "Has earlier version":
				p.Add(v.ID, r.Value)
			}
		}
	}
	p.Complete()
	return p
}
