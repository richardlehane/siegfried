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

// Provides configuration structures and helpers for the Siegfried
// Wikidata functionality.

package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/richardlehane/siegfried/pkg/config/internal/wikidatasparql"
)

// Wikidata configuration fields. NB. In alphabetical order.
var wikidata = struct {
	// archive formats that Siegfried should be able to decompress via
	// the Wikidata identifier.
	arc    string
	arc1_1 string
	gzip   string
	tar    string
	warc   string
	// debug provides a way for users to output errors and warnings
	// associated with Wikidata records.
	debug bool
	// definitions stores the name of the file for the Wikidata
	// signature definitions. The definitions file is the raw SPARQL
	// output from Wikidata which will then be processed into an
	// identifier that can be consumed by Siegfried.
	definitions string
	// endpoint stores the URL of the SPARQL endpoint to pull
	// definitions from.
	endpoint string
	// filemode describes the file-mode we want to use to access the
	// Wikidata definitions file.
	filemode os.FileMode
	// namespace acts as a flag to tell us that we're using the Wikidata
	// identifier and describes and distinguishes it in reports.
	namespace string
	// nopronom determines whether the identifier will be build without
	// patterns from PRONOM sources outside of Wikidata.
	nopronom bool
	// propPronom should be the URL needed to resolve PRONOM encoded
	// signatures in the Wikidata identifier.
	propPronom string
	// propBOF should be the URL needed to resolve signatures to BOF
	// in the Wikidata identifier.
	propBOF string
	// propEOF should be the URL needed to resolve signatures to EOF
	// in the Wikidata identifier.
	propEOF string
	// revisionHistoryLen provides a way to configure the amount of
	// history returned from a Wikibase instance. More history will
	// slow down query time. Less history will speed it up.
	revisionHistoryLen int
	// revisionHistoryThreads provides a way to configure the number of
	// threads used to download revision history from a Wikibase instance.
	// Theoretically this value can speed up the Wikidata harvest process
	// but it isn't guaranteed.
	revisionHistoryThreads int
	// sparqlParam refers to the SPARQL parameter (?param) that returns
	// the QID for the record that we want to return revision history
	// and permalink for. E.g. ?uriLabl may return QID: Q12345. This
	// will then be used to query Wikibase for its revision history.
	// This should be the subject/IRI of the file format record in
	// Wikidata.
	sparqlParam string
	// wikidatahome describes the name of the wikidata directory withing
	// $SFHOME.
	wikidatahome string
	// wikibaseSparql is a placeholder for custom queries to be provided
	// to a custom Wikibase instance.
	wikibaseSparql string
	// wikibaseSparqlFile points to the wikibase.sparql file required to
	// query a custom Wikibase endpoint.
	wikibaseSparqlFile string
	// wikibasePropsFile is a JSON file that stores all the properties
	// needed for Roy to interpret a custom Wikibase SPARQL result.
	wikibasePropsFile string
	// wikibaseURL is the base URL needed by Wikibase for permalinks to
	// resolve and for revision history to be retrieved. The URL will be
	// augmented with /w/index.php or /w/api.php via Wikiprov (so do not
	// add this!). Wikiprov is the package used to make revision history
	// and permalinks happen.
	wikibaseURL string
}{
	arc:                    "Q7978505",
	arc1_1:                 "Q27824065",
	gzip:                   "Q27824060",
	tar:                    "Q283579",
	warc:                   "Q10287816",
	definitions:            "wikidata-definitions-3.0.0",
	endpoint:               "https://query.wikidata.org/sparql",
	filemode:               0644,
	propPronom:             "http://www.wikidata.org/entity/Q35432091",
	propBOF:                "http://www.wikidata.org/entity/Q35436009",
	propEOF:                "http://www.wikidata.org/entity/Q1148480",
	revisionHistoryLen:     5,
	revisionHistoryThreads: 10,
	sparqlParam:            "uri",
	wikidatahome:           "wikidata",
	wikibaseSparql:         "",
	wikibaseSparqlFile:     "wikibase.sparql",
	wikibasePropsFile:      "wikibase.json",
	wikibaseURL:            "https://www.wikidata.org/",
}

// WikidataHome describes where files needed by Siegfried and Roy for
// its Wikidata component resides.
func WikidataHome() string {
	return filepath.Join(siegfried.home, wikidata.wikidatahome)
}

// Namespace to be used in the Siegfried identification reports.
const wikidataNamespace = "wikidata"

// SetWikidataNamespace will set the Wikidata namespace. One reason
// this isn't set already is that Roy's idiom is to use it as a signal
// to say this identifier is ON/OFF and should be used, i.e. when
// this function is called, we want to use a Wikidata identifier.
func SetWikidataNamespace() func() private {
	return func() private {
		loc.fdd = ""     // reset loc to avoid pollution
		mimeinfo.mi = "" // reset mimeinfo to avoid pollution
		wikidata.namespace = wikidataNamespace
		return private{}
	}
}

