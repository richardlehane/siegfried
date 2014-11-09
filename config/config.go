// Copyright 2014 Richard Lehane. All rights reserved.
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

// Package config sets up defaults used by both the SF and R2D2 tools
// Core siegfried defaults are here, PRONOM-ish defaults are in the pronom file, signature defaults in the signature file.
// Defaults set at runtime (Home location) are in the "default" file.
// All config options can be overriden with build flags e.g. the brew and archivematica files.
package config

import (
	"net/http"
	"path/filepath"
	"time"
)

var Siegfried = struct {
	Version          [3]int // Siegfried version (i.e. of the sf tool)
	Home             string // Home directory used by both sf and r2d2 tools
	Signature        string // Name of signature file
	SignatureVersion int    // Version of the signature file (this is used for the update service)
	// Defaults for processing bytematcher signatures. These control the segmentation.
	Distance  int // The acceptable distance between two frames before they will be segmented (default is 8192)
	Range     int // The acceptable range between two frames before they will be segmented (default is 0-2049)
	Choices   int // The acceptable number of plain sequences generated from a single segment
	VarLength int // The acceptable length of a variable byte sequence (longer the better to reduce false matches)
	// Config for using the update service.
	UpdateURL       string // URL for the update service (a JSON file that indicates whether update necessary and where can be found)
	UpdateTimeout   time.Duration
	UpdateTransport *http.Transport
}{
	Version:          [3]int{0, 6, 0},
	Signature:        "pronom.gob",
	SignatureVersion: 5,
	Distance:         8192,
	Range:            2049,
	Choices:          64,
	VarLength:        1,
	UpdateURL:        "http://www.itforarchivists.com/siegfried/update",
	UpdateTimeout:    30 * time.Second,
	UpdateTransport:  &http.Transport{Proxy: http.ProxyFromEnvironment},
}

func Signature() string {
	return filepath.Join(Siegfried.Home, Siegfried.Signature)
}
