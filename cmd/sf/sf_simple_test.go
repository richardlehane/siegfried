// Test Siegfried against DROID's simplified syntax introduced in
// earlier versions of the tooling and can be created via
// sigtool:
//
//   https://github.com/digital-preservation/droid#sigtool

package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var simpleDROIDTests = filepath.Join("..", "roy", "data", "droid_simple")
var simpleDROIDtestData = filepath.Join("..", "roy", "data", "droid_simple", "sample_files")

// const simpleTestDefinitions = "simple-test-definitions"
const simpleDefinitionsBaseDir = "simpleDefinitionsBaseDir"

var (
	simpleDefinitions = flag.String(
		"",
		simpleDROIDTests,
		"creates a flag var that is compatible with the config functions...",
	)
)

func resetSFSimple() {
	// use if any state needs resetting in the tests below.
}

func setupSimpleDROIDTests() (*siegfried.Siegfried, error) {
	resetSFSimple()
	config.SetHome(*simpleDefinitions)
	opts := []config.Option{
		config.SetNoReports(),
		config.SetDroid("acdc_standard.xml"),
		config.SetContainer("acdc_container.xml"),
	}
	identifier, err := pronom.New(opts...)
	if err != nil {
		return nil, err
	}
	var simpleDROIDSiegfried = siegfried.New()
	simpleDROIDSiegfried.Add(identifier)
	return simpleDROIDSiegfried, nil
}

// wbIdentificationTests provides our structure for table driven tests.
type simpleDROIDTest struct {
	fname          string
	label          string
	puid           string
	extMatch       bool
	byteMatch      bool
	containerMatch bool
	error          bool
	hasExt         bool
}

var simpleDROIDTestsOK = []simpleDROIDTest{
	// Standard file.
	simpleDROIDTest{
		fname:          filepath.Join(simpleDROIDtestData, "simple_file.skel"),
		label:          "",
		puid:           "skel/1",
		extMatch:       false,
		byteMatch:      true,
		containerMatch: false,
		error:          false,
		hasExt:         true,
	},
	// Container file.
	simpleDROIDTest{
		fname:          filepath.Join(simpleDROIDtestData, "container_file.skelc"),
		label:          "",
		puid:           "skel/2",
		extMatch:       false,
		byteMatch:      true,
		containerMatch: false,
		error:          false,
		hasExt:         true,
	},
}

var simpleDROIDTestsNegative = []simpleDROIDTest{
	// No beginning.
	simpleDROIDTest{
		fname:          filepath.Join(simpleDROIDtestData, "simple_no_bof.skel"),
		label:          "",
		puid:           "UNKNOWN",
		extMatch:       false,
		byteMatch:      false,
		containerMatch: false,
		error:          false,
		hasExt:         true,
	},
	// No middle.
	simpleDROIDTest{
		fname:          filepath.Join(simpleDROIDtestData, "simple_no_mid.skel"),
		label:          "",
		puid:           "UNKNOWN",
		extMatch:       false,
		byteMatch:      false,
		containerMatch: false,
		error:          false,
		hasExt:         true,
	},
	// No end.
	simpleDROIDTest{
		fname:          filepath.Join(simpleDROIDtestData, "simple_no_eof.skel"),
		label:          "",
		puid:           "UNKNOWN",
		extMatch:       false,
		byteMatch:      false,
		containerMatch: false,
		error:          false,
		hasExt:         true,
	},
	// Text file.
	simpleDROIDTest{
		fname:          filepath.Join(simpleDROIDtestData, "plain_text.skel"),
		label:          "",
		puid:           "UNKNOWN",
		extMatch:       false,
		byteMatch:      false,
		containerMatch: false,
		error:          false,
		hasExt:         true,
	},
	// Zip file.
	simpleDROIDTest{
		fname:          filepath.Join(simpleDROIDtestData, "empty_zip.skelc"),
		label:          "",
		puid:           "x-fmt/263",
		extMatch:       false,
		byteMatch:      true,
		containerMatch: false,
		error:          false,
		hasExt:         true,
	},
}

// Run the simple DROID tests.
func TestSimpleBasic(t *testing.T) {
	sf, err := setupSimpleDROIDTests()
	if err != nil {
		t.Error(err)
	}
	for _, test := range simpleDROIDTestsOK {
		simpleDROIDRunner(sf, test, t)
	}
	for _, test := range simpleDROIDTestsNegative {
		simpleDROIDRunner(sf, test, t)
	}
}

// Compare the test results against those that we expect.
func simpleDROIDRunner(wdSiegfried *siegfried.Siegfried, test simpleDROIDTest, t *testing.T) {
	file, err := os.Open(test.fname)
	if err != nil {
		t.Fatalf("failed to open %v, got: %v", test.fname, err)
	}
	defer file.Close()
	res, err := wdSiegfried.Identify(file, test.fname, "")
	if err != nil && !test.error {
		t.Fatal(err)
	}
	if len(res) > 1 {
		t.Errorf("Match length greater than one: '%d'", len(res))
	}
	namespace := res[0].Values()[0]
	if namespace != "pronom" {
		t.Errorf("Namespace error, expected: '%s' received: '%s'",
			wikidataNamespace, namespace,
		)
	}

	id := res[0].Values()[1]
	_ = res[0].Values()[2] // label
	basis := res[0].Values()[5]

	if id != test.puid {
		t.Errorf(
			"QID match different than anticipated: '%s' expected '%s'",
			id,
			test.puid,
		)
	}
	if test.extMatch && !strings.Contains(basis, extensionMatch) {
		t.Errorf(
			"Extension match not returned by identifier: %s",
			basis,
		)
	}
	if test.byteMatch && !strings.Contains(basis, byteMatch) {
		t.Errorf(
			"Byte match not returned by identifier: %s",
			basis,
		)
	}
	if test.containerMatch && !strings.Contains(basis, containerMatch) {
		t.Errorf(
			"Container match not returned by identifier: %s",
			basis,
		)
	}

}
