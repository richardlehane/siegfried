//go:build !brew && !archivematica && !js
// +build !brew,!archivematica,!js

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
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// the default Home location is a "siegfried" folder in the user's application data folder, which can be overridden by setting the SIEGFRIED_HOME environment variable
func init() {
	if home, ok := os.LookupEnv("SIEGFRIED_HOME"); ok {
		siegfried.home = home
	} else {
		// if a home directory already exists in the legacy location continue using it, otherwise default to a XDG-aware OS-specific application data directory
		siegfried.home = filepath.Join(xdg.Home, "siegfried")
		if _, err := os.Stat(siegfried.home); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				siegfried.home = filepath.Join(xdg.DataHome, "siegfried")
			} else {
				log.Fatal(err)
			}
		}
	}
}
