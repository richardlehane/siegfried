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

package main

import (
	"bufio"
	"flag"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/internal/siegreader"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
	/*// Uncomment to build with profiler
	"net/http"
	_ "net/http/pprof"
	*/)

// defaults
const (
	maxProcs     = 1024
	defaultProcs = 8

	fileString = "[FILE]"
	errString  = "[ERROR]"
	warnString = "[WARN]"
	timeString = "[TIME]"
)

// flags
var (
	update    = flag.Bool("update", false, "update or install the default signature file")
	version   = flag.Bool("version", false, "display version information")
	logf      = flag.String("log", "error", "log errors, warnings, debug or slow output, knowns or unknowns to stderr or stdout e.g. -log error,warn,unknown,stdout")
	nr        = flag.Bool("nr", false, "prevent automatic directory recursion")
	csvo      = flag.Bool("csv", false, "CSV output format")
	jsono     = flag.Bool("json", false, "JSON output format")
	droido    = flag.Bool("droid", false, "DROID CSV output format")
	sig       = flag.String("sig", config.SignatureBase(), "set the signature file")
	home      = flag.String("home", config.Home(), "override the default home directory")
	serve     = flag.String("serve", "", "start siegfried server e.g. -serve localhost:5138")
	multi     = flag.Int("multi", defaultProcs, "set number of parallel file ID processes")
	archive   = flag.Bool("z", false, "scan archive formats (zip, tar, gzip, warc, arc)")
	hashf     = flag.String("hash", "", "calculate file checksum with hash algorithm; options "+hashChoices)
	throttlef = flag.Duration("throttle", 0, "set a time to wait between scanning files e.g. 50ms")
)

var throttle *time.Ticker

type context struct {
	s  *siegfried.Siegfried
	w  writer
	wg *sync.WaitGroup
	// opts
	h hash.Hash
	z bool
	// info
	path string
	mime string
	mod  string
	sz   int64
	// results
	res chan results
}

type results struct {
	err error
	cs  []byte
	ids []core.Identification
}

func newContext(s *siegfried.Siegfried, w writer, wg *sync.WaitGroup, h hash.Hash, z bool, path, mime, mod string, sz int64) *context {
	return &context{
		s:    s,
		w:    w,
		wg:   wg,
		h:    h,
		z:    z,
		path: path,
		mime: mime,
		mod:  mod,
		sz:   sz,
		res:  make(chan results, 1),
	}
}

func printer(ctxts chan *context, lg *logger) {
	// helpers for logging
	abs := func(p string) string {
		np, _ := filepath.Abs(p)
		if np == "" {
			return p
		}
		return np
	}
	printFile := func(done bool, p string) bool {
		if !done {
			fmt.Fprintf(lg.w, "%s %s\n", fileString, abs(p))
		}
		return true
	}
	for ctx := range ctxts {
		var fp bool // just print FILE once in log
		// log progress
		if lg.progress {
			fp = printFile(fp, ctx.path)
		}
		// block on the results
		res := <-ctx.res
		// log error
		if lg.e && res.err != nil {
			fp = printFile(fp, ctx.path)
			fmt.Fprintf(lg.w, "%s %v\n", errString, res.err)
		}
		// log warnings, known, unknown and report matches for slow or debug
		if lg.warn || lg.known || lg.unknown {
			var kn bool
			for _, id := range res.ids {
				if id.Known() {
					kn = true
				}
				if lg.warn {
					if w := id.Warn(); w != "" {
						fp = printFile(fp, ctx.path)
						fmt.Fprintf(lg.w, "%s %s\n", warnString, w)
					}
				}
			}
			if (lg.known && kn) || (lg.unknown && !kn) {
				fmt.Fprintln(lg.w, abs(ctx.path))
			}
		}
		// write the result
		ctx.w.writeFile(ctx.path, ctx.sz, ctx.mod, res.cs, res.err, res.ids)
		ctx.wg.Done()
	}
}

// identify() defined in longpath.go and longpath_windows.go

func readFile(ctx *context, ctxts chan *context) {
	f, err := os.Open(ctx.path)
	if err != nil {
		f, err = retryOpen(ctx.path, err) // retry open in case is a windows long path error
		if err != nil {
			ctx.res <- results{err, nil, nil}
			ctx.wg.Add(1)
			ctxts <- ctx
			return
		}
	}
	identifyRdr(f, ctx, ctxts)
	f.Close()
}

func identifyFile(ctx *context, ctxts chan *context) {
	ctx.wg.Add(1)
	ctxts <- ctx
	if ctx.z || config.Slow() || config.Debug() {
		readFile(ctx, ctxts)
		return
	}
	go func() {
		ctx.wg.Add(1)
		readFile(ctx, ctxts)
		ctx.wg.Done()
	}()
}

