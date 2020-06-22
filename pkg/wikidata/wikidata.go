package wikidata

import (
	"fmt"
	"os"

	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/pronom"
	"github.com/richardlehane/siegfried/pkg/wikidata/internal/mappings"
)

// WIKIDATA TODO: Suggest a different name for parseable? Parseable is
// described as:
//
//      Parseable is something we can parse to derive filename,
//      MIME, XML and byte signatures.
//
type wikidataFDDs struct {
	formats   []mappings.Wikidata
	parseable identifier.Parseable
	identifier.Blank
}

// Load Wikidata report from disk and
func newWikidata() (identifier.Parseable, []string, error) {
	/*
	   We're going to load up a Wikidata "report" here. It's going to do a
	   few things, but first, it's going to read the Wikidata IDs available.
	   It's going to then read how those are mapped to PRONOM signatures.

	   It's then going to load a PRONOM signature into memory (unless
	   noPRONOM is set) and then we'll return some form of mapping.

	   And for now, we're also going to try and pass about a slice  of PUIDs to
	   work with in this identifier. I think it likely that we'll want to
	   rework that before finalizing this effort.

	   But first, let's load the mappings...
	*/
	var puids []string
	var wikiParseable identifier.Parseable = identifier.Blank{}
	var err error
	// Open Wikidata report from disk and read into a report structure.
	openWikidata()
	// If we are using PRONOM signatures we start to load those here.
	if config.WDNoPRONOM() {
		fmt.Fprintf(
			os.Stderr,
			"Roy (Wikidata): Not building identifiers set from PRONOM\n",
		)
	} else {
		fmt.Fprintf(
			os.Stderr,
			"Roy (Wikidata): Building identifiers set from PRONOM\n",
		)
		wikiParseable, err = pronom.NewPronom()
		if err != nil {
			return nil, []string{}, err
		}
		// WIKIDATA TODO: This structure gives us access to the PRONOM
		// identifiers we want to work with in this identifier.
		_, puids, _ = wikiParseable.Signatures()
	}
	return wikidataFDDs{
		reportMappings,     // formats.
		wikiParseable,      // parseable.
		identifier.Blank{}, // anonymous field.
	}, puids, nil
}
