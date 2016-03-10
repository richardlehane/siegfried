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
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"

	/*// Uncomment to build with profiler
	"net/http"
	_ "net/http/pprof"
	*/)

const PROCS = -1

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
	multi     = flag.Int("multi", 1, "set number of file ID processes")
	archive   = flag.Bool("z", false, "scan archive formats (zip, tar, gzip, warc, arc)")
	hashf     = flag.String("hash", "", "calculate file checksum with hash algorithm; options "+hashChoices)
	throttlef = flag.Duration("throttle", 0, "set a time to wait between scanning files e.g. 50ms")
)

var throttle *time.Ticker

func writeError(w writer, path string, sz int64, mod string, err error) {
	w.writeFile(path, sz, mod, nil, err, nil)
	// log the error too
	lg.set(path)
	lg.err(err)
	lg.reset()
}

type res struct {
	path string
	sz   int64
	mod  string
	c    iterableID
	err  error
}

func printer(w writer, resc chan chan res, wg *sync.WaitGroup) {
	for rr := range resc {
		r := <-rr
		w.writeFile(r.path, r.sz, r.mod, nil, r.err, r.c)
		wg.Done()
	}
}

func multiIdentifyP(w writer, s *siegfried.Siegfried, r string, norecurse bool) error {
	wg := &sync.WaitGroup{}
	runtime.GOMAXPROCS(PROCS)
	resc := make(chan chan res, *multi)
	go printer(w, resc, wg)
	wf := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walking %s; got %v", path, err)
		}
		if info.IsDir() {
			if norecurse && path != r {
				return filepath.SkipDir
			}
			if *droido {
				wg.Add(1)
				rchan := make(chan res, 1)
				resc <- rchan
				go func() {
					rchan <- res{path, -1, info.ModTime().String(), nil, nil} // write directory with a -1 size for droid output only
				}()
			}
			return nil
		}
		wg.Add(1)
		rchan := make(chan res, 1)
		resc <- rchan
		go func() {
			f, err := os.Open(path)
			if err != nil {
				rchan <- res{path, 0, "", nil, err.(*os.PathError).Err} // return summary error only
				return
			}
			c, err := s.Identify(f, path, "")
			if c == nil {
				f.Close()
				rchan <- res{path, 0, "", nil, err}
				return
			}
			ids := makeIdSlice(idChan(c))
			f.Close()
			rchan <- res{path, info.Size(), info.ModTime().Format(time.RFC3339), ids, err}
		}()
		return nil
	}
	err := filepath.Walk(r, wf)
	wg.Wait()
	close(resc)
	return err
}

// multiIdentifyS() defined in longpath.go and longpath_windows.go

func identifyFile(w writer, s *siegfried.Siegfried, path string, sz int64, mod string) {
	f, err := os.Open(path)
	if err != nil {
		f, err = retryOpen(path, err) // retry open in case is a windows long path error
		if err != nil {
			writeError(w, path, sz, mod, err.(*os.PathError).Err) // write summary error
			return
		}
	}
	identifyRdr(w, s, f, sz, path, "", mod)
	f.Close()
}

func identifyRdr(w writer, s *siegfried.Siegfried, r io.Reader, sz int64, path, mime, mod string) {
	lg.set(path)
	c, err := s.Identify(r, path, mime)
	lg.err(err)
	if c == nil {
		w.writeFile(path, sz, mod, nil, err, nil)
		lg.reset()
		return
	}
	var b *siegreader.Buffer
	var cs []byte
	if checksum != nil {
		b = s.Buffer()
		var i int64
		l := checksum.BlockSize()
		for ; ; i += int64(l) {
			buf, _ := b.Slice(i, l)
			if buf == nil {
				break
			}
			checksum.Write(buf)
		}
		cs = checksum.Sum(nil)
		checksum.Reset()
	}
	a := w.writeFile(path, sz, mod, cs, err, idChan(c))
	lg.reset()
	if !*archive || a == config.None {
		return
	}
	var d decompressor
	if b == nil {
		b = s.Buffer()
	}
	switch a {
	case config.Zip:
		d, err = newZip(siegreader.ReaderFrom(b), path, sz)
	case config.Gzip:
		d, err = newGzip(b, path)
	case config.Tar:
		d, err = newTar(siegreader.ReaderFrom(b), path)
	case config.ARC:
		d, err = newARC(siegreader.ReaderFrom(b), path)
	case config.WARC:
		d, err = newWARC(siegreader.ReaderFrom(b), path)
	}
	if err != nil {
		writeError(w, path, sz, mod, fmt.Errorf("failed to decompress, got: %v", err))
		return
	}
	for err = d.next(); err == nil; err = d.next() {
		if *droido {
			for _, v := range d.dirs() {
				w.writeFile(v, -1, "", nil, nil, nil)
			}
		}
		identifyRdr(w, s, d.reader(), d.size(), d.path(), d.mime(), d.mod())
	}
}

