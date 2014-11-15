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
		infos[r.p[i]] = FormatInfo{v.Name, v.Version, v.MIME()}
	}
	return infos
}

func (r *reports) extensions() ([][]string, []string) {
	exts := make([][]string, len(r.r))
	for i, v := range r.r {
		exts[i] = v.Extensions
	}
	return exts, r.p
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
			s, err := parseSig(puid, v)
			if err != nil {
				return nil, nil, err
			}
			sigs = append(sigs, s)
			puids = append(puids, puid)
		}
	}
	return sigs, puids, nil
}

const (
	bofstring = "Absolute from BOF"
	eofstring = "Absolute from EOF"
	varstring = "Variable"
)

// parse sig takes a signature from a report and returns a BM signature
func parseSig(puid string, s mappings.Signature) (frames.Signature, error) {
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
		}
		// parse the hexstring
		toks, lmin, lmax, err := parseHex(puid, bs.Hex)
		if err != nil {
			return nil, err
		}
		// create a new signature for this set of tokens
		tokSig := make(frames.Signature, len(toks))
		// check position and add patterns to signature
		switch bs.Position {
		case bofstring:
			if toks[0].min == 0 && toks[0].max == 0 {
				toks[0].min, toks[0].max = min, max
			}
			tokSig[0] = frames.NewFrame(frames.BOF, toks[0].pat, toks[0].min, toks[0].max)
			if len(toks) > 1 {
				for i, tok := range toks[1:] {
					tokSig[i+1] = frames.NewFrame(frames.PREV, tok.pat, tok.min, tok.max)
				}
			}
		case varstring:
			if max == 0 {
				max = -1
			}
			if toks[0].min == 0 && toks[0].max == 0 {
				toks[0].min, toks[0].max = min, max
			}
			if toks[0].min == toks[0].max {
				toks[0].max = -1
			}
			tokSig[0] = frames.NewFrame(frames.BOF, toks[0].pat, toks[0].min, toks[0].max)
			if len(toks) > 1 {
				for i, tok := range toks[1:] {
					tokSig[i+1] = frames.NewFrame(frames.PREV, tok.pat, tok.min, tok.max)
				}
			}
		case eofstring:
			if len(toks) > 1 {
				for i, tok := range toks[:len(toks)-1] {
					tokSig[i] = frames.NewFrame(frames.SUCC, tok.pat, toks[i+1].min, toks[i+1].max)
				}
			}
			// handle edge case where there is a {x-y} at end of EOF seq e.g. x-fmt/263
			if lmin != 0 || lmax != 0 {
				min, max = lmin, lmax
			}
			tokSig[len(toks)-1] = frames.NewFrame(frames.EOF, toks[len(toks)-1].pat, min, max)
		default:
			return nil, errors.New("Pronom parse error: invalid ByteSequence position " + bs.Position)
		}
		// add the segment (tokens signature) to the complete signature
		sig = appendSig(sig, tokSig, bs.Position)
	}
	return sig, nil
}

// an intermediary structure before creating a bytematcher.Frame
type token struct {
	min, max int
	pat      patterns.Pattern
}

// helper funcs
func decodeHex(hx string) []byte {
	buf, _ := hex.DecodeString(hx) // ignore err, the hex string has been lexed
	return buf
}

func decodeNum(num string) (int, error) {
	if strings.TrimSpace(num) == "" {
		return 0, nil
	}
	return strconv.Atoi(num)
}

