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
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/bytematcher/patterns"
)

func magics(m []string) ([]frames.Signature, error) {
	hx, ascii, hxx, asciix, err := characterise(m)
	if err != nil {
		return nil, err
	}
	if len(hx) > 0 {
		sigs := make([]frames.Signature, len(hx))
		for i, v := range hx {
			byts, offs, masks, err := dehex(v, hxx[i])
			if err != nil {
				return nil, err
			}
			sigs[i] = make(frames.Signature, len(byts))
			for ii, vv := range byts {
				rel := frames.BOF
				if ii > 0 {
					rel = frames.PREV
				}
				var pat patterns.Pattern
				if masks[ii] {
					pat = patterns.Mask(vv[0])
				} else {
					pat = patterns.Sequence(vv)
				}
				sigs[i][ii] = frames.NewFrame(rel, pat, offs[ii], offs[ii])
			}
		}
		return sigs, nil
	} else if len(ascii) > 0 {
		sigs := make([]frames.Signature, len(ascii))
		for i, v := range ascii {
			pat := patterns.Sequence(v)
			sigs[i] = frames.Signature{frames.NewFrame(frames.BOF, pat, asciix[i], asciix[i])}
		}
		return sigs, nil
	}
	return nil, nil
}

// return raw hex and ascii signatures (ascii to be used only if no hex)
func characterise(m []string) ([]string, []string, []int, []int, error) {
	hx, ascii := []string{}, []string{}
	hxx, asciix := []int{}, []int{}
	for _, v := range m {
		v = strings.Replace(v, " (ASCII: \"1↵00:\")", "", 1)
		tokens := strings.SplitN(v, ": ", 2)
		switch len(tokens) {
		case 1:
			if v == "Not applicable" { // special case fdd000230
				continue
			}
			_, _, _, err := dehex(v, 0)
			if err == nil {
				hx = append(hx, v)
				hxx = append(hxx, 0)
				continue
			}
			ascii = append(ascii, v)
			asciix = append(asciix, 0)
		case 2:
			switch strings.TrimSpace(tokens[0]) { // special case fdd000147
			case "Hex", "HEX":
				hx = append(hx, tokens[1])
				hxx = append(hxx, 0)
			case "ASCII":
				ascii = append(ascii, tokens[1])
				asciix = append(asciix, 0)
			case "12 byte string": // special case fdd000127
				hx = append(hx, strings.TrimSuffix(strings.TrimPrefix(tokens[1], "X'"), "'"))
				hxx = append(hxx, 0)
			case "Hex (position 25)": // special case fdd000126
				hx = append(hx, tokens[1])
				hxx = append(hxx, 25)
			case "EBCDIC": // special case fdd000468 (skip for now)
			case "Byte 0": // skip fdd000585
			default:
				return hx, ascii, hxx, asciix, fmt.Errorf("loc: can't characterise signature (value: %v), unexpected label %s", v, tokens[0])
			}
		default:
			return hx, ascii, hxx, asciix, fmt.Errorf("loc: can't characterise signature (value: %v), unexpected token number %d", v, len(tokens))
		}
	}
	return hx, ascii, hxx, asciix, nil
}

func dehex(h string, off int) ([][]byte, []int, []bool, error) { // return bytes, offsets, and masks
	repl := strings.NewReplacer("0x", "", " ", "", "\n", "", "\t", "", "\r", "", "{8}", "xxxxxxxxxxxxxxxx", "{20 bytes of Hex 20}", "2020202020202020202020202020202020202020") // special case fdd000245 {8}, fdd000342 (nl tab within the hex)
	h = repl.Replace(h)
	if len(h)%2 != 0 {
		return nil, nil, nil, fmt.Errorf("loc: can't dehex %s", h)
	}
	h = strings.ToLower(h)
	var (
		idx   int
		byts  [][]byte = [][]byte{{}}
		offs  []int    = []int{0}
		masks []bool   = []bool{false}
	)
	for i := 0; i < len(h); i += 2 {
		switch {
		case h[i:i+2] == "xx":
			if off == 0 && i > 0 {
				idx++
				byts = append(byts, []byte{})
				offs = append(offs, 0)
				masks = append(masks, false)
			}
			off++
		case h[i] == 'x' || h[i+1] == 'x':
			if len(byts[idx]) > 0 {
				idx++
				byts = append(byts, []byte{})
				offs = append(offs, 0)
				masks = append(masks, false)
			}
			if off > 0 {
				offs[idx] = off
				off = 0
			}
			masks[idx] = true
			if h[i] == 'x' {
				byts[idx] = append(byts[idx], '0', h[i+1])
			} else {
				byts[idx] = append(byts[idx], h[i], '0')
			}
			if i+2 < len(h) {
				idx++
				byts = append(byts, []byte{})
				offs = append(offs, 0)
				masks = append(masks, false)
			}
		default:
			if off > 0 {
				offs[idx] = off
				off = 0
			}
			byts[idx] = append(byts[idx], h[i], h[i+1])
		}
	}
	for i, s := range byts {
		byt, err := hex.DecodeString(string(s))
		if err != nil {
			return nil, nil, nil, err
		}
		byts[i] = byt
	}
	return byts, offs, masks, nil
}
