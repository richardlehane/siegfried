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
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/bytematcher/patterns"
	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/mimeinfo/internal/mappings"
)

func versions() []string {
	return nil
}

type mimeinfo struct {
	m []mappings.MIMEType
	identifier.Blank
}

func newMIMEInfo(path string) (identifier.Parseable, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	mi := &mappings.MIMEInfo{}
	err = xml.Unmarshal(buf, mi)
	if err != nil {
		return nil, err
	}
	index := make(map[string]int)
	errs := []string{}
	for i, v := range mi.MIMETypes {
		if _, ok := index[v.MIME]; ok {
			errs = append(errs, v.MIME)
		}
		index[v.MIME] = i
	}
	if len(errs) > 0 {
		return nil, errors.New("Can't parse mimeinfo file, duplicated IDs: " + strings.Join(errs, ", "))
	}
	for i, v := range mi.MIMETypes {
		if len(v.SuperiorClasses) == 1 && v.SuperiorClasses[0].SubClassOf != config.TextMIME() { // subclasses of text/plain shouldn't inherit text magic
			sup := index[v.SuperiorClasses[0].SubClassOf]
			if len(mi.MIMETypes[sup].XMLPattern) > 0 {
				mi.MIMETypes[i].XMLPattern = append(mi.MIMETypes[i].XMLPattern, mi.MIMETypes[sup].XMLPattern...)
			}
			if len(mi.MIMETypes[sup].Magic) > 0 {
				nm := make([]mappings.Magic, len(mi.MIMETypes[sup].Magic))
				copy(nm, mi.MIMETypes[sup].Magic)
				for i, w := range nm {
					if len(w.Priority) > 0 {
						num, err := strconv.Atoi(w.Priority)
						if err == nil {
							nm[i].Priority = strconv.Itoa(num - 1)
							continue
						}
					}
					nm[i].Priority = "49"
				}
				mi.MIMETypes[i].Magic = append(mi.MIMETypes[i].Magic, nm...)
			}
		}
	}
	return mimeinfo{mi.MIMETypes, identifier.Blank{}}, nil
}

func (mi mimeinfo) IDs() []string {
	ids := make([]string, len(mi.m))
	for i, v := range mi.m {
		ids[i] = v.MIME
	}
	return ids
}

type formatInfo struct {
	comment      string
	text         bool
	globWeights  []int
	magicWeights []int
}

func (f formatInfo) String() string {
	return f.comment
}

// turn generic FormatInfo into mimeinfo formatInfo
func infos(m map[string]identifier.FormatInfo) map[string]formatInfo {
	i := make(map[string]formatInfo, len(m))
	for k, v := range m {
		i[k] = v.(formatInfo)
	}
	return i
}

func textMIMES(m map[string]identifier.FormatInfo) []string {
	ret := make([]string, 1, len(m))
	ret[0] = config.TextMIME() // first one is the default
	for k, v := range m {
		if v.(formatInfo).text {
			ret = append(ret, k)
		}
	}
	return ret
}

func (mi mimeinfo) Infos() map[string]identifier.FormatInfo {
	fmap := make(map[string]identifier.FormatInfo, len(mi.m))
	for _, v := range mi.m {
		fi := formatInfo{}
		if len(v.Comment) > 0 {
			fi.comment = v.Comment[0]
		} else if len(v.Comments) > 0 {
			fi.comment = v.Comments[0]
		}
		var magicWeight int
		for _, mg := range v.Magic {
			magicWeight += len(mg.Matches)
		}
		fi.globWeights, fi.magicWeights = make([]int, len(v.Globs)), make([]int, 0, magicWeight)
		for i, w := range v.Globs {
			if len(w.Weight) > 0 {
				num, err := strconv.Atoi(w.Weight)
				if err == nil {
					fi.globWeights[i] = num
					continue
				}
			}
			fi.globWeights[i] = 50
		}
		for _, w := range v.Magic {
			weight := 50
			if len(w.Priority) > 0 {
				if num, err := strconv.Atoi(w.Priority); err == nil {
					weight = num
				}
			}
			for _, s := range w.Matches {
				ss, _ := toSigs(s)
				for _, sig := range ss {
					if sig != nil {
						fi.magicWeights = append(fi.magicWeights, weight)
					}
				}
			}
		}
		if len(v.SuperiorClasses) == 1 && v.SuperiorClasses[0].SubClassOf == config.TextMIME() {
			fi.text = true
		}
		fmap[v.MIME] = fi
	}
	return fmap
}

