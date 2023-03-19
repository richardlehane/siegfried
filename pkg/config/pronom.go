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

package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var pronom = struct {
	name             string
	droid            string   // name of droid file e.g. DROID_SignatureFile_V78.xml
	container        string   // e.g. container-signature-19770502.xml
	reports          string   // directory where PRONOM reports are stored
	noclass          bool     // omit class from the format info
	doubleup         bool     // include byte signatures for formats that also have container signatures
	extendc          []string //container extensions
	changesURL       string
	harvestURL       string
	harvestTimeout   time.Duration
	harvestThrottle  time.Duration
	harvestTransport *http.Transport
	// archive puids
	zip    string
	tar    string
	gzip   string
	arc    string
	arc1_1 string
	warc   string
	// text puid
	text string
}{
	name:             "pronom",
	reports:          "pronom",
	changesURL:       "http://www.nationalarchives.gov.uk/aboutapps/pronom/release-notes.xml",
	harvestURL:       "http://www.nationalarchives.gov.uk/pronom/",
	harvestTimeout:   120 * time.Second,
	harvestTransport: &http.Transport{Proxy: http.ProxyFromEnvironment},
	zip:              "x-fmt/263",
	tar:              "x-fmt/265",
	gzip:             "x-fmt/266",
	arc:              "x-fmt/219",
	arc1_1:           "fmt/410",
	warc:             "fmt/289",
	text:             "x-fmt/111",
}

// GETTERS

// Droid returns the location of the DROID signature file.
// If not set, infers the latest file.
func Droid() string {
	if pronom.droid == "" {
		droid, err := latest("DROID_SignatureFile_V", ".xml")
		if err != nil {
			return ""
		}
		return filepath.Join(siegfried.home, droid)
	}
	if filepath.Dir(pronom.droid) == "." {
		return filepath.Join(siegfried.home, pronom.droid)
	}
	return pronom.droid
}

// DroidBase returns the base filename of the DROID signature file.
// If not set, infers the latest file.
func DroidBase() string {
	if pronom.droid == "" {
		droid, err := latest("DROID_SignatureFile_V", ".xml")
		if err != nil {
			return ""
		}
		return droid
	}
	return pronom.droid
}

// Container returns the location of the DROID container signature file.
// If not set, infers the latest file.
func Container() string {
	if pronom.container == "" {
		container, err := latest("container-signature-", ".xml")
		if err != nil {
			return ""
		}
		return filepath.Join(siegfried.home, container)
	}
	if filepath.Dir(pronom.container) == "." {
		return filepath.Join(siegfried.home, pronom.container)
	}
	return pronom.container
}

// ContainerBase returns the base filename of the DROID container signature file.
// If not set, infers the latest file.
func ContainerBase() string {
	if pronom.container == "" {
		container, err := latest("container-signature-", ".xml")
		if err != nil {
			return ""
		}
		return container
	}
	return pronom.container
}

func latest(prefix, suffix string) (string, error) {
	var hits []string
	var ids []int
	files, err := ioutil.ReadDir(siegfried.home)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		nm := f.Name()
		if strings.HasPrefix(nm, prefix) && strings.HasSuffix(nm, suffix) {
			hits = append(hits, nm)
			id, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(nm, prefix), suffix))
			if err != nil {
				return "", err
			}
			ids = append(ids, id)
		}
	}
	if len(hits) == 0 {
		return "", fmt.Errorf("Config: no file in %s with prefix %s", siegfried.home, prefix)
	}
	if len(hits) == 1 {
		return hits[0], nil
	}
	max, idx := ids[0], 0
	for i, v := range ids[1:] {
		if v > max {
			max = v
			idx = i + 1
		}
	}
	return hits[idx], nil
}

// Reports returns the location of the PRONOM reports directory.
func Reports() string {
	if pronom.reports == "" {
		return ""
	}
	return filepath.Join(siegfried.home, pronom.reports)
}

// NoClass reports whether the noclass flag has been set. This will cause class to be omitted from format infos
func NoClass() bool {
	return pronom.noclass
}

// DoubleUp reports whether the doubleup flag has been set. This will cause byte signatures to be built for formats where container signatures are also provided.
func DoubleUp() bool {
	return pronom.doubleup
}

// ExcludeDoubles takes a slice of puids and a slice of container puids and excludes those that are in the container slice, if nodoubles is set.
func ExcludeDoubles(puids, cont []string) []string {
	return exclude(puids, cont)
}

// ExtendC reports whether a set of container signature extensions has been provided.
func ExtendC() []string {
	return extensionPaths(pronom.extendc)
}

// ChangesURL returns the URL for the PRONOM release notes.
func ChangesURL() string {
	return pronom.changesURL
}

// HarvestOptions reports the PRONOM url, timeout and transport.
func HarvestOptions() (string, time.Duration, time.Duration, *http.Transport) {
	return pronom.harvestURL, pronom.harvestTimeout, pronom.harvestThrottle, pronom.harvestTransport
}

// ZipPuid reports the puid for a zip archive.
func ZipPuid() string {
	return pronom.zip
}

// TextPuid reports the puid for a text file.
func TextPuid() string {
	return pronom.text
}

// SETTERS

// SetDroid sets the name and/or location of the DROID signature file.
// I.e. can provide a full path or a filename relative to the HOME directory.
func SetDroid(d string) func() private {
	return func() private {
		pronom.droid = d
		return private{}
	}
}

// SetContainer sets the name and/or location of the DROID container signature file.
// I.e. can provide a full path or a filename relative to the HOME directory.
func SetContainer(c string) func() private {
	return func() private {
		pronom.container = c
		return private{}
	}
}

// SetNoReports instructs roy to build from the DROID signature file alone (and not from the PRONOM reports).
func SetNoReports() func() private {
	return func() private {
		pronom.reports = ""
		return private{}
	}
}

// SetNoClass causes class to be omitted from the format info
func SetNoClass() func() private {
	return func() private {
		pronom.noclass = true
		return private{}
	}
}

// SetDoubleUp causes byte signatures to be built for formats where container signatures are also provided.
func SetDoubleUp() func() private {
	return func() private {
		pronom.doubleup = true
		return private{}
	}
}

// SetExtendC adds container extension signatures to the build.
func SetExtendC(l []string) func() private {
	return func() private {
		pronom.extendc = l
		return private{}
	}
}

// unlike other setters, these are only relevant in the roy tool so can't be converted to the Option type

// SetHarvestTimeout sets a time limit on PRONOM harvesting.
func SetHarvestTimeout(d time.Duration) {
	pronom.harvestTimeout = d
}

// SetHarvestThrottle sets a throttle value for downloading DROID reports.
func SetHarvestThrottle(d time.Duration) {
	pronom.harvestThrottle = d
}

// SetHarvestTransport sets the PRONOM harvesting transport.
func SetHarvestTransport(t *http.Transport) {
	pronom.harvestTransport = t
}
