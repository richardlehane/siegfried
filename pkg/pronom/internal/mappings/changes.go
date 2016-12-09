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

package mappings

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
)

type Releases struct {
	XMLName  xml.Name  `xml:"release_notes"`
	Releases []Release `xml:"release_note"`
}

type Release struct {
	ReleaseDate   string    `xml:"release_date"`
	SignatureName string    `xml:"signature_filename"`
	Outlines      []Outline `xml:"release_outline"`
}

type Outline struct {
	Typ   string `xml:"name,attr"`
	Puids []Puid `xml:"format>puid"`
}

type Puid struct {
	Typ string `xml:"type,attr"`
	Val string `xml:",chardata"`
}

type KeyVal struct {
	Key string
	Val []string
}

// OrderedMap define an ordered map
type OrderedMap []KeyVal

// Implement the json.Marshaler interface
func (omap OrderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{")
	for i, kv := range omap {
		if i != 0 {
			buf.WriteString(",")
		}
		// marshal key
		key, err := json.MarshalIndent(kv.Key, "", "  ")
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(":")
		// marshal value
		val, err := json.MarshalIndent(kv.Val, "", "  ")
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}
	buf.WriteString("}")
	return buf.Bytes(), nil
}
