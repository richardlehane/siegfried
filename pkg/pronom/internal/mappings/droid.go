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

// Package mappings contains struct mappings to unmarshal three
// different PRONOM XML formats: the signature file format, the report
// format, and the container format.
package mappings

import (
	"encoding/xml"
)

// Droid Signature File

// Droid describes the basic top-level structure of a DROID signature
// file.
type Droid struct {
	XMLName     xml.Name            `xml:"FFSignatureFile"`
	Version     int                 `xml:",attr"`
	Signatures  []InternalSignature `xml:"InternalSignatureCollection>InternalSignature"`
	FileFormats []FileFormat        `xml:"FileFormatCollection>FileFormat"`
}

// InternalSignature describes the structure of the InternalSignature
// section of a DROID signature file.
type InternalSignature struct {
	ID            int       `xml:"ID,attr"`
	ByteSequences []ByteSeq `xml:"ByteSequence"`
}

// ByteSeq describes the structure of the ByteSequence sections of a
// DROID signature file.
type ByteSeq struct {
	Reference    string        `xml:"Reference,attr"`
	SubSequences []SubSequence `xml:"SubSequence"`
}

// SubSequence describes the structure of the SubSequence sections of a
// DROID signature file.
type SubSequence struct {
	Position        int    `xml:",attr"`
	SubSeqMinOffset string `xml:",attr"` // and empty int values are unmarshalled to 0
	SubSeqMaxOffset string `xml:",attr"` // uses string rather than int because value might be empty
	Sequence        string
	LeftFragments   []Fragment `xml:"LeftFragment"`
	RightFragments  []Fragment `xml:"RightFragment"`
}

// Fragment describes the structure of the fragment sections of a DROID
// signature file.
type Fragment struct {
	Value     string `xml:",chardata"`
	MinOffset string `xml:",attr"`
	MaxOffset string `xml:",attr"`
	Position  int    `xml:",attr"`
}

// FileFormat describes the structure of the FileFormat section of a
// DROID signature file.
type FileFormat struct {
	XMLName    xml.Name `xml:"FileFormat"`
	ID         int      `xml:"ID,attr"`
	Puid       string   `xml:"PUID,attr"`
	Name       string   `xml:",attr"`
	Version    string   `xml:",attr"`
	MIMEType   string   `xml:",attr"`
	FormatType string   "xml:-"
	Extensions []string `xml:"Extension"`
	Signatures []int    `xml:"InternalSignatureID"`
	Priorities []int    `xml:"HasPriorityOverFileFormatID"`
}
