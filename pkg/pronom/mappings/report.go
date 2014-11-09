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
	XMLName     xml.Name        `xml:"PRONOM-Report"`
	Description string          `xml:"report_format_detail>FileFormat>FormatDescription"`
	Signatures  []Signature     `xml:"report_format_detail>FileFormat>InternalSignature"`
	Relations   []RelatedFormat `xml:"report_format_detail>FileFormat>RelatedFormat"`
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
	return trim("Pos", bs.Position) + trim("Off", bs.Offset) + trim("Max", bs.MaxOffset) + trim("Hex", bs.Hex)
}

type RelatedFormat struct {
	Type string `xml:"RelationshipType"`
	ID   int    `xml:"RelatedFormatID"`
}
