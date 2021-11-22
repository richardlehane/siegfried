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

// Perform some linting on the SPARQL we receive from Wikidata. This is
// all preliminary stuff where we will still need to wrangle the
// signatures to be useful in aggregate. Using that as a rule then we
// only do enough work here to make that wrangling a bit easier later
// on.

package wikidata

import (
	"fmt"
	"strconv"

	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/wikidata/internal/converter"
)

// Create a map to store linting results per Wikidata URI.
var linter = make(map[string]map[lintingResult]bool)

// lintingResult provides a data structure to store information about
// errors encountered while trying to process Wikidata records.
type lintingResult struct {
	URI      string  // URI of the Wikidata record.
	Value    linting // Linting error.
	Critical bool    // Critical, true or false.
}

// addLinting adds linting errors to our linter map when the function
// is called.
func addLinting(uri string, value linting) {
	if value == nle {
		return
	}
	critical := false
	switch value {
	case offWDE02:
	case relWDE02:
	case heuWDE01:
		critical = true
	}
	linting := lintingResult{}
	linting.URI = uri
	linting.Value = value
	linting.Critical = critical
	if linter[uri] == nil {
		lMap := make(map[lintingResult]bool)
		lMap[linting] = critical
		linter[uri] = lMap
		return
	}
	linter[uri][linting] = critical
}

// lintingToString will output our linting errors in an easy to consume
// slice.
func lintingToString() []string {
	var lintingMessages []string
	for _, result := range linter {
		for res := range result {
			s := fmt.Sprintf(
				"%s: URI: %s Critical: %t", lintingLookup(res.Value),
				res.URI,
				res.Critical,
			)
			lintingMessages = append(lintingMessages, s)
		}
	}
	return lintingMessages
}

// countLintingErrors will count all the linting errors returned during
// processing. It will return two counts, that of all the records with
// at least one error, and that of all the individual errors.
func countLintingErrors() (int, int, int) {
	var recordCount, individualCount, badHeuristicCount int
	for _, result := range linter {
		recordCount++
		for res := range result {
			if res.Value == heuWDE01 || res.Value == heuWDE02 {
				badHeuristicCount++
			}
			individualCount++
		}
	}
	return recordCount, individualCount, badHeuristicCount
}

type linting int

// nle provides a nil for no lint errors.
const nle = noLintingError

// Linting enumerator. This approach feels like it might be a little
// old fashioned but it lets us capture as many of the data issues we're
// seeing in Wikidata as they come up so that they can be fixed. Once
// we can find better control of the source data I think we'll be able
// to get rid of this and use a much simpler approach for compiling
// the set of signatures for the identifier.
const (
	noLintingError linting = iota // noLintingError encodes No linting error.

	// Offset based linting issues.
	offWDE01 // offWDE01 encodes ErrNoOffset
	offWDE02 // offWDE02 encodes ErrCannotParseOffset
	offWDE03 // offWDE03 encodes ErrBlankNodeOffset

	// Relativity based linting issues.
	relWDE01 // relWDE01 encodes ErrEmptyStringRelativity
	relWDE02 // relWDE02 encodes ErrUnknownRelativity

	// Encoding based linting issues.
	encWDE01 // encWDE01 encodes ErrNoEncoding

	// Provenance based linting issues.
	proWDE01 // proWDE01 encodes ErrNoProvenance
	proWDE02 // proWDE02 encodes ErrNoDate

	// Sequence based linting issues.
	seqWDE01 // seqWDE01 encodes ErrDuplicateSequence

	// Heuristic errors. We have to give up on this record.
	heuWDE01 // heuWDE01 encodes ErrNoHeuristic
	heuWDE02 // heuWDE02 encodes ErrCannotProcessSequence
)

// lintingLookup returns a plain-text string for the type of errors or
// issues that we encounter when trying to process Wikidata records
// into an identifier.
func lintingLookup(lint linting) string {
	switch lint {
	case offWDE01:
		return "Linting: WARNING no offset"
	case offWDE02:
		return "Linting: ERROR cannot parse offset"
	case offWDE03:
		return "Linting: ERROR blank node returned for offset"
	case relWDE01:
		return "Linting: WARNING no relativity"
	case relWDE02:
		return "Linting: ERROR unknown relativity"
	case encWDE01:
		return "Linting: WARNING no encoding"
	case seqWDE01:
		return "Linting: ERROR duplicate sequence"
	case proWDE01:
		return "Linting: WARNING no provenance"
	case proWDE02:
		return "Linting: WARNING no provenance date"
	case heuWDE01:
		return "Linting: ERROR bad heuristic"
	case heuWDE02:
		return "Linting: ERROR cannot process sequence"
	case noLintingError:
		return "Linting: INFO no linting errors"
	}
	return "Linting: ERROR unknown linting error"
}

