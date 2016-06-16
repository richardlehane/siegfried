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

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/parseable"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/loc/mappings"
)

type fdds []mappings.FDD

func newLOC(path string) (parseable.Parseable, error) {
	rc, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	ret := make(fdds, 0, len(rc.File))
	for _, f := range rc.File {
		dir, nm := filepath.Split(f.Name)
		if dir == "fddXML/" && nm != "" && filepath.Ext(nm) == ".xml" {
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
			ret = append(ret, res)
		}
	}
	return ret, nil
}

const dateFmt = "2006-01-02"

func (f fdds) Updated() time.Time {
	t, _ := time.Parse(dateFmt, "2000-01-01")
	for _, v := range f {
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
	ids := make([]string, len(f))
	for i, v := range f {
		ids[i] = v.ID
	}
	return ids
}

type formatInfo struct {
	name string
}

// turn generic FormatInfo into fdd formatInfo
func infos(m map[string]parseable.FormatInfo) map[string]formatInfo {
	i := make(map[string]formatInfo, len(m))
	for k, v := range m {
		i[k] = v.(formatInfo)
	}
	return i
}

func (f fdds) Infos() map[string]parseable.FormatInfo {
	fmap := make(map[string]parseable.FormatInfo, len(f))
	for _, v := range f {
		fi := formatInfo{name: v.Name}
		fmap[v.ID] = fi
	}
	return fmap
}

func (f fdds) Globs() ([]string, []string) {
	globs, ids := make([]string, 0, len(f)), make([]string, 0, len(f))
	for _, v := range f {
		for _, w := range v.Extensions {
			globs, ids = append(globs, "*."+w), append(ids, v.ID)
		}
	}
	return globs, ids
}

func (f fdds) MIMEs() ([]string, []string) {
	mimes, ids := make([]string, 0, len(f)), make([]string, 0, len(f))
	for _, v := range f {
		for _, w := range v.MIMEs {
			mimes, ids = append(mimes, w), append(ids, v.ID)
		}
	}
	return mimes, ids
}

// slice of root/NS
func (f fdds) XMLs() ([][2]string, []string) {
	return nil, nil
}

func (f fdds) Signatures() ([]frames.Signature, []string, error) {
	var errs []error
	sigs, ids := make([]frames.Signature, 0, len(f)), make([]string, 0, len(f))
	for _, v := range f {
		ss, e := magics(v.Magics)
		if e != nil {
			errs = append(errs, e)
		}
		if ss == nil {
			continue
		}
		for _, s := range ss {
			sigs = append(sigs, s)
			ids = append(ids, v.ID)
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

func (f fdds) Priorities() priority.Map {
	p := make(priority.Map)
	for _, v := range f {
		for _, r := range v.Relations {
			switch r.Typ {
			case "Subtype of":
				p.Add(v.ID, r.Value)
			}
		}
	}
	p.Complete()
	return p
}
