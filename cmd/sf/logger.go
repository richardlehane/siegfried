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
	"path/filepath"
	"strings"
	"time"

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/core"
)

// TODO: slow and debug are non-parallel, but other logging functions should work in parallel modes

var lg *logger

type logger struct {
	progress, e, warn, debug, slow, known, unknown bool
	w                                              io.Writer
	start                                          time.Time
	// mutate each file
	path string
	u    bool
	fp   bool // file name already printed
}

func newLogger(opts string) error {
	l := &logger{w: os.Stderr}
	for _, o := range strings.Split(opts, ",") {
		switch o {
		case "stderr":
		case "stdout", "out", "o":
			l.w = os.Stdout
		case "progress", "p":
			l.progress = true
		case "time", "t":
			l.start = time.Now()
		case "error", "err", "e":
			l.e = true
		case "warning", "warn", "w":
			l.warn = true
		case "debug", "d":
			l.debug, l.progress = true, true
			config.SetDebug()
		case "slow", "s":
			l.slow, l.progress = true, true
			config.SetSlow()
		case "unknown", "u":
			l.unknown = true
		case "known", "k":
			l.known = true
		default:
			return fmt.Errorf("unknown -log input %s; expect be comma-separated list of stdout,out,o,progress,p,error,err,e,warning,warn,w,debug,d,slow,s,unknown,u,known,k", opts)
		}
	}
	if config.Debug() || config.Slow() {
		config.SetOut(l.w)
	}
	lg = l
	return nil
}

var (
	fileString = "[FILE]"
	errString  = "[ERROR]"
	warnString = "[WARN]"
	timeString = "[TIME]"
)

func (l *logger) printElapsed() {
	if l == nil || l.start.IsZero() {
		return
	}
	fmt.Fprintf(l.w, "%s %v\n", timeString, time.Since(l.start))
}

func (l *logger) printFile() {
	if !l.fp {
		fmt.Fprintf(l.w, "%s %s\n", fileString, l.path)
		l.fp = true
	}
}

func (l *logger) set(path string) {
	if l == nil {
		return
	}
	l.path, _ = filepath.Abs(path)
	if l.path == "" {
		l.path = path
	}
	if l.progress {
		l.printFile()
	}
}

func (l *logger) err(err error) {
	if l != nil && l.e && err != nil {
		l.printFile()
		fmt.Fprintf(l.w, "%s %v\n", errString, err)
	}
}

func (l *logger) id(i core.Identification) {
	if l == nil {
		return
	}
	if (l.unknown || l.known) && i.Known() {
		l.u = true
	}
	if l.warn {
		if w := i.Warn(); w != "" {
			l.printFile()
			fmt.Fprintf(l.w, "%s %s\n", warnString, w)
		}
	}
	if l.slow || l.debug {
		fmt.Fprintf(l.w, "matched: %s\n", i.String())
	}
}

func (l *logger) reset() {
	if l == nil {
		return
	}
	if (l.known && l.u) || (l.unknown && !l.u) {
		fmt.Fprintln(l.w, l.path)
	}
	l.u, l.fp = false, false
	l.path = ""
}
