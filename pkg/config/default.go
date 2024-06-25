//go:build !js && !brew

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
	"os"
	"path/filepath"
	"strings"
)

// the default Home location is a "siegfried" folder in the user's application data folder, which can be overridden by setting the SIEGFRIED_HOME environment variable
func defaultHome() string {
	// users can override the default home location with the SIEGFRIED_HOME env var
	if siegfried_home, ok := os.LookupEnv("SIEGFRIED_HOME"); ok {
		return siegfried_home
	}
	user_home, _ := os.UserHomeDir()
	// XDG-aware OS-specific application data directory is the default, return it first if it exists
	xdg_home := filepath.Join(userDataDir(user_home), "siegfried")
	if _, err := os.Stat(xdg_home); err == nil {
		return xdg_home
	}
	// if a home directory already exists in the legacy location continue using it
	legacy_home := filepath.Join(user_home, "siegfried")
	if _, err := os.Stat(legacy_home); err == nil {
		return legacy_home
	}
	// check if XDG_DATA_DIRS exist
	if data_dir, found := xdgDataDirs(user_home); found {
		return data_dir
	}
	// if no directories exist, xdg_home is the default
	return xdg_home
}

func xdgPath(home string, defaultPath string) string {
	dataHome, found := os.LookupEnv("XDG_DATA_HOME")
	if found && dataHome != "" {
		if strings.HasPrefix(dataHome, "~") {
			dataHome = filepath.Join(home, strings.TrimPrefix(dataHome, "~"))
		}
		// environment variable might contain variables like $HOME itself, let's expand
		dataHome = os.ExpandEnv(dataHome)
	}
	// XDG Base Directory Specification demands relative paths to be ignored, fall back to default in that case
	if filepath.IsAbs(dataHome) {
		return dataHome
	} else if filepath.IsAbs(defaultPath) {
		return defaultPath
	} else {
		return filepath.Join(home, defaultPath)
	}
}

func xdgDataDirs(home string) (string, bool) {
	data_dirs, found := os.LookupEnv("XDG_DATA_DIRS")
	if found && data_dirs != "" {
		for _, dir := range strings.Split(data_dirs, ":") {
			if strings.HasPrefix(dir, "~") {
				dir = filepath.Join(home, strings.TrimPrefix(dir, "~"))
			}
			// environment variable might contain variables like $HOME itself, let's expand
			dir = os.ExpandEnv(dir)
			data_dir := filepath.Join(dir, "siegfried")
			if _, err := os.Stat(data_dir); err == nil {
				return data_dir, true
			}
		}
		return "", false
	}
	//if XDG_DATA_DIRS unset or empty, try defaults /usr/local/share/ and /usr/share/
	if _, err := os.Stat(filepath.Join("/usr/local/share", "siegfried")); err == nil {
		return filepath.Join("/usr/local/share", "siegfried"), true
	}
	if _, err := os.Stat(filepath.Join("/usr/share", "siegfried")); err == nil {
		return filepath.Join("/usr/share", "siegfried"), true
	}
	return "", false
}
