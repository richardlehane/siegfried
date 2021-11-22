package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/wikidata"
	"github.com/richardlehane/siegfried/pkg/writer"
)

// Path components associated with the Roy command folder.
const wikibaseTestDefinitions = "custom-wikibase-test-definitions"
const wikibaseCustomSkeletons = "wikibase"

func setupWikibase() (*siegfried.Siegfried, error) {
	config.SetWikidataEndpoint("https://query.wikidata.org/sparql")
	// resetWikidata sets state for the standard Wikidata tests because
	// it isn't set during normal runtime like the Wikibase config. This
	// function is good to call here anyway as it helps verify that
	// runtime config for Wikibase regardless.
	resetWikidata()
	var wbSiegfried *siegfried.Siegfried
	wbSiegfried = siegfried.New()
	config.SetHome(*wikidataDefinitions)
	config.SetWikidataDefinitions(wikibaseTestDefinitions)
	opts := []config.Option{config.SetWikidataNamespace()}
	opts = append(opts, config.SetWikidataNoPRONOM())
	wbIdentifier, err := wikidata.New(opts...)
	if err != nil {
		return wbSiegfried, err
	}
	wbSiegfried.Add(wbIdentifier)
	return wbSiegfried, nil
}

// wbIdentificationTests provides our structure for table driven tests.
type wbIdentificationTests struct {
	fname          string
	label          string
	qid            string
	extMatch       bool
	byteMatch      bool
	containerMatch bool
	error          bool
	hasExt         bool
}

var wbSkeletonSamples = []wbIdentificationTests{
	wbIdentificationTests{
		filepath.Join(wikibaseCustomSkeletons, "badf00d.badf00d"),
		"FFIFF", "Q6", false, true, false, false, true},
	wbIdentificationTests{
		filepath.Join(wikibaseCustomSkeletons, "ba53ba11.ff2"),
		"FFIIFF", "Q7", true, true, false, false, true},
	wbIdentificationTests{
		filepath.Join(wikibaseCustomSkeletons, "FFIIIFF"),
		"FFIIIFF", "Q9", true, true, false, false, false},
	wbIdentificationTests{
		filepath.Join(wikibaseCustomSkeletons, "FITS"),
		"Flexible Image Transport System (FITS), Version 3.0", "Q8",
		true, true, false, false, false,
	},
}

// TestWikidataBasic will perform some rudimentary tests using some
// simple Skeleton files and the Wikibase identifier.
func TestWikibaseBasic(t *testing.T) {
	wbSiegfried, err := setupWikibase()
	if err != nil {
		t.Error(err)
	}
	for _, test := range wbSkeletonSamples {
		path := filepath.Join(siegfriedTestData, wikidataTestData, test.fname)
		wbSiegfriedRunner(wbSiegfried, path, test, t)
	}
}

func wbSiegfriedRunner(wbSiegfried *siegfried.Siegfried, path string, test wbIdentificationTests, t *testing.T) {
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open %v, got: %v", path, err)
	}
	defer file.Close()
	res, err := wbSiegfried.Identify(file, path, "")
	if err != nil && !test.error {
		t.Fatal(err)
	}
	if len(res) > 1 {
		t.Errorf("Match length greater than one: '%d'", len(res))
	}
	namespace := res[0].Values()[0]
	if namespace != wikidataNamespace {
		t.Errorf("Namespace error, expected: '%s' received: '%s'",
			wikidataNamespace, namespace,
		)
	}
	// res is a an array of JSON values. We're interested in the first
	// result (index 0), and then the following fields
	id := res[0].Values()[1]
	label := res[0].Values()[2]
	permalink := res[0].Values()[4]
	basis := res[0].Values()[6]
	warning := res[0].Values()[7]
	if id != test.qid {
		t.Errorf(
			"QID match different than anticipated: '%s' expected '%s'",
			id,
			test.qid,
		)
	}
	if label != test.label {
		t.Errorf(
			"Label match different than anticipated: '%s' expected '%s'",
			label,
			test.label,
		)
	}
	const placeholderPermalink = "http://wikibase.example.com/w/index.php?oldid=58&title=Item%3AQ6"
	if permalink != placeholderPermalink {
		t.Errorf(
			"There has been a problem parsing the permalink for '%s' from Wikidata/Wikiprov: %s",
			test.qid,
			permalink,
		)
	}
	if test.extMatch && !strings.Contains(basis, extensionMatch) {
		if test.hasExt {
			t.Errorf(
				"Extension match not returned by identifier: %s",
				basis,
			)
		}
	}
	if test.byteMatch && !strings.Contains(basis, byteMatch) {
		t.Errorf(
			"Byte match not returned by identifier: %s",
			basis,
		)
	}
	if !test.extMatch && !strings.Contains(warning, extensionMismatch) {
		t.Errorf(
			"Expected an extension mismatch but it wasn't returned: %s",
			warning,
		)
	}
	// Implement a basic Writer test for some of the data coming out of
	// the Wikidata identifier. CSV and YAML will need a little more
	// thought.
	var w writer.Writer
	buf := new(bytes.Buffer)
	w = writer.JSON(buf)
	w.Head(
		"path/to/file",
		time.Now(),
		time.Now(),
		[3]int{0, 0, 0},
		wbSiegfried.Identifiers(),
		wbSiegfried.Fields(),
		"md5",
	)
	w.File("testName", 10, "testMod", []byte("d41d8c"), nil, res)
	w.Tail()
	if !json.Valid([]byte(buf.String())) {
		t.Fatalf("Output from JSON writer is invalid: %s", buf.String())
	}
}
