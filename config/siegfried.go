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

// Core siegfried defaults
package config

import (
	"net/http"
	"path/filepath"
	"time"
)

var siegfried = struct {
	version   [3]int // Siegfried version (i.e. of the sf tool)
	home      string // Home directory used by both sf and roy tools
	signature string // Name of signature file
	magic     []byte // Magic bytes to ID signature file
	// Defaults for processing bytematcher signatures. These control the segmentation.
	distance int // The acceptable distance between two frames before they will be segmented (default is 8192)
	rng      int // The acceptable range between two frames before they will be segmented (default is 0-2049)
	choices  int // The acceptable number of plain sequences generated from a single segment
	// Config for using the update service.
	updateURL       string // URL for the update service (a JSON file that indicates whether update necessary and where can be found)
	updateTimeout   time.Duration
	updateTransport *http.Transport
	// Archivematica format policy registry service
	fpr string
	// DEBUG and SLOW modes
	debug      bool
	slow       bool
	checkpoint int64
}{
	version:         [3]int{1, 4, 1},
	signature:       "pronom.sig",
	magic:           []byte{'s', 'f', 0x00, 0xFF},
	distance:        8192,
	rng:             4096,
	choices:         128,
	updateURL:       "http://www.itforarchivists.com/siegfried/update",
	updateTimeout:   30 * time.Second,
	updateTransport: &http.Transport{Proxy: http.ProxyFromEnvironment},
	fpr:             "/tmp/siegfried",
	checkpoint:      524288, // point at which to report slow signatures (must be power of two)
}

// GETTERS

func Version() [3]int {
	return siegfried.version
}

func Home() string {
	return siegfried.home
}

func Signature() string {
	if filepath.Dir(siegfried.signature) == "." {
		return filepath.Join(siegfried.home, siegfried.signature)
	}
	return siegfried.signature
}

func SignatureBase() string {
	return siegfried.signature
}

func Magic() []byte {
	return siegfried.magic
}

func Distance() int {
	return siegfried.distance
}

func Range() int {
	return siegfried.rng
}

func Choices() int {
	return siegfried.choices
}

func UpdateOptions() (string, time.Duration, *http.Transport) {
	return siegfried.updateURL, siegfried.updateTimeout, siegfried.updateTransport
}

func Fpr() string {
	return siegfried.fpr
}

func Debug() bool {
	return siegfried.debug
}

func Slow() bool {
	return siegfried.slow
}

func Checkpoint(i int64) bool {
	return i == siegfried.checkpoint
}

// SETTERS

func SetHome(h string) {
	siegfried.home = h
}

func SetSignature(s string) {
	siegfried.signature = s
}

func SetDistance(i int) func() private {
	return func() private {
		siegfried.distance = i
		return private{}
	}
}

func SetRange(i int) func() private {
	return func() private {
		siegfried.rng = i
		return private{}
	}
}

func SetChoices(i int) func() private {
	return func() private {
		siegfried.choices = i
		return private{}
	}
}

func SetDebug() {
	siegfried.debug = true
}

func SetSlow() {
	siegfried.slow = true
}
