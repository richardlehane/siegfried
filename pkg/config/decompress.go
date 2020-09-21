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
	"fmt"
)

// Archive is a file format capable of decompression by sf.
type Archive int

// Archive type enum.
const (
	None Archive = iota // None means the format cannot be decompressed by sf.
	Zip                 // Zip describes a Zip type archive.
	Gzip                // Gzip describes a Gzip type archive.	.
	Tar                 // Tar describes a Tar type archive
	ARC                 // ARC describes an ARC web archive.
	WARC                // WARC describes a WARC web archive.
)

const (
	zipArc  = "zip"
	tarArc  = "tar"
	gzipArc = "gzip"
	warcArc = "warc"
	arcArc  = "arc"
)

// ArcZipTypes returns a string array with all Zip identifiers Siegfried
// can match and decompress.
func ArcZipTypes() []string {
	return []string{
		pronom.zip,
		mimeinfo.zip,
		loc.zip,
	}
}

// ArcGzipTypes returns a string array with all Gzip identifiers
// Siegfried can match and decompress.
func ArcGzipTypes() []string {
	return []string{
		pronom.gzip,
		mimeinfo.gzip,
		wikidata.gzip,
	}
}

// ArcTarTypes returns a string array with all Tar identifiers Siegfried
// can match and decompress.
func ArcTarTypes() []string {
	return []string{
		pronom.tar,
		mimeinfo.tar,
		wikidata.tar,
	}
}

// ArcArcTypes returns a string array with all Arc identifiers Siegfried
// can match and decompress.
func ArcArcTypes() []string {
	return []string{
		pronom.arc,
		pronom.arc1_1,
		mimeinfo.arc,
		loc.arc,
		wikidata.arc,
		wikidata.arc1_1,
	}
}

// ArcWarcTypes returns a string array with all Arc identifiers
// Siegfried can match and decompress.
func ArcWarcTypes() []string {
	return []string{
		pronom.warc,
		mimeinfo.warc,
		loc.warc,
		wikidata.warc,
	}
}

// ListAllArcTypes returns a list of archive file-format extensions that
// can be used to filter the files Siegfried will decompress to identify
// the contents of.
func ListAllArcTypes() string {
	return fmt.Sprintf("%s, %s, %s, %s, %s",
		zipArc,
		tarArc,
		gzipArc,
		warcArc,
		arcArc,
	)
}

func (a Archive) String() string {
	switch a {
	case Zip:
		return "zip"
	case Gzip:
		return "gzip"
	case Tar:
		return "tar"
	case ARC:
		return "ARC"
	case WARC:
		return "WARC"
	}
	return ""
}
