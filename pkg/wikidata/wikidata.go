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

// Package wikidata contains the majority of the functions needed to
// build a Wikidata identifier (compiled signature file) compatible with
// Siegfried. Package Wikidata then also contains the majority of the
// functions required to enable Siegfried to consume that same
// identifier. The ability to do this is enabled by implementing
// Siegfried's Identifier and Parseable interfaces.
package wikidata

import (
	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

// wikidataDefinitions contains the file format information retrieved
// from the Wikidata definitions file, e.g. the name, URI, extensions,
// PUIDs that are associated with a Wikidata record. The structure also
// contains an implementation of parseable which we will attempt to
// satisfy, and a default implementation of parseable to take care of
// parts of the interface we don't complete (I think!).
type wikidataDefinitions struct {
	formats   []wikidataRecord
	parseable identifier.Parseable
	identifier.Blank
}

// newWikidata will call the functions required to load Wikidata
// definitions from disk and parse them into an identifier compatible
// structure. newWikidata will also add the data needed to also use
// native PRONOM identification patterns before finally collecting a
// list of PUIDs to be used in constructing provenance about each /
// signature.
func newWikidata() (identifier.Parseable, []string, error) {
	var puids []string
	var wikiParseable identifier.Parseable = identifier.Blank{}
	var err error
	// Process Wikidata report from disk and read into a report
	// structure.
	reportMappings, err := createMappingFromWikidata()
	if err != nil {
		return nil, []string{}, err
	}
	if config.GetWikidataNoPRONOM() {
		logln(
			"Roy (Wikidata): Not building identifiers set from PRONOM",
		)
	} else {
		logln(
			"Roy (Wikidata): Building identifiers set from PRONOM",
		)
		wikiParseable, err = pronom.NewPronom()
		if err != nil {
			return nil, []string{}, err
		}
		// Collect the PRONOM identifiers we want to work with in this
		// identifier for use in generating provenance that will be
		// displayed in the source field.
		_, puids, _ = wikiParseable.Signatures()
	}
	return wikidataDefinitions{
		reportMappings,     // Wikidata formats.
		wikiParseable,      // Implementation of Parseable.
		identifier.Blank{}, // Blank Parseable implementation.
	}, puids, nil
}
