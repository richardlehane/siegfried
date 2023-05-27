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
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/wikidata/internal/converter"
	"github.com/richardlehane/siegfried/pkg/wikidata/internal/mappings"

	"github.com/ross-spencer/wikiprov/pkg/spargo"
	"github.com/ross-spencer/wikiprov/pkg/wikiprov"
)

type wikidataMappings = map[string]mappings.Wikidata

// Alias our spargo Item for ease of referencing.
type wikidataItem = []map[string]spargo.Item

// Alias our wikiprov Provenance structure.
type wikiProv = []wikiprov.Provenance

// wikiItemProv helpfully collects item and provenance data.
type wikiItemProv struct {
	items wikidataItem
	prov  wikiProv
}

// Signature provides an alias for mappings.Signature for convenience.
type Signature = mappings.Signature

// Fields which are used in the Wikidata SPARQL query which we will
// access via JSON mapping.
const (
	uriField         = "uri"
	formatLabelField = "uriLabel"
	puidField        = "puid"
	locField         = "ldd"
	extField         = "extension"
	mimeField        = "mimetype"
	signatureField   = "sig"
	offsetField      = "offset"
	encodingField    = "encoding"
	relativityField  = "relativity"
	dateField        = "date"
	referenceField   = "referenceLabel"
)

// helper functions to control logging output
// Verbose output is default for roy but not when running tests etc.
func logf(format string, v ...any) {
	if config.Verbose() {
		log.Printf(format, v...)
	}
}
func logln(v ...any) {
	if config.Verbose() {
		log.Println(v...)
	}
}

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

// endpointJSON provides a helper to us to read the harvest results from a
// Wikibase and read the endpoint specifically. This is a special use feature
// of Roy/Siegfried and doesn't yet exist in the Wikiprov internals.
type endpointJSON struct {
	Endpoint string `json:"endpoint"`
}

// ErrNoEndpoint provides a method of validating the error received from
// this package when the custom SPARQL endpoint cannot be read from
// the harvest data.
var ErrNoEndpoint = errors.New("Endpoint in custom Wikibase sparql results not set")

// customEndpoint checks whether or not a custom "endpoint" is set in
// the Wikidata harvest results. If the endpoint doesn't match the default
// for Wikidata then we have a signal that we need to do more work to
// default endpoint then we need to do more work to make things run.
func customEndpoint(jsonFile []byte) (bool, error) {
	var endpoint endpointJSON
	err := json.Unmarshal([]byte(jsonFile), &endpoint)
	if err != nil {
		return false, fmt.Errorf(
			"Cannot parse JSON in Wikidata file: %w",
			err,
		)
	}
	if endpoint.Endpoint == "" {
		return false, fmt.Errorf("%w", ErrNoEndpoint)
	}
	if endpoint.Endpoint != config.WikidataEndpoint() {
		return true, nil
	}
	return false, nil
}

// openWikidata accesses the signatures definitions we harvested from
// Wikidata which are stored in SPARQL JSON and initiates their
// processing into the structures required by Roy to process into an
// identifier to be consumed by Siegfried.
func openWikidata() (wikiItemProv, error) {
	path := config.WikidataDefinitionsPath()
	logf("Roy (Wikidata): Opening Wikidata definitions: %s\n", path)
	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		return wikiItemProv{}, fmt.Errorf(
			"cannot open Wikidata file (check, or try harvest again): %w",
			err,
		)
	}
	custom, err := customEndpoint(jsonFile)
	if err != nil {
		return wikiItemProv{}, err
	}
	if custom {
		logln("Roy (Wikidata): Using a custom endpoint for results")
		err := setCustomWikibaseProperties()
		if err != nil {
			return wikiItemProv{}, fmt.Errorf("setting custom Wikibase properties: %w", err)
		}
		logf(
			"Roy (Wikidata): Custom PRONOM encoding loaded; config: '%s' => local: '%s'",
			config.WikibasePronom(),
			converter.GetPronomEncoding(),
		)
		logf(
			"Roy (Wikidata): Custom BOF loaded; config: '%s' => local: '%s'",
			config.WikibaseBOF(),
			relativeBOF,
		)
		logf(
			"Roy (Wikidata): Custom EOF loaded; config: '%s' => local: '%s'",
			config.WikibaseEOF(),
			relativeEOF,
		)
	}
	var sparqlReport spargo.WikiProv
	err = json.Unmarshal(jsonFile, &sparqlReport)
	if err != nil {
		return wikiItemProv{}, fmt.Errorf(
			"cannot open Wikidata file: %w",
			err,
		)
	}
	return wikiItemProv{
		items: sparqlReport.Binding.Bindings,
		prov:  sparqlReport.Provenance,
	}, nil
}

