package wikidata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/wikidata/internal/mappings"

	"github.com/ross-spencer/spargo/pkg/spargo"
)

const formatField = "format"
const puidField = "puid"
const locField = "ldd"
const extField = "extension"
const mimeField = "mimetype"

func getID(wikidataURI string) string {
	splitURI := strings.Split(wikidataURI, "/")
	return splitURI[len(splitURI)-1]
}

// Source strings from Wikidata. We may have a use-case to normalize them, but
// also, we can do this via ShEx on a long-enough time-line.
const (
	// Wikidata source strings are those returned by Wikidata specifically.
	prePronom  = "PRONOM"
	preKessler = "Gary Kessler's File Signature Table"

	// Normalized source strings are those that we want to return from the
	// Wikidata identifier to the user so that they can be parsed consistently
	// by the consumer.
	tnaPronom = "PRONOM (TNA)"
	wdPronom  = "PRONOM (Wikidata)"
	wdNone    = "Wikidata reference is empty"

	// Provenance to include in source information when the PRONOM signatures
	// are being used to compliment those in the Wikidata identifier.
	pronomOfficial          = "PRONOM (Official (%s))"
	pronomOfficialContainer = "PRONOM (Official container ID)"
)

func parseProvenance(p string) string {
	if p == prePronom {
		p = wdPronom
	}
	if p == "" {
		p = wdNone
	}
	return p
}

func newSignature(wdRecord map[string]spargo.Item) mappings.Signature {
	tmpWD := mappings.Signature{}
	tmpWD.Signature = wdRecord["sig"].Value
	tmpWD.Provenance = parseProvenance(wdRecord["referenceLabel"].Value)
	tmpWD.Date = wdRecord["date"].Value
	tmpWD.Encoding = wdRecord["encodingLabel"].Value
	tmpWD.Relativity = wdRecord["relativityLabel"].Value
	return tmpWD
}

// Create a newRecord with fields from the query sent to Wikidata.
//
//		"format"	<-- Wikidata URI.
//		"formatLabel"	<-- Format name.
//		"puid"	<-- PUID returned by Wikidata.
//		"extension"	<-- Format extension.
//		"mimetype"	<-- MimeType as recorded by Wikidata.
//
//		WIKIDATA TODO: Let's begin with a count of Wikidata signatures
//			  A format might have multiple signatures that can be used to
//			  match a record. Signatures might have multiple forms, e.g. Hex,
//			  or PRONOM regular expression.
//
//		"sig"	<-- Signature in Wikidata.
//		"referenceLabel"	<-- Signature provenance.
//		"date"	<-- Date the signature was submitted.
//		"encodingLabel"	<-- Encoding used for a Signature.
//		"offset"	<-- Offset relative to a position in a file.
//		"relativityLabel" 	<-- Direction from which to measure an offset for a signature.
//
func newRecord(wdRecord map[string]spargo.Item) mappings.Wikidata {
	sig := false
	if wdRecord["sig"].Value != "" {
		sig = true
	}
	wd := mappings.Wikidata{}

	wd.ID = getID(wdRecord["format"].Value)
	wd.Name = wdRecord["formatLabel"].Value
	wd.URI = wdRecord["format"].Value

	wd.PRONOM = append(wd.PRONOM, wdRecord["puid"].Value)
	wd.LOC = append(wd.LOC, wdRecord["ldd"].Value)
	wd.Extension = append(wd.Extension, wdRecord["extension"].Value)
	wd.Mimetype = append(wd.Mimetype, wdRecord["mimetype"].Value)

	if sig == true {
		wd.Signatures = append(wd.Signatures, newSignature(wdRecord))
	}

	return wd
}

func contains(items []string, item string) bool {
	for i := range items {
		if items[i] == item {
			return true
		}
	}
	return false
}

func updateSignatures(wd *mappings.Wikidata, wdRecord map[string]spargo.Item) {
	found := false
	for _, s := range wd.Signatures {
		if s.Signature == wdRecord["sig"].Value {
			found = true
		}
	}
	if found == false {
		wd.Signatures = append(wd.Signatures, newSignature(wdRecord))
	}
}

