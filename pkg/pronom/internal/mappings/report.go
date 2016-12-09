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

package mappings

import (
	"encoding/xml"
	"strings"
)

// PRONOM Report

type Report struct {
	XMLName     xml.Name           `xml:"PRONOM-Report"`
	Id          int                `xml:"report_format_detail>FileFormat>FormatID"`
	Name        string             `xml:"report_format_detail>FileFormat>FormatName"`
	Version     string             `xml:"report_format_detail>FileFormat>FormatVersion"`
	Description string             `xml:"report_format_detail>FileFormat>FormatDescription"`
	Families    string             `xml:"report_format_detail>FileFormat>FormatFamilies"`
	Types       string             `xml:"report_format_detail>FileFormat>FormatTypes"`
	Identifiers []FormatIdentifier `xml:"report_format_detail>FileFormat>FileFormatIdentifier"`
	Signatures  []Signature        `xml:"report_format_detail>FileFormat>InternalSignature"`
	Extensions  []string           `xml:"report_format_detail>FileFormat>ExternalSignature>Signature"`
	Relations   []RelatedFormat    `xml:"report_format_detail>FileFormat>RelatedFormat"`
}

type Signature struct {
	ByteSequences []ByteSequence `xml:"ByteSequence"`
}

func (s Signature) String() string {
	var p string
	for i, v := range s.ByteSequences {
		if i > 0 {
			p += "\n"
		}
		p += v.String()
	}
	return p
}

type ByteSequence struct {
	Position    string `xml:"PositionType"`
	Offset      string
	MaxOffset   string
	IndirectLoc string `xml:"IndirectOffsetLocation"`
	IndirectLen string `xml:"IndirectOffsetLength"`
	Endianness  string
	Hex         string `xml:"ByteSequenceValue"`
}

func trim(label, s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	return label + ":" + s + " "
}

func (bs ByteSequence) String() string {
	return trim("Pos", bs.Position) + trim("Min", bs.Offset) + trim("Max", bs.MaxOffset) + trim("Hex", bs.Hex)
}

type RelatedFormat struct {
	Typ string `xml:"RelationshipType"`
	Id  int    `xml:"RelatedFormatID"`
}

func appendUniq(is []int, i int) []int {
	for _, v := range is {
		if i == v {
			return is
		}
	}
	return append(is, i)
}

func (r *Report) Superiors() []int {
	sups := []int{}
	for _, v := range r.Relations {
		if v.Typ == "Has lower priority than" || v.Typ == "Is supertype of" {
			sups = appendUniq(sups, v.Id)
		}
	}
	return sups
}

func (r *Report) Subordinates() []int {
	subs := []int{}
	for _, v := range r.Relations {
		if v.Typ == "Has priority over" || v.Typ == "Is subtype of" {
			subs = appendUniq(subs, v.Id)
		}
	}
	return subs
}

type FormatIdentifier struct {
	Typ string `xml:"IdentifierType"`
	Id  string `xml:"Identifier"`
}

func (r *Report) MIME() string {
	for _, v := range r.Identifiers {
		if v.Typ == "MIME" {
			return v.Id
		}
	}
	return ""
}

func (r *Report) Label(puid string) string {
	name, version := strings.TrimSpace(r.Name), strings.TrimSpace(r.Version)
	switch {
	case name == "" && version == "":
		return puid
	case name == "":
		return puid + " (" + version + ")"
	case version == "":
		return puid + " (" + name + ")"
	}
	return puid + " (" + name + " " + version + ")"
}
