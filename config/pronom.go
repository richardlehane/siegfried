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

var Pronom = struct {
	Droid            string // name of droid file e.g. DROID_SignatureFile_V78.xml
	Container        string // e.g. container-signature-19770502.xml
	Reports          string // directory, within home, where PRONOM reports are stored
	HarvestUrl       string
	HarvestTimeout   time.Duration
	HarvestTransport *http.Transport
}{
	Reports:          "pronom",
	HarvestUrl:       "http://apps.nationalarchives.gov.uk/pronom/",
	HarvestTimeout:   120 * time.Second,
	HarvestTransport: &http.Transport{Proxy: http.ProxyFromEnvironment},
}

// Convenience funcs to give full paths to Droid, Container and Reports

func Droid() string {
	return filepath.Join(Siegfried.Home, Pronom.Droid)
}

func Container() string {
	return filepath.Join(Siegfried.Home, Pronom.Container)
}

func Reports() string {
	return filepath.Join(Siegfried.Home, Pronom.Reports)
}

// Scan the Home directory for the most recent DROID & container files
func SetLatest() error {
	droid, err := latest("DROID_SignatureFile_V", ".xml")
	if err != nil {
		return err
	}
	Pronom.Droid = droid
	container, err := latest("container-signature-", ".xml")
	if err != nil {
		return err
	}
	Pronom.Container = container
	return nil
}

func latest(prefix, suffix string) (string, error) {
	var hits []string
	var ids []int
	files, err := ioutil.ReadDir(Siegfried.Home)
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
		return "", fmt.Errorf("Config: no file in %s with prefix %s", Siegfried.Home, prefix)
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
