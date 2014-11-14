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

// This file contains struct mappings to unmarshal three different PRONOM XML formats: the signature file format, the report format, and the container format
package mappings

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// Droid Signature File

type Droid struct {
	XMLName     xml.Name     `xml:"FFSignatureFile"`
	Version     int          `xml:",attr"`
	Signatures  []Signature  `xml:"InternalSignatureCollection>InternalSignature"`
	FileFormats []FileFormat `xml:"FileFormatCollection>FileFormat"`
}

func (d Droid) String() string {
	buf := new(bytes.Buffer)
	for _, v := range d.FileFormats {
		fmt.Fprintln(buf, v)
	}
	return buf.String()
}

type InternalSignature struct {
	Id            int       `xml:"ID,attr"`
	Specificity   string    `xml:",attr"`
	ByteSequences []ByteSeq `xml:"ByteSequence"`
}

type ByteSeq struct {
	Reference    string        `xml:"Reference,attr"`
	SubSequences []SubSequence `xml:"SubSequence"`
}

type SubSequence struct {
	Position        int    `xml:",attr"`
	SubSeqMinOffset string `xml:",attr"` // and empty int values are unmarshalled to 0
	SubSeqMaxOffset string `xml:",attr"` // uses string rather than int because value might be empty
	Sequence        string
	LeftFragments   []Fragment `xml:"LeftFragment"`
	RightFragments  []Fragment `xml:"RightFragment"`
}

type Fragment struct {
	Value     string `xml:",chardata"`
	MinOffset string `xml:",attr"`
	MaxOffset string `xml:",attr"`
	Position  int    `xml:",attr"`
}

type FileFormat struct {
	XMLName    xml.Name `xml:"FileFormat"`
	Id         int      `xml:",attr"`
	Puid       string   `xml:"PUID,attr"`
	Name       string   `xml:",attr"`
	Version    string   `xml:",attr"`
	MIMEType   string   `xml:",attr"`
	Extensions []string `xml:"Extension"`
	Priorities []int    `xml:"HasPriorityOverFileFormatID"`
}

func (f FileFormat) String() string {
	null := func(s string) string {
		if strings.TrimSpace(s) == "" {
			return "NULL"
		}
		return s
	}
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "Puid: %s; Name: %s; Version: %s; Ext(s): %s\n", f.Puid, f.Name, f.Version, strings.Join(f.Extensions, ", "))
	fmt.Fprintln(buf, f.Description)
	return buf.String()
}
