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

package mappings

import "encoding/xml"

type MIMEInfo struct {
	XMLName   xml.Name   `xml:"mime-info"`
	MIMETypes []MIMEType `xml:"mime-type"`
}

type MIMEType struct {
	MIME  string `xml:"type,attr"`
	Globs []struct {
		Pattern string `xml:"pattern,attr"`
		Weight  string `xml:"weight,attr"`
	} `xml:"glob"`
	XMLPattern []struct {
		Local string `xml:"localName,attr"`
		NS    string `xml:"namespaceURI,attr"`
	} `xml:"root-XML"`
	Magic   []Magic `xml:"magic"`
	Aliases []struct {
		Alias string `xml:"type,attr"`
	} `xml:"alias"`
	SuperiorClasses []struct {
		SubClassOf string `xml:"type,attr"`
	} `xml:"sub-class-of"`
	Comment  []string `xml:"_comment"`
	Comments []string `xml:"comment"`
	Acronym  []string `xml:"acronym"`
	Superior bool     `xml:"-"`
}

type Magic struct {
	Matches  []Match `xml:"match"`
	Priority string  `xml:"priority,attr"`
}

type Match struct {
	Typ     string  `xml:"type,attr"`
	Offset  string  `xml:"offset,attr"`
	Value   string  `xml:"value,attr"`
	Mask    string  `xml:"mask,attr"`
	Matches []Match `xml:"match"`
}

// Some string methods just for debugging the mappings - delete once confirmed
type stringMaker [][]string

func (sm stringMaker) stringify() string {
	if len(sm) == 0 {
		return ""
	}
	str := "["
	for i, item := range sm {
		if i > 0 {
			str += " | "
		}
		for j, field := range item {
			if j > 0 && field != "" {
				str += "; "
			}
			str += field
		}
	}
	return str + "]"
}

func matchString(m Match) string {
	str := "{"
	str += "type:" + m.Typ
	str += ",off:" + m.Offset
	str += ",pat:" + m.Value
	if m.Mask != "" {
		str += ",mask:" + m.Mask
	}
	if len(m.Matches) > 0 {
		str += " ==> "
		for i, sub := range m.Matches {
			if i > 0 {
				str += " | "
			}
			str += matchString(sub)
		}
	}
	return str + "}"
}

func (m MIMEType) String() string {
	glob := make(stringMaker, len(m.Globs))
	for i, g := range m.Globs {
		glob[i] = []string{g.Pattern, g.Weight}
	}
	xmlPat := make(stringMaker, len(m.XMLPattern))
	for i, x := range m.XMLPattern {
		xmlPat[i] = []string{x.Local, x.NS}
	}
	var magic string
	if len(m.Magic) > 0 {
		magic = "["
		for i, mg := range m.Magic {
			if i > 0 {
				magic += " || "
			}
			if mg.Priority != "" {
				magic += "priority: " + mg.Priority + " "
			}
			for _, mt := range mg.Matches {
				magic += matchString(mt)
			}
		}
		magic += "]"
	}
	return m.MIME + " " + glob.stringify() + " " + xmlPat.stringify() + magic
}