// GetWikidataNamespace will return the Wikidata namespace field to the
// caller.
func GetWikidataNamespace() string {
	return wikidata.namespace
}

// SetWikidataDebug turns linting messages on when compiling the
// identifier
func SetWikidataDebug() func() private {
	wikidata.debug = true
	return SetWikidataNamespace()
}

// WikidataDebug will return the status of the debug flag, i.e.
// true for debug linting messages, false for none.
func WikidataDebug() bool {
	return wikidata.debug
}

// SetWikidataDefinitions is a setter to enable us to elect to use a
// different signature file name, e.g. as a setter during testing.
func SetWikidataDefinitions(definitions string) {
	wikidata.definitions = definitions
}

// WikidataDefinitionsFile returns the name of the file used to store
// the signature definitions.
func WikidataDefinitionsFile() string {
	return wikidata.definitions
}

// WikidataDefinitionsPath is a helper for convenience from callers to
// point directly at the definitions path for reading/writing as
// required.
func WikidataDefinitionsPath() string {
	return filepath.Join(WikidataHome(), WikidataDefinitionsFile())
}

// WikidataFileMode returns the file-mode required to save the
// definitions file.
func WikidataFileMode() os.FileMode {
	return wikidata.filemode
}

// SetWikidataEndpoint enables the use of another Wikibase instance if
// one is available. If there is an error with the URL then summary
// information will be returned to the caller and the default endpoint
// will be used.
func SetWikidataEndpoint(endpoint string) (func() private, error) {
	_, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return func() private { return private{} }, fmt.Errorf(
			"URL provided is invalid: '%w' default Wikidata Query Service will be used instead",
			err,
		)
	}
	wikidata.endpoint = endpoint
	return func() private {
		return private{}
	}, err
}

// WikidataEndpoint returns the SPARQL endpoint to call when harvesting
// Wikidata definitions.
func WikidataEndpoint() string {
	return wikidata.endpoint
}

// WikidataSPARQL returns the SPARQL query required to harvest Wikidata
// definitions.
func WikidataSPARQL() string {
	if wikidata.wikibaseSparql != "" {
		return wikidata.wikibaseSparql
	}
	return wikidatasparql.WikidataSPARQL()
}

// WikidataLang returns the language we want to return results in from
// Wikidata.
func WikidataLang() string {
	return wikidatasparql.WikidataLang()
}

// SetWikidataLang sets the language that we want to return results in
// from Wikidata. The default is en.
func SetWikidataLang(lang string) {
	wikidatasparql.SetWikidataLang(lang)
}

// SetWikidataNoPRONOM will turn native PRONOM patterns off in the final
// identifier output.
func SetWikidataNoPRONOM() func() private {
	return func() private {
		wikidata.nopronom = true
		return private{}
	}
}

// SetWikidataPRONOM will turn native PRONOM patterns on in the final
// identifier output.
func SetWikidataPRONOM() func() private {
	return func() private {
		wikidata.nopronom = false
		return private{}
	}
}

// GetWikidataNoPRONOM will tell the caller whether or not to use native
// PRONOM patterns inside the identifier.
func GetWikidataNoPRONOM() bool {
	return wikidata.nopronom
}

// SetWikibaseURL lets the default value for the Wikibase URL to be
// overridden. The URL should be that which enables permalinks to be
// returned from Wikibase, e.g. for Wikidata this URL needs to be:
//
// e.g. https://www.wikidata.org/w/index.php
//
func SetWikibaseURL(baseURL string) (func() private, error) {
	_, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return func() private { return private{} }, fmt.Errorf(
			"URL provided is invalid: '%w' default Wikibase URL be used instead but may not work",
			err,
		)
	}
	wikidata.wikibaseURL = baseURL
	return func() private {
		return private{}
	}, err
}

// WikidataWikibaseURL returns the SPARQL endpoint to call when harvesting
// Wikidata definitions.
func WikidataWikibaseURL() string {
	return wikidata.wikibaseURL
}

// WikidataSPARQLRevisionParam returns the SPARQL parameter (?param) that
// returns the QID for the record that we want to return revision
// history and permalink for. E.g. ?uriLabl may return QID: Q12345.
// This will then be used to query Wikibase for its revision history.
func WikidataSPARQLRevisionParam() string {
	return wikidata.sparqlParam
}

// GetWikidataRevisionHistoryLen will return the length of the Wikibase
// history to retrieve to the caller.
func GetWikidataRevisionHistoryLen() int {
	return wikidata.revisionHistoryLen
}

