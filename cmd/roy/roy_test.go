package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/fs"
	"strings"
	"testing"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/loc"
	"github.com/richardlehane/siegfried/pkg/mimeinfo"
	"github.com/richardlehane/siegfried/pkg/pronom"
	"github.com/richardlehane/siegfried/pkg/sets"
	wd "github.com/richardlehane/siegfried/pkg/wikidata"
)

var testhome = flag.String("home", "data", "override the default home directory")

func TestDefault(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	p, err := pronom.New()
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(p)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoc(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	l, err := loc.New(config.SetLOC(""))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(l)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTika(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	m, err := mimeinfo.New(config.SetMIMEInfo("tika"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(m)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFreedesktop(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	m, err := mimeinfo.New(config.SetMIMEInfo("freedesktop"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(m)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWikidata(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	config.SetWikidataDefinitions("wikidata-test-definitions")
	m, err := wd.New(config.SetWikidataNamespace())
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(m)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWikibaseNoEndpoint(t *testing.T) {
	config.SetHome(*testhome)
	config.SetWikidataDefinitions("custom-wikibase-test-definitions-no-endpoint")
	_, err := wd.New(config.SetWikidataNamespace())
	if !errors.Is(err, wd.ErrNoEndpoint) {
		t.Fatalf(
			"Expected 'ErrNoEndpoint' trying to open custom Wikibase definitions, but got: '%s'",
			err,
		)
	}
}

func TestWikibaseNoProps(t *testing.T) {
	config.SetHome(*testhome)
	config.SetWikibasePropsPath("/path/does/not/exist.json")
	config.SetWikidataDefinitions("custom-wikibase-test-definitions")
	_, err := wd.New(config.SetWikidataNamespace())
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf(
			"Expected an error trying to open custom Wikibase properties, but got: '%s'",
			err,
		)
	}
}

func TestWikibase(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	// Default wouldn't normally need to be set, but may be overridden
	// through other tests.
	config.SetWikibasePropsPath("wikibase.json")
	config.SetWikidataDefinitions("custom-wikibase-test-definitions")
	m, err := wd.New(config.SetWikidataNamespace())
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(m)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPronomTikaLoc(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	p, err := pronom.New(config.Clear())
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(p)
	if err != nil {
		t.Fatal(err)
	}
	m, err := mimeinfo.New(config.SetMIMEInfo("tika"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(m)
	if err != nil {
		t.Fatal(err)
	}
	l, err := loc.New(config.SetLOC(""))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(l)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeluxe(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	p, err := pronom.New(config.Clear())
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(p)
	if err != nil {
		t.Fatal(err)
	}
	m, err := mimeinfo.New(config.SetMIMEInfo("tika"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(m)
	if err != nil {
		t.Fatal(err)
	}
	f, err := mimeinfo.New(config.SetMIMEInfo("freedesktop"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(f)
	if err != nil {
		t.Fatal(err)
	}
	l, err := loc.New(config.SetLOC(""))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(l)
	if err != nil {
		t.Fatal(err)
	}
}

func TestArchivematica(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	p, err := pronom.New(
		config.SetName("archivematica"),
		config.SetExtend(sets.Expand("archivematica-fmt2.xml,archivematica-fmt3.xml,archivematica-fmt4.xml,archivematica-fmt5.xml")))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(p)
	if err != nil {
		t.Fatal(err)
	}
}

// TestAddEndpoint makes sure that valid JSON is still output when we
// add the endpoint to the SPARQL JSON.
func TestAddEndpoint(t *testing.T) {
	simpleJSON := `
{
  "key_one": "value_one",
  "key_two": "value_two"
}
`
	resJSON := `
{
  "endpoint": "http://example.com:8834/proxy/wdqs/bigdata/namespace/wdq/sparql?",
  "key_one": "value_one",
  "key_two": "value_two"
}
`
	res := addEndpoint(
		simpleJSON,
		"http://example.com:8834/proxy/wdqs/bigdata/namespace/wdq/sparql?",
	)
	// Try to see if adding endpoint works, and is equal to our sample
	// JSON before checking whether or not it is valid.
	if res != resJSON {
		t.Errorf(
			"Replacement result '%s' does not match what was expected '%s'",
			res,
			resJSON,
		)
	}
	valid := json.Valid([]byte(res))
	if !valid {
		t.Fatalf("Add endpoint returned invalid JSON: '%s'", res)
	}
	// Lets flatten the JSON structure a bit and see if we can cause
	// more problems this way,
	res = addEndpoint(
		strings.ReplaceAll(simpleJSON, "\n", ""),
		"http://example.com:8834/proxy/wdqs/bigdata/namespace/wdq/sparql?",
	)
	if strings.ReplaceAll(res, "\n", "") !=
		strings.ReplaceAll(resJSON, "\n", "") {
		t.Errorf(
			"Replacement result '%s' does not match what was expected '%s'",
			res,
			resJSON,
		)
	}
	valid = json.Valid([]byte(res))
	if !valid {
		t.Fatalf(
			"Add endpoint returned invalid JSON: '%s'",
			res,
		)
	}
}