// preProcessedSequence gives us a way to hold temporary information
// about the signature associated with a record.
type preProcessedSequence struct {
	signature  string
	offset     string
	relativity string
	encoding   string
}

// Relativities as encoded in Wikidata records. IRIs from Wikidata mean
// that we don't need to encode i18n differences. IRIs must have
// http:// scheme, and link to the data entity, i.e. not the "page",
// e.g.
//
//    * BOF data entity: http://www.wikidata.org/entity/Q35436009
//    * BOF page: https://www.wikidata.org/wiki/Q35436009
//

var relativeBOF string = config.WikibaseBOF()
var relativeEOF string = config.WikibaseEOF()

// GetBOFandEOFFromConfig will read the current value of the BOF/EOF
// properties from the configuration, e.g. after being updated using a
// custom SPARQL query.
func GetBOFandEOFFromConfig() {
	relativeBOF = config.WikibaseBOF()
	relativeEOF = config.WikibaseEOF()
}

// GetPronomURIFromConfig will read the current value of the PRONOM
// properties from the configuration, e.g. after being updated using a
// custom SPARQL query.
func GetPronomURIFromConfig() {
	converter.GetPronomURIFromConfig()
}

// validateAndReturnProvenance performs some arbitrary validation on
// provenance as recorded by Wikidata and let's us know any issues
// with it. Right now we can only really say if the provenance field
// is empty, it's not going to be very useful to us.
func validateAndReturnProvenance(value string) (string, linting) {
	if value == "" {
		return value, proWDE01
	}
	return value, nle
}

// validateAndReturnDate will perform some validation on the provenance
// date we are able to access from Wikidata records. If the value is
// blank for example, it will return a linting warning.
func validateAndReturnDate(value string) (string, linting) {
	if value == "" {
		return value, proWDE02
	}
	return value, nle
}

// validateAndReturnEncoding asks whether the encoding we can access
// from Wikidata is known to Siegfried. If it isn't then we know for
// now that we cannot handle it. If we cannot handle it, we either need
// to correct the Wikidata record, or add capability to Siegfried or
// the converter package.
func validateAndReturnEncoding(value string) (int, linting) {
	encoding := converter.LookupEncoding(value)
	if encoding == converter.UnknownEncoding {
		return encoding, encWDE01
	}
	return encoding, nle
}

// validateAndReturnRelativity will return a string and an error based
// on whether the relativity of a format identification pattern, e.g.
// BOF, EOF is known. If it isn't then it makes it more difficult to
// process in Roy/Siegfried.
func validateAndReturnRelativity(value string) (string, linting, error) {
	const unknownRelativity = "Received an unknown relativity"
	if value == "" {
		// Assume beginning of file.
		return relativeBOF, relWDE01, nil
	} else if value == relativeBOF {
		return relativeBOF, nle, nil
	} else if value == relativeEOF {
		return relativeEOF, nle, nil
	}
	return value, relWDE02, fmt.Errorf("%s: '%s'", unknownRelativity, value)
}

// validateAndReturnOffset will return an integer and an error based on
// whether we can use the offset delivered by Wikidata.
func validateAndReturnOffset(value string, nodeType string) (int, linting) {
	const blankNodeType = "bnode"
	const blankNodeErr = "Received a blank node type instead of offset"
	var offset int
	if value == "" {
		return offset, nle
	} else if nodeType == blankNodeType {
		return offset, offWDE03
	}
	offset, err := strconv.Atoi(value)
	if err != nil {
		return offset, offWDE02
	}
	return offset, nle
}

// validateAndReturnSignature calls the converter functions to normalize
// our signature. We need to do this so that we can compare signatures
// and remove duplicates and identify other errors.
func validateAndReturnSignature(value string, encoding int) (string, linting, error) {
	value, _, _, err := converter.Parse(value, encoding)
	if err != nil {
		return value, heuWDE02, err
	}
	return value, nle, nil
}
