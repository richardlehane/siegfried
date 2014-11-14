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
	"encoding/xml"
	"errors"
	"strconv"
	"strings"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/pronom/mappings"
)

const (
	bofstring = "Absolute from BOF"
	eofstring = "Absolute from EOF"
	varstring = "Variable"
)

// Returns slice of bytematcher signatures, slice of bytematcher puids, slice of extensions, slice of extensionmatcher puids, a priority map, and a format info map
func ParseDroid(d *mappings.Droid, ids map[int]string) ([]frames.Signature, []string, [][]string, []string, priority.Map, map[string]FormatInfo, error) {
	sigs := make([]frames.Signature, 0, 700)
	bpuids := make([]string, 0, 700)
	exts := make([][]string, 0, 700)
	epuids := make([]string, 0, 700)
	pMap := make(priority.Map)
	infos := make(map[string]FormatInfo)
	for _, f := range d.FileFormats {
		// Bytematcher & bytematcher puids
		puid := f.Puid
		for _, s := range f.Signatures {
			sig, err := parseSig(puid, s)
			if err != nil {
				return nil, nil, nil, nil, nil, nil, err
			}
			sigs = append(sigs, sig)
			bpuids = append(bpuids, puid)
		}
		// Extension & extensionmatcher puids
		if len(f.Extensions) > 0 {
			exts = append(exts, f.Extensions)
			epuids = append(epuids, puid)
		}
		// Priorities
		for _, v := range f.Priorities {
			subordinate := ids[v]
			pMap.Add(subordinate, puid)
		}
		// Format infos
		infos[f.Puid] = FormatInfo{f.Name, f.Version, f.MIMEType}
	}
	return sigs, bpuids, exts, epuids, pMap, infos, nil
}

// Returns slice of bytematcher signatures, slice of bytematcher puids, slice of extensions, slice of extensionmatcher puids, a priority map, and a format info map
func ParseReport(r *mappings.Report) ([]frames.Signature, []string, [][]string, []string, priority.Map, map[string]FormatInfo) {
	return nil, nil, nil, nil, nil, nil
}

func (p *pronom) Parse() ([]frames.Signature, error) {
	sigs := make([]frames.Signature, 0, 700)
	for _, f := range p.droid.FileFormats {
		puid := f.Puid
		for _, s := range f.Signatures {
			sig, err := parseSig(puid, s)
			if err != nil {
				return nil, err
			}
			sigs = append(sigs, sig)
		}
	}
	return sigs, nil
}

func ParsePuid(f string) ([]frames.Signature, error) {
	buf, err := get(f)
	if err != nil {
		return nil, err
	}
	rep := new(mappings.Report)
	if err = xml.Unmarshal(buf, rep); err != nil {
		return nil, err
	}
	sigs := make([]frames.Signature, len(rep.Signatures))
	for i, v := range rep.Signatures {
		s, err := parseSig(f, v)
		if err != nil {
			return nil, err
		}
		sigs[i] = s
	}
	return sigs, nil
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
	// if s1 already has an EOF segment, prepend that s2 segment before it, but after any preceding segments
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
		if sub.RightFragment.Value != "" {
			min, _ = decodeNum(sub.RightFragment.MinOffset)
			if sub.RightFragment.MinOffset == "" {
				max = -1
			} else {
				max, _ = decodeNum(sub.RightFragment.MaxOffset)
			}
			fragpats, err := parseContainerSeq(puid, sub.RightFragment.Value)
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

func parseDroidSeq(puid, seq string, eof false) ([]frames.Frame, error) {
	typ := frames.PREV
	if eof {
		typ = frames.SUCC
	}
	frames := make([]frames.Frame, 0, 10)
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
			frames = append(frames, frames.NewFrame(typ, patterns.Sequence(decodeHex(i.val)), 0, 0))
		case itemNotText:
			frames = append(frames, frames.NewFrame(typ, NotSequence(decodeHex(i.val)), 0, 0))
		// parse range types
		case itemRangeStart, itemNotRangeStart:
			rangeStart = i.val
		case itemRangeEnd:
			frames = append(frames, frames.NewFrame(typ, Range{decodeHex(rangeStart), decodeHex(i.val)}, 0, 0))
		case itemNotRangeEnd:
			frames = append(frames, frames.NewFrame(typ, NotRange{decodeHex(rangeStart), decodeHex(i.val)}, 0, 0))
		}
	}
	return frames, nil
}

