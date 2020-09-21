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
	"io"
	"net/http"
	"path/filepath"
	"time"
)

var siegfried = struct {
	version   [3]int // Siegfried version (i.e. of the sf tool)
	home      string // Home directory used by both sf and roy tools
	signature string // Name of signature file
	conf      string // Name of the conf file
	magic     []byte // Magic bytes to ID signature file
	// Defaults for processing bytematcher signatures. These control the segmentation.
	distance   int // The acceptable distance between two frames before they will be segmented (default is 8192)
	rng        int // The acceptable range between two frames before they will be segmented (default is 0-2049)
	choices    int // The acceptable number of plain sequences generated from a single segment
	cost       int // The acceptable cost of a signature segement, in terms of the times that it might match in a worst case
	repetition int // The acceptable repetition within a signature segment, used in combination with cost to determine segmentation.
	// Config for using the update service.
	updateURL       string // URL for the update service (a JSON file that indicates whether update necessary and where can be found)
	updateTimeout   time.Duration
	updateTransport *http.Transport
	// Archivematica format policy registry service
	fpr string
	// DEBUG and SLOW modes
	debug      bool
	slow       bool
	out        io.Writer
	checkpoint int64
	userAgent  string
}{
	version:         [3]int{1, 8, 0},
	signature:       "default.sig",
	conf:            "sf.conf",
	magic:           []byte{'s', 'f', 0x00, 0xFF},
	distance:        8192,
	rng:             4096,
	choices:         128,
	cost:            25600000,
	repetition:      4,
	updateURL:       "https://www.itforarchivists.com/siegfried/update", // "http://localhost:8081/siegfried/update",
	updateTimeout:   30 * time.Second,
	updateTransport: &http.Transport{Proxy: http.ProxyFromEnvironment},
	fpr:             "/tmp/siegfried",
	checkpoint:      524288, // point at which to report slow signatures (must be power of two)
	userAgent:       "siegfried/siegbot (+https://github.com/richardlehane/siegfried)",
}

// GETTERS

// Version reports the siegfried version.
func Version() [3]int {
	return siegfried.version
}

// Home reports the siegfried HOME location (e.g. /usr/home/siegfried).
func Home() string {
	return siegfried.home
}

// Home makes a path local to Home() if it is relative
func Local(base string) string {
	if filepath.Dir(base) == "." {
		return filepath.Join(siegfried.home, base)
	}
	return base
}

// Signature returns the path to the siegfried signature file.
func Signature() string {
	return Local(siegfried.signature)
}

// SignatureBase returns the filename of the siegfried signature file.
func SignatureBase() string {
	return siegfried.signature
}

// Conf returns the path to the siegfried configuration file.
func Conf() string {
	return Local(siegfried.conf)
}

// Magic returns the magic string encoded at the start of a siegfried signature file.
func Magic() []byte {
	return siegfried.magic
}

// Distance is a bytematcher setting. It controls the absolute widths at which segments in signatures are split.
// E.g. if segments are separated by a minimum of 50 and maximum of 100 bytes, the distance is 100.
// A short distance means a smaller Aho Corasick search tree and more patterns to follow-up.
// A long distance means a larger Aho Corasick search tree and more signatures immediately satisfied without follow-up pattern matching.
func Distance() int {
	return siegfried.distance
}

// Range is a bytematcher setting. It controls the relative widths at which segments in signatures are split.
// E.g. if segments are separated by a minimum of 50 and maximum of 100 bytes, the range is 50.
// A small range means a smaller Aho Corasick search tree and more patterns to follow-up.
// A large range means a larger Aho Corasick search tree and more signatures immediately satisfied without follow-up pattern matching.
func Range() int {
	return siegfried.rng
}

// Choices is a bytematcher setting. It controls the number of tolerable strings produced by processing signature segments.
// E.g. signature has two adjoining frames ("PDF") and ("1.1" OR "1.2") it can be processed into two search strings: "PDF1.1" and "PDF1.2".
// A low number of choices means a smaller Aho Corasick search tree and more patterns to follow-up.
// A large of choices means a larger Aho Corasick search tree and more signatures immediately satisfied without follow-up pattern matching.
func Choices() int {
	return siegfried.choices
}

// Choices is a bytematcher setting. It controls the number of tolerable matches in a worst case scenario for a signature segement.
// If this cost is exceeded, then segmentation won't happen and the choices/range/distance preferences will be ignored.
func Cost() int {
	return siegfried.cost
}

// Repetitition is a bytematcher setting. It is used in combination with Cost to determine segmentation.
func Repetition() int {
	return siegfried.repetition
}

// UpdateOptions returns the update URL, timeout and transport for the sf -update command.
func UpdateOptions() (string, time.Duration, *http.Transport) {
	return siegfried.updateURL, siegfried.updateTimeout, siegfried.updateTransport
}

// Fpr reports whether sf is being run in -fpr (Archivematica format policy registry) mode.
func Fpr() string {
	return siegfried.fpr
}

// Debug reports whether debug logging is activated.
func Debug() bool {
	return siegfried.debug
}

// Slow reports whether slow logging is activated.
func Slow() bool {
	return siegfried.slow
}

// Out reports the target for logging messages (STDOUT or STDIN).
func Out() io.Writer {
	return siegfried.out
}

// Checkpoint reports the offset at which slow logging should trigger.
func Checkpoint(i int64) bool {
	return i == siegfried.checkpoint
}

// UserAgent returns the siegbot User-Agent string for http requests.
func UserAgent() string {
	return siegfried.userAgent
}

// SETTERS

// SetHome sets the siegfried HOME location (e.g. /usr/home/siegfried).
func SetHome(h string) {
	siegfried.home = h
}

// SetSignature sets the signature filename or filepath.
func SetSignature(s string) {
	siegfried.signature = s
}

// SetConf sets the configuration filename or filepath.
func SetConf(s string) {
	siegfried.conf = s
}

// SetDistance sets the distance variable for the bytematcher.
func SetDistance(i int) func() private {
	return func() private {
		siegfried.distance = i
		return private{}
	}
}

// SetRange sets the range variable for the bytematcher.
func SetRange(i int) func() private {
	return func() private {
		siegfried.rng = i
		return private{}
	}
}

// SetChoices sets the choices variable for the bytematcher.
func SetChoices(i int) func() private {
	return func() private {
		siegfried.choices = i
		return private{}
	}
}

// SetCost sets the cost variable for the bytematcher.
func SetCost(i int) func() private {
	return func() private {
		siegfried.cost = i
		return private{}
	}
}

// SetRepetition sets the repetitition variable for the bytematcher.
func SetRepetition(i int) func() private {
	return func() private {
		siegfried.repetition = i
		return private{}
	}
}

// SetDebug sets degub logging on.
func SetDebug() {
	siegfried.debug = true
}

// SetSlow sets slow logging on.
func SetSlow() {
	siegfried.slow = true
}

// SetOut sets the target for logging.
func SetOut(o io.Writer) {
	siegfried.out = o
}
