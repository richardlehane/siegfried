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
	fdd  string
	def  string
	zip  string
	gzip string
	tar  string
	arc  string
	warc string
	text string
}{
	def:  "fddXML.zip",
	zip:  "application/zip",
	gzip: "application/gzip",
	tar:  "application/x-tar",
	arc:  "application/x-arc",
	warc: "application/x-warc",
	text: "text/plain",
}

// LOC returns the location of the LOC signature file.
func LOC() string {
	if loc.fdd == "" {
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

func TextLOC() string {
	return loc.text
}

func SetLOC(fdd string) func() private {
	return func() private {
		loc.fdd = fdd
		return private{}
	}
}
