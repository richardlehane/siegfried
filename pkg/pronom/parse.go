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
	"encoding/hex"
	"errors"
	"strconv"
	"strings"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/pronom/mappings"
)

// a parseable is something we can parse (either a DROID signature file or PRONOM report file)
// to derive extension and bytematcher signatures
type parseable interface {
	puids() []string
	infos() map[string]FormatInfo
	extensions() ([][]string, []string)
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

func (j *joint) infos() map[string]FormatInfo {
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
// limit and exclude filters control what parts of a parseable we show
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

func (f *filter) infos() map[string]FormatInfo {
	ret, infos := make(map[string]FormatInfo), f.p.infos()
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

// REPORTS
type reports struct {
	p  []string
	r  []*mappings.Report
	ip map[int]string
}

func (r *reports) puids() []string {
	return r.p
}

func (r *reports) infos() map[string]FormatInfo {
	infos := make(map[string]FormatInfo)
	for i, v := range r.r {
		infos[r.p[i]] = FormatInfo{v.Name, strings.TrimSpace(v.Version), v.MIME()}
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

func (d *droid) infos() map[string]FormatInfo {
	infos := make(map[string]FormatInfo)
	for _, v := range d.FileFormats {
		infos[v.Puid] = FormatInfo{v.Name, v.Version, v.MIMEType}
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

// PROCESSING
// (take PRONOM/DROID/CONTAINER XML and make frames.Signature(s))

const (
	pronombof = "Absolute from BOF"
	pronomeof = "Absolute from EOF"
	pronomvry = "Variable"
	droidbof  = "BOFoffset"
	droideof  = "EOFoffset"
)

// helper
func decodeNum(num string) (int, error) {
	if strings.TrimSpace(num) == "" {
		return 0, nil
	}
	return strconv.Atoi(num)
}

// PRONOM
func processPRONOM(puid string, s mappings.Signature) (frames.Signature, error) {
	sig := make(frames.Signature, 0, 1)
	for _, bs := range s.ByteSequences {
		// check if <Offset> or <MaxOffset> elements are present
		min, err := decodeNum(bs.Offset)
		if err != nil {
			return nil, err
		}
		max, err := decodeNum(bs.MaxOffset)
		if err != nil {
			return nil, err
		}
		// lack of a max offset implies a fixed offset for BOF and EOF seqs (not VAR)
		if max == 0 {
			max = min
		} else {
			max = max + min // the max offset in a PRONOM report is relative to the "offset" value, not to the BOF/EOF
		}
		var eof bool
		if bs.Position == pronomeof {
			eof = true
		}
		// parse the hexstring
		seg, lmin, lmax, err := process(puid, bs.Hex, eof)
		if err != nil {
			return nil, err
		}
		// check position and add patterns to signature
		switch bs.Position {
		case pronombof:
			if seg[0].Min() != 0 || seg[0].Max() != 0 {
				min, max = seg[0].Min(), seg[0].Max()
			}
			seg[0] = frames.NewFrame(frames.BOF, seg[0].Pat(), min, max)
		case pronomvry:
			if max == 0 {
				max = -1
			}
			if seg[0].Min() != 0 || seg[0].Max() != 0 {
				min, max = seg[0].Min(), seg[0].Max()
			}
			if min == max {
				max = -1
			}
			seg[0] = frames.NewFrame(frames.BOF, seg[0].Pat(), min, max)
		case pronomeof:
			if len(seg) > 1 {
				for i, f := range seg[:len(seg)-1] {
					seg[i] = frames.NewFrame(frames.SUCC, f.Pat(), seg[i+1].Min(), seg[i+1].Max())
				}
			}
			// handle edge case where there is a {x-y} at end of EOF seq e.g. x-fmt/263
			if lmin != 0 || lmax != 0 {
				min, max = lmin, lmax
			}
			seg[len(seg)-1] = frames.NewFrame(frames.EOF, seg[len(seg)-1].Pat(), min, max)
		default:
			return nil, errors.New("Pronom parse error: invalid ByteSequence position " + bs.Position)
		}
		// add the segment to the complete signature
		sig = appendSig(sig, seg, bs.Position)
	}
	return sig, nil
}

// merge two segments into a signature. Provide s2's pos
func appendSig(s1, s2 frames.Signature, pos string) frames.Signature {
	if len(s1) == 0 {
		return s2
	}
	// if s2 is an EOF - just append it
	if pos == pronomeof || pos == droideof {
		return append(s1, s2...)
	}
	// if s1 has an EOF segment, and s2 is a BOF or Var, prepend that s2 segment before it, but after any preceding segments
	for i, f := range s1 {
		orientation := f.Orientation()
		if orientation == frames.SUCC || orientation == frames.EOF {
			s3 := make(frames.Signature, len(s1)+len(s2))
			copy(s3, s1[:i])
			copy(s3[i:], s2)
			copy(s3[i+len(s2):], s1[i:])
			return s3
		}
	}
	// default is just to append it
	return append(s1, s2...)
}

// DROID & Container
func processDROID(puid string, s []mappings.ByteSeq) (frames.Signature, error) {
	var sig frames.Signature
	for _, b := range s {
		var eof, vry bool
		ref := b.Reference
		if ref == droideof {
			eof = true
		} else if ref == "" {
			vry = true
		}
		for _, ss := range b.SubSequences {
			ns, err := processSubSequence(puid, ss, eof, vry)
			if err != nil {
				return nil, err
			}
			sig = appendSig(sig, ns, ref)
		}
	}
	return sig, nil
}

func processSubSequence(puid string, ss mappings.SubSequence, eof, vry bool) (frames.Signature, error) {
	sig, _, _, err := process(puid, ss.Sequence, eof)
	if err != nil {
		return nil, err
	}
	if len(ss.LeftFragments) > 0 {
		sig, err = appendFragments(puid, sig, ss.LeftFragments, true, eof)
		if err != nil {
			return nil, err
		}
	}
	if len(ss.RightFragments) > 0 {
		sig, err = appendFragments(puid, sig, ss.RightFragments, false, eof)
		if err != nil {
			return nil, err
		}
	}
	if ss.Position > 1 {
		vry = true
	}
	calcOffset := func(minS, maxS string, vry bool) (int, int, error) {
		min, err := decodeNum(minS)
		if err != nil {
			return 0, 0, err
		}
		if maxS == "" {
			if vry {
				return min, -1, nil
			}
			return min, 0, nil
		}
		max, err := decodeNum(maxS)
		if err != nil {
			return 0, 0, err
		}
		return min, max, nil
	}
	min, max, err := calcOffset(ss.SubSeqMinOffset, ss.SubSeqMaxOffset, vry)
	if err != nil {
		return nil, err
	}
	if eof {
		if ss.Position == 1 {
			sig[len(sig)-1] = frames.NewFrame(frames.EOF, sig[len(sig)-1].Pat(), min, max)
		} else {
			sig[len(sig)-1] = frames.NewFrame(frames.SUCC, sig[len(sig)-1].Pat(), min, max)
		}
	} else {
		if ss.Position == 1 {
			sig[0] = frames.NewFrame(frames.BOF, sig[0].Pat(), min, max)
		} else {
			sig[0] = frames.NewFrame(frames.PREV, sig[0].Pat(), min, max)
		}
	}
	return sig, nil
}

// append a slice of fragments (left or right) to the central droid sequence
func appendFragments(puid string, sig frames.Signature, frags []mappings.Fragment, left, eof bool) (frames.Signature, error) {
	// First off, group the fragments:
	// droid fragments (right or left) can share positions. If such fragments have same offsets, they are a patterns.Choice. If not, then err.
	var maxPos int
	for _, f := range frags {
		if f.Position == 0 {
			return nil, errors.New("Pronom: encountered fragment without a position, puid " + puid)
		}
		if f.Position > maxPos {
			maxPos = f.Position
		}
	}
	fs := make([][]mappings.Fragment, maxPos)
	for _, f := range frags {
		fs[f.Position-1] = append(fs[f.Position-1], f)
	}
	for _, r := range fs {
		max, min := r[0].MaxOffset, r[0].MinOffset
		for _, v := range r {
			if v.MaxOffset != max || v.MinOffset != min {
				return nil, errors.New("Pronom: encountered fragments at same positions with different offsets, puid " + puid)
			}
		}
	}
	typ := frames.PREV
	if eof {
		typ = frames.SUCC
	}
	var choice patterns.Choice
	offs := make([][2]int, len(fs))
	ns := make([]frames.Signature, len(fs))
	//l := len(sig)
	// iterate over the grouped fragments
	for i, v := range fs {
		if len(v) > 1 {
			choice = patterns.Choice{}
			for _, c := range v {
				pats, _, _, err := process(puid, c.Value, eof)
				if err != nil {
					return nil, err
				}
				if len(pats) > 1 {
					list := make(patterns.List, len(pats))
					for i, v := range pats {
						list[i] = v.Pat()
					}
					choice = append(choice, list)
				} else {
					choice = append(choice, pats[0].Pat())
				}
			}
			ns[i] = frames.Signature{frames.NewFrame(typ, choice, 0, 0)}
		} else {
			pats, _, _, err := process(puid, v[0].Value, eof)
			if err != nil {
				return nil, err
			}
			ns[i] = pats
		}
		min, err := decodeNum(v[0].MinOffset)
		if err != nil {
			return nil, err
		}
		var max int
		if v[0].MaxOffset == "" {
			max = -1
		} else {
			max, err = decodeNum(v[0].MaxOffset)
			if err != nil {
				return nil, err
			}
		}
		offs[i] = [2]int{min, max}
	}
	// Now make the frames by adding in offset information (if left fragments, this needs to be taken from their neighbour)
	if left {
		if eof {
			for i, v := range ns {
				v[len(v)-1] = frames.NewFrame(frames.SUCC, v[len(v)-1].Pat(), offs[i][0], offs[i][1])
				sig = append(v, sig...)
			}
		} else {
			for i, v := range ns {
				sig[0] = frames.NewFrame(frames.PREV, sig[0].Pat(), offs[i][0], offs[i][1])
				sig = append(v, sig...)
			}
		}
	} else {
		if eof {
			for i, v := range ns {
				sig[len(sig)-1] = frames.NewFrame(frames.SUCC, sig[len(sig)-1].Pat(), offs[i][0], offs[i][1])
				sig = append(sig, v...)
			}

		} else {
			for i, v := range ns {
				v[0] = frames.NewFrame(frames.PREV, v[0].Pat(), offs[i][0], offs[i][1])
				sig = append(sig, v...)
			}
		}
	}
	return sig, nil
}

// Shared code for processing raw lex outputs in PRONOM/Container pattern language

func process(puid, seq string, eof bool) (frames.Signature, int, int, error) {
	typ := frames.PREV
	if eof {
		typ = frames.SUCC
	}
	var min, max int
	l := lexPRONOM(puid, seq)
	sig := frames.Signature{}
	for i := l.nextItem(); i.typ != itemEOF; i = l.nextItem() {
		switch i.typ {
		case itemError:
			return nil, 0, 0, errors.New("parse error " + puid + ": " + i.String())
		case itemWildSingle:
			min++
			max++
		case itemWildStart:
			min, _ = decodeNum(i.val)
		case itemCurlyRight: //detect {n} wildcards by checking if the max value has been set
			if max == 0 {
				max = min
			}
		case itemWildEnd:
			if i.val == "*" {
				max = -1
			} else {
				max, _ = decodeNum(i.val)
			}
		case itemWild:
			max = -1
		case itemEnterGroup:
			pat, err := processGroup(l)
			if err != nil {
				return nil, 0, 0, errors.New("parse error " + puid + ": " + err.Error())
			}
			sig = append(sig, frames.NewFrame(typ, pat, min, max))
			min, max = 0, 0
		case itemUnprocessedText:
			sig = append(sig, frames.NewFrame(typ, patterns.Sequence(processText(i.val)), min, max))
			min, max = 0, 0
		}
	}
	return sig, min, max, nil
}

func processText(hx string) []byte {
	var buf []byte
	l := lexText(hx)
	for i := range l.items {
		switch i.typ {
		case itemHexText:
			byts, _ := hex.DecodeString(i.val)
			buf = append(buf, byts...)
		case itemQuoteText:
			buf = append(buf, []byte(i.val)...)
		case itemError:
			panic(i.val)
		case itemEOF:
			return buf
		}
	}
	// ignore err, the hex string has been lexed
	return buf
}

//
func processGroup(l *lexer) (patterns.Pattern, error) {
	var (
		list                    patterns.List   // bucket to stuff patterns into
		choice                  patterns.Choice // bucket to stuff choices into
		val                     []byte          // bucket to stuff text values
		not, mask, anyMask, rng bool            // retains state from previous tokens
	)
	// when commit a pattern (to the list), go back to zero state
	reset := func() {
		val = []byte{}
		not, mask, anyMask, rng = false, false, false, false
	}
	// make a pattern based on the current state
	makePat := func() patterns.Pattern {
		if len(val) == 0 {
			return nil
		}
		var pat patterns.Pattern
		switch {
		case mask:
			pat = Mask(val[0])
		case anyMask:
			pat = AnyMask(val[0])
		default:
			pat = patterns.Sequence(val)
		}
		if not {
			pat = patterns.Not{pat}
		}
		reset()
		return pat
	}
	// add patterns to the choice
	addChoice := func() (patterns.Choice, error) {
		switch len(list) {
		case 0:
			return nil, errors.New(l.name + " has choice marker without preceding pattern")
		case 1:
			choice = append(choice, list[0])
		default:
			choice = append(choice, list)
		}
		list = patterns.List{}
		return choice, nil
	}
	for {
		i := <-l.items
		switch i.typ {
		default:
			return nil, errors.New(l.name + " encountered unexpected token " + i.val)
		case itemEnterGroup: // recurse e.g. for a range nested within a choice
			if pat := makePat(); pat != nil {
				list = append(list, pat)
			}
			pat, err := processGroup(l)
			if err != nil {
				return nil, err
			}
			list = append(list, pat)
		case itemExitGroup:
			if pat := makePat(); pat != nil {
				list = append(list, pat)
			}
			if len(choice) > 0 {
				return addChoice()
			} else {
				switch len(list) {
				case 0:
					return nil, errors.New(l.name + " has group with no legal pattern")
				case 1:
					return list[0], nil
				default:
					return list, nil
				}
			}
		case itemRangeMarker:
			rng = true
		case itemChoiceMarker:
			if pat := makePat(); pat != nil {
				list = append(list, pat)
			}
			_, err := addChoice()
			if err != nil {
				return nil, err
			}
		case itemNotMarker:
			not = true
		case itemMaskMarker:
			mask = true
		case itemAnyMaskMarker:
			anyMask = true
		case itemUnprocessedText:
			v := processText(i.val)
			// if it is a range, we need values before and after the range marker, so add it here
			if rng {
				r := Range{val, v}
				if not {
					list = append(list, patterns.Not{r})
				} else {
					list = append(list, r)
				}
				reset()
			} else {
				val = v
			}
		}
	}
}