// parse hexstrings - puids are passed in for error reporting
func parseHex(puid, hx string) ([]token, int, int, error) {
	tokens := make([]token, 0, 10)
	var choice patterns.Choice // common bucket for stuffing choices into
	var rangeStart string
	var min, max int
	l := sigLex(puid, hx)
	for i := l.nextItem(); i.typ != itemEOF; i = l.nextItem() {
		switch i.typ {
		case itemError:
			return nil, 0, 0, errors.New(i.String())
		// parse simple types
		case itemText:
			tokens = append(tokens, token{min, max, patterns.Sequence(decodeHex(i.val))})
		case itemNotText:
			tokens = append(tokens, token{min, max, NotSequence(decodeHex(i.val))})
		// parse range types
		case itemRangeStart, itemNotRangeStart, itemRangeStartChoice, itemNotRangeStartChoice:
			rangeStart = i.val
		case itemRangeEnd:
			tokens = append(tokens, token{min, max, Range{decodeHex(rangeStart), decodeHex(i.val)}})
		case itemNotRangeEnd:
			tokens = append(tokens, token{min, max, NotRange{decodeHex(rangeStart), decodeHex(i.val)}})
		// parse choice types
		case itemParensLeft:
			choice = make(patterns.Choice, 0, 2)
		case itemTextChoice:
			choice = append(choice, patterns.Sequence(decodeHex(i.val)))
		case itemNotTextChoice:
			choice = append(choice, NotSequence(decodeHex(i.val)))
		case itemRangeEndChoice:
			choice = append(choice, Range{decodeHex(rangeStart), decodeHex(i.val)})
		case itemNotRangeEndChoice:
			choice = append(choice, NotRange{decodeHex(rangeStart), decodeHex(i.val)})
		case itemParensRight:
			tokens = append(tokens, token{min, max, choice})
		// parse wild cards
		case itemWildSingle:
			min++
			max++
		case itemWildStart:
			min, _ = decodeNum(i.val)
		case itemCurlyRight: //detect {n} wildcards (i.e. not ranges) by checking if the max value has been set
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
		}
		// if we've appended a pattern, reset min and max
		switch i.typ {
		case itemText, itemNotText, itemRangeEnd, itemNotRangeEnd, itemParensRight:
			min, max = 0, 0
		}
	}
	return tokens, min, max, nil
}

