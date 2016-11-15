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
	"io"
	"os"
	"strings"
	"time"

	"github.com/richardlehane/siegfried/pkg/config"
)

type logger struct {
	progress, e, warn, known, unknown bool
	w                                 io.Writer
	start                             time.Time
}

func newLogger(opts string) (*logger, error) {
	lg := &logger{w: os.Stderr}
	if opts == "" {
		return lg, nil
	}
	for _, o := range strings.Split(opts, ",") {
		switch o {
		case "stderr":
		case "stdout", "out", "o":
			lg.w = os.Stdout
		case "progress", "p":
			lg.progress = true
		case "time", "t":
			lg.start = time.Now()
		case "error", "err", "e":
			lg.e = true
		case "warning", "warn", "w":
			lg.warn = true
		case "debug", "d":
			config.SetDebug()
		case "slow", "s":
			config.SetSlow()
		case "unknown", "u":
			lg.unknown = true
		case "known", "k":
			lg.known = true
		default:
			return nil, fmt.Errorf("unknown -log input %s; expect be comma-separated list of stdout,out,o,progress,p,error,err,e,warning,warn,w,debug,d,slow,s,unknown,u,known,k", opts)
		}
	}
	if config.Debug() || config.Slow() {
		lg.progress = false // progress reported internally
		config.SetOut(lg.w)
	}
	return lg, nil
}
