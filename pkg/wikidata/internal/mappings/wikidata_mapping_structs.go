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

// Package mappings provides data structures and helpers that describe
// Wikidata signature resources that we want to work with.
package mappings

import (
	"encoding/json"
	"fmt"
)

// Wikidata stores information about something which constitutes a
// format resource in Wikidata. I.e. Anything which has a URI and
// describes a file-format.
type Wikidata struct {
	ID                string      // Wikidata short name, e.g. Q12345 can be appended to a URI to be dereferenced.
	Name              string      // Name of the format as described in Wikidata.
	URI               string      // URI is the absolute URL in Wikidata terms that can be dereferenced.
	PRONOM            []string    // 1:1 mapping to PRONOM wherever possible.
	Extension         []string    // Extension returned by Wikidata.
	Mimetype          []string    // Mimetype as recorded by Wikidata.
	Signatures        []Signature // Signature associated with a record which we will convert to a new Type.
	Permalink         string      // Permalink associated with a record when the definitions were downloaded.
	RevisionHistory   string      // RevisionHistory is a human readable block of JSON for use in roy inspect functions.
	disableSignatures bool        // If a bad heuristic was found we can't reliably add signatures to the record.
}

// Signature describes a complete signature resource, i.e. a way to
// identify a file format using Wikidata information.
type Signature struct {
	ByteSequences []ByteSequence // A signature is made up of multiple byte sequences that encode a position and a pattern, e.g. BOF and EOF.
	Source        string         // Source (provenance) of the signature in Wikidata.
	Date          string         // Date the signature was submitted.
}

// ByteSequence describes a sequence that goes into a signature, where
// a signature is made up of 1..* sequences. Usually up to three.
type ByteSequence struct {
	Signature  string // Signature byte sequence.
	Offset     int    // Offset used by the signature.
	Encoding   int    // Signature encoding, e.g. Hexadecimal, ASCII, PRONOM.
	Relativity string // Position relative to beginning or end of file, or elsewhere.
}

// Serialize the signature component of our record to a string for
// debugging purposes.
func (signature Signature) String() string {
	report, err := json.MarshalIndent(signature, "", "  ")
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s", report)
}

// Serialize the byte sequence component of our record to a string for
// debugging purposes.
func (byteSequence ByteSequence) String() string {
	report, err := json.MarshalIndent(byteSequence, "", "  ")
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s", report)
}

// DisableSignatures is used when processing Wikidata records when a
// critical error is discovered with a record that needs to be looked
// into beyond what Roy can do for us.
func (wikidata *Wikidata) DisableSignatures() {
	wikidata.disableSignatures = true
}

// SignaturesDisabled tells us whether the signatures are disabled for
// a given record.
func (wikidata Wikidata) SignaturesDisabled() bool {
	return wikidata.disableSignatures
}

// PUIDs enables the Wikidata format records to be mapped to existing
// PRONOM records when run in PRONOM mode, i.e. not just with Wikidata
// signatures.
func (wikidata Wikidata) PUIDs() []string {
	var puids []string
	for _, puid := range wikidata.PRONOM {
		puids = append(puids, puid)
	}
	return puids
}

// NewWikidata creates new map for adding Wikidata records to.
func NewWikidata() map[string]Wikidata {
	return make(map[string]Wikidata)
}