// GetWikidataRevisionHistoryThreads will return the number of threads
// to use to retrieve Wikibase history to the caller.
func GetWikidataRevisionHistoryThreads() int {
	return wikidata.revisionHistoryThreads
}

// WikibaseSparqlFile returns the file path expected for a custom
// Wikibase sparql query file. This file is needed to query a custom
// instance in the majority of cases. It is unlikely a host Wikibase
// will use the same configured properties and entities.
func WikibaseSparqlFile() string {
	return filepath.Join(WikidataHome(), wikidata.wikibaseSparqlFile)
}

// SetWikibaseSparql sets the SPARQL placeholder for custom Wikibase
// queries.
func SetWikibaseSparql(query string) func() private {
	wikidata.wikibaseSparql = query
	return func() private {
		return private{}
	}
}

// checkWikibaseURL guides the user to set the Wikibase/Wikimedia server
// URL when using a custom endpoint. This URL should be a valid
// Wikimedia URL. Roy will connect to the Wikimedia API via this URL.
func checkWikibaseURL(customEndpointURL string, customWikibaseURL string) error {
	if customWikibaseURL == WikidataWikibaseURL() {
		return fmt.Errorf(
			"Wikibase server URL for '%s' needs to be configured, can't be: '%s'",
			customEndpointURL,
			customWikibaseURL,
		)
	}
	return nil
}

// SetCustomWikibaseEndpoint sets a custom Wikibase endpoint if provided
// by the caller.
func SetCustomWikibaseEndpoint(customEndpointURL string, customWikibaseURL string) error {
	err := checkWikibaseURL(customEndpointURL, customWikibaseURL)
	if err != nil {
		return err
	}
	_, err = SetWikidataEndpoint(customEndpointURL)
	if err != nil {
		return err
	}
	_, err = SetWikibaseURL(customWikibaseURL)
	if err != nil {
		return err
	}
	return nil
}

// SetCustomWikibaseQuery checks for a custom query file and then sets
// the configuration to point to that file if it finds one.
func SetCustomWikibaseQuery() error {
	wikibaseSparqlPath := WikibaseSparqlFile()
	sparqlFile, err := os.ReadFile(wikibaseSparqlPath)
	if os.IsNotExist(err) {
		return fmt.Errorf(
			"Setting custom Wikibase SPARQL: cannot find file '%s' in '%s': %w",
			wikibaseSparqlPath,
			WikidataHome(),
			err,
		)
	}
	if err != nil {
		return fmt.Errorf(
			"Setting custom Wikibase SPARQL: unexpected error opening '%s' has occurred: %w",
			wikibaseSparqlPath,
			err,
		)
	}
	// Read the contents and assign to our query field in our config.
	SetWikibaseSparql(string(sparqlFile))
	return nil
}

// WikibasePropsPath returns the file path expected for the
// configuration needed to tell roy how to interpret results from a
// custom Wikibase query.
//
// Example:
//
//		{
//		   "PronomProp": "http://wikibase.example.com/entity/Q2",
//		   "BofProp": "http://wikibase.example.com/entity/Q3",
//		   "EofProp": "http://wikibase.example.com/entity/Q4"
//		}
//
func WikibasePropsPath() string {
	return filepath.Join(WikidataHome(), wikidata.wikibasePropsFile)
}

// SetWikibasePropsPath allows the WikidataPropsPath to be overwritten,
// e.g. for testing, and if it becomes needed, in the primary Roy code base.
func SetWikibasePropsPath(propsPath string) func() private {
	wikidata.wikibasePropsFile = propsPath
	return func() private {
		return private{}
	}
}

// WikibasePronom will return the configured PRONOM property from the
// Wikibase configuration.
func WikibasePronom() string {
	return wikidata.propPronom
}

// WikibaseBOF will return the configured BOF property from the Wikibase
// configuration.
func WikibaseBOF() string {
	return wikidata.propBOF
}

// WikibaseEOF will return the configured EOF property from the Wikibase
// configuration.
func WikibaseEOF() string {
	return wikidata.propEOF
}

// testPropURL provides a helper for testing the properties being set by
// SetProps.
func testPropURL(propURL string) error {
	_, err := url.ParseRequestURI(propURL)
	if err != nil {
		return fmt.Errorf(
			"URL provided '%s' is invalid: '%w' custom Wikibase instances need this value to be correct",
			propURL,
			err,
		)
	}
	return nil
}

// SetProps will set the three minimum properties needed to run Roy/SF
// with a custom Wikibase instance.
func SetProps(pronom string, bof string, eof string) error {
	if err := testPropURL(pronom); err != nil {
		return err
	}
	if err := testPropURL(bof); err != nil {
		return err
	}
	if err := testPropURL(eof); err != nil {
		return err
	}
	wikidata.propPronom = pronom
	wikidata.propBOF = bof
	wikidata.propEOF = eof
	return nil
}
