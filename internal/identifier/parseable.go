// Copyright 2016 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is dis  tributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package identifier

import (
	"sort"
	"strings"

	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/pkg/config"
)

// FormatInfo is Identifier-specific information to be retained for the Identifier.
type FormatInfo interface {
	String() string
}

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
	Texts() []string                                             // IDs for textmatcher
	Priorities() priority.Map                                    // priority map
}

type inspectErr []string

func (ie inspectErr) Error() string {
	return "can't find " + strings.Join(ie, ", ")
}

// inspect returns string representations of the format signatures within a parseable
func inspect(p Parseable, ids ...string) (string, error) {
	var (
		ie                   = inspectErr{}
		fmts                 = make([]string, 0, len(ids))
		gs, gids             = p.Globs()
		ms, mids             = p.MIMEs()
		xs, xids             = p.XMLs()
		bs, bids, _          = p.Signatures()
		zns, zbs, zids, _    = p.Zips()
		msns, msbs, msids, _ = p.MSCFBs()
		rs, rids             = p.RIFFs()
		tids                 = p.Texts()
		pm                   = p.Priorities()
	)
	has := func(ss []string, s string) bool {
		for _, v := range ss {
			if s == v {
				return true
			}
		}
		return false
	}
	get := func(ss, rs []string, s string) []string {
		ret := make([]string, 0, len(ss))
		for i, v := range ss {
			if s == v {
				ret = append(ret, rs[i])
			}
		}
		return ret
	}
	getX := func(ss []string, rs [][2]string, s string) []string {
		ret := make([]string, 0, len(ss))
		for i, v := range ss {
			if s == v {
				ret = append(ret, "root: "+rs[i][0]+"; ns: "+rs[i][1])
			}
		}
		return ret
	}
	getS := func(ss []string, rs []frames.Signature, s string) []string {
		ret := make([]string, 0, len(ss))
		for i, v := range ss {
			if s == v {
				ret = append(ret, rs[i].String())
			}
		}
		return ret
	}
	getC := func(ss []string, cn [][]string, cb [][]frames.Signature, s string) []string {
		ret := make([]string, 0, len(ss))
		for i, v := range ss {
			if s == v {
				cret := make([]string, len(cn[i]))
				for j, n := range cn[i] {
					cret[j] = n
					if cb[i][j] != nil {
						cret[j] += " | " + cb[i][j].String()
					}
				}
				ret = append(ret, strings.Join(cret, "; "))
			}
		}
		return ret
	}
	getR := func(ss []string, rs [][4]byte, s string) []string {
		ret := make([]string, 0, len(ss))
		for i, v := range ss {
			if s == v {
				ret = append(ret, string(rs[i][:]))
			}
		}
		return ret
	}
	for _, id := range ids {
		lines := make([]string, 0, 10)
		info, ok := p.Infos()[id]
		if ok {
			if strings.Contains(info.String(), "\n") { // for wikidata output
				lines = append(lines, id, info.String(), "Signatures:")
			} else {
				lines = append(lines, strings.ToUpper(info.String()+" ("+id+")"))
			}
			if has(gids, id) {
				lines = append(lines, "globs: "+strings.Join(get(gids, gs, id), ", "))
			}
			if has(mids, id) {
				lines = append(lines, "mimes: "+strings.Join(get(mids, ms, id), ", "))
			}
			if has(xids, id) {
				lines = append(lines, "xmls: "+strings.Join(getX(xids, xs, id), ", "))
			}
			if has(bids, id) {
				lines = append(lines, "sigs: "+strings.Join(getS(bids, bs, id), "\n      "))
			}
			if has(zids, id) {
				lines = append(lines, "zip sigs: "+strings.Join(getC(zids, zns, zbs, id), "\n          "))
			}
			if has(msids, id) {
				lines = append(lines, "mscfb sigs: "+strings.Join(getC(msids, msns, msbs, id), "\n           "))
			}
			if has(rids, id) {
				lines = append(lines, "riffs: "+strings.Join(getR(rids, rs, id), ", "))
			}
			if has(tids, id) {
				lines = append(lines, "text signature")
			}
			// Priorities
			ps, ok := pm[id]
			if ok && len(ps) > 0 {
				lines = append(lines, "superiors: "+strings.Join(ps, ", "))
			} else {
				lines = append(lines, "superiors: none")
			}
		} else {
			ie = append(ie, id)
		}
		fmts = append(fmts, strings.Join(lines, "\n"))
	}
	if len(ie) > 0 {
		return strings.Join(fmts, "\n\n"), ie
	}
	return strings.Join(fmts, "\n\n"), nil
}

// Blank parseable can be embedded within other parseables in order to include default nil implementations of the interface
type Blank struct{}

