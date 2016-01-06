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

// Archive is a file format capable of decompression by sf.
type Archive int

const (
	None Archive = iota // None means the format cannot be decompressed by sf.
	Zip
	Gzip
	Tar
	ARC
	WARC
)

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