func identifyRdr(r io.Reader, ctx *context, ctxts chan *context) {
	b, berr := ctx.s.Buffer(r)
	defer ctx.s.Put(b)
	c, err := ctx.s.IdentifyBuffer(b, berr, ctx.path, ctx.mime)
	if c == nil {
		ctx.res <- results{err, nil, nil}
		return
	}
	// calculate checksum
	var cs []byte
	if ctx.h != nil {
		var i int64
		l := ctx.h.BlockSize()
		for ; ; i += int64(l) {
			buf, _ := b.Slice(i, l)
			if buf == nil {
				break
			}
			ctx.h.Write(buf)
		}
		cs = ctx.h.Sum(nil)
	}
	// decompress if an archive format
	ids, arc := sliceIDs(c)
	ctx.res <- results{err, cs, ids}
	if !ctx.z || arc == config.None {
		return
	}
	var d decompressor
	switch arc {
	case config.Zip:
		d, err = newZip(siegreader.ReaderFrom(b), ctx.path, ctx.sz)
	case config.Gzip:
		d, err = newGzip(b, ctx.path)
	case config.Tar:
		d, err = newTar(siegreader.ReaderFrom(b), ctx.path)
	case config.ARC:
		d, err = newARC(siegreader.ReaderFrom(b), ctx.path)
	case config.WARC:
		d, err = newWARC(siegreader.ReaderFrom(b), ctx.path)
	}
	if err != nil {
		ctx.res <- results{fmt.Errorf("failed to decompress, got: %v", err), nil, nil}
		return
	}
	for err = d.next(); err == nil; err = d.next() {
		if _, ok := ctx.w.(*droidWriter); ok {
			for _, v := range d.dirs() {
				dctx := newContext(ctx.s, ctx.w, ctx.wg, nil, false, v, "", "", -1)
				dctx.res <- results{nil, nil, nil}
				ctx.wg.Add(1)
				ctxts <- dctx
			}
		}
		nctx := newContext(ctx.s, ctx.w, ctx.wg, nil, false, d.path(), d.mime(), d.mod(), d.size())
		nctx.wg.Add(1)
		ctxts <- nctx
		identifyRdr(d.reader(), nctx, ctxts)
	}
}

func main() {
	flag.Parse()
	/*//UNCOMMENT TO RUN PROFILER
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()*/
	// configure home and signature if not default
	if *home != config.Home() {
		config.SetHome(*home)
	}
	if *sig != config.SignatureBase() {
		config.SetSignature(*sig)
	}
	// handle -update
	if *update {
		msg, err := updateSigs()
		if err != nil {
			log.Fatalf("[FATAL] failed to update signature file, %v", err)
		}
		fmt.Println(msg)
		return
	}
	// handle -hash error
	checksum := getHash(*hashf)
	if *hashf != "" && checksum == nil {
		log.Fatalf("[FATAL] invalid hash type; choose from %s", hashChoices)
	}
	// load and handle signature errors
	s, err := siegfried.Load(config.Signature())
	if err != nil {
		log.Fatalf("[FATAL] error loading signature file, got: %v", err)
	}
	// handle -version
	if *version {
		version := config.Version()
		fmt.Printf("siegfried %d.%d.%d\n%s", version[0], version[1], version[2], s)
		return
	}
	// handle -fpr
	if *fprflag {
		log.Printf("FPR server started at %s. Use CTRL-C to quit.\n", config.Fpr())
		serveFpr(config.Fpr(), s)
		return
	}
	// check -multi
	if *multi > maxProcs || *multi < 1 {
		*multi = defaultProcs
	}
	// start logger
	lg, err := newLogger(*logf)
	if err != nil {
		log.Fatalln(err)
	}
	if config.Slow() || config.Debug() {
		if *serve != "" || *fprflag {
			log.Fatalln("[FATAL] debug and slow logging cannot be run in server mode")
		}
	}
	// start throttle
	if *throttlef != 0 {
		throttle = time.NewTicker(*throttlef)
		defer throttle.Stop()
	}
	// start the printer
	ctxts := make(chan *context, *multi)
	go printer(ctxts, lg)
	// handle -serve
	if *serve != "" {
		log.Printf("Starting server at %s. Use CTRL-C to quit.\n", *serve)
		listen(*serve, s, ctxts)
		return
	}
	// handle no file/directory argument
	if flag.NArg() != 1 {
		close(ctxts)
		log.Fatalln("[FATAL] expecting a single file or directory argument")
	}
	// set default writer
	var w writer
	switch {
	case *csvo:
		w = newCSV(os.Stdout)
	case *jsono:
		w = newJSON(os.Stdout)
	case *droido:
		w = newDroid(os.Stdout)
		if len(s.Fields()) != 1 || len(s.Fields()[0]) != 7 {
			close(ctxts)
			log.Fatalln("[FATAL] DROID output is limited to signature files with a single PRONOM identifier")
		}
	default:
		w = newYAML(os.Stdout)
	}
	// overrite writer with nil writer if logging is to stdout
	if lg != nil && lg.w == os.Stdout {
		w = logWriter{}
	}
	wg := &sync.WaitGroup{}
	w.writeHead(s, *hashf)
	// support reading list files from stdin
	if flag.Arg(0) == "-" {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			info, err := os.Stat(scanner.Text())
			if err != nil {
				info, err = retryStat(scanner.Text(), err)
			}
			if err != nil || info.IsDir() {
				ctx := newContext(s, w, wg, nil, false, scanner.Text(), "", "", 0)
				ctx.res <- results{fmt.Errorf("failed to identify %s (in scanning mode, inputs must all be files and not directories), got: %v", scanner.Text(), err), nil, nil}
				wg.Add(1)
				ctxts <- ctx
			} else {
				identifyFile(newContext(s, w, wg, checksum, *archive, scanner.Text(), "", info.ModTime().Format(time.RFC3339), info.Size()), ctxts)
			}
		}
	} else {
		err = identify(ctxts, wg, s, w, flag.Arg(0), "", checksum, *archive, *nr)
	}
	wg.Wait()
	close(ctxts)
	w.writeTail()
	// log time elapsed
	if !lg.start.IsZero() {
		fmt.Fprintf(lg.w, "%s %v\n", timeString, time.Since(lg.start))
	}
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
