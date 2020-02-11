package config

import (
	"fmt"
	"strings"
	"testing"
)

/* TestReturnWDSparql ensures that the language `<<lang>>` template is present
in the SPARQL query required to harvest Wikidata results. The language template
is a functional component required to generate results in different languages
including English. */
func TestReturnWDSparql(t *testing.T) {
	res := WikidataSPARQL()
	if strings.Contains(res, wikidata.languageTemplate) {
		t.Errorf(
			"Lang replacement template is missing from SPARQL request string",
		)
	}
	res = WikidataSPARQL()
	defaultLang := fmt.Sprintf("\"[AUTO_LANGUAGE],%s\".", wikidata.wikidataLang)
	if !strings.Contains(res, defaultLang) {
		t.Errorf(
			"Default language `en` missing from SPARQL request string",
		)
	}
	// Change the language string and ensure that replacement occurs.
	newLang := "jp"
	newLangReplacement := fmt.Sprintf("\"[AUTO_LANGUAGE],%s\".", newLang)
	SetWikidataLang(newLangReplacement)
	res = WikidataSPARQL()
	if !strings.Contains(res, newLang) {
		t.Errorf(
			"Language replacement hasn't been done in returned SPARQL request string",
		)
	}
}
