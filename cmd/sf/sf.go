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
	"sync"
	"time"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/internal/checksum"
	"github.com/richardlehane/siegfried/internal/logger"
	"github.com/richardlehane/siegfried/internal/siegreader"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/reader"
	"github.com/richardlehane/siegfried/pkg/writer"
	/*// Uncomment to build with profiler
	"net/http"
	_ "net/http/pprof"
	*/)

// defaults
const maxMulti = 1024

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
	multi     = flag.Int("multi", 1, "set number of parallel file ID processes")
	archive   = flag.Bool("z", false, "scan archive formats (zip, tar, gzip, warc, arc)")
	hashf     = flag.String("hash", "", "calculate file checksum with hash algorithm; options "+checksum.HashChoices)
	throttlef = flag.Duration("throttle", 0, "set a time to wait between scanning files e.g. 50ms")
	replay    = flag.Bool("replay", false, "replay one (or more) results files to change output or logging e.g. sf -replay -csv results.yaml")
	list      = flag.Bool("f", false, "scan one (or more) lists of filenames e.g. sf -f myfiles.txt")
	name      = flag.String("name", "", "provide a filename when scanning a stream e.g. sf -name myfile.txt -")
)

var (
	throttle *time.Ticker
	ctxPool  *sync.Pool
)

type WalkError struct {
	path string
	err  error
}

func (we WalkError) Error() string {
	return fmt.Sprintf("walking %s; got %v", we.path, we.err)
}

func setCtxPool(s *siegfried.Siegfried, wg *sync.WaitGroup, w writer.Writer, d, z bool, h checksum.HashTyp) {
	ctxPool = &sync.Pool{
		New: func() interface{} {
			return &context{
				s:   s,
				wg:  wg,
				w:   w,
				d:   d,
				z:   z,
				h:   checksum.MakeHash(h),
				res: make(chan results, 1),
			}
		},
	}
}

type getFn func(string, string, string, int64) *context

func getCtx(path, mime, mod string, sz int64) *context {
	c := ctxPool.Get().(*context)
	if c.h != nil {
		c.h.Reset()
	}
	c.path, c.mime, c.mod, c.sz = path, mime, mod, sz
	return c
}

