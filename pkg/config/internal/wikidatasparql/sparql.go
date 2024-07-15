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

package wikidatasparql

// wikidatasparql encapsulates SPARQL functions required for generating
// the Wikidata identifier in Roy.

import (
	"strconv"
	"strings"
)

// sigLenTemplate gives us a field which we can replace with a
// min signature length value of our own choosing.
const sigLenTemplate = "<<siglen>>"

// Default signature length to return from Wikidata.
var wikidataSigLen = 6

// languateTemplate gives us a field which we can replace with a
// language code of our own configuration.
const languageTemplate = "<<lang>>"

// Number of replacements to make when replacing the SPARQL fields with
// the values that we have configured.
const numberReplacements = 1

// Default language for the Wikidata SPARQL query.
var wikidataLang = "en"

// sparql represents the query required to pull all file format records
// and signatures from the Wikidata query service.
const sparql = `
# Return all file format records from Wikidata.
SELECT DISTINCT ?uri ?uriLabel ?puid ?extension ?mimetype ?encoding ?referenceLabel ?date ?relativity ?offset ?sig WHERE {
  { ?uri (wdt:P31/(wdt:P279*)) wd:Q235557. }
  UNION
  { ?uri (wdt:P31/(wdt:P279*)) wd:Q26085352. }
  FILTER(EXISTS { ?uri (wdt:P2748|wdt:P1195|wdt:P1163|ps:P4152) _:b2. })
  FILTER((STRLEN(?sig)) >= <<siglen>> )
  OPTIONAL { ?uri wdt:P2748 ?puid. }
  OPTIONAL { ?uri wdt:P1195 ?extension. }
  OPTIONAL { ?uri wdt:P1163 ?mimetype. }
  OPTIONAL {
    ?uri p:P4152 ?object.
    OPTIONAL { ?object pq:P3294 ?encoding. }
    OPTIONAL { ?object ps:P4152 ?sig. }
    OPTIONAL { ?object pq:P2210 ?relativity. }
    OPTIONAL { ?object pq:P4153 ?offset. }
    OPTIONAL {
      ?object prov:wasDerivedFrom ?provenance.
      OPTIONAL {
        ?provenance pr:P248 ?reference;
          pr:P813 ?date.
      }
    }
  }
  SERVICE wikibase:label { bd:serviceParam wikibase:language "[AUTO_LANGUAGE], <<lang>>". }
}
ORDER BY (?uri)
`

// WikidataSPARQL returns the SPARQL query needed to pull file-format
// signatures from Wikidata replacing various template values as we
// go.
func WikidataSPARQL() string {
	wdSparql := strings.Replace(sparql, languageTemplate, wikidataLang, numberReplacements)
	wdSparql = strings.Replace(wdSparql, sigLenTemplate, strconv.Itoa(wikidataSigLen), numberReplacements)
	return wdSparql
}

// WikidataSigLen returns the minimum signature length we want the Wikidata
// SPARQL query to return.
func WikidataSigLen() int {
	return wikidataSigLen
}

// SetWikidataSigLen sets the minimum signature length we want the Wikidata
// SPARQL query to return.
func SetWikidataSigLen(len int) {
	wikidataSigLen = len
}

// WikidataLang will return to the caller the ISO language code
// currently configured for this module.
func WikidataLang() string {
	return wikidataLang
}

// SetWikidataLang will set the Wikidata language to one supplied by
// the user. The language should be an ISO language code such as fr.
// de. jp. etc.
func SetWikidataLang(lang string) {
	wikidataLang = lang
}
