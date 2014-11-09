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

import "encoding/xml"

// Container

type Container struct {
	XMLName             xml.Name             `xml:"ContainerSignatureMapping"`
	ContainerSignatures []ContainerSignature `xml:"ContainerSignatures>ContainerSignature"`
	FormatMappings      []FormatMapping      `xml:"FileFormatMappings>FileFormatMapping"`
	TriggerPuids        []TriggerPuid        `xml:"TriggerPuids>TriggerPuid"`
}

type ContainerSignature struct {
	Id            int    `xml:",attr"`
	ContainerType string `xml:",attr"`
	Description   string
	Files         []File `xml:"Files>File"`
}

type File struct {
	Path      string
	Signature InternalSignature `xml:"BinarySignatures>InternalSignatureCollection>InternalSignature"`
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
	RightFragment   Fragment `xml:"RightFragment"`
}

type Fragment struct {
	Value     string `xml:",chardata"`
	MinOffset string `xml:",attr"`
	MaxOffset string `xml:",attr"`
}

type FormatMapping struct {
	Id   int    `xml:"signatureId,attr"`
	Puid string `xml:",attr"`
}

type TriggerPuid struct {
	ContainerType string `xml:",attr"`
	Puid          string `xml:",attr"`
}
