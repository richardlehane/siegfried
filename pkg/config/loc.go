// Copyright 2015 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import "path/filepath"

var loc = struct {
	fdd      string
	def      string // default
	nopronom bool
	name     string
	zip      string
	gzip     string // n/a
	tar      string // n/a
	arc      string
	warc     string
	text     string // n/a
}{
	def:  "fddXML.zip",
	name: "loc",
	zip:  "fdd000354",
	arc:  "fdd000235",
	warc: "fdd000236",
}

// LOC returns the location of the LOC signature file.
func LOC() string {
	if loc.fdd == loc.def {
		return filepath.Join(siegfried.home, loc.def)
	}
	if filepath.Dir(loc.fdd) == "." {
		return filepath.Join(siegfried.home, loc.fdd)
	}
	return loc.fdd
}

func ZipLOC() string {
	return loc.zip
}

func NoPRONOM() bool {
	return loc.nopronom
}

func SetNoPRONOM() func() private {
	return func() private {
		loc.nopronom = true
		return private{}
	}
}

func SetLOC(fdd string) func() private {
	return func() private {
		wikidata.namespace = "" // reset wikidata to prevent pollution
		mimeinfo.mi = ""        // reset mimeinfo to prevent pollution
		if fdd == "" {
			fdd = loc.def
		}
		loc.fdd = fdd
		return private{}
	}
}
