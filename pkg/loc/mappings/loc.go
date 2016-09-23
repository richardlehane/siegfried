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
	"fmt"
	"strings"
)

type FDD struct {
	ID         string     `xml:"id,attr"`
	Name       string     `xml:"titleName,attr"`
	LongName   string     `xml:"identificationAndDescription>fullName"`
	Extensions []string   `xml:"fileTypeSignifiers>signifiersGroup>filenameExtension>sigValues>sigValue"`
	MIMEs      []string   `xml:"fileTypeSignifiers>signifiersGroup>internetMediaType>sigValues>sigValue"`
	Magics     []string   `xml:"fileTypeSignifiers>signifiersGroup>magicNumbers>sigValues>sigValue"`
	Others     []Other    `xml:"fileTypeSignifiers>signifiersGroup>other"`
	Relations  []Relation `xml:"identificationAndDescription>relationships>relationship"`
	Updates    []string   `xml:"properties>updates>date"`
	Links      []string   `xml:"usefulReferences>urls>url>urlReference>link"`
}

type Other struct {
	Tag    string   `xml:"tag"`
	Values []string `xml:"values>sigValues>sigValue"`
}

func (o Other) String() string {
	return fmt.Sprintf("[tag: %s; vals: %s]", o.Tag, strings.Join(o.Values, ","))
}

func ostr(os []Other) []string {
	ret := make([]string, len(os))
	for i, v := range os {
		ret[i] = v.String()
	}
	return ret
}

type Relation struct {
	Typ   string `xml:"typeOfRelationship"`
	Value string `xml:"relatedTo>id"`
}

func (r Relation) String() string {
	return fmt.Sprintf("[typ: %s; val: %s]", r.Typ, r.Value)
}

func rstr(rs []Relation) []string {
	ret := make([]string, len(rs))
	for i, v := range rs {
		ret[i] = v.String()
	}
	return ret
}

func (f FDD) String() string {
	return fmt.Sprintf("ID: %s\nName: %s\nLong Name: %s\nExts: %s\nMIMEs: %s\nMagics: %s\nOthers: %s\nRelations: %s\nPUIDs: %s",
		f.ID,
		f.Name,
		f.LongName,
		strings.Join(f.Extensions, ", "),
		strings.Join(f.MIMEs, ", "),
		strings.Join(f.Magics, ", "),
		strings.Join(ostr(f.Others), ", "),
		strings.Join(rstr(f.Relations), ", "),
		strings.Join(f.PUIDs(), ", "),
	)
}

func (f FDD) PUIDs() []string {
	var puids []string
	for _, l := range f.Links {
		if strings.HasPrefix(l, "http://apps.nationalarchives.gov.uk/pronom/") {
			puids = append(puids, strings.TrimPrefix(l, "http://apps.nationalarchives.gov.uk/pronom/"))
		} else if strings.HasPrefix(l, "http://www.nationalarchives.gov.uk/pronom/") {
			puids = append(puids, strings.TrimPrefix(l, "http://www.nationalarchives.gov.uk/pronom/"))
		}
	}
	return puids
}