type context struct {
	s  *siegfried.Siegfried
	wg *sync.WaitGroup
	w  writer.Writer
	d  bool // droid
	// opts
	z bool
	h hash.Hash
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

func printer(ctxts chan *context, lg *logger.Logger) {
	for ctx := range ctxts {
		lg.Progress(ctx.path)
		// block on the results
		res := <-ctx.res
		lg.Error(ctx.path, res.err)
		lg.IDs(ctx.path, res.ids)
		// write the result
		ctx.w.File(ctx.path, ctx.sz, ctx.mod, res.cs, res.err, res.ids)
		ctx.wg.Done()
		ctxPool.Put(ctx) // return the context to the pool
	}
}

// identify() defined in longpath.go and longpath_windows.go

func readFile(ctx *context, ctxts chan *context, gf getFn) {
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
	identifyRdr(f, ctx, ctxts, gf)
	f.Close()
}

func identifyFile(ctx *context, ctxts chan *context, gf getFn) {
	ctx.wg.Add(1)
	ctxts <- ctx
	if *multi == 1 || ctx.z || config.Slow() || config.Debug() {
		readFile(ctx, ctxts, gf)
		return
	}
	go func() {
		ctx.wg.Add(1)
		readFile(ctx, ctxts, gf)
		ctx.wg.Done()
	}()
}

func identifyRdr(r io.Reader, ctx *context, ctxts chan *context, gf getFn) {
	s := ctx.s
	b, berr := s.Buffer(r)
	defer s.Put(b)
	ids, err := s.IdentifyBuffer(b, berr, ctx.path, ctx.mime)
	if ids == nil {
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
	if !ctx.z {
		ctx.res <- results{err, cs, ids}
		return
	}
	arc := isArc(ids)
	if arc == config.None {
		ctx.res <- results{err, cs, ids}
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
		ctx.res <- results{fmt.Errorf("failed to decompress, got: %v", err), cs, ids}
		return
	}
	// send the result
	ctx.res <- results{err, cs, ids}
	// decompress and recurse
	for err = d.next(); err == nil; err = d.next() {
		if ctx.d {
			for _, v := range d.dirs() {
				dctx := gf(v, "", "", -1)
				dctx.res <- results{nil, nil, nil}
				dctx.wg.Add(1)
				ctxts <- dctx
			}
		}
		nctx := gf(d.path(), d.mime(), d.mod(), d.size())
		nctx.wg.Add(1)
		ctxts <- nctx
		identifyRdr(d.reader(), nctx, ctxts, gf)
	}
}

func openFile(path string) (*os.File, error) {
	if path == "-" {
		return os.Stdin, nil
	}
	return os.Open(path)
}

var firstReplay sync.Once

func replayFile(path string, ctxts chan *context, w writer.Writer) error {
	f, err := openFile(path)
	if err != nil {
		return err
	}
	defer f.Close()
	rdr, err := reader.New(f, path)
	if err != nil {
		return fmt.Errorf("[FATAL] error reading results file %s; got %v\n", path, err)
	}
	firstReplay.Do(func() {
		hd := rdr.Head()
		w.Head(hd.SignaturePath, hd.Scanned, hd.Created, hd.Version, hd.Identifiers, hd.Fields, hd.HashHeader)
	})
	var rf reader.File
	for rf, err = rdr.Next(); err == nil; rf, err = rdr.Next() {
		ctx := getCtx(rf.Path, "", rf.Mod, rf.Size)
		ctx.res <- results{rf.Err, rf.Hash, rf.IDs}
		ctx.wg.Add(1)
		ctxts <- ctx
	}
	if err != nil && err != io.EOF {
		return fmt.Errorf("[FATAL] error reading results file %s; got %v\n", path, err)
	}
	return nil
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
		msg, err := updateSigs(*sig, flag.Args())
		if err != nil {
			log.Fatalf("[FATAL] failed to update signature file, %v", err)
		}
		fmt.Println(msg)
		return
	}
	// handle -hash error
	hashT := checksum.GetHash(*hashf)
	if *hashf != "" && hashT < 0 {
		log.Fatalf("[FATAL] invalid hash type; choose from %s", checksum.HashChoices)
	}
	// load and handle signature errors
	var (
		s   *siegfried.Siegfried
		err error
	)
	if !*replay || *version || *fprflag || *serve != "" {
		s, err = siegfried.Load(config.Signature())
	}
	if err != nil {
		log.Fatalf("[FATAL] error loading signature file, got: %v", err)
	}
	// handle -version
	if *version {
		version := config.Version()
		fmt.Printf("siegfried %d.%d.%d\n", version[0], version[1], version[2])
		fmt.Printf("%s (%s)\nidentifiers: \n", config.Signature(), s.C.Format(time.RFC3339))
		for _, id := range s.Identifiers() {
			fmt.Printf("  - %s: %s\n", id[0], id[1])
		}
		return
	}
	// handle -fpr
	if *fprflag {
		log.Printf("FPR server started at %s. Use CTRL-C to quit.\n", config.Fpr())
		serveFpr(config.Fpr(), s)
		return
	}
	// check -multi
	if *multi > maxMulti || *multi < 1 || (*archive && *multi > 1) {
		log.Println("[WARN] -multi must be > 0 and =< 1024. If -z, -multi must be 1. Resetting -multi to 1")
		*multi = 1
	}
	// start logger
	lg, err := logger.New(*logf)
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
	lenCtxts := *multi
	if lenCtxts == 1 {
		lenCtxts = 8
	}
	ctxts := make(chan *context, lenCtxts)
	go printer(ctxts, lg)
	// set default writer
	var w writer.Writer
	var d bool
	switch {
	case lg.IsOut():
		w = writer.Null()
	case *csvo:
		w = writer.CSV(os.Stdout)
	case *jsono:
		w = writer.JSON(os.Stdout)
	case *droido:
		if len(s.Fields()) != 1 || len(s.Fields()[0]) != 7 {
			close(ctxts)
			log.Fatalln("[FATAL] DROID output is limited to signature files with a single PRONOM identifier")
		}
		w = writer.Droid(os.Stdout)
		d = true
	default:
		w = writer.YAML(os.Stdout)
	}
	// setup default waitgroup
	wg := &sync.WaitGroup{}
	// setup context pool
	setCtxPool(s, wg, w, d, *archive, hashT)
	// handle -serve
	if *serve != "" {
		log.Printf("Starting server at %s. Use CTRL-C to quit.\n", *serve)
		listen(*serve, s, ctxts)
		return
	}
	// handle no file/directory argument
	if flag.NArg() < 1 {
		close(ctxts)
		log.Fatalln("[FATAL] expecting one or more file or directory arguments (or '-' to scan stdin)")
	}
	if !*replay {
		w.Head(config.SignatureBase(), time.Now(), s.C, config.Version(), s.Identifiers(), s.Fields(), hashT.String())
	}
	for _, v := range flag.Args() {
		if *list {
			f, err := openFile(v)
			if err != nil {
				break
			}
			scanner := bufio.NewScanner(f)
			var info os.FileInfo
			for scanner.Scan() {
				info, err = os.Stat(scanner.Text())
				if err != nil {
					info, err = retryStat(scanner.Text(), err)
				}
				if err != nil || info.IsDir() {
					ctx := getCtx(scanner.Text(), "", "", 0)
					ctx.res <- results{fmt.Errorf("failed to identify %s (in scanning mode, inputs must all be files and not directories), got: %v", scanner.Text(), err), nil, nil}
					ctx.wg.Add(1)
					ctxts <- ctx
					err = nil // don't log.Fatal this error, report it in the results
				} else if *replay {
					err = replayFile(scanner.Text(), ctxts, w)
					if err != nil {
						break
					}
				} else {
					identifyFile(getCtx(scanner.Text(), "", info.ModTime().Format(time.RFC3339), info.Size()), ctxts, getCtx)
				}
			}
			f.Close()
		} else if *replay {
			err = replayFile(v, ctxts, w)
		} else if v == "-" {
			ctx := getCtx(*name, "", "", 0)
			ctx.wg.Add(1)
			ctxts <- ctx
			identifyRdr(os.Stdin, ctx, ctxts, getCtx)
		} else {
			err = identify(ctxts, v, "", *nr, d, getCtx)
		}
		if err != nil {
			break
		}
	}
	wg.Wait()
	close(ctxts)
	w.Tail()
	// log time elapsed and chart
	lg.Close()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
