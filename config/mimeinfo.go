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

var mimeinfo = struct {
	mi   string
	zip  string
	gzip string
	tar  string
	arc  string
	warc string
	text string
}{
	zip:  "application/zip",
	gzip: "application/gzip",
	tar:  "application/x-tar",
	arc:  "application/x-arc",
	warc: "application/x-warc",
	text: "text/plain",
}

// MIMEInfo returns the location of the MIMEInfo signature file.
func MIMEInfo() string {
	return filepath.Join(siegfried.home, mimeinfo.mi)
}

func ZipMIME() string {
	return mimeinfo.zip
}

func TextMIME() string {
	return mimeinfo.text
}

func SetMIMEInfo(mi string) func() private {
	return func() private {
		switch mi {
		case "tika", "tika-mimetypes.xml":
			mimeinfo.mi = "tika-mimetypes.xml"
			if identifier.name == "" {
				identifier.name = "tika"
			}
		case "freedesktop", "freedesktop.org", "freedesktop.org.xml":
			mimeinfo.mi = "freedesktop.org.xml"
			if identifier.name == "" {
				identifier.name = "freedesktop.org"
			}
		}
		return private{}
	}
}