func (b Blank) IDs() []string                                             { return nil }
func (b Blank) Infos() map[string]FormatInfo                              { return nil }
func (b Blank) Globs() ([]string, []string)                               { return nil, nil }
func (b Blank) MIMEs() ([]string, []string)                               { return nil, nil }
func (b Blank) XMLs() ([][2]string, []string)                             { return nil, nil }
func (b Blank) Signatures() ([]frames.Signature, []string, error)         { return nil, nil, nil }
func (b Blank) Zips() ([][]string, [][]frames.Signature, []string, error) { return nil, nil, nil, nil }
func (b Blank) MSCFBs() ([][]string, [][]frames.Signature, []string, error) {
	return nil, nil, nil, nil
}
func (b Blank) RIFFs() ([][4]byte, []string) { return nil, nil }
func (b Blank) Texts() []string              { return nil }
func (b Blank) Priorities() priority.Map     { return nil }

// Joint allows two parseables to be logically joined.
type joint struct {
	a, b Parseable
}

// Join two Parseables.
func Join(a, b Parseable) joint {
	return joint{a, b}
}

// IDs returns a list of all the IDs in an identifier.
func (j joint) IDs() []string {
	ids := make([]string, len(j.a.IDs()), len(j.a.IDs())+len(j.b.IDs()))
	copy(ids, j.a.IDs())
	for _, id := range j.b.IDs() {
		var present bool
		for _, ida := range j.a.IDs() {
			if id == ida {
				present = true
				break
			}
		}
		if !present {
			ids = append(ids, id)
		}
	}
	return ids
}

