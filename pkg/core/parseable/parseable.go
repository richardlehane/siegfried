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

package parseable

import (
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/priority"
)

// FormatInfo is Identifier-specific information to be retained for the Identifier.
type FormatInfo interface{}

// Parseable is something we can parse to derive filename, MIME, XML and byte signatures.
type Parseable interface {
	IDs() []string                                               // list of all IDs in identifier
	Infos() map[string]FormatInfo                                // identifier specific information
	Globs() ([]string, []string)                                 // signature set and corresponding IDs for globmatcher
	MIMEs() ([]string, []string)                                 // signature set and corresponding IDs for mimematcher
	XMLs() ([][2]string, []string)                               // signature set and corresponding IDs for xmlmatcher
	Signatures() ([]frames.Signature, []string, error)           // signature set and corresponding IDs for bytematcher
	Zips() ([][]string, [][]frames.Signature, []string, error)   // signature set and corresponding IDs for container matcher - Zip
	MSCFBs() ([][]string, [][]frames.Signature, []string, error) // signature set and corresponding IDs for container matcher - MSCFB
	RIFFs() ([][4]byte, []string)                                // signature set and corresponding IDs for riffmatcher
	Priorities() priority.Map                                    // priority map
}

// Blank parseable can be embedded within other parseables in order to include default nil implementations of the interface
type Blank struct{}

func (b Blank) IDs() []string                                               { return nil }
func (b Blank) Infos() map[string]FormatInfo                                { return nil }
func (b Blank) Globs() ([]string, []string)                                 { return nil, nil }
func (b Blank) MIMEs() ([]string, []string)                                 { return nil, nil }
func (b Blank) XMLs() ([][2]string, []string)                               { return nil, nil }
func (b Blank) Signatures() ([]frames.Signature, []string, error)           { return nil, nil, nil }
func (b Blank) Zips() ([][]string, [][]frames.Signature, []string, error)   { return nil, nil, nil, nil }
func (b Blank) MSCFBs() ([][]string, [][]frames.Signature, []string, error) { return nil, nil, nil, nil }
func (b Blank) RIFFs() ([][4]byte, []string)                                { return nil, nil }
func (b Blank) Priorities() priority.Map                                    { return nil }

// Joint allows two parseables to be logically joined.
type joint struct {
	a, b Parseable
}

// Join two Parseables.
func Join(a, b Parseable) *joint {
	return &joint{a, b}
}

// IDs returns a list of all the IDs in an identifier.
func (j *joint) IDs() []string {
	return append(j.a.IDs(), j.b.IDs()...)
}

// Infos returns a map of identifier specific information.
func (j *joint) Infos() map[string]FormatInfo {
	infos := j.a.Infos()
	for k, v := range j.b.Infos() {
		infos[k] = v
	}
	return infos
}

func joinStrings(a func() ([]string, []string), b func() ([]string, []string)) ([]string, []string) {
	c, d := a()
	e, f := b()
	return append(c, e...), append(d, f...)
}

// Globs returns a signature set with corresponding IDs for the globmatcher.
func (j *joint) Globs() ([]string, []string) {
	return joinStrings(j.a.Globs, j.b.Globs)
}

// MIMEs returns a signature set with corresponding IDs for the mimematcher.
func (j *joint) MIMEs() ([]string, []string) {
	return joinStrings(j.a.MIMEs, j.b.MIMEs)
}

// XMLs returns a signature set with corresponding IDs for the xmlmatcher.
func (j *joint) XMLs() ([][2]string, []string) {
	a, b := j.a.XMLs()
	c, d := j.b.XMLs()
	return append(a, c...), append(b, d...)
}

// Signatures returns a signature set with corresponding IDs and weights for the bytematcher.
func (j *joint) Signatures() ([]frames.Signature, []string, error) {
	s, p, err := j.a.Signatures()
	if err != nil {
		return nil, nil, err
	}
	t, q, err := j.b.Signatures()
	if err != nil {
		return nil, nil, err
	}
	return append(s, t...), append(p, q...), nil
}

// Priorities returns a priority map.
func (j *joint) Priorities() priority.Map {
	ps := j.a.Priorities()
	for k, v := range j.b.Priorities() {
		for _, w := range v {
			ps.Add(k, w)
		}
	}
	return ps
}