// merge two segments into a signature. Provide s2's pos
func appendSig(s1, s2 frames.Signature, pos string) frames.Signature {
	if len(s1) == 0 {
		return s2
	}
	// if s2 is an EOF - just append it
	if pos == eofstring {
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
	exts := make([][]string, len(d.FileFormats))
	for i, v := range d.FileFormats {
		exts[i] = v.Extensions
	}
	return exts, d.puids()
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
	sigs, puids := make([]frames.Signature, len(d.Signatures)), make([]string, len(d.Signatures))
	// first a map of internal sig ids to bytesequences
	seqs := make(map[int][]mappings.ByteSeq)
	for _, v := range d.Signatures {
		seqs[v.Id] = v.ByteSequences
	}
	m := d.puidsInternalIds()
	var i int
	var err error
	for _, v := range d.puids() {
		for _, w := range m[v] {
			sigs[i], err = parseByteSeqs(v, seqs[w])
			if err != nil {
				return nil, nil, err
			}
			puids[i] = v
			i++
		}
	}
	return sigs, puids, err
}

func parseByteSeqs(puid string, bs []mappings.ByteSeq) (frames.Signature, error) {
	return nil, nil
}

func parseSeq(puid, seq string, eof bool) ([]frames.Frame, error) {
	typ := frames.PREV
	if eof {
		typ = frames.SUCC
	}
	fs := make([]frames.Frame, 0, 10)
	var rangeStart string
	l := droidLex(puid, seq)
	for i := l.nextItem(); i.typ != itemEOF; i = l.nextItem() {
		switch i.typ {
		default:
			return nil, errors.New(i.String())
		case itemError:
			return nil, errors.New(i.String())
		// parse simple types
		case itemText:
			fs = append(fs, frames.NewFrame(typ, patterns.Sequence(decodeHex(i.val)), 0, 0))
		case itemNotText:
			fs = append(fs, frames.NewFrame(typ, NotSequence(decodeHex(i.val)), 0, 0))
		// parse range types
		case itemRangeStart, itemNotRangeStart:
			rangeStart = i.val
		case itemRangeEnd:
			fs = append(fs, frames.NewFrame(typ, Range{decodeHex(rangeStart), decodeHex(i.val)}, 0, 0))
		case itemNotRangeEnd:
			fs = append(fs, frames.NewFrame(typ, NotRange{decodeHex(rangeStart), decodeHex(i.val)}, 0, 0))
		}
	}
	return fs, nil
}

// droid fragments (right or left) can share positions. If such fragments have same offsets, they are a patterns.Choice. If not, then err.
func groupFragments(puid string, fs []mappings.Fragment) ([][]mappings.Fragment, error) {
	//var min, max string
	var maxPos int
	for _, f := range fs {
		if f.Position == 0 {
			return nil, errors.New("Pronom: encountered fragment without a position, puid " + puid)
		}
		if f.Position > maxPos {
			maxPos = f.Position
		}
	}
	ret := make([][]mappings.Fragment, maxPos)
	for _, f := range fs {
		ret[f.Position] = append(ret[f.Position], f)
	}
	for _, r := range ret {
		max, min := r[0].MaxOffset, r[0].MinOffset
		for _, v := range r {
			if v.MaxOffset != max || v.MinOffset != min {
				return nil, errors.New("Pronom: encountered fragments at same positions with different offsets, puid " + puid)
			}
		}
	}
	return ret, nil
}

// append a slice of fragments (left or right) to the central droid sequence
func appendFragments(puid string, f []frames.Frame, frags []mappings.Fragment, left, eof bool) ([]frames.Frame, error) {
	fs, err := groupFragments(puid, frags)
	if err != nil {
		return nil, err
	}
	typ := frames.PREV
	if eof {
		typ = frames.SUCC
	}
	var choice patterns.Choice
	offs := make([][2]int, len(fs))
	nfs := make([][]frames.Frame, len(fs))
	l := len(f)
	// iterate over the grouped fragments
	for i, v := range fs {
		if len(v) > 1 {
			choice = patterns.Choice{}
			for _, c := range v {
				pats, err := parseSeq(puid, c.Value, eof)
				if err != nil {
					return nil, err
				}
				if len(pats) > 1 {
					return nil, errors.New("Pronom: encountered multiple patterns within a single choice, puid " + puid)
				}
				choice = append(choice, pats[0].Pat())
			}
			nfs[i] = []frames.Frame{frames.NewFrame(typ, choice, 0, 0)}
			l++ // only one choice added
		} else {
			pats, err := parseSeq(puid, v[0].Value, eof)
			if err != nil {
				return nil, err
			}
			nfs[i] = pats
			l += len(pats) // can have multiple patterns within a fragment
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

	} else {

	}

	ret := make([]frames.Frame, l)
	return ret, nil
}

// CONTAINER

func parseContainerSeq(puid, seq string) ([]patterns.Pattern, error) {
	pats := make([]patterns.Pattern, 0, 10)
	var insideBracket bool
	var choiceMode bool
	var rangeMode bool
	var choice patterns.Choice // common bucket for stuffing choices into
	var firstBit []byte        // first text within brackets (could be a range or a choice)
	sequence := make(patterns.Sequence, 0)
	l := conLex(puid, seq)
	for i := l.nextItem(); i.typ != itemEOF; i = l.nextItem() {
		switch i.typ {
		case itemError:
			return nil, errors.New(i.String())
		case itemText:
			if insideBracket {
				if choiceMode {
					choice = append(choice, patterns.Sequence(decodeHex(i.val)))
				} else if rangeMode {
					pats = append(pats, Range{firstBit, decodeHex(i.val)})
				} else {
					firstBit = decodeHex(i.val)
				}
			} else {
				sequence = append(sequence, decodeHex(i.val)...)
			}
		case itemQuoteText:
			if insideBracket {
				if choiceMode {
					choice = append(choice, patterns.Sequence(i.val))
				} else if rangeMode {
					pats = append(pats, Range{firstBit, []byte(i.val)})
				} else {
					firstBit = []byte(i.val)
				}
			} else {
				sequence = append(sequence, []byte(i.val)...)
			}
		case itemSpace:
			if insideBracket {
				if !choiceMode {
					choiceMode = true
					choice = patterns.Choice{patterns.Sequence(firstBit)}
				}
			}
		case itemSlash:
			if insideBracket {
				rangeMode = true
			} else {
				return nil, errors.New("Pronom parse error: unexpected slash in container (appears outside brackets)")
			}
		case itemColon:
			if insideBracket {
				rangeMode = true
			} else {
				return nil, errors.New("Pronom parse error: unexpected colon in container (appears outside brackets)")
			}
		case itemBracketLeft:
			if len(sequence) > 0 {
				pats = append(pats, sequence)
				sequence = make(patterns.Sequence, 0)
			}
			insideBracket = true
		case itemBracketRight:
			if choiceMode {
				pats = append(pats, choice)
			}
			insideBracket, choiceMode, rangeMode = false, false, false
		}
	}
	if len(sequence) > 0 {
		pats = append(pats, sequence)
	}
	return pats, nil
}

// Container signatures are simpler than regular Droid signatures
// No BOF/EOF/VAR - all are BOF.
// Min and Max Offsets usually provided. Lack of a Max Offset implies a Variable sequence.
// No wildcards within sequences: multiple subsequences with new offsets are used instead.
func parseContainerSig(puid string, s mappings.InternalSignature) (frames.Signature, error) {
	// some sigs only have paths, this is OK
	if s.ByteSequences == nil {
		return nil, nil
	}
	sig := make(frames.Signature, 0, 1)
	// Return an error for multiple byte sequences
	if len(s.ByteSequences) > 1 {
		return nil, errors.New("Pronom parse error: unexpected multiple byte sequences in container sig for puid " + puid)
	}
	bs := s.ByteSequences[0]
	// Return an error for non-BOF sequence
	if bs.Reference != "" && bs.Reference != "BOFoffset" {
		return nil, errors.New("Pronom parse error: unexpected reference in container sig for puid " + puid + "; bad reference is " + bs.Reference)
	}
	var prevPos int
	for i, sub := range bs.SubSequences {
		// Return an error if the positions don't increment.
		if sub.Position < prevPos {
			return nil, errors.New("Pronom parse error: container sub-sequences out of order for puid " + puid)
		}
		prevPos = sub.Position
		var typ frames.OffType
		if i == 0 {
			typ = frames.BOF
		} else {
			typ = frames.PREV
		}
		var min, max int
		min, _ = decodeNum(sub.SubSeqMinOffset)
		if sub.SubSeqMaxOffset == "" {
			max = -1
		} else {
			max, _ = decodeNum(sub.SubSeqMaxOffset)
		}
		pats, err := parseContainerSeq(puid, sub.Sequence)
		if err != nil {
			return nil, err
		}
		sig = append(sig, frames.NewFrame(typ, pats[0], min, max))
		if len(pats) > 1 {
			for _, v := range pats[1:] {
				sig = append(sig, frames.NewFrame(frames.PREV, v, 0, 0))
			}
		}
		if len(sub.RightFragments) > 0 {
			min, _ = decodeNum(sub.RightFragments[0].MinOffset)
			if sub.RightFragments[0].MinOffset == "" {
				max = -1
			} else {
				max, _ = decodeNum(sub.RightFragments[0].MaxOffset)
			}
			fragpats, err := parseContainerSeq(puid, sub.RightFragments[0].Value)
			if err != nil {
				return nil, err
			}
			sig = append(sig, frames.NewFrame(frames.PREV, fragpats[0], min, max))
			if len(fragpats) > 1 {
				for _, v := range fragpats[1:] {
					sig = append(sig, frames.NewFrame(frames.PREV, v, 0, 0))
				}
			}
		}

	}
	return sig, nil
}
