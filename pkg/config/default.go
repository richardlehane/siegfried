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
	"log"
	"os/user"
	"path/filepath"
)

// the default Home location is a "siegfried" folder in the user's $HOME
func init() {
	current, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	siegfried.home = filepath.Join(current.HomeDir, "siegfried")
}
