// Copyright 2015 Richard Lehane. All rights reserved.
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

package main

import (
	"os"
	"path/filepath"
	"strings"
)

// longpath code derived from https://github.com/docker/docker/tree/master/pkg/longpath
// prefix is the longpath prefix for Windows file paths.
const (
	prefix = `\\?\`
	lplen  = 240
)

func longpath(path string) string {
	if !strings.HasPrefix(path, prefix) {
		if strings.HasPrefix(path, `\\`) {
			// This is a UNC path, so we need to add 'UNC' to the path as well.
			path = prefix + `UNC` + path[1:]
		} else {
			abs, err := filepath.Abs(path)
			if err != nil {
				return path
			}
			path = prefix + abs
		}
	}
	return path
}

// attempt to reconstitute original path
func shortpath(long, short string) string {
	if short == "" {
		return long
	}
	i := strings.Index(long, short)
	if i == -1 {
		return long
	}
	return long[i:]
}

func retryStat(path string, err error) (os.FileInfo, error) {
	if len(path) < lplen || strings.HasPrefix(path, prefix) { // already a long path - no point retrying
		return nil, err
	}
	info, e := os.Lstat(longpath(path)) // filepath.Walk uses Lstat not Stat
	if e != nil {
		return nil, err
	}
	return info, nil
}

func retryOpen(path string, err error) (*os.File, error) {
	if len(path) < lplen || strings.HasPrefix(path, prefix) { // already a long path - no point retrying
		return nil, err
	}
	file, e := os.Open(longpath(path))
	if e != nil {
		return nil, err
	}
	return file, nil
}
