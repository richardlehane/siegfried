package config

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TestProps makes sure the properties do not skew without deliberate
// consideration when doing so.
func TestProps(t *testing.T) {
	pronom := "http://www.wikidata.org/entity/Q35432091"
	bof := "http://www.wikidata.org/entity/Q35436009"
	eof := "http://www.wikidata.org/entity/Q1148480"
	if WikibasePronom() != pronom {
		t.Errorf(
			"Pronom property '%s' is not '%s'",
			WikibasePronom(),
			pronom,
		)
	}
	if WikibaseBOF() != bof {
		t.Errorf(
			"BOF property '%s' is not '%s'",
			WikibaseBOF(),
			bof,
		)
	}
	if WikibaseEOF() != eof {
		t.Errorf(
			"EOF property '%s' is not '%s'",
			WikibaseEOF(),
			eof,
		)
	}
}

// TestSetCustomWikibaseQuery provides a way to verify some of the basic
// handling required for updating our SPARQL query for a custom Wikibase.
func TestSetCustomWikibaseQuery(t *testing.T) {
	var testSPARQL = "select ?s ?p ?o where { ?s ?p ?o. }"
	tempDir, _ := ioutil.TempDir("", "wikidata-test-dir-*")
	defer os.RemoveAll(tempDir)
	err := os.Mkdir(filepath.Join(tempDir, "wikidata"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	SetHome(tempDir)
	customSPARQLFile := filepath.Join(tempDir, "wikidata", "wikibase.sparql")
	err = ioutil.WriteFile(customSPARQLFile, []byte(testSPARQL), 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = SetCustomWikibaseQuery()
	if err != nil {
		t.Errorf(
			"Unexpected error setting custom wikibase query %s",
			err,
		)
	}
	if WikidataSPARQL() != testSPARQL {
		t.Errorf(
			"Query not updated from custom SPARQL as expected: '%s'",
			WikidataSPARQL(),
		)
	}
	err = os.Remove(customSPARQLFile)
	err = SetCustomWikibaseQuery()
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf(
			"Expected error loading wikibase.sparql but received: %s",
			err,
		)
	}
}