// A format record has some repeating properties. updateRecord manages those
// exceptions and adds them to the list if it doesn't already exist.
func updateRecord(wdRecord map[string]spargo.Item, wd mappings.Wikidata) mappings.Wikidata {
	if contains(wd.PRONOM, wdRecord[puidField].Value) == false {
		wd.PRONOM = append(wd.PRONOM, wdRecord[puidField].Value)
	}
	if contains(wd.LOC, wdRecord[locField].Value) == false {
		wd.LOC = append(wd.LOC, wdRecord[locField].Value)
	}
	if contains(wd.Extension, wdRecord[extField].Value) == false {
		wd.Extension = append(wd.Extension, wdRecord[extField].Value)
	}
	if contains(wd.Mimetype, wdRecord[mimeField].Value) == false {
		wd.Mimetype = append(wd.Mimetype, wdRecord[mimeField].Value)
	}
	if wdRecord["sig"].Value != "" {
		updateSignatures(&wd, wdRecord)
	}
	return wd
}

func countSignatures() int {
	var count int
	for _, wd := range mappings.WikidataMapping {
		if len(wd.Signatures) > 0 {
			count++
		}
	}
	return count
}

func openWikidata() {
	path := config.WikidataDefinitionsPath()
	fmt.Fprintf(os.Stderr, "Roy: Opening Wikidata report: %s\n", path)

	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Errorf("Roy: Cannot open Wikidata file: %s", err)
	}

	var sparqlReport spargo.SPARQLResult
	err = json.Unmarshal(jsonFile, &sparqlReport)
	if err != nil {
		fmt.Errorf("Roy: Cannot open Wikidata file: %s", err)
	}

	results := sparqlReport.Results.Bindings
	fmt.Fprintf(os.Stderr, "Roy: Original SPARQL results count: %d\n", len(results))
	for _, wdRecord := range results {
		id := getID(wdRecord[formatField].Value)
		if mappings.WikidataMapping[id].ID == "" {
			mappings.WikidataMapping[id] = newRecord(wdRecord)
		} else {
			mappings.WikidataMapping[id] = updateRecord(wdRecord, mappings.WikidataMapping[id])
		}
	}

	before := countSignatures()

	// Temporary: Remove this to bring in all Signatures.
	//
	// 1. For ease of management, deal with records with just one signature.
	// 2. Validate those signatures:
	//		* Currently validating for Hex only and assuming BOF.
	//		* ...
	//
	for k, wd := range mappings.WikidataMapping {
		if len(wd.Signatures) > 1 {
			fmt.Fprintf(
				os.Stderr,
				"Roy: Removing binary signatures for: %s (> 1 (%d))\n",
				wd.ID, len(wd.Signatures),
			)
			mappings.WikidataMapping[k] = mappings.DeleteSignatures(&wd)
		}
		if len(wd.Signatures) > 0 {
			for signature := range wd.Signatures {
				_, err := isHex(wd.Signatures[signature].Signature)
				if err != nil {
					fmt.Fprintf(
						os.Stderr,
						"Roy: Removing binary signatures for: %s (invalid hex: %s) (%s)\n",
						wd.ID, wd.Signatures[signature].Signature,
						err,
					)
					mappings.WikidataMapping[k] = mappings.DeleteSignatures(&wd)
				}
			}
		}
	}

	fmt.Fprintf(os.Stderr, "Roy: Condensed SPARQL results: %d\n", len(mappings.WikidataMapping))
	fmt.Fprintf(os.Stderr, "Roy: Number of anticipated signatures before: %d\n", before)
	fmt.Fprintf(os.Stderr, "Roy: Number of anticipated signatures after: %d\n", countSignatures())
	fmt.Fprintf(os.Stderr, "Roy: Report generation complete...\n")

	createReportMapping()
}

// WIKIDATA TODO: Another use for ShEX. We need to ensure the data quality is
// consistent so that it can be used reliably. Do we validate via Wikidata and
// do we do it in Roy too to be safe? What other strategies are in use in
// Siegfried.
func isHex(signature string) (bool, error) {
	if len(signature)%2 != 0 {
		return false, fmt.Errorf("Length of HEX is uneven")
	}
	for i := 0; i < len(signature); i += 2 {
		byte := signature[i : i+2]
		_, err := strconv.ParseUint(byte, 16, 64)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

// Basic mapping to load into newWikidata that we will then map to and return
// when PRONOM identifies the files...
var reportMappings = []mappings.Wikidata{
	/*
	   {"Q12345", "PNG", "http://wikidata.org/q12345", "fmt/11", "png"},
	   {"Q23456", "FLAC", "http://wikidata.org/q23456", "fmt/279", "flac"},
	   {"Q34567", "ICO", "http://wikidata.org/q34567", "x-fmt/418", "ico"},
	   {"Q45678", "SIARD", "http://wikidata.org/q45678", "fmt/995", "siard"},
	*/
}

func createReportMapping() {
	for _, wd := range mappings.WikidataMapping {
		reportMappings = append(reportMappings, wd)
	}

}