func (mi mimeinfo) Globs() ([]string, []string) {
	globs, ids := make([]string, 0, len(mi.m)), make([]string, 0, len(mi.m))
	for _, v := range mi.m {
		for _, w := range v.Globs {
			globs, ids = append(globs, w.Pattern), append(ids, v.MIME)
		}
	}
	return globs, ids
}

func (mi mimeinfo) MIMEs() ([]string, []string) {
	mimes, ids := make([]string, 0, len(mi.m)), make([]string, 0, len(mi.m))
	for _, v := range mi.m {
		mimes, ids = append(mimes, v.MIME), append(ids, v.MIME)
		for _, w := range v.Aliases {
			mimes, ids = append(mimes, w.Alias), append(ids, v.MIME)
		}
	}
	return mimes, ids
}

func (mi mimeinfo) Texts() []string {
	return textMIMES(mi.Infos())
}

// slice of root/NS
func (mi mimeinfo) XMLs() ([][2]string, []string) {
	xmls, ids := make([][2]string, 0, len(mi.m)), make([]string, 0, len(mi.m))
	for _, v := range mi.m {
		for _, w := range v.XMLPattern {
			xmls, ids = append(xmls, [2]string{w.Local, w.NS}), append(ids, v.MIME)
		}
	}
	return xmls, ids
}

