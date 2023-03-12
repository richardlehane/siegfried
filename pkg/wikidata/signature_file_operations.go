// Copyright 2020 Ross Spencer, Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

// Satisfies the Identifier interface: functions responsible for PERSIST
// (saving and loading data structures to the Siegfried signature file).
// Also creates the structures we are going to use to inspect the
// signature file which also get converted to Siegfried result sets,
// including provenance and revision history.

package wikidata

import (
	"fmt"
	"strings"
	"time"

	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"

	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/internal/persist"

	"github.com/richardlehane/siegfried/pkg/wikidata/internal/mappings"
)

// Alias for the FormatInfo interface in Parseable to make it easier to
// reference.
type parseableFormatInfo = map[string]identifier.FormatInfo

// Save will write a Wikidata identifier to the Siegfried signature
// file using the persist package to save primitives in the identifier's
// data structure.
func (i *Identifier) Save(ls *persist.LoadSaver) {

	// Save the Wikidata magic enum from core.
	ls.SaveByte(core.Wikidata)

	// Save the no. formatInfo entries to read.
	ls.SaveSmallInt(len(i.infos))

	// Save the information in the formatInfo records.
	for idx, value := range i.infos {
		ls.SaveString(idx)
		ls.SaveString(value.name)
		ls.SaveString(value.uri)
		ls.SaveString(value.mime)
		ls.SaveStrings(value.sources)
		ls.SaveString(value.permalink)
		ls.SaveString(value.revisionHistory)
	}
	i.Base.Save(ls)
}

// Load back into memory from the signature file the same information
// that we wrote to the file using Save().
func Load(ls *persist.LoadSaver) core.Identifier {
	i := &Identifier{}
	le := ls.LoadSmallInt()
	i.infos = make(map[string]formatInfo)
	for j := 0; j < le; j++ {
		i.infos[ls.LoadString()] = formatInfo{
			ls.LoadString(),  // name.
			ls.LoadString(),  // URI.
			ls.LoadString(),  // mime.
			ls.LoadStrings(), // sources.
			ls.LoadString(),  // permalink.
			ls.LoadString(),  // revision history.
		}
	}
	i.Base = identifier.Load(ls)
	return i
}

// formatInfo can hold absolutely anything and can be used to return
// that information to the user. So we could also map to PUID or LoC
// identifier here if there was a strong link. Other information might
// exist in Wikidata (does exist in Wikidata) that we might want to map
// here, e.g. file formats capable of rendering the identified file.
type formatInfo struct {
	// name is the Name as retrieved from Wikidata. Name usually
	// incorporates version too in Wikidata which is why there isn't a
	// separate field for that.
	name string
	// uri is Wikidata IRI, e.g. http://www.wikidata.org/entity/Q1069215
	uri string
	// mime is a semi-colon separated list of the MIMETypes which are
	// also associated with a file format.
	mime string
	// sources describes the source of a signature retrieved from
	// Wikidata.
	sources []string
	// permalink refers to the Wikibase permalink for a Wikidata record.
	// The data at the permalink represents the specific version of the
	// record used to derive the information used by Siegfried, e.g.
	// signature definition.
	permalink string
	// revisionHistory refers to a bigger chunk of JSON which can be
	// displayed to a user to describe the history of a format
	// definition.
	revisionHistory string
}

// infos turns the generic formatInfo into the structure that will be
// written into the Siegfried identifier.
func infos(formatInfoMap parseableFormatInfo) map[string]formatInfo {
	idx := make(map[string]formatInfo, len(formatInfoMap))
	for key, value := range formatInfoMap {
		idx[key] = value.(formatInfo)
	}
	return idx
}

// Serialize formatInfo as a string for the roy --inspect function and
// debugging.
func (f formatInfo) String() string {
	sources := ""
	if len(f.sources) > 0 {
		sources = strings.Join(f.sources, " ")
	}
	return fmt.Sprintf(
		"Name: '%s'\nMIMEType: '%s'\nSources: '%s' \nRevision History: %s\n---",
		f.name,
		f.mime,
		sources,
		f.revisionHistory,
	)
}

// Infos arranges summary information about formats within an Identifier
// into a structure suitable for output in a Siegfried signature file.
//
// Infos provides a mechanism for a placing any other information about
// formats that you'd like to talk about in an identifier.
func (wdd wikidataDefinitions) Infos() parseableFormatInfo {
	logf(
		"Roy (Wikidata): In Infos()... length formats: '%d' no-pronom: '%t'\n",
		len(wdd.formats),
		config.GetWikidataNoPRONOM(),
	)
	formatInfoMap := make(
		map[string]identifier.FormatInfo, len(wdd.formats),
	)
	for _, value := range wdd.formats {
		var mime = value.Mimetype[0]
		if len(value.Mimetype) > 1 {
			// Q24907733 is a good example with mimes:
			//
			// `image/heif-sequence`; `image/heif`;
			// `image/heic-sequence`; `image/heic...`
			//
			for idx, value := range value.Mimetype {
				if idx == 0 {
					continue
				}
				mime = fmt.Sprintf("%s; %s", mime, value)
			}
		}
		sources := prepareSources(value)
		fi := formatInfo{
			name:            value.Name,
			uri:             value.URI,
			mime:            mime,
			sources:         sources,
			permalink:       value.Permalink,
			revisionHistory: value.RevisionHistory,
		}
		formatInfoMap[value.ID] = fi
	}
	return formatInfoMap
}

// prepareSources prepares a slice of sources that will be used to
// return some sort of source information (provenance of datum in
// Wikidata) for positive matches returned by the Wikidata identifier.
// We need to return a slice here as order of processing is important
// and matches the order which they will be processed into the
// identifier in the other identifier functions.
//
// We also want to take into account the native PRONOM sources here.
//
// What is also strange about this function is that it happens before
// Parseable processes the signatures for this identifier and so there
// is a potential for things to go wrong. If we start seeing issues in
// Parseable with the data, we might also need to consider how we
// interact with signature sources to modify them on the fly.
//
// We currently return a slice of strings but in-time there might well
// be value in using a struct here. The struct could then encode
// "source" information in different fields and could in future encode
// information of increasing complexity. The structure would also need
// to be compatible with Siegfried's/Roy's persist package.
func prepareSources(wdMapping mappings.Wikidata) []string {

	// Output the source date consistently.
	const provDateFormat = "2006-01-02"

	sources := []string{}

	// We need at least one signature to write a source for.
	if len(wdMapping.Signatures) > 0 {
		// Records like MACH-0 (Q2627217) are good examples of records
		// with multiple signatures that can potentially have different
		// sources.
		for idx := range wdMapping.Signatures {
			prov := wdMapping.Signatures[idx].Source
			date := wdMapping.Signatures[idx].Date
			if date != "" {
				date, _ := time.Parse(time.RFC3339, date)
				prov = fmt.Sprintf("%s (source date: %s)", prov, date.Format(provDateFormat))
			}
			sources = append(sources, prov)

		}
	}
	if !config.GetWikidataNoPRONOM() {
		// Bring PRONOM sources into the identifier.
		for _, exid := range wdMapping.PRONOM {
			// We have PRONOM identifiers to work with so we need to extend the
			// slice further.
			for _, value := range sourcePuids {
				if exid == value {
					sources = append(sources, fmt.Sprintf(pronomOfficial, exid))
				}
			}
		}
	}
	return sources
}