func (j *joint) Zips() ([][]string, [][]frames.Signature, []string, error) { return nil, nil, nil, nil }
func (j *joint) MSCFBs() ([][]string, [][]frames.Signature, []string, error) {
	return nil, nil, nil, nil
}
func (j *joint) RIFFs() ([][4]byte, []string) { return nil, nil }

// Filtered allows us to apply limit and exclude filters to a parseable (in both cases - provide the list of ids we want to show).
type filtered struct {
	ids []string
	p   Parseable
}

// Filter restricts a Parseable to the supplied ids. Enables limit and exclude filters.
func Filter(ids []string, p Parseable) *filtered {
	return &filtered{ids, p}
}

// IDs returns a list of all the IDs in an identifier.
func (f *filtered) IDs() []string {
	ret := make([]string, 0, len(f.ids))
	for _, v := range f.p.IDs() {
		for _, w := range f.ids {
			if v == w {
				ret = append(ret, v)
				break
			}
		}
	}
	return ret
}

// Infos returns a map of identifier specific information.
func (f *filtered) Infos() map[string]FormatInfo {
	ret, infos := make(map[string]FormatInfo), f.p.Infos()
	for _, v := range f.IDs() {
		ret[v] = infos[v]
	}
	return ret
}

func filterStrings(a func() ([]string, []string), ids []string) ([]string, []string) {
	ret, retp := make([]string, 0, len(ids)), make([]string, 0, len(ids))
	e, p := a()
	for i, v := range p {
		for _, w := range ids {
			if v == w {
				ret, retp = append(ret, e[i]), append(retp, v)
				break
			}
		}
	}
	return ret, retp
}

// Globs returns a signature set with corresponding IDs for the globmatcher.
func (f *filtered) Globs() ([]string, []string) {
	return filterStrings(f.p.Globs, f.IDs())
}

// MIMEs returns a signature set with corresponding IDs for the mimematcher.
func (f *filtered) MIMEs() ([]string, []string) {
	return filterStrings(f.p.MIMEs, f.IDs())
}

// XMLs returns a signature set with corresponding IDs for the xmlmatcher.
func (f *filtered) XMLs() ([][2]string, []string) {
	ret, retp := make([][2]string, 0, len(f.IDs())), make([]string, 0, len(f.IDs()))
	e, p := f.p.XMLs()
	for i, v := range p {
		for _, w := range f.IDs() {
			if v == w {
				ret, retp = append(ret, e[i]), append(retp, v)
				break
			}
		}
	}
	return ret, retp
}

// Signatures returns a signature set with corresponding IDs and weights for the bytematcher.
func (f *filtered) Signatures() ([]frames.Signature, []string, error) {
	s, p, err := f.p.Signatures()
	if err != nil {
		return nil, nil, err
	}
	ret, retp := make([]frames.Signature, 0, len(f.IDs())), make([]string, 0, len(f.IDs()))
	for i, v := range p {
		for _, w := range f.IDs() {
			if v == w {
				ret, retp = append(ret, s[i]), append(retp, v)
				break
			}
		}
	}
	return ret, retp, nil
}

func (f *filtered) Zips() ([][]string, [][]frames.Signature, []string, error) {
	return nil, nil, nil, nil
}
func (f *filtered) MSCFBs() ([][]string, [][]frames.Signature, []string, error) {
	return nil, nil, nil, nil
}
func (f *filtered) RIFFs() ([][4]byte, []string) { return nil, nil }

// Priorities returns a priority map.
func (f *filtered) Priorities() priority.Map {
	m := f.p.Priorities()
	return m.Filter(f.IDs())
}

// Mirror reverses the PREV wild segments within signatures as SUCC/EOF wild segments so they match at EOF as well as BOF.
type Mirror struct{ Parseable }

// Signatures returns a signature set with corresponding IDs and weights for the bytematcher.
func (m *Mirror) Signatures() ([]frames.Signature, []string, error) {
	sigs, ids, err := m.Parseable.Signatures()
	if err != nil {
		return sigs, ids, err
	}
	rsigs := make([]frames.Signature, 0, len(sigs)+100)
	rids := make([]string, 0, len(sigs)+100)
	for i, v := range sigs {
		rsigs = append(rsigs, v)
		rids = append(rids, ids[i])
		mirror := v.Mirror()
		if mirror != nil {
			rsigs = append(rsigs, mirror)
			rids = append(rids, ids[i])
		}
	}
	return rsigs, rids, nil
}
