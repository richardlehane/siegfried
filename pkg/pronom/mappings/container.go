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
	Signature InternalSignature `xml:"BinarySignatures>InternalSignatureCollection>InternalSignature"` // see Droid mapping file
}

type FormatMapping struct {
	Id   int    `xml:"signatureId,attr"`
	Puid string `xml:",attr"`
}

type TriggerPuid struct {
	ContainerType string `xml:",attr"`
	Puid          string `xml:",attr"`
}

func (c *Container) Puids() []string {
	if c == nil {
		return []string{}
	}
	ids := make([]int, len(c.ContainerSignatures))
	for i, v := range c.ContainerSignatures {
		ids[i] = v.Id
	}
	hasId := func(id int) bool {
		for _, v := range ids {
			if id == v {
				return true
			}
		}
		return false
	}
	puids := make([]string, 0, len(c.FormatMappings))
	addPuid := func(p string) {
		for _, v := range puids {
			if v == p {
				return
			}
		}
		puids = append(puids, p)
	}
	for _, v := range c.FormatMappings {
		if hasId(v.Id) {
			addPuid(v.Puid)
		}
	}
	return puids
}
