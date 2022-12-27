package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var DataPath string = filepath.Join("..", "..", "cmd", "roy", "data")

// setMinimalParams configures our tests to use the minimal fixtures in
// the test data directory.
func setMinimalParams() {
	config.SetDroid("DROID_minimal.xml")
	config.SetPRONOMReportsDir("pronom_minimal")
}

// resetParams undoes the config used by PRONOM minimal tests.
func resetMinimalParams() {
	config.SetDroid("")
	config.SetPRONOMReportsDir("pronom")
}

// pronomIdentificationTests provides our structure for table driven tests.
type pronomIdentificationTests struct {
	identiifer string
	puid       string
	label      string
	version    string
	mime       string
	types      string
	details    string
	error      string
}

var pronomIDs = []pronomIdentificationTests{
	pronomIdentificationTests{
		"pronom",
		"fmt/1",
		"Broadcast WAVE",
		"0 Generic",
		"audio/x-wav",
		"Audio",
		"extension match wav; byte match at [[0 12] [32 356]]",
		"",
	},
	pronomIdentificationTests{
		"pronom",
		"fmt/11",
		"Portable Network Graphics",
		"1.0",
		"image/png",
		"Image (Raster)",
		"extension match png; byte match at [[0 16] [16 12]]",
		"",
	},
	pronomIdentificationTests{
		"pronom",
		"fmt/14",
		"Acrobat PDF 1.X - Portable Document Format",
		"1.0",
		"application/pdf",
		"Page Description",
		"extension match pdf; byte match at [[0 8] [8 5]]",
		"",
	},
	pronomIdentificationTests{
		"pronom",
		"fmt/3",
		"Graphics Interchange Format",
		"87a",
		"image/gif",
		"Image (Raster)",
		"extension match gif; byte match at [[0 6] [6 1]]",
		"",
	},
	pronomIdentificationTests{
		"pronom",
		"fmt/5",
		"Audio/Video Interleaved Format",
		"",
		"video/x-msvideo",
		"Audio, Video",
		"extension match avi; byte match at [[0 12] [32 16] [68 12]]",
		"",
	},
}

// TestPronom looks to see if PRONOM identification results for a
// minimized PRONOM dataset are correct and contain the information we
// anticipate.
func TestPronom(t *testing.T) {

	var sf *siegfried.Siegfried
	sf = siegfried.New()

	config.SetHome(DataPath)
	if config.Reports() == "" {
		t.Error("PRONOM reports are not set, this set requires reports")
	}

	setMinimalParams()

	identifier, err := pronom.New()
	if err != nil {
		t.Errorf("Error creating new PRONOM identifier: %s", err)
	}

	sf.Add(identifier)

	testFiles := filepath.Join(DataPath, "testdata", "minimal-pronom-skeletons")

	testDirListing, err := os.ReadDir(testFiles)
	if err != nil {
		t.Fatalf("Error reading test files directory: %s", err)
	}

	const resultLen int = 8
	results := make([]pronomIdentificationTests, 0)

	for _, val := range testDirListing {
		testFilePath := filepath.Join(testFiles, val.Name())
		file, err := os.Open(testFilePath)
		if err != nil {
			t.Fatalf("Problem opening test file: %s", err)
		}
		defer file.Close()
		res, _ := sf.Identify(file, testFilePath, "")
		result := res[0].Values()
		if len(result) != resultLen {
			t.Errorf("Result len: %d not %d", len(result), resultLen)
		}

		idResult := pronomIdentificationTests{
			result[0], // identifier
			result[1], // PUID
			result[2], // label
			result[3], // version
			result[4], // mime
			result[5], // types
			result[6], // details
			result[7], // error
		}
		results = append(results, idResult)
	}

	// Sort expected results and received results to make them
	// comparable.
	sort.Slice(pronomIDs, func(i, j int) bool {
		return pronomIDs[i].puid < pronomIDs[j].puid
	})
	sort.Slice(results, func(i, j int) bool {
		return results[i].puid < results[j].puid
	})

	// Compare results on a result by result basis.
	for idx, res := range results {
		if !reflect.DeepEqual(res, pronomIDs[idx]) {
			t.Errorf("Results not equal for %s", res.puid)
		}
	}
	resetMinimalParams()
}
