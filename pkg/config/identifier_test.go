// Copyright 2020 Ross Spencer, Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

// Satisfies the Parseable interface to enable Roy to process Wikidata
// signatures into a Siegfried compatible identifier.

package config

import (
	"testing"
)

// Valid archive UIDs.
var proZipUID = "x-fmt/263"
var locArcUID = "fdd000235"
var mimeTarUID = "application/x-tar"
var mimeWarcUID = "application/x-warc"
var mimeGzipUID = "application/gzip"

// Non-archive UID.
var nonArcUID = "fmt/1000"

// arcTest defines the structure needed for our table driven testing.
type arcTest struct {
	uid    string  // A UID (PUID, FDD) that identifies a zip-type file.
	result Archive // The anticipated result from our test.
}

// isArcTests provide us a slice of tests and results to loop through.
var isArcTests = []arcTest{
	// Positive tests should return valid Archive values.
	arcTest{proZipUID, Zip},
	arcTest{mimeTarUID, Tar},
	arcTest{mimeGzipUID, Gzip},
	arcTest{mimeWarcUID, WARC},
	arcTest{locArcUID, ARC},
	// Negative tests should all return None.
	arcTest{nonArcUID, None},
}

// TestIsArchive tests cases whether we return the correct result when
// testing whether something is an Archive.
func TestIsArchive(t *testing.T) {
	for _, test := range isArcTests {
		arc := IsArchive(test.uid)
		if arc != test.result {
			t.Errorf(
				"Unexpected test result '%s', expected '%s'",
				arc, test.result,
			)
		}
	}
}

var arcTypes = [...]Archive{Zip, Gzip, Tar, ARC, WARC}

const noneType = None

// TestIsArchiveGreaterThanNone is a test to to ensure that legacy
// functions relying on `id.Archive() > config.None`.
func TestIsArchiveGreaterThanNone(t *testing.T) {
	for _, item := range arcTypes {
		if item <= None {
			t.Errorf("Archive is evaluating less than 0")
		}
	}
	if noneType != 0 {
		t.Errorf("Archive 0 type should equal zero not %d", noneType)
	}
}
