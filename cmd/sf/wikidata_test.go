package main

import (
	"flag"
	"fmt"

	"os"

	"path/filepath"
	"strings"
	"testing"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/wikidata"
)

// Path components associated with the Roy command folder.
const wikidataTestDefinitions = "wikidata-test-definitions"
const wikidataDefinitionsBaseDir = "definitionsBaseDir"

var royTestData = filepath.Join("..", "roy", "data")

// Path components within the Siegfried command folder.
const siegfriedTestData = "testdata"
const wikidataTestData = "wikidata"
const wikidataPRONOMSkeletons = "pro"
const wikidataCustomSkeletons = "wd"
const wikidataArcSkeletons = "arc"

var (
	wikidataDefinitions = flag.String(
		wikidataDefinitionsBaseDir,
		royTestData,
		"Creates an flag var that is compatible with the config functions...",
	)
)

var wdSiegfried *siegfried.Siegfried

func setupWikidata(pronomx bool, opts ...config.Option) error {
	if opts == nil && wdSiegfried != nil {
		return fmt.Errorf(
			"Wikidata setup options are not properly configured",
		)
	}
	wdSiegfried = siegfried.New()
	config.SetHome(*wikidataDefinitions)
	config.SetWikidataNamespace()
	config.SetWikidataDefinitions(wikidataTestDefinitions)
	if pronomx != true {
		opts = append(opts, config.SetWikidataNoPRONOM())
	} else {
		opts = append(opts, config.SetWikidataPRONOM())
	}
	identifier, err := wikidata.New(opts...)
	if err != nil {
		return err
	}
	wdSiegfried.Add(identifier)
	return nil
}

// identificationTests provides our structure for table driven tests.
type identificationTests struct {
	fname     string
	qid       string
	extMatch  bool
	byteMatch bool
	error     bool
}

var skeletonSamples = []identificationTests{
	identificationTests{
		filepath.Join(wikidataPRONOMSkeletons, "fmt-11-signature-id-58.png"),
		"Q178051", true, true, false},
	identificationTests{
		filepath.Join(wikidataPRONOMSkeletons, "fmt-279-signature-id-295.flac"),
		"Q27881556", true, true, false},
	identificationTests{
		filepath.Join(wikidataCustomSkeletons, "Q10287816.gz"),
		"Q10287816", true, true, false},
	identificationTests{
		filepath.Join(wikidataCustomSkeletons, "Q28205479.info"),
		"Q28205479", true, true, false},
}

// Rudimentary consts that can help us determine the method of
// identification. Can also add "container name" here for when we want
// to validate PRONOM alongside Wikidata.
const extensionMatch = "extension match"
const byteMatch = "byte match"

// TestWikidataBasic will perform some rudimentary tests using some
// simple Skeleton files and the Wikidata identifier without PRONOM.
func TestWikidataBasic(t *testing.T) {
	pronom := false
	err := setupWikidata(pronom)
	if err != nil {
		t.Error(err)
	}
	for _, test := range skeletonSamples {
		path := filepath.Join(siegfriedTestData, wikidataTestData, test.fname)
		siegfriedRunner(path, test, t)
	}
	wdSiegfried = nil
}

var archiveSamples = []identificationTests{
	identificationTests{
		filepath.Join(wikidataArcSkeletons, "fmt-289-signature-id-305.warc"),
		"Q7978505", true, true, false},
	identificationTests{
		filepath.Join(wikidataArcSkeletons, "fmt-410-signature-id-580.arc"),
		"Q27824065", true, true, false},
	identificationTests{
		filepath.Join(wikidataArcSkeletons, "x-fmt-219-signature-id-525.arc"),
		"Q27824060", true, true, false},
	identificationTests{
		filepath.Join(wikidataArcSkeletons, "x-fmt-265-signature-id-265.tar"),
		"Q283579", true, true, false},
	identificationTests{
		filepath.Join(wikidataArcSkeletons, "x-fmt-266-signature-id-201.gz"),
		"Q10287816", true, true, false},
}

func TestArchives(t *testing.T) {
	pronom := true
	err := setupWikidata(pronom)
	if err != nil {
		t.Error(err)
	}
	for _, test := range archiveSamples {
		path := filepath.Join(siegfriedTestData, wikidataTestData, test.fname)
		siegfriedRunner(path, test, t)
	}
	wdSiegfried = nil
}

func siegfriedRunner(path string, test identificationTests, t *testing.T) {
	file, err := os.Open(path)
	if err != nil {
		t.Errorf("failed to open %v, got: %v", path, err)
	}
	defer file.Close()
	res, err := wdSiegfried.Identify(file, path, "")
	if err != nil && !test.error {
		t.Fatal(err)
	}
	if len(res) > 1 {
		t.Errorf("Match length greater than one: '%d'", len(res))
	}
	id := res[0].Values()[1]
	basis := res[0].Values()[5]
	if id != test.qid {
		t.Errorf(
			"QID match different than anticipated: '%s' expected '%s'",
			id,
			test.qid,
		)
	}
	if test.extMatch && !strings.Contains(basis, extensionMatch) {
		t.Errorf("Extension match not returned by identifier: %s", basis)
	}
	if test.byteMatch && !strings.Contains(basis, byteMatch) {
		t.Errorf("Byte match not returned by identifier: %s", basis)
	}
}