// Infos returns a map of identifier specific information.
func (j joint) Infos() map[string]FormatInfo {
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
func (j joint) Globs() ([]string, []string) {
	return joinStrings(j.a.Globs, j.b.Globs)
}

// MIMEs returns a signature set with corresponding IDs for the mimematcher.
func (j joint) MIMEs() ([]string, []string) {
	return joinStrings(j.a.MIMEs, j.b.MIMEs)
}

// XMLs returns a signature set with corresponding IDs for the xmlmatcher.
func (j joint) XMLs() ([][2]string, []string) {
	a, b := j.a.XMLs()
	c, d := j.b.XMLs()
	return append(a, c...), append(b, d...)
}

// Signatures returns a signature set with corresponding IDs for the bytematcher.
func (j joint) Signatures() ([]frames.Signature, []string, error) {
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
func (j joint) Priorities() priority.Map {
	ps := j.a.Priorities()
	for k, v := range j.b.Priorities() {
		for _, w := range v {
			ps.Add(k, w)
		}
	}
	return ps
}

func (j joint) Zips() ([][]string, [][]frames.Signature, []string, error) {
	n, s, i, err := j.a.Zips()
	if err != nil {
		return nil, nil, nil, err
	}
	m, q, k, err := j.b.Zips()
	if err != nil {
		return nil, nil, nil, err
	}
	return append(n, m...), append(s, q...), append(i, k...), nil
}

func (j joint) MSCFBs() ([][]string, [][]frames.Signature, []string, error) {
	n, s, i, err := j.a.MSCFBs()
	if err != nil {
		return nil, nil, nil, err
	}
	m, q, k, err := j.b.MSCFBs()
	if err != nil {
		return nil, nil, nil, err
	}
	return append(n, m...), append(s, q...), append(i, k...), nil
}
func (j joint) RIFFs() ([][4]byte, []string) {
	a, b := j.a.RIFFs()
	c, d := j.b.RIFFs()
	return append(a, c...), append(b, d...)
}

func (j joint) Texts() []string {
	txts := make([]string, len(j.a.Texts()), len(j.a.Texts())+len(j.b.Texts()))
	copy(txts, j.a.Texts())
	for _, t := range j.a.Texts() {
		var present bool
		for _, u := range j.b.Texts() {
			if t == u {
				present = true
				break
			}
		}
		if !present {
			txts = append(txts, t)
		}
	}
	return txts
}

// Filtered allows us to apply limit and exclude filters to a parseable (in both cases - provide the list of ids we want to show).
type filtered struct {
	ids []string
	p   Parseable
}

// Filter restricts a Parseable to the supplied ids. Enables limit and exclude filters.
func Filter(ids []string, p Parseable) filtered {
	return filtered{ids, p}
}

// IDs returns a list of all the IDs in an identifier.
func (f filtered) IDs() []string {
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
func (f filtered) Infos() map[string]FormatInfo {
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
func (f filtered) Globs() ([]string, []string) {
	return filterStrings(f.p.Globs, f.IDs())
}

// MIMEs returns a signature set with corresponding IDs for the mimematcher.
func (f filtered) MIMEs() ([]string, []string) {
	return filterStrings(f.p.MIMEs, f.IDs())
}

// XMLs returns a signature set with corresponding IDs for the xmlmatcher.
func (f filtered) XMLs() ([][2]string, []string) {
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
func (f filtered) Signatures() ([]frames.Signature, []string, error) {
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

func (f filtered) Zips() ([][]string, [][]frames.Signature, []string, error) {
	n, s, i, err := f.p.Zips()
	if err != nil {
		return nil, nil, nil, err
	}
	nret, sret, iret := make([][]string, 0, len(f.IDs())), make([][]frames.Signature, 0, len(f.IDs())), make([]string, 0, len(f.IDs()))
	for idx, v := range i {
		for _, w := range f.IDs() {
			if v == w {
				nret, sret, iret = append(nret, n[idx]), append(sret, s[idx]), append(iret, v)
				break
			}
		}
	}
	return nret, sret, iret, nil
}

func (f filtered) MSCFBs() ([][]string, [][]frames.Signature, []string, error) {
	n, s, i, err := f.p.MSCFBs()
	if err != nil {
		return nil, nil, nil, err
	}
	nret, sret, iret := make([][]string, 0, len(f.IDs())), make([][]frames.Signature, 0, len(f.IDs())), make([]string, 0, len(f.IDs()))
	for idx, v := range i {
		for _, w := range f.IDs() {
			if v == w {
				nret, sret, iret = append(nret, n[idx]), append(sret, s[idx]), append(iret, v)
				break
			}
		}
	}
	return nret, sret, iret, nil
}

func (f filtered) RIFFs() ([][4]byte, []string) {
	ret, retp := make([][4]byte, 0, len(f.IDs())), make([]string, 0, len(f.IDs()))
	r, p := f.p.RIFFs()
	for i, v := range p {
		for _, w := range f.IDs() {
			if v == w {
				ret, retp = append(ret, r[i]), append(retp, v)
				break
			}
		}
	}
	return ret, retp
}

func (f filtered) Texts() []string {
	txts := make([]string, 0, len(f.p.Texts()))
	for _, t := range f.p.Texts() {
		for _, u := range f.IDs() {
			if t == u {
				txts = append(txts, t)
				break
			}
		}
	}
	return txts
}

// Priorities returns a priority map.
func (f filtered) Priorities() priority.Map {
	m := f.p.Priorities()
	return m.Filter(f.IDs())
}

// Mirror reverses the PREV wild segments within signatures as SUCC/EOF wild segments so they match at EOF as well as BOF.
type Mirror struct{ Parseable }

// Signatures returns a signature set with corresponding IDs and weights for the bytematcher.
func (m Mirror) Signatures() ([]frames.Signature, []string, error) {
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

type noName struct{ Parseable }

func (nn noName) Globs() ([]string, []string) { return nil, nil }

type noMIME struct{ Parseable }

func (nm noMIME) MIMEs() ([]string, []string) { return nil, nil }

type noXML struct{ Parseable }

func (nx noXML) XMLs() ([][2]string, []string) { return nil, nil }

type noByte struct{ Parseable }

func (nb noByte) Signatures() ([]frames.Signature, []string, error) { return nil, nil, nil }

type noContainers struct{ Parseable }

func (nc noContainers) Zips() ([][]string, [][]frames.Signature, []string, error) {
	return nil, nil, nil, nil
}

func (nc noContainers) MSCFBs() ([][]string, [][]frames.Signature, []string, error) {
	return nil, nil, nil, nil
}

type noRIFF struct{ Parseable }

func (nr noRIFF) RIFFs() ([][4]byte, []string) { return nil, nil }

type noText struct{ Parseable }

func (nt noText) Texts() []string { return nil }

type noPriority struct{ Parseable }

func (np noPriority) Priorities() priority.Map { return nil }

// sorted sorts signatures by their index so that runs of signatures
// e.g. fmt/1, fmt/1, fmt/2, fmt/1 can be properly placed.
type sorted struct{ Parseable }

func (s sorted) Signatures() ([]frames.Signature, []string, error) {
	sigs, ids, err := s.Parseable.Signatures()
	if err != nil {
		return sigs, ids, err
	}
	retSigs := make([]frames.Signature, len(sigs))
	retIds := make([]string, len(ids))
	copy(retIds, ids)
	sort.Strings(retIds)
	var last string
	var nth int
	for i, this := range retIds {
		if this == last {
			nth++
		} else {
			nth = 0
			last = this
		}
		var cursor int
		for j, str := range ids {
			if this != str {
				continue
			}
			if cursor == nth {
				retSigs[i] = sigs[j]
				break
			}
			cursor++
		}
	}
	return retSigs, retIds, nil
}

func ApplyConfig(p Parseable) Parseable {
	if config.NoName() {
		p = noName{p}
	}
	if config.NoMIME() {
		p = noMIME{p}
	}
	if config.NoXML() {
		p = noXML{p}
	}
	if config.NoByte() {
		p = noByte{p}
	}
	if config.NoContainer() {
		p = noContainers{p}
	}
	if config.NoRIFF() {
		p = noRIFF{p}
	}
	if config.NoText() {
		p = noText{p}
	}
	if config.NoPriority() {
		p = noPriority{p}
	}
	// mirror PREV wild segments into EOF if maxBof and maxEOF set
	if config.MaxBOF() > 0 && config.MaxEOF() > 0 {
		p = Mirror{p}
	}
	if config.HasLimit() || config.HasExclude() {
		ids := p.IDs()
		if config.HasLimit() {
			ids = config.Limit(ids)
		} else {
			ids = config.Exclude(ids)
		}
		p = Filter(ids, p)
	}
	// Sort Parseable so runs of signatures are contiguous.
	p = sorted{p}
	return p
}
