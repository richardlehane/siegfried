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

// Registration of different file-format signature sequence encodings
// that we might discover e.g. from sources such as Wikidata.

package converter

import (
	"github.com/richardlehane/siegfried/pkg/config"
)

// Encoding enumeration to return unambiguous values for encoding from
// the mapping lookup below.
const (
	// UnknownEncoding provides us with a default to work with.
	UnknownEncoding = iota
	// HexEncoding describes magic numbers written in plain-hexadecimal.
	HexEncoding
	// PronomEncoding describe PRONOM based file format signatures.
	PronomEncoding
	// PerlEncoding describe PERL regular expression encoded signatures.
	PerlEncoding
	// ASCIIEncoding encoded patterns are those written entirely in plain ASCII.
	ASCIIEncoding
	// GUIDEncoding are globally unique identifiers.
	GUIDEncoding
)

// Encoding constants. IRIs from Wikidata mean that we don't need to
// encode i18n differences. IRIs must have http:// scheme, and link to
// the data entity, i.e. not the "page", e.g.
//
//    * Hex data entity: http://www.wikidata.org/entity/Q82828
//    * Hex page: https://www.wikidata.org/wiki/Q82828
//
const (
	// Hexadecimal.
	hexadecimal = "http://www.wikidata.org/entity/Q82828"
	// Globally unique identifier.
	guid = "http://www.wikidata.org/entity/Q254972"
	// ASCII.
	ascii = "http://www.wikidata.org/entity/Q8815"
	// Perl compatible regular expressions 2.
	perl = "http://www.wikidata.org/entity/Q98056596"
	// Unknown encoding.
	unknown = "unknown encoding"
)

// PRONOM internal signature. This is not a constant as it can be read
// into Roy from wikibase.json.
var pronom string = config.WikibasePronom()

// GetPronomURIFromConfig will read the current value of the PRONOM
// property from the configuration, e.g. after being updated using a
// custom SPARQL query.
func GetPronomURIFromConfig() {
	pronom = config.WikibasePronom()
}

// GetPronomEncoding returns the PRONOM encoding URI as is set locally in
// the package.
func GetPronomEncoding() string { return pronom }

// LookupEncoding will return a best-guess encoding type for a supplied
// encoding string.
func LookupEncoding(encoding string) int {
	encoding = encoding
	switch encoding {
	case hexadecimal:
		return HexEncoding
	case pronom:
		return PronomEncoding
	case perl:
		return PerlEncoding
	case ascii:
		return ASCIIEncoding
	case guid:
		return GUIDEncoding
	}
	return UnknownEncoding
}

// ReverseEncoding can provide a human readable string for us if we
// ever need it, e.g. if we need to debug this module.
func ReverseEncoding(encoding int) string {
	switch encoding {
	case HexEncoding:
		return hexadecimal
	case PronomEncoding:
		return pronom
	case PerlEncoding:
		return perl
	case ASCIIEncoding:
		return ascii
	case GUIDEncoding:
		return guid
	}
	return unknown
}
