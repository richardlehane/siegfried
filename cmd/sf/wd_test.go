package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/wikidata"
)

const wdTestDefinitions = "wikidata-test-definitions"
const wdRoyData = "../roy/data"
const wdTestData = "testdata"
const wdWikidataSamples = "wikidata"
const wdPronom = "pro"
const wdNative = "wd"

// WIKIDATA TODO: We don't have to define these anew as they are defined in
// the sf_test runner. What separation of responsibilities do we want in this code?
var (
	wdtesthome = flag.String("wdtesthome", wdRoyData, "override the default home directory")
	wdtestdata = flag.String("wdtestdata", filepath.Join(".", wdTestData), "override the default test data directory")
)

var wdSiegfried *siegfried.Siegfried

func setupWikidata(opts ...config.Option) error {
	if opts == nil && s != nil {
		return nil
	}
	var err error
	wdSiegfried = siegfried.New()
	// Our Wikidata test signature file will be in
	// ../roy/data/wikidata/wikidata-definitions-*
	config.SetHome(*wdtesthome)
	config.SetWikidataDefinitions(wdTestDefinitions)
	p, err := wikidata.New()
	if err != nil {
		return err
	}
	return wdSiegfried.Add(p)
}

type PronomQname struct {
	fname          string
	qname          string
	extMatch       bool
	byteMatch      bool
	containerMatch bool
	error          bool
}

var skeletonSamples = []PronomQname{
	PronomQname{"fmt-11-signature-id-58.png", "Q27229565", true, true, false, false},
	PronomQname{"fmt-111-signature-id-170.nul", "Q5156830", false, true, false, true},
	PronomQname{"fmt-136-container-signature-id-6000.odt", "Q27203404", true, true, true, false},
	PronomQname{"fmt-279-signature-id-295.flac", "Q27881556", true, true, false, false},
	PronomQname{"fmt-412-container-signature-id-1030.docx", "Q3033641", true, true, true, false},
	PronomQname{"fmt-412-container-signature-id-1040.docx", "Q3033641", true, true, true, false},
	PronomQname{"fmt-412-container-signature-id-1050.docx", "Q3033641", true, true, true, false},

	// WIKIDATA TODO: This is going to fail intermittently 'Q28052851' or
	// 'Q475488'. (Both have fmt/483 as their listed identifier... need to
	// resolve) PronomQname{"fmt-483-container-signature-id-14010.epub",
	// "Q28052851", true, true, false},

	PronomQname{"fmt-494-container-signature-id-16000.docx", "Q3033641", true, false, true, false},

	PronomQname{"x-fmt-263-signature-id-200.zip", "Q136218", true, true, false, false},
	PronomQname{"x-fmt-418-signature-id-440.ico", "Q729366", true, true, false, false},
}

const extensionMatch = "extension match"
const byteMatch = "byte match"
const containerMatch = "container name"

// TestWDPRONOMSuite trials the Wikidata identifier against samples in the
// PRONOM test suite.
func TestWDPRONOMSuite(t *testing.T) {
	err := setupWikidata()
	if err != nil {
		t.Error(err)
	}
	for _, v := range skeletonSamples {
		path := filepath.Join(*testdata, wdWikidataSamples, wdPronom, v.fname)
		file, err := os.Open(path)
		if err != nil {
			t.Errorf("failed to open %v, got: %v", path, err)
		}
		defer file.Close()
		c, err := wdSiegfried.Identify(file, path, "")
		if err != nil && !v.error {
			t.Fatal(err)
		}
		if len(c) > 1 {
			t.Errorf("Match length greater than one: '%d'", len(c))
		}
		id := c[0].String()
		basis := c[0].Values()[5]
		if id != v.qname {
			t.Errorf("Qname match different than anticipated: '%s' expected '%s'", id, v.qname)
		}
		if v.extMatch && !strings.Contains(basis, extensionMatch) {
			t.Errorf("Extension match not returned by identifier: %s", basis)
		}
		if v.byteMatch && !strings.Contains(basis, byteMatch) {
			t.Errorf("Byte match not returned by identifier: %s", basis)
		}
		if v.containerMatch && !strings.Contains(basis, containerMatch) {
			t.Errorf("Container match not returned by identifier: %s", basis)
		}
	}
}
