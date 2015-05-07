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

var decompress struct {
	mode bool
	zip  string
	tar  string
	gzip string
}

func Decompress() bool {
	return decompress.mode
}

func IsCompress(s string) bool {
	switch s {
	default:
		return false
	case decompress.zip, decompress.tar, decompress.gzip:
		return true
	}
}

func IsZip(s string) bool {
	return s == decompress.zip
}

func IsTar(s string) bool {
	return s == decompress.tar
}

func IsGzip(s string) bool {
	return s == decompress.gzip
}

// SETTERS

func SetDecompress() func() private {
	return func() private {
		decompress.zip, decompress.tar, decompress.gzip = pronom.zip, pronom.tar, pronom.gzip
		decompress.mode = true
		return private{}
	}
}
