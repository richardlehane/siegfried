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

package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/richardlehane/siegfried/cmd/internal/chart"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/sets"
)

const (
	fileString = "[FILE]"
	errString  = "[ERROR]"
	warnString = "[WARN]"
	timeString = "[TIME]"
)

// Logger logs characteristics of the matching process depending on options set by user.
type Logger struct {
	progress, e, warn, known, unknown bool
	fmts                              map[string]bool
	cht                               map[string]map[string]int
	w                                 io.Writer
	start                             time.Time
	// mutate
	fp bool
}

// New creates a new Logger.
func New(opts string) (*Logger, error) {
	lg := &Logger{w: os.Stderr}
	if opts == "" {
		return lg, nil
	}
	var items []string
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
		case "chart", "c":
			lg.cht = make(map[string]map[string]int)
		default:
			items = append(items, o)
		}
	}
	if len(items) > 0 {
		lg.fmts = make(map[string]bool)
		for _, v := range sets.Sets(items...) {
			lg.fmts[v] = true
		}
	}
	if config.Debug() || config.Slow() {
		lg.progress = false // progress reported internally
		config.SetOut(lg.w)
	}
	return lg, nil
}

// IsOut reports if the logger is writing to os.Stdout
func (lg *Logger) IsOut() bool {
	return lg.w == os.Stdout
}

// Elapsed logs time elapsed since logger created.
func (lg *Logger) Elapsed() {
	if !lg.start.IsZero() {
		fmt.Fprintf(lg.w, "%s %v\n", timeString, time.Since(lg.start))
	}
}

// Chart prints a chart of formats matched
func (lg *Logger) Chart() {
	if lg.cht == nil {
		return
	}
	sections := make([]string, 0, len(lg.cht))
	for k, _ := range lg.cht {
		sections = append(sections, k)
	}
	sort.Strings(sections)
	fieldT, fieldR := make(map[string]int), make(map[int][]string)
	for _, m := range lg.cht {
		for k, v := range m {
			fieldT[k] += v
		}
	}
	for k, v := range fieldT {
		fieldR[v] = append(fieldR[v], k)
	}
	fields, totals := make([]string, 0, len(fieldT)), make([]int, 0, len(fieldR))
	for k, v := range fieldR {
		totals = append(totals, k)
		sort.Strings(v)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(totals)))
	for _, k := range totals {
		fields = append(fields, fieldR[k]...)
	}
	fmt.Fprint(lg.w, chart.Chart("[Chart]", sections, fields, map[string]bool{}, lg.cht))
}

// Close prints and chart and time elapsed
func (lg *Logger) Close() {
	lg.Chart()
	lg.Elapsed()
}

// Progress prints file name and resets.
func (lg *Logger) Progress(p string) {
	lg.fp = false
	if lg.progress {
		lg.fp = printFile(lg.fp, lg.w, p)
	}
}

// Error logs errors.
func (lg *Logger) Error(p string, e error) {
	if lg.e && e != nil {
		lg.fp = printFile(lg.fp, lg.w, p)
		fmt.Fprintf(lg.w, "%s %v\n", errString, e)
	}
}

// IDs logs warnings, known, unknown and reports matches against supplied formats.
func (lg *Logger) IDs(p string, ids []core.Identification) {
	if !lg.warn && !lg.known && !lg.unknown && lg.fmts == nil && lg.cht == nil {
		return
	}
	var kn bool
	for _, id := range ids {
		if id.Known() {
			kn = true
		}
		if lg.warn {
			if w := id.Warn(); w != "" {
				lg.fp = printFile(lg.fp, lg.w, p)
				fmt.Fprintf(lg.w, "%s %s\n", warnString, w)
			}
		}
		if lg.fmts[id.String()] {
			fmt.Fprintln(lg.w, abs(p))
		}
		if lg.cht != nil {
			if lg.cht[id.Values()[0]] == nil {
				lg.cht[id.Values()[0]] = make(map[string]int)
			}
			lg.cht[id.Values()[0]][id.String()]++
		}
	}
	if (lg.known && kn) || (lg.unknown && !kn) {
		fmt.Fprintln(lg.w, abs(p))
	}
}

// helpers
func abs(p string) string {
	np, _ := filepath.Abs(p)
	if np == "" {
		return p
	}
	return np
}

func printFile(done bool, w io.Writer, p string) bool {
	if !done {
		fmt.Fprintf(w, "%s %s\n", fileString, abs(p))
	}
	return true
}
