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
	"strings"
)

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
	#
	select distinct ?uri ?uriLabel ?puid ?extension ?mimetype ?encoding ?referenceLabel ?date ?relativity ?offset ?sig
	where
	{
	  ?uri wdt:P31/wdt:P279* wd:Q235557.               # Return records of type File Format.
	  optional { ?uri wdt:P2748 ?puid.      }          # PUID is used to map to PRONOM signatures proper.
	  optional { ?uri wdt:P1195 ?extension. }
	  optional { ?uri wdt:P1163 ?mimetype.  }
	  optional { ?uri p:P4152 ?object;                 # Format identification pattern statement.
	    optional { ?object pq:P3294 ?encoding.   }     # We don't always have an encoding.
	    optional { ?object ps:P4152 ?sig.        }     # We always have a signature.
	    optional { ?object pq:P2210 ?relativity. }     # Relativity to beginning or end of file.
	    optional { ?object pq:P4153 ?offset.     }     # Offset relative to the relativity.
	    optional { ?object prov:wasDerivedFrom ?provenance;
	       optional { ?provenance pr:P248 ?reference;
	                              pr:P813 ?date.
	                }
	    }
	  }
	  service wikibase:label { bd:serviceParam wikibase:language "[AUTO_LANGUAGE], <<lang>>". }
	}
	order by ?uri
 	`

// WikidataSPARQL returns the SPARQL query needed to pull file-format
// signatures from Wikidata replacing various template values as we
// go.
func WikidataSPARQL() string {
	return strings.Replace(sparql, languageTemplate, wikidataLang, numberReplacements)
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
