// +build !windows

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
	"time"
)

func retryOpen(path string, err error) (*os.File, error) {
	return nil, err
}

func retryStat(path string, err error) (os.FileInfo, error) {
	return nil, err
}

func identify(ctxts chan *context, root, orig string, coerr, norecurse, droid bool, gf getFn) error {
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if *throttlef > 0 {
			<-throttle.C
		}
		if err != nil {
			if coerr {
				printFile(ctxts, gf(path, "", "", 0), WalkError{path, err})
				return nil
			}
			return WalkError{path, err}
		}
		if info.IsDir() {
			if norecurse && path != root {
				return filepath.SkipDir
			}
			if droid {
				printFile(ctxts, gf(path, "", info.ModTime().Format(time.RFC3339), -1), nil)
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			printFile(ctxts, gf(path, "", info.ModTime().Format(time.RFC3339), info.Size()), ModeError(info.Mode()))
			return nil
		}
		// zero user read permissions mask, octal 400 (decimal 256)
		if info.Mode()&(256) == 0 {
			printFile(ctxts, gf(path, "", info.ModTime().Format(time.RFC3339), info.Size()), ModeError(info.Mode()))
			return nil
		}
		identifyFile(gf(path, "", info.ModTime().Format(time.RFC3339), info.Size()), ctxts, gf)
		return nil
	}
	return filepath.Walk(root, walkFunc)
}
