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

// Structures and helpers for Wikidata processing and validation.

package wikidata

import (
	"encoding/json"
	"fmt"

	"github.com/richardlehane/siegfried/pkg/config"
)

// Summary of the identifier once processed.
type Summary struct {
	AllSparqlResults               int      // All rows of data returned from our SPARQL request.
	CondensedSparqlResults         int      // All unique records once the SPARQL is processed.
	SparqlRowsWithSigs             int      // All SPARQL rows with signatures (SPARQL necessarily returns duplicates).
	RecordsWithPotentialSignatures int      // Records that have signatures that can be processed.
	FormatsWithBadHeuristics       int      // Formats that have bad heuristics that we can't process.
	RecordsWithSignatures          int      // Records remaining that were processed.
	MultipleSequences              int      // Records that have been parsed out into multiple signatures per record.
	AllLintingMessages             []string // All linting messages returned.
	AllLintingMessageCount         int      // Count of all linting messages output.
	RecordCountWithLintingMessages int      // A count of the records that have linting messages to investigate.
}

// String will serialize the summary report as JSON to be printed.
func (summary Summary) String() string {
	report, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s", report)
}

// analyseWikidataRecords will parse the processed Wikidata mapping and
// populate the summary structure to enable us to report on the identifier.
func analyseWikidataRecords(wikidataMapping wikidataMappings, summary *Summary) {
	recordsWithLinting, allLinting, badHeuristics := countLintingErrors()
	summary.RecordCountWithLintingMessages = recordsWithLinting
	summary.AllLintingMessageCount = allLinting
	summary.FormatsWithBadHeuristics = badHeuristics
	for _, wd := range wikidataMapping {
		if len(wd.Signatures) > 0 {
			summary.RecordsWithSignatures++
		}
		for _, sigs := range wd.Signatures {
			if len(sigs.ByteSequences) > 1 {
				summary.MultipleSequences++
			}
		}
	}
	if config.WikidataDebug() {
		summary.AllLintingMessages = lintingToString()
	} else {
		const debugMessage = "Use the `-wikidataDebug` flag to build the identifier to see linting messages"
		summary.AllLintingMessages = []string{debugMessage}
	}
}
