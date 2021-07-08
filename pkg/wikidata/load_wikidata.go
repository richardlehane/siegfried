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

// Accesses the harvested signature definitions from Wikidata and
// processes them into mappings structures which will be processed by
// Roy to create the identifier that will be consumed by Siegfried.

package wikidata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/wikidata/internal/mappings"

	"github.com/ross-spencer/spargo/pkg/spargo"
)

// Alias for the mappings.WikidataMapping structure so that it is easy
// to reference below.
var wikidataMapping = mappings.WikidataMapping

// Alias our spargo Item for ease of referencing.
type wikidataItem = []map[string]spargo.Item

// Signature provides an alias for mappings.Signature for convenience.
type Signature = mappings.Signature

// Fields which are used in the Wikidata SPARQL query which we will
// access via JSON mapping.
const uriField = "uri"
const formatLabelField = "uriLabel"
const puidField = "puid"
const locField = "ldd"
const extField = "extension"
const mimeField = "mimetype"
const signatureField = "sig"
const offsetField = "offset"
const encodingField = "encodingLabel"
const relativityField = "relativityLabel"
const dateField = "date"
const referenceField = "referenceLabel"

// getID returns the QID from the IRI of the record that we're
// processing.
func getID(wikidataURI string) string {
	splitURI := strings.Split(wikidataURI, "/")
	return splitURI[len(splitURI)-1]
}

// contains will look for the appearance of a string  item in slice of
// strings items.
func contains(items []string, item string) bool {
	for i := range items {
		if items[i] == item {
			return true
		}
	}
	return false
}

// openWikidata accesses the signatures definitions we harvested from
// Wikidata which are stored in SPARQL JSON and initiates their
// processing into the structures required by Roy to process into an
// identifier to be consumed by Siegfried.
func openWikidata() wikidataItem {
	path := config.WikidataDefinitionsPath()
	log.Printf(
		"Roy (Wikidata): Opening Wikidata definitions: %s\n", path,
	)
	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Errorf(
			"Roy (Wikidata): Cannot open Wikidata file: %s",
			err,
		)
	}
	var sparqlReport spargo.SPARQLResult
	err = json.Unmarshal(jsonFile, &sparqlReport)
	if err != nil {
		fmt.Errorf(
			"Roy (Wikidata): Cannot open Wikidata file: %s",
			err,
		)
	}
	results := sparqlReport.Results.Bindings
	return results
}

// processWikidata iterates over the Wikidata signature definitions and
// creates or updates records as it goes. The global wikidataMapping
// stores the Roy ready definitions to turn into an identifier. The
// summary data structure is returned to the caller so that it can be
// used to replay the results of processing, e.g. so the caller can
// access the stored linting results.
func processWikidata(results wikidataItem) Summary {
	var summary Summary
	var expectedRecordsWithSignatures = make(map[string]bool)
	for _, item := range results {
		id := getID(item[uriField].Value)
		if item[signatureField].Value != "" {
			summary.SparqlRowsWithSigs++
			expectedRecordsWithSignatures[item[uriField].Value] = true
		}
		if wikidataMapping[id].ID == "" {
			add := addSignatures(results, id)
			wikidataMapping[id] = newRecord(item, add)
		} else {
			wikidataMapping[id] =
				updateRecord(item, wikidataMapping[id])
		}
	}
	summary.AllSparqlResults = len(results)
	summary.CondensedSparqlResults = len(wikidataMapping)
	summary.RecordsWithPotentialSignatures =
		len(expectedRecordsWithSignatures)
	return summary
}

// createMappingFromWikidata encapsulates the functions needed to load
// parse, and process the Wikidata records from our definitions file.
// After processing the summary results are output by Roy.
func createMappingFromWikidata() []wikidataRecord {
	results := openWikidata()
	summary := processWikidata(results)
	analyseWikidataRecords(&summary)
	mapping := createReportMapping()
	// Output our summary before leaving the function. Consciously
	// output to stdout, but we may want to output to stderr.
	fmt.Println(summary)
	return mapping
}

// createReportMapping iterates over our Wikidata records to return a
// mapping `reportMappings` that can later be used to map PRONOM
// signatures into the Wikidata identifier. reportMappings is used to
// map Wikidata identifiers to PRONOM so that PRONOM native patterns can
// be incorporated into the identifier when it is first created.
func createReportMapping() []wikidataRecord {
	var reportMappings = []wikidataRecord{
		/* Examples:
		   {"Q12345", "PNG", "http://wikidata.org/q12345", "fmt/11", "png"},
		   {"Q23456", "FLAC", "http://wikidata.org/q23456", "fmt/279", "flac"},
		   {"Q34567", "ICO", "http://wikidata.org/q34567", "x-fmt/418", "ico"},
		   {"Q45678", "SIARD", "http://wikidata.org/q45678", "fmt/995", "siard"},
		*/
	}
	for _, wd := range wikidataMapping {
		reportMappings = append(reportMappings, wd)
	}
	return reportMappings
}
