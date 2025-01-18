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

	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/bytematcher/patterns"
	"github.com/richardlehane/siegfried/pkg/pronom/internal/mappings"
)

// This code produces siegfried bytematcher signatures from the relevant parts of PRONOM, Droid and Container XML signature files

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

// PROCompatSequence (compatibility) provides access to the PRONON
// primitive mappings.ByteSequence for custom identifier types that
// want to make use of PRONOM's level of expression.
type PROCompatSequence = mappings.ByteSequence

// BeginningOfFile provides access to PRONOM's BOF const.
const BeginningOfFile = pronombof

// EndOfFile provides access to PRONOM's EOF const.
const EndOfFile = pronomeof

// FormatPRONOM is an external helper function for enabling the
// processing of a significant number of signature types compatible with
// the PRONOM standard from plain-old hex, to more complex PRONOM regex.
func FormatPRONOM(id string, ps []PROCompatSequence) (frames.Signature, error) {
	signature := mappings.Signature{}
	signature.ByteSequences = ps
	return processPRONOM(id, signature)
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
			if seg[0].Min != 0 || seg[0].Max != 0 {
				// some signatures may begin with offsets e.g. {0-8} see e.g. fmt/1741
				min, max = seg[0].Min+min, seg[0].Max+max
			}
			seg[0] = frames.NewFrame(frames.BOF, seg[0].Pattern, min, max)
		case pronomvry:
			if max == 0 {
				max = -1
			}
			if seg[0].Min != 0 || seg[0].Max != 0 {
				// this seems iffy?
				min, max = seg[0].Min, seg[0].Max
			}
			if min == max {
				max = -1
			}
			seg[0] = frames.NewFrame(frames.BOF, seg[0].Pattern, min, max)
		case pronomeof:
			if len(seg) > 1 {
				for i, f := range seg[:len(seg)-1] {
					seg[i] = frames.NewFrame(frames.SUCC, f.Pattern, seg[i+1].Min, seg[i+1].Max)
				}
			}
			// handle edge case where there is a {x-y} at end of EOF seq e.g. x-fmt/263
			if lmin != 0 || lmax != 0 {
				min, max = lmin, lmax
			}
			seg[len(seg)-1] = frames.NewFrame(frames.EOF, seg[len(seg)-1].Pattern, min, max)
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
		switch ref {
		case droideof:
			eof = true
		case "Variable", "":
			vry = true
		}
		var zeroIndexed bool // fmt/1190 bug in containers: https://github.com/richardlehane/siegfried/issues/175
		for _, ss := range b.SubSequences {
			if ss.Position == 0 {
				zeroIndexed = true
			}
			if zeroIndexed {
				ss.Position += 1
			}
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
			return min, min, nil // if not var - max should be at least min (which is prob 0)
		}
		max, err := decodeNum(maxS)
		if err != nil {
			return 0, 0, err
		}
		if max == 0 { // fix bug fmt/837 where has a min but no max
			max = min
		}
		return min, max, nil
	}
	min, max, err := calcOffset(ss.SubSeqMinOffset, ss.SubSeqMaxOffset, vry)
	if err != nil {
		return nil, err
	}
	if eof {
		if ss.Position == 1 {
			sig[len(sig)-1] = frames.NewFrame(frames.EOF, sig[len(sig)-1].Pattern, min, max)
		} else {
			sig[len(sig)-1] = frames.NewFrame(frames.SUCC, sig[len(sig)-1].Pattern, min, max)
		}
	} else {
		if ss.Position == 1 {
			sig[0] = frames.NewFrame(frames.BOF, sig[0].Pattern, min, max)
		} else {
			sig[0] = frames.NewFrame(frames.PREV, sig[0].Pattern, min, max)
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
						list[i] = v.Pattern
					}
					choice = append(choice, list)
				} else {
					choice = append(choice, pats[0].Pattern)
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
				v[len(v)-1] = frames.NewFrame(frames.SUCC, v[len(v)-1].Pattern, offs[i][0], offs[i][1])
				sig = append(v, sig...)
			}
		} else {
			for i, v := range ns {
				sig[0] = frames.NewFrame(frames.PREV, sig[0].Pattern, offs[i][0], offs[i][1])
				sig = append(v, sig...)
			}
		}
	} else {
		if eof {
			for i, v := range ns {
				sig[len(sig)-1] = frames.NewFrame(frames.SUCC, sig[len(sig)-1].Pattern, offs[i][0], offs[i][1])
				sig = append(sig, v...)
			}

		} else {
			for i, v := range ns {
				v[0] = frames.NewFrame(frames.PREV, v[0].Pattern, offs[i][0], offs[i][1])
				sig = append(sig, v...)
			}
		}
	}
	return sig, nil
}

// Shared code for processing raw lex outputs in PRONOM/Container pattern language

func process(puid, seq string, eof bool) (frames.Signature, int, int, error) {
	if seq == "" {
		return nil, 0, 0, errors.New("parse error " + puid + ": empty sequence")
	}
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

// groups are chunks of PRONOM/Droid patterns delimited by parentheses or brackets
// these chunks represent any non-sequence pattern (choices, ranges, bitmasks, not-patterns etc.)
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
			pat = patterns.Mask(val[0])
		case anyMask:
			pat = patterns.AnyMask(val[0])
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
