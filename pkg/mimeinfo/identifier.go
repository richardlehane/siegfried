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

package mimeinfo

import (
	"fmt"

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/mimematcher"
	"github.com/richardlehane/siegfried/pkg/core/namematcher"
	"github.com/richardlehane/siegfried/pkg/core/parseable"
	"github.com/richardlehane/siegfried/pkg/core/persist"
	"github.com/richardlehane/siegfried/pkg/core/textmatcher"
	"github.com/richardlehane/siegfried/pkg/core/xmlmatcher"
)

func init() {
	core.RegisterIdentifier(core.MIMEInfo, Load)
}

type Identifier struct {
	p          parseable.Parseable
	name       string
	details    string
	zipDefault bool
	infos      map[string]formatInfo
	gstart     int
	gids       []string
	mstart     int
	mids       []string
	xstart     int
	xids       []string
	bstart     int
	bids       []string
	tstart     int
}

func (i *Identifier) Save(ls *persist.LoadSaver) {
	ls.SaveByte(core.MIMEInfo)
	ls.SaveString(i.name)
	ls.SaveString(i.details)
	ls.SaveBool(i.zipDefault)
	ls.SaveSmallInt(len(i.infos))
	for k, v := range i.infos {
		ls.SaveString(k)
		ls.SaveString(v.comment)
		ls.SaveInts(v.globWeights)
		ls.SaveInts(v.magicWeights)
	}
	ls.SaveInt(i.gstart)
	ls.SaveStrings(i.gids)
	ls.SaveInt(i.mstart)
	ls.SaveStrings(i.mids)
	ls.SaveInt(i.xstart)
	ls.SaveStrings(i.xids)
	ls.SaveInt(i.bstart)
	ls.SaveStrings(i.bids)
	ls.SaveSmallInt(i.tstart)
}

func Load(ls *persist.LoadSaver) core.Identifier {
	i := &Identifier{}
	i.name = ls.LoadString()
	i.details = ls.LoadString()
	i.zipDefault = ls.LoadBool()
	i.infos = make(map[string]formatInfo)
	le := ls.LoadSmallInt()
	for j := 0; j < le; j++ {
		i.infos[ls.LoadString()] = formatInfo{
			ls.LoadString(),
			ls.LoadInts(),
			ls.LoadInts(),
		}
	}
	i.gstart = ls.LoadInt()
	i.gids = ls.LoadStrings()
	i.mstart = ls.LoadInt()
	i.mids = ls.LoadStrings()
	i.xstart = ls.LoadInt()
	i.xids = ls.LoadStrings()
	i.bstart = ls.LoadInt()
	i.bids = ls.LoadStrings()
	i.tstart = ls.LoadSmallInt()
	return i
}

func contains(ss []string, str string) bool {
	for _, s := range ss {
		if s == str {
			return true
		}
	}
	return false
}

func New(opts ...config.Option) (*Identifier, error) {
	for _, v := range opts {
		v()
	}
	mi, err := newMIMEInfo()
	if err != nil {
		return nil, err
	}
	id := &Identifier{
		p:       mi,
		name:    config.Name(),
		details: config.Details(),
		infos:   infos(mi.Infos()),
	}
	if contains(mi.IDs(), config.ZipMIME()) {
		id.zipDefault = true
	}
	return id, nil
}

func (i *Identifier) Add(m core.Matcher, t core.MatcherType) error {
	switch t {
	default:
		return fmt.Errorf("MIMEInfo: unknown matcher type %d", t)
	case core.NameMatcher:
		if !config.NoName() {
			var globs []string
			globs, i.gids = i.p.Globs()
			l, err := m.Add(namematcher.SignatureSet(globs), nil)
			if err != nil {
				return err
			}
			i.gstart = l - len(i.gids)
			return nil
		}
	case core.MIMEMatcher:
		if !config.NoMIME() {
			var mimes []string
			mimes, i.mids = i.p.MIMEs()
			l, err := m.Add(mimematcher.SignatureSet(mimes), nil)
			if err != nil {
				return err
			}
			i.mstart = l - len(i.mids)
			return nil
		}
	case core.XMLMatcher:
		if !config.NoXML() {
			var xmls [][2]string
			xmls, i.xids = i.p.XMLs()
			l, err := m.Add(xmlmatcher.SignatureSet(xmls), nil)
			if err != nil {
				return err
			}
			i.xstart = l - len(i.xids)
			return nil
		}
	case core.ContainerMatcher:
		return nil
	case core.ByteMatcher:
		var sigs []frames.Signature
		var err error
		sigs, i.bids, err = i.p.Signatures()
		if err != nil {
			return err
		}
		l, err := m.Add(bytematcher.SignatureSet(sigs), nil)
		if err != nil {
			return err
		}
		i.bstart = l - len(i.bids)
	case core.TextMatcher:
		if !config.NoText() && contains(i.p.IDs(), config.TextMIME()) {
			l, _ := m.Add(textmatcher.SignatureSet{}, nil)
			i.tstart = l
		}
	}
	return nil
}

func (i *Identifier) Name() string {
	return i.name
}

func (i *Identifier) Details() string {
	return i.details
}

func (i *Identifier) String() string {
	str := fmt.Sprintf("Name: %s\nDetails: %s\n", i.name, i.details)
	str += fmt.Sprintf("Number of filename signatures: %d \n", len(i.gids))
	str += fmt.Sprintf("Number of MIME signatures: %d \n", len(i.mids))
	str += fmt.Sprintf("Number of XML signatures: %d \n", len(i.xids))
	str += fmt.Sprintf("Number of byte signatures: %d \n", len(i.bids))
	return str
}

func (i *Identifier) Recognise(m core.MatcherType, idx int) (bool, string) {
	switch m {
	default:
		return false, ""
	case core.NameMatcher:
		if idx >= i.gstart && idx < i.gstart+len(i.gids) {
			idx = idx - i.gstart
			return true, i.name + ": " + i.gids[idx]
		}
		return false, ""
	case core.MIMEMatcher:
		if idx >= i.mstart && idx < i.mstart+len(i.mids) {
			idx = idx - i.mstart
			return true, i.name + ": " + i.mids[idx]
		}
		return false, ""
	case core.XMLMatcher:
		if idx >= i.xstart && idx < i.xstart+len(i.xids) {
			idx = idx - i.xstart
			return true, i.name + ": " + i.xids[idx]
		}
		return false, ""
	case core.ContainerMatcher:
		return false, ""
	case core.ByteMatcher:
		if idx >= i.bstart && idx < i.bstart+len(i.bids) {
			return true, i.name + ": " + i.bids[idx]
		}
		return false, ""
	case core.TextMatcher:
		if idx == i.tstart {
			return true, i.name + ": " + config.TextPuid()
		}
		return false, ""
	}
}

func (i *Identifier) Recorder() core.Recorder {
	return nil
}