func (mi mimeinfo) Signatures() ([]frames.Signature, []string, error) {
	var errs []error
	sigs, ids := make([]frames.Signature, 0, len(mi.m)), make([]string, 0, len(mi.m))
	for _, v := range mi.m {
		for _, w := range v.Magic {
			for _, s := range w.Matches {
				ss, err := toSigs(s)
				for _, sig := range ss {
					if sig != nil {
						sigs, ids = append(sigs, sig), append(ids, v.MIME)
					}
				}
				if err != nil {
					errs = append(errs, err)
				}
			}
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

func toSigs(m mappings.Match) ([]frames.Signature, error) {
	f, err := toFrames(m)
	if err != nil || f == nil {
		return nil, err
	}
	if len(m.Matches) == 0 {
		return []frames.Signature{frames.Signature(f)}, nil
	}
	subs := make([][]frames.Signature, 0, len(m.Matches))
	for _, m2 := range m.Matches {
		frs, err := toSigs(m2)
		if err != nil {
			return nil, err
		}
		if frs != nil {
			subs = append(subs, frs)
		}
	}
	var l, idx int
	for _, v := range subs {
		l += len(v)
	}
	ss := make([]frames.Signature, l)
	for _, v := range subs {
		for _, w := range v {
			ss[idx] = append(frames.Signature(f), w...)
			idx++
		}
	}
	return ss, nil
}

func toFrames(m mappings.Match) ([]frames.Frame, error) {
	pat, min, max, err := toPattern(m)
	if err != nil || pat == nil {
		return nil, err
	}
	mask, ok := pat.(Mask)
	if !ok {
		return []frames.Frame{frames.NewFrame(frames.BOF, pat, min, max)}, nil
	}
	pats, ints := unmask(mask)
	f := []frames.Frame{frames.NewFrame(frames.BOF, pats[0], min+ints[0], max+ints[0])}
	if len(pats) > 1 {
		for i, p := range pats[1:] {
			f = append(f, frames.NewFrame(frames.PREV, p, ints[i+1], ints[i+1]))
		}
	}
	return f, nil
}

func toPattern(m mappings.Match) (patterns.Pattern, int, int, error) {
	min, max, err := toOffset(m.Offset)
	if err != nil {
		return nil, min, max, err
	}
	var pat patterns.Pattern
	switch m.Typ {
	case "byte":
		i, err := strconv.ParseInt(m.Value, 0, 16)
		if err != nil {
			return nil, min, max, err
		}
		pat = Int8(i)
	case "big16":
		i, err := strconv.ParseInt(m.Value, 0, 32)
		if err != nil {
			return nil, min, max, err
		}
		pat = Big16(i)
	case "little16":
		i, err := strconv.ParseInt(m.Value, 0, 32)
		if err != nil {
			return nil, min, max, err
		}
		pat = Little16(i)
	case "host16":
		i, err := strconv.ParseInt(m.Value, 0, 32)
		if err != nil {
			return nil, min, max, err
		}
		pat = Host16(i)
	case "big32":
		i, err := strconv.ParseInt(m.Value, 0, 64)
		if err != nil {
			return nil, min, max, err
		}
		pat = Big32(i)
	case "little32":
		i, err := strconv.ParseInt(m.Value, 0, 64)
		if err != nil {
			return nil, min, max, err
		}
		pat = Little32(i)
	case "host32":
		i, err := strconv.ParseInt(m.Value, 0, 64)
		if err != nil {
			return nil, min, max, err
		}
		pat = Host32(i)
	case "string", "": // if no type given, assume string
		pat = patterns.Sequence(unquote(m.Value))
	case "stringignorecase":
		pat = IgnoreCase(unquote(m.Value))
	case "unicodeLE":
		uints := utf16.Encode([]rune(string(unquote(m.Value))))
		buf := make([]byte, len(uints)*2)
		for i, u := range uints {
			binary.LittleEndian.PutUint16(buf[i*2:], u)
		}
		pat = patterns.Sequence(buf)
	case "regex":
		return nil, min, max, nil // ignore regex magic
	default:
		return nil, min, max, errors.New("unknown magic type: " + m.Typ + " val: " + m.Value)
	}
	if len(m.Mask) > 0 {
		pat = Mask{pat, unquote(m.Mask)}
	}
	return pat, min, max, err
}

func toOffset(off string) (int, int, error) {
	var min, max int
	var err error
	if off == "" {
		return min, max, err
	}
	idx := strings.IndexByte(off, ':')
	switch {
	case idx < 0:
		min, err = strconv.Atoi(off)
		max = min
	case idx == 0:
		max, err = strconv.Atoi(off[1:])
	default:
		min, err = strconv.Atoi(off[:idx])
		if err == nil {
			max, err = strconv.Atoi(off[idx+1:])
		}
	}
	return min, max, err
}

var (
	rpl = strings.NewReplacer("\\ ", " ", "\\n", "\n", "\\t", "\t", "\\r", "\r", "\\b", "\b", "\\f", "\f", "\\v", "\v", "\\\\", "\\")
	rgx = regexp.MustCompile(`\\([0-9]{1,3}|x[0-9A-Fa-f]{1,2})`)
)

func numReplace(b []byte) []byte {
	var i uint64
	var err error
	if b[1] == 'x' {
		i, err = strconv.ParseUint(string(b[2:]), 16, 8)
	} else {
		// octal
		if len(b) == 4 {
			i, err = strconv.ParseUint(string(b[1:]), 8, 8)
		} else { // decimal
			i, err = strconv.ParseUint(string(b[1:]), 10, 8)
		}
	}
	if err != nil {
		panic(b)
	}
	return []byte{byte(i)}
}

func unquote(input string) []byte {
	// deal with hex first
	if len(input) > 2 && input[:2] == "0x" {
		h, err := hex.DecodeString(input[2:])
		if err == nil {
			return h
		} else {
			panic(input + " " + err.Error())
		}
	}
	return rgx.ReplaceAllFunc([]byte(rpl.Replace(input)), numReplace)
}