func groupFragments(puid string, fs []Fragment) ([][]Fragment, error) {
	var min, max string
	var maxPos int
	for _, f := range fs {
		if f.Position == 0 {
			return nil, errors.New("Pronom: encountered fragment without a position, puid " + puid)
		}
		if f.Position > maxPos {
			maxPos = f.Position
		}
	}
	ret := make([][]Fragment, maxPos)
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

func appendFragments(puid string, f []frames.Frame, frags []Fragment, left, eof bool) ([]frames.Frame, err) {
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
	for i, v := range fs {
		if len(v) > 1 {
			choice = patterns.Choice{}
			for _, c := range v {
				pats, err := parseDroidSeq(puid, c.Value, eof)
				if pats > 1 {
					return nil, errors.New("Pronom: encountered multiple patterns within a single choice, puid " + puid)
				}
				choice = append(choice, pats[0].Pat())
			}
			nfs[i] = []frames.Frame{frames.New(typ, choice, 0, 0)}
			l++ // only one choice added
		} else {
			pats, err := parseDroidSeq(puid, v[0].Value, eof)
			if err != nil {
				return nil, err
			}
			nfs[i] = pats
			l += len(pats)
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

// Container signatures are simpler than regular Droid signatures
// No BOF/EOF/VAR - all are BOF.
// Min and Max Offsets usually provided. Lack of a Max Offset implies a Variable sequence.
// No wildcards within sequences: multiple subsequences with new offsets are used instead.
func parseDroidSig(puid string, s mappings.InternalSignature) (frames.Signature, error) {
	sig := make(frames.Signature, 0, 1)
	// Return an error for multiple byte sequences
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
		if sub.RightFragment.Value != "" {
			min, _ = decodeNum(sub.RightFragment.MinOffset)
			if sub.RightFragment.MinOffset == "" {
				max = -1
			} else {
				max, _ = decodeNum(sub.RightFragment.MaxOffset)
			}
			fragpats, err := parseContainerSeq(puid, sub.RightFragment.Value)
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

/*
7B5C7274(66|6631)5C(616E7369|6D6163|7063|706361)5C616E7369637067{3-*}5C737473686664626368{1-4}5C73747368666C6F6368{1-4}5C737473686668696368{1-4}5C73747368666269

 <InternalSignature ID="26" Specificity="Specific">
            <ByteSequence Reference="BOFoffset">
                <SubSequence MinFragLength="8" Position="1"
                    SubSeqMaxOffset="0" SubSeqMinOffset="0">
                    <Sequence>5C616E7369637067</Sequence>
                    <DefaultShift>9</DefaultShift>
                    <Shift Byte="5C">8</Shift>
                    <Shift Byte="61">7</Shift>
                    <Shift Byte="63">3</Shift>
                    <Shift Byte="67">1</Shift>
                    <Shift Byte="69">4</Shift>
                    <Shift Byte="6E">6</Shift>
                    <Shift Byte="70">2</Shift>
                    <Shift Byte="73">5</Shift>
                    <LeftFragment MaxOffset="0" MinOffset="0" Position="1">616E7369</LeftFragment>
                    <LeftFragment MaxOffset="0" MinOffset="0" Position="1">6D6163</LeftFragment>
                    <LeftFragment MaxOffset="0" MinOffset="0" Position="1">7063</LeftFragment>
                    <LeftFragment MaxOffset="0" MinOffset="0" Position="1">706361</LeftFragment>
                    <LeftFragment MaxOffset="0" MinOffset="0" Position="2">5C</LeftFragment>
                    <LeftFragment MaxOffset="0" MinOffset="0" Position="3">66</LeftFragment>
                    <LeftFragment MaxOffset="0" MinOffset="0" Position="3">6631</LeftFragment>
                    <LeftFragment MaxOffset="0" MinOffset="0" Position="4">7B5C7274</LeftFragment>
                </SubSequence>
                <SubSequence MinFragLength="0" Position="2" SubSeqMinOffset="3">
                    <Sequence>5C737473686664626368</Sequence>
                    <DefaultShift>11</DefaultShift>
                    <Shift Byte="5C">10</Shift>
                    <Shift Byte="62">3</Shift>
                    <Shift Byte="63">2</Shift>
                    <Shift Byte="64">4</Shift>
                    <Shift Byte="66">5</Shift>
                    <Shift Byte="68">1</Shift>
                    <Shift Byte="73">7</Shift>
                    <Shift Byte="74">8</Shift>
                    <RightFragment MaxOffset="4" MinOffset="1" Position="1">5C73747368666C6F6368</RightFragment>
                    <RightFragment MaxOffset="4" MinOffset="1" Position="2">5C737473686668696368</RightFragment>
                    <RightFragment MaxOffset="4" MinOffset="1" Position="3">5C73747368666269</RightFragment>
                </SubSequence>
            </ByteSequence>


 	3C2F(48544D4C|68746D6C|424F4459|626F6479)3E

            <InternalSignature ID="41" Specificity="Specific">
            <ByteSequence Reference="BOFoffset">
                <SubSequence MinFragLength="0" Position="1"
                    SubSeqMaxOffset="1024" SubSeqMinOffset="0">
                    <Sequence>3C</Sequence>
                    <DefaultShift>2</DefaultShift>
                    <Shift Byte="3C">1</Shift>
                    <RightFragment MaxOffset="0" MinOffset="0" Position="1">48544D4C</RightFragment>
                    <RightFragment MaxOffset="0" MinOffset="0" Position="1">68746D6C</RightFragment>
                </SubSequence>
            </ByteSequence>
            <ByteSequence Reference="EOFoffset">
                <SubSequence MinFragLength="5" Position="1"
                    SubSeqMaxOffset="1024" SubSeqMinOffset="0">
                    <Sequence>3C2F</Sequence>
                    <DefaultShift>-3</DefaultShift>
                    <Shift Byte="2F">-2</Shift>
                    <Shift Byte="3C">-1</Shift>
                    <RightFragment MaxOffset="0" MinOffset="0" Position="1">424F4459</RightFragment>
                    <RightFragment MaxOffset="0" MinOffset="0" Position="1">48544D4C</RightFragment>
                    <RightFragment MaxOffset="0" MinOffset="0" Position="1">626F6479</RightFragment>
                    <RightFragment MaxOffset="0" MinOffset="0" Position="1">68746D6C</RightFragment>
                    <RightFragment MaxOffset="0" MinOffset="0" Position="2">3E</RightFragment>
                </SubSequence>
            </ByteSequence>
        </InternalSignature>

        <InternalSignature ID="105" Specificity="Specific">
            <ByteSequence>
                <SubSequence MinFragLength="0" Position="1" SubSeqMinOffset="0">
                    <Sequence>300D0A53454354494F4E0D0A2020320D0A4845414445520D0A</Sequence>
                    <DefaultShift>26</DefaultShift>
                    <Shift Byte="0A">1</Shift>
                    <Shift Byte="0D">2</Shift>
                    <Shift Byte="20">12</Shift>
                    <Shift Byte="30">25</Shift>
                    <Shift Byte="32">11</Shift>
                    <Shift Byte="41">6</Shift>
                    <Shift Byte="43">20</Shift>
                    <Shift Byte="44">5</Shift>
                    <Shift Byte="45">4</Shift>
                    <Shift Byte="48">8</Shift>
                    <Shift Byte="49">18</Shift>
                    <Shift Byte="4E">16</Shift>
                    <Shift Byte="4F">17</Shift>
                    <Shift Byte="52">3</Shift>
                    <Shift Byte="53">22</Shift>
                    <Shift Byte="54">19</Shift>
                </SubSequence>
                <SubSequence MinFragLength="0" Position="2" SubSeqMinOffset="0">
                    <Sequence>390D0A24414341445645520D0A2020310D0A4143</Sequence>
                    <DefaultShift>21</DefaultShift>
                    <Shift Byte="0A">3</Shift>
                    <Shift Byte="0D">4</Shift>
                    <Shift Byte="20">6</Shift>
                    <Shift Byte="24">17</Shift>
                    <Shift Byte="31">5</Shift>
                    <Shift Byte="39">20</Shift>
                    <Shift Byte="41">2</Shift>
                    <Shift Byte="43">1</Shift>
                    <Shift Byte="44">13</Shift>
                    <Shift Byte="45">11</Shift>
                    <Shift Byte="52">10</Shift>
                    <Shift Byte="56">12</Shift>
                    <RightFragment MaxOffset="0" MinOffset="0" Position="1">31303031</RightFragment>
                    <RightFragment MaxOffset="0" MinOffset="0" Position="1">322E3231</RightFragment>
                    <RightFragment MaxOffset="0" MinOffset="0" Position="1">322E3232</RightFragment>
                    <RightFragment MaxOffset="0" MinOffset="0" Position="2">0D0A</RightFragment>
                </SubSequence>
                <SubSequence MinFragLength="0" Position="3" SubSeqMinOffset="0">
                    <Sequence>300D0A454E445345430D0A</Sequence>
                    <DefaultShift>12</DefaultShift>
                    <Shift Byte="0A">1</Shift>
                    <Shift Byte="0D">2</Shift>
                    <Shift Byte="30">11</Shift>
                    <Shift Byte="43">3</Shift>
                    <Shift Byte="44">6</Shift>
                    <Shift Byte="45">4</Shift>
                    <Shift Byte="4E">7</Shift>
                    <Shift Byte="53">5</Shift>
                </SubSequence>
            </ByteSequence>
*/