// processWikidata iterates over the Wikidata signature definitions and
// creates or updates records as it goes. The global wikidataMapping
// stores the Roy ready definitions to turn into an identifier. The
// summary data structure is returned to the caller so that it can be
// used to replay the results of processing, e.g. so the caller can
// access the stored linting results.
func processWikidata(itemProv wikiItemProv) (Summary, wikidataMappings) {
	var wikidataMapping = mappings.NewWikidata()
	var summary Summary
	var expectedRecordsWithSignatures = make(map[string]bool)
	for _, item := range itemProv.items {
		id := getID(item[uriField].Value)
		if item[signatureField].Value != "" {
			summary.SparqlRowsWithSigs++
			expectedRecordsWithSignatures[item[uriField].Value] = true
		}
		if wikidataMapping[id].ID == "" {
			okayToAdd := addSignatures(itemProv.items, id)
			wikidataMapping[id] = newRecord(item, itemProv.prov, okayToAdd)
		} else {
			wikidataMapping[id] =
				updateRecord(item, wikidataMapping[id])
		}
	}
	summary.AllSparqlResults = len(itemProv.items)
	summary.CondensedSparqlResults = len(wikidataMapping)
	summary.RecordsWithPotentialSignatures =
		len(expectedRecordsWithSignatures)
	return summary, wikidataMapping
}

// createMappingFromWikidata encapsulates the functions needed to load
// parse, and process the Wikidata records from our definitions file.
// After processing the summary results are output by Roy.
func createMappingFromWikidata() ([]wikidataRecord, error) {
	itemProv, err := openWikidata()
	if err != nil {
		return []wikidataRecord{}, err
	}
	summary, wikidataMapping := processWikidata(itemProv)
	analyseWikidataRecords(wikidataMapping, &summary)
	reportMapping := createReportMapping(wikidataMapping)
	// Log summary before leaving the function.
	logf("%s\n", summary)
	return reportMapping, nil
}

// createReportMapping iterates over our Wikidata records to return a
// mapping `reportMappings` that can later be used to map PRONOM
// signatures into the Wikidata identifier. reportMappings is used to
// map Wikidata identifiers to PRONOM so that PRONOM native patterns can
// be incorporated into the identifier when it is first created.
func createReportMapping(wikidataMapping wikidataMappings) []wikidataRecord {
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

// There are three sets of properties to lookup in a configuration
// file, those for PRONOM, BOF and EOF values.
const (
	pronomProp = "PronomProp"
	bofProp    = "BofProp"
	eofProp    = "EofProp"
)

// setCustomWikibaseProperties sets the properties needed by Roy to
// parse the results coming from a custom Wikibase endpoint.
func setCustomWikibaseProperties() error {
	logln("Roy (Wikidata): Looking for existence of wikibase.json in Siegfried home")
	wikibasePropsPath := config.WikibasePropsPath()
	propsFile, err := os.ReadFile(wikibasePropsPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf(
				"cannot find file '%s' in '%s': %w",
				wikibasePropsPath,
				config.WikidataHome(),
				err,
			)
		}

		return fmt.Errorf(
			"a different error handling '%s' has occurred: %w",
			wikibasePropsPath,
			err,
		)
	}
	propsMap := make(map[string]string)
	if err == nil {
		err = json.Unmarshal(propsFile, &propsMap)
		if err != nil {
			return err
		}
	}
	pronom := propsMap[pronomProp]
	bof := propsMap[bofProp]
	eof := propsMap[eofProp]
	// Set the properties globally in the config and then request they are
	// re-read from the module so that they are updated prior to building the
	// signature file.
	err = config.SetProps(pronom, bof, eof)
	if err != nil {
		return err
	}
	GetPronomURIFromConfig()
	GetBOFandEOFFromConfig()
	logf(
		"Roy (Wikidata): Properties set for PRONOM: '%s', BOF: '%s', EOF: '%s'",
		pronom,
		bof,
		eof,
	)
	return nil
}
