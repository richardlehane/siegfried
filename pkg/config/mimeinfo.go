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

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

var mimeinfo = struct {
	mi       string
	name     string
	versions string
	zip      string
	gzip     string
	tar      string
	arc      string
	warc     string
	text     string
}{
	versions: "mime-info.json",
	zip:      "application/zip",
	gzip:     "application/gzip",
	tar:      "application/x-tar",
	arc:      "application/x-arc",
	warc:     "application/x-warc",
	text:     "text/plain",
}

// MIMEInfo returns the location of the MIMEInfo signature file.
func MIMEInfo() string {
	if filepath.Dir(mimeinfo.mi) == "." {
		return filepath.Join(siegfried.home, mimeinfo.mi)
	}
	return mimeinfo.mi
}

func MIMEVersion() []string {
	byt, err := ioutil.ReadFile(filepath.Join(siegfried.home, mimeinfo.versions))
	m := make(map[string][]string)
	if err == nil {
		err = json.Unmarshal(byt, &m)
		if err == nil {
			return m[mimeinfo.mi]
		}
	}
	return nil
}

func ZipMIME() string {
	return mimeinfo.zip
}

func TextMIME() string {
	return mimeinfo.text
}

func SetMIMEInfo(mi string) func() private {
	return func() private {
		wikidata.namespace = "" // reset wikidata to prevent pollution
		loc.fdd = ""            // reset loc to prevent pollution
		switch mi {
		case "tika", "tika-mimetypes.xml":
			mimeinfo.mi = "tika-mimetypes.xml"
			mimeinfo.name = "tika"
		case "freedesktop", "freedesktop.org", "freedesktop.org.xml":
			mimeinfo.mi = "freedesktop.org.xml"
			mimeinfo.name = "freedesktop.org"
		default:
			mimeinfo.mi = mi
			mimeinfo.name = "mimeinfo"
		}
		return private{}
	}
}
