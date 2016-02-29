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

package mimeinfo

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"strings"

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/parseable"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/mimeinfo/mappings"
)

type mimeinfo []mappings.MIMEType

func (mi mimeinfo) identifier() *Identifier { return nil }

func newMIMEInfo() (mimeinfo, error) {
	buf, err := ioutil.ReadFile(config.MIMEInfo())
	if err != nil {
		return nil, err
	}
	mi := &mappings.MIMEInfo{}
	err = xml.Unmarshal(buf, mi)
	if err != nil {
		return nil, err
	}
	return mi.MIMETypes, nil
}

func (mi mimeinfo) IDs() []string {
	ids := make([]string, len(mi))
	for i, v := range mi {
		ids[i] = v.MIME
	}
	return ids
}

type formatInfo struct {
	globWeights  []int
	magicWeights []int
}

func (mi mimeinfo) Infos() map[string]parseable.FormatInfo { return nil }

func (mi mimeinfo) Globs() ([][]string, []string) {
	globs, ids := make([][]string, 0, len(mi)), make([]string, 0, len(mi))
	for _, v := range mi {
		if len(v.Globs) > 0 {
			g := make([]string, len(v.Globs))
			for i, w := range v.Globs {
				g[i] = w.Pattern
			}
			globs, ids = append(globs, g), append(ids, v.MIME)
		}
	}
	return globs, ids
}

func (mi mimeinfo) MIMEs() ([][]string, []string) {
	mimes, ids := make([][]string, len(mi)), make([]string, len(mi))
	for idx, v := range mi {
		r := make([]string, len(v.Aliases)+1)
		r[0] = v.MIME
		for i, w := range v.Aliases {
			r[i+1] = w.Alias
		}
		mimes[idx], ids[idx] = r, v.MIME
	}
	return mimes, ids
}

// slice of root/NS
func (mi mimeinfo) XMLs() ([][][2]string, []string) {
	xmls, ids := make([][][2]string, 0, len(mi)), make([]string, 0, len(mi))
	for _, v := range mi {
		if len(v.XMLPattern) > 0 {
			x := make([][2]string, len(v.XMLPattern))
			for i, w := range v.XMLPattern {
				x[i] = [2]string{w.Local, w.NS}
			}
			xmls, ids = append(xmls, x), append(ids, v.MIME)
		}
	}
	return xmls, ids
}

func (mi mimeinfo) Signatures() ([]frames.Signature, []string, error) {
	var errStrs []string
	sigs, ids := make([]frames.Signature, 0, len(mi)), make([]string, 0, len(mi))
	var err error
	if len(errStrs) > 0 {
		err = errors.New(strings.Join(errStrs, "; "))
	}
	return sigs, ids, err
}

// we don't create a priority map for mimeinfo
func (mi mimeinfo) Priorities() priority.Map { return nil }
