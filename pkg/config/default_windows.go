//go:build !brew && !archivematica && !js

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

	"golang.org/x/sys/windows"
)

func userDataDir(home string) string {
	path, _ := windows.KnownFolderPath(windows.FOLDERID_LocalAppData, windows.KF_FLAG_DEFAULT|windows.KF_FLAG_DONT_VERIFY)
	if path == "" {
		path, _ = windows.KnownFolderPath(windows.FOLDERID_LocalAppData, windows.KF_FLAG_DEFAULT_PATH|windows.KF_FLAG_DONT_VERIFY)
	}
	if path == "" {
		dataDir, found := os.LookupEnv("LOCALAPPDATA")
		if found && dataDir != "" {
			path = dataDir
		} else {
			path = filepath.Join("AppData", "Local")
		}
	}
	return xdgPath(home, path)
}
