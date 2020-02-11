package config

import (
	"os"
	"path/filepath"
	"strings"
)

var wikidata = struct {
	definitions string
	name        string
	nopronom    bool
	filemode    os.FileMode

	wikidataLang     string
	languageTemplate string

	endpoint string
	sparql   string

	wdSourceField bool
}{
	// WIKIDATA TODO: identify something helpful for versioning Wikidata
	// definitions file.
	definitions: "wikidata-definitions-0.0.3-beta-playable-demo-1.0.0",

	name:     "wikidata",
	filemode: 0644,

	wikidataLang:     "en",
	languageTemplate: "<<lang>>",

	endpoint: "https://query.wikidata.org/sparql",
	sparql: `
	SELECT DISTINCT ?format ?formatLabel ?puid ?ldd ?extension ?mimetype ?sig ?referenceLabel ?date ?encodingLabel ?offset ?relativityLabel WHERE
	{
	  ?format wdt:P31/wdt:P279* wd:Q235557.
	  OPTIONAL { ?format wdt:P2748 ?puid. }
	  OPTIONAL { ?format wdt:P3266 ?ldd }
	  OPTIONAL { ?format wdt:P1195 ?extension }
	  OPTIONAL { ?format wdt:P1163 ?mimetype }
	  OPTIONAL { ?format wdt:P4152 ?sig }
	  OPTIONAL {
	     ?format p:P4152 ?object.
	     ?object prov:wasDerivedFrom ?provenance.
	     ?provenance pr:P248 ?reference;
	        pr:P813 ?date.
	  }
	  OPTIONAL {
	     ?format p:P4152 ?object.
	     ?object pq:P3294 ?encoding.
	     ?object pq:P4153 ?offset.
	  }
	  OPTIONAL {
	     ?format p:P4152 ?object.
	     ?object pq:P2210 ?relativity.
	  }
	  SERVICE wikibase:label { bd:serviceParam wikibase:language "[AUTO_LANGUAGE], <<lang>>". }
	}
	order by ?format
	`,
}

func WDNoPRONOM() bool {
	return wikidata.nopronom
}

// WIKIDATA TODO: Learn exactly the technique being used here to make this
// field work as anticipated. As below with Wikidata source field.
func WDSetNoPRONOM() func() private {
	wikidata.nopronom = true
	return func() private {
		return private{}
	}
}

// SetWikidata ...
func SetWikidata() func() private {
	return func() private {
		return private{}
	}
}

// WIKIDATA TODO: Learn exactly the technique being used here to make this
// field work as anticipated. As above with SetNoPRONOM.
func SetWikidataSourceField() func() private {
	wikidata.wdSourceField = true
	return func() private {
		return private{}
	}
}

func GetWikidataSourceField() bool {
	return wikidata.wdSourceField
}

/* SetWikidataDefinitions is a setter to enable us to elect to use a different
signature file name, e.g. for testing. */
func SetWikidataDefinitions(definitions string) {
	wikidata.definitions = definitions
}

/* WikidataDefinitionsFile returns the name of the file used to store the
Signature definitions. */
func WikidataDefinitionsFile() string {
	return wikidata.definitions
}

/* WikidataHome describes where files needed by Siegfried and Roy for its
Wikidata component resides. */
func WikidataHome() string {
	return filepath.Join(siegfried.home, wikidata.name)
}

/* WikidataDefinitionsPath is a helper for convenience from callers to point
directly at the definitions path for reading/writing as required. */
func WikidataDefinitionsPath() string {
	return filepath.Join(WikidataHome(), WikidataDefinitionsFile())
}

// WikidataFileMode ...
func WikidataFileMode() os.FileMode {
	return wikidata.filemode
}

// WikidataEndpoint ...
func WikidataEndpoint() string {
	return wikidata.endpoint
}

// WikidataSPARQL ...
func WikidataSPARQL() string {
	return strings.Replace(wikidata.sparql, wikidata.languageTemplate, wikidata.wikidataLang, 1)
}

// WikidataLang ...
func WikidataLang() string {
	return wikidata.wikidataLang
}

// SetWikidataLang ...
func SetWikidataLang(lang string) {
	wikidata.wikidataLang = lang
}
