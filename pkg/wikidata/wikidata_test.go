package wikidata

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardlehane/siegfried/pkg/config"
)

type idTestsStruct struct {
	uri string
	res string
}

var idTests = []idTestsStruct{
	idTestsStruct{"http://www.wikidata.org/entity/Q1023647", "Q1023647"},
	idTestsStruct{"http://www.wikidata.org/entity/Q336284", "Q336284"},
	idTestsStruct{"http://www.wikidata.org/entity/Q9296340", "Q9296340"},
}

// TestGetID is a rudimentary test to make sure that we can retrieve
// QUIDs reliably from a Wikidata URI.
func TestGetID(t *testing.T) {
	for _, v := range idTests {
		res := getID(v.uri)
		if res != v.res {
			t.Errorf("Expected to generate QID '%s' but received '%s'", v.res, res)
		}
	}
}

// TestOpenWikidata simply verifies the different anticipated behavior
// of openWikidata so that behavior is predictable for the end user.
func TestOpenWikidata(t *testing.T) {

	var testDefinitions = `
{
  "endpoint": "%replaceMe%",
  "head": {},
  "results": {
    "bindings": []
  },
  "provenance": []
}
`

	var testJSON = `
{
 "PronomProp": "http://wikibase.example.com/entity/Q2",
 "BofProp": "http://wikibase.example.com/entity/Q3",
 "EofProp": "http://wikibase.example.com/entity/Q4"
}
`
	var replaceMe = "%replaceMe%"
	var defaultEndpoint = "https://query.wikidata.org/sparql"
	tempDir, _ := ioutil.TempDir("", "wikidata-test-dir-*")
	defer os.RemoveAll(tempDir)
	err := os.Mkdir(filepath.Join(tempDir, "wikidata"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	config.SetHome(tempDir)
	defsWithCustomEndpoint := filepath.Join(tempDir, "wikidata", "wikidata-definitions")
	err = ioutil.WriteFile(defsWithCustomEndpoint, []byte(testDefinitions), 0755)
	if err != nil {
		t.Fatal(err)
	}
	config.SetWikidataDefinitions("wikidata-definitions")
	// At this point wikibase.json doesn't exist and so we want to
	// receive an error.
	_, err = openWikidata()
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf(
			"Expected error loading wikibase.json but received: %s",
			err,
		)
	}
	// Now make sure wikibase.json exists, there should be no error.
	wikibaseJSON := filepath.Join(tempDir, "wikidata", "wikibase.json")
	err = ioutil.WriteFile(wikibaseJSON, []byte(testJSON), 0755)
	if err != nil {
		t.Fatal(err)
	}
	// Wikidata should now open without error and complete processing.
	// Even though we have a definitions file with no actionable data,
	// it's not actually an error.
	_, err = openWikidata()
	if err != nil {
		t.Fatal(err)
	}
	// Test defaults, by first removing custom JSON and then setting up
	// a default definitions file.
	err = os.Remove(wikibaseJSON)
	if err != nil {
		t.Fatal(err)
	}
	defaultDefinitions := strings.Replace(testDefinitions, replaceMe, defaultEndpoint, 1)
	config.SetHome(tempDir)
	defsWithDefaultEndpoint := filepath.Join(tempDir, "wikidata", "wikidata-definitions")
	err = ioutil.WriteFile(defsWithDefaultEndpoint, []byte(defaultDefinitions), 0755)
	if err != nil {
		t.Fatal(err)
	}
	// Open the good definitions now, there should be no issue.
	_, err = openWikidata()
	if err != nil {
		t.Fatal(err)
	}
	// Finally let's put some really un-useful information into the path
	// for the definitions and try that.
	config.SetWikidataDefinitions("/path/🙅/does/🙅/not/🙅/exist/🙅/")
	_, err = openWikidata()
	if err == nil {
		t.Errorf(
			"Anticipating error with non-existent path, but got: 'nil'",
		)
	}
}