func main() {

	flag.Parse()

	/*//UNCOMMENT TO RUN PROFILER
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()*/

	if *version {
		version := config.Version()
		fmt.Printf("siegfried %d.%d.%d\n", version[0], version[1], version[2])
		s, err := siegfried.Load(config.Signature())
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(s)
		return
	}

	if *home != config.Home() {
		config.SetHome(*home)
	}

	if *sig != config.SignatureBase() {
		config.SetSignature(*sig)
	}

	if *update {
		msg, err := updateSigs()
		if err != nil {
			log.Fatalf("[FATAL] failed to update signature file, %v", err)
		}
		fmt.Println(msg)
		return
	}

	// during parallel scanning or in server mode, unsafe to access the last read buffer - so can't unzip or hash
	if *multi > 1 || *serve != "" {
		if *archive {
			log.Fatalln("[FATAL] cannot scan archive formats when running in parallel or server mode")
		}
		if *hashf != "" {
			log.Fatalln("[FATAL] cannot calculate file checksum when running in parallel or server mode")
		}
	}

	if *logf != "" {
		if *multi > 1 && *logf != "error" {
			log.Fatalln("[FATAL] cannot log in parallel mode")
		}
		if err := newLogger(*logf); err != nil {
			log.Fatalln(err)
		}
	}

	if err := setHash(); err != nil {
		log.Fatal(err)
	}

	if *serve != "" || *fprflag {
		s, err := siegfried.Load(config.Signature())
		if err != nil {
			log.Fatalf("[FATAL] error loading signature file, got: %v", err)
		}
		if *serve != "" {
			log.Printf("Starting server at %s. Use CTRL-C to quit.\n", *serve)
			listen(*serve, s)
			return
		}
		log.Printf("FPR server started at %s. Use CTRL-C to quit.\n", config.Fpr())
		serveFpr(config.Fpr(), s)
		return
	}

	if flag.NArg() != 1 {
		log.Fatalln("[FATAL] expecting a single file or directory argument")
	}

	s, err := siegfried.Load(config.Signature())
	if err != nil {
		log.Fatalf("[FATAL] error loading signature file, got: %v", err)
	}

	var w writer
	switch {
	case *csvo:
		w = newCSV(os.Stdout)
	case *jsono:
		w = newJSON(os.Stdout)
	case *droido:
		w = newDroid(os.Stdout)
		if len(s.Fields()) != 1 || len(s.Fields()[0]) != 7 {
			log.Fatalln("[FATAL] DROID output is limited to signature files with a single PRONOM identifier")
		}
	default:
		w = newYAML(os.Stdout)
	}

	if lg != nil && lg.w == os.Stdout {
		w = logWriter{}
	}

	// support reading list files from stdin
	if flag.Arg(0) == "-" {
		w.writeHead(s)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			info, err := os.Stat(scanner.Text())
			if err != nil {
				info, err = retryStat(scanner.Text(), err)
			}
			if err != nil || info.IsDir() {
				writeError(w, scanner.Text(), 0, "", fmt.Errorf("failed to identify %s (in scanning mode, inputs must all be files and not directories), got: %v", scanner.Text(), err))
			} else {
				identifyFile(w, s, scanner.Text(), info.Size(), info.ModTime().Format(time.RFC3339))
			}
		}
		w.writeTail()
		lg.printElapsed()
		os.Exit(0)
	}

	info, err := os.Stat(flag.Arg(0))
	if err != nil {
		info, err = retryStat(flag.Arg(0), err)
		if err != nil {
			log.Fatalf("[FATAL] cannot get info for %v, got: %v", flag.Arg(0), err)
		}
	}

	if info.IsDir() {
		w.writeHead(s)
		if *multi > 16 {
			*multi = 16
		}
		if *multi > 1 {
			err = multiIdentifyP(w, s, flag.Arg(0), *nr)
		} else {
			if *throttlef != 0 {
				throttle = time.NewTicker(*throttlef)
				defer throttle.Stop()
			}
			err = multiIdentifyS(w, s, flag.Arg(0), "", *nr)
		}
		w.writeTail()
		if err != nil {
			log.Fatalf("[FATAL] %v\n", err)
		}
		lg.printElapsed()
		os.Exit(0)
	}
	w.writeHead(s)
	identifyFile(w, s, flag.Arg(0), info.Size(), info.ModTime().Format(time.RFC3339))
	w.writeTail()
	lg.printElapsed()
	os.Exit(0)
}
