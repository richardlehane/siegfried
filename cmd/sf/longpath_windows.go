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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/richardlehane/siegfried"
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

func multiIdentifyS(w writer, s *siegfried.Siegfried, root, orig string, norecurse bool) error {
	wf := func(path string, info os.FileInfo, err error) error {
		var retry bool
		var lp, sp string
		if *throttlef > 0 {
			<-throttle.C
		}
		if err != nil {
			info, err = retryStat(path, err) // retry stat in case is a windows long path error
			if err != nil {
				return fmt.Errorf("walking %s; got %v", path, err)
			}
			lp, sp = longpath(path), path
			retry = true
		}
		if info.IsDir() {
			if norecurse && path != root {
				return filepath.SkipDir
			}
			if retry { // if a dir long path, restart the recursion with a long path as the new root
				return multiIdentifyS(w, s, lp, sp, norecurse)
			}
			if *droido {
				w.writeFile(shortpath(path, orig), -1, info.ModTime().Format(time.RFC3339), nil, nil, nil) // write directory with a -1 size for droid output only
			}
			return nil
		}
		identifyFile(w, s, shortpath(path, orig), info.Size(), info.ModTime().Format(time.RFC3339))
		return nil
	}
	return filepath.Walk(root, wf)
}
