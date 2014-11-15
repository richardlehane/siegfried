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
	droid            string // name of droid file e.g. DROID_SignatureFile_V78.xml
	container        string // e.g. container-signature-19770502.xml
	reports          string // directory where PRONOM reports are stored
	extend           []string
	single           string
	harvestURL       string
	harvestTimeout   time.Duration
	harvestTransport *http.Transport
}{
	reports:          "pronom",
	harvestURL:       "http://apps.nationalarchives.gov.uk/pronom/",
	harvestTimeout:   120 * time.Second,
	harvestTransport: &http.Transport{Proxy: http.ProxyFromEnvironment},
}

// GETTERS

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
			idx = i
		}
	}
	return hits[idx], nil
}

func Reports() string {
	if filepath.Dir(pronom.reports) == "." {
		return filepath.Join(siegfried.home, pronom.reports)
	}
	return pronom.reports
}

func Extend() []string {
	ret := make([]string, len(pronom.extend))
	for i, v := range pronom.extend {
		if filepath.Dir(v) == "." {
			ret[i] = filepath.Join(Reports(), v)
		} else {
			ret[i] = v
		}
	}
	return ret
}

func Single() (bool, string) {
	if pronom.single == "" {
		return false, ""
	}
	return true, pronom.single
}

func HarvestOptions() (string, time.Duration, *http.Transport) {
	return pronom.harvestURL, pronom.harvestTimeout, pronom.harvestTransport
}

// SETTERS

func SetDroid(d string) func() private {
	return func() private {
		pronom.droid = d
		return private{}
	}
}

func SetContainer(c string) func() private {
	return func() private {
		pronom.container = c
		return private{}
	}
}

func SetReports(r string) func() private {
	return func() private {
		pronom.reports = r
		return private{}
	}
}

func SetExtend(e string) func() private {
	return func() private {
		pronom.extend = strings.Split(e, ",")
		return private{}
	}
}

func SetSingle(s string) func() private {
	return func() private {
		pronom.single = s
		return private{}
	}
}

// unlike other setters, these are only relevant in the roy tool so can't be converted to the Option type

func SetHarvestTimeout(d time.Duration) {
	pronom.harvestTimeout = d
}

func SetHarvestTransport(t *http.Transport) {
	pronom.harvestTransport = t
}
