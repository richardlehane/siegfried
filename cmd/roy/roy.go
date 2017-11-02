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
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/internal/chart"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/loc"
	"github.com/richardlehane/siegfried/pkg/mimeinfo"
	"github.com/richardlehane/siegfried/pkg/pronom"
	"github.com/richardlehane/siegfried/pkg/reader"
	"github.com/richardlehane/siegfried/pkg/sets"
)

var usage = `
Usage:
   roy build -help
   roy add -help 
   roy harvest -help
   roy inspect -help
   roy sets -help
   roy compare -help
`

var inspectUsage = `
Usage of inspect:
   roy inspect
      Inspect the default signature file.
   roy inspect SIGNATURE 
      Inspect a named signature file e.g. roy inspect archivematica.sig
   roy inspect MATCHER
      Inspect  contents of a matcher e.g. roy inspect bytematcher.
      Short aliases work too e.g. roy inspect bm
      Current matchers are bytematcher (or bm), containermatcher (cm),
      xmlmatcher (xm), riffmatcher (rm), namematcher (nm), textmatcher (tm).
   roy inspect INTEGER
      Identify the signatures related to the numerical hits reported by the 
      sf debug and slow flags (sf -log d,s). E.g. roy inspect 100
      To inspect hits within containermatchers, give the index for the
      container type with the -ct flag, and the name of the container 
      sub-folder with the -cn flag.
      The container types are 0 for XML and 1 for MSCFB.
      E.g. roy inspect -ct 0 -cn [Content_Types].xml 0
   roy inspect FMT
      Inspect a file format signature e.g. roy inspect fmt/40
      MIME-info and LOC FDD file format signatures can be inspected too. 
      Also accepts comma separated lists of formats or format sets.
      E.g. roy inspect fmt/40,fmt/41 or roy inspect @pdfa
   roy inspect priorities
      Create a graph of priority relations (in graphviz dot format).
      The graph is built from the set of defined priority relations.
      Short alias is roy inspect p.
      View graph with a command e.g. roy inspect p | dot -Tpng -o priorities.png
      If you don't have dot installed, can use http://www.webgraphviz.com/.
   roy inspect missing-priorities
      Create a graph of relations that can be inferred from byte signatures,
      but that are not in the set of defined priority relations.
      Short alias is roy inspect mp.
      View graph with a command e.g. roy inspect mp | dot -Tpng -o missing.png
   roy inspect implicit-priorities
      Create a graph of relations that can be inferred from byte signatures,
      regardless of whether they are in the set of defined priority relations.
      Short alias is roy inspect ip.
      View graph with a command e.g. roy inspect ip | dot -Tpng -o implicit.png
   roy inspect releases
      Summary view of a PRONOM release-notes.xml file (which must be in your 
      siegfried home directory). 

Additional flags:
   The roy inspect FMT and roy inspect priorities sub-commands both accept
   the following flags. These flags mirror the equivalent flags for the
   roy build subcommand and you can find more detail with roy build -help.
   -extend, -extendc
      Add additional extension and container extension signature files. 
      Useful for inspecting test signatures during development.
      E.g. roy inspect -extend my-groovy-sig.xml dev/1
   -limit, -exclude
      Limit signatures to a comma-separated list of formats (or sets).
      Useful for priority graphs.
      E.g. roy inspect -limit @pdfa priorities
   -mi, -loc, -fdd
      Specify particular MIME-info or LOC FDD signature files for inspecting
      formats or viewing priorities.
   -reports
      Build from PRONOM reports files (rather than just using the DROID XML
      file as input). A bit slower but can be more accurate for a small set
      of formats like FLAC.
   -home
      Use a different siegfried home directory.
`

var (
	// BUILD, ADD flag sets
	build       = flag.NewFlagSet("build | add", flag.ExitOnError)
	home        = build.String("home", config.Home(), "override the default home directory")
	droid       = build.String("droid", config.Droid(), "set name/path for DROID signature file")
	mi          = build.String("mi", "", "set name/path for MIMEInfo signature file")
	fdd         = build.String("fdd", "", "set name/path for LOC FDD signature file")
	locfdd      = build.Bool("loc", false, "build a LOC FDD signature file")
	nopronom    = build.Bool("nopronom", false, "don't include PRONOM sigs with LOC signature file")
	container   = build.String("container", config.Container(), "set name/path for Droid Container signature file")
	name        = build.String("name", "", "set identifier name")
	details     = build.String("details", config.Details(), "set identifier details")
	extend      = build.String("extend", "", "comma separated list of additional signatures")
	extendc     = build.String("extendc", "", "comma separated list of additional container signatures")
	include     = build.String("limit", "", "comma separated list of PRONOM signatures to include")
	exclude     = build.String("exclude", "", "comma separated list of PRONOM signatures to exclude")
	bof         = build.Int("bof", 0, "define a maximum BOF offset")
	eof         = build.Int("eof", 0, "define a maximum EOF offset")
	noeof       = build.Bool("noeof", false, "ignore EOF segments in signatures")
	multi       = build.String("multi", "", "control how identifiers treat multiple results")
	nobyte      = build.Bool("nobyte", false, "skip byte signatures")
	nocontainer = build.Bool("nocontainer", false, "skip container signatures")
	notext      = build.Bool("notext", false, "skip text matcher")
	noname      = build.Bool("noname", false, "skip filename matcher")
	nomime      = build.Bool("nomime", false, "skip MIME matcher")
	noxml       = build.Bool("noxml", false, "skip XML matcher")
	noriff      = build.Bool("noriff", false, "skip RIFF matcher")
	noreports   = build.Bool("noreports", false, "build directly from DROID file rather than PRONOM reports")
	doubleup    = build.Bool("doubleup", false, "include byte signatures for formats that also have container signatures")
	rng         = build.Int("range", config.Range(), "define a maximum range for segmentation")
	distance    = build.Int("distance", config.Distance(), "define a maximum distance for segmentation")
	choices     = build.Int("choices", config.Choices(), "define a maximum number of choices for segmentation")

	// HARVEST
	harvest           = flag.NewFlagSet("harvest", flag.ExitOnError)
	harvestHome       = harvest.String("home", config.Home(), "override the default home directory")
	harvestDroid      = harvest.String("droid", config.Droid(), "set name/path for DROID signature file")
	harvestChanges    = harvest.Bool("changes", false, "harvest the latest PRONOM release-notes.xml file")
	_, htimeout, _, _ = config.HarvestOptions()
	timeout           = harvest.Duration("timeout", htimeout, "set duration before timing-out harvesting requests e.g. 120s")
	throttlef         = harvest.Duration("throttle", 0, "set a time to wait HTTP requests e.g. 50ms")

	// INSPECT (roy inspect | roy inspect fmt/121 | roy inspect usr/local/mysig.sig | roy inspect 10)
	inspect        = flag.NewFlagSet("inspect", flag.ExitOnError)
	inspectHome    = inspect.String("home", config.Home(), "override the default home directory")
	inspectReports = inspect.Bool("reports", false, "build signatures from PRONOM reports (rather than DROID xml)")
	inspectExtend  = inspect.String("extend", "", "comma separated list of additional signatures")
	inspectExtendc = inspect.String("extendc", "", "comma separated list of additional container signatures")
	inspectInclude = inspect.String("limit", "", "when inspecting priorities, comma separated list of PRONOM signatures to include")
	inspectExclude = inspect.String("exclude", "", "when inspecting priorities, comma separated list of PRONOM signatures to exclude")
	inspectMI      = inspect.String("mi", "", "set name/path for MIMEInfo signature file to inspect")
	inspectFDD     = inspect.String("fdd", "", "set name/path for LOC FDD signature file to inspect")
	inspectLOC     = inspect.Bool("loc", false, "inspect a LOC FDD signature file")
	inspectCType   = inspect.Int("ct", 0, "provide container type to inspect container hits")
	inspectCName   = inspect.String("cn", "", "provide container name to inspect container hits")

	// SETS
	setsf       = flag.NewFlagSet("sets", flag.ExitOnError)
	setsHome    = setsf.String("home", config.Home(), "override the default home directory")
	setsDroid   = setsf.String("droid", config.Droid(), "set name/path for DROID signature file")
	setsChanges = setsf.Bool("changes", false, "create a pronom-changes.json sets file")
	setsList    = setsf.String("list", "", "expand comma separated list of format sets")

	// COMPARE
	comparef    = flag.NewFlagSet("compare", flag.ExitOnError)
	compareJoin = comparef.Int("join", 0, "control which field(s) are used to link results files. Default is 0 (full file path). Other options are 1 (filename), 2, (filename + size), 3 (filename + modified), 4 (filename + hash), 5 (hash)")
)

func savereps() error {
	file, err := os.Open(config.Reports())
	if err != nil {
		err = os.Mkdir(config.Reports(), os.ModePerm)
		if err != nil {
			return fmt.Errorf("roy: error making reports directory")
		}
	}
	file.Close()
	errs := pronom.Harvest()
	if len(errs) > 0 {
		return fmt.Errorf("roy: errors saving reports to disk")
	}
	return nil
}

func makegob(s *siegfried.Siegfried, opts []config.Option) error {
	var id core.Identifier
	var err error
	if *mi != "" {
		id, err = mimeinfo.New(opts...)
	} else if *locfdd || *fdd != "" {
		id, err = loc.New(opts...)
	} else {
		id, err = pronom.New(opts...)
	}
	if err != nil {
		return err
	}
	err = s.Add(id)
	if err != nil {
		return err
	}
	return s.Save(config.Signature())
}

func inspectSig(t core.MatcherType) error {
	if *inspectHome != config.Home() {
		config.SetHome(*inspectHome)
	}
	s, err := siegfried.Load(config.Signature())
	if err == nil {
		fmt.Print(s.Inspect(t))
	}
	return err
}

func inspectFmts(fmts []string) error {
	var id core.Identifier
	var err error
	fs := sets.Expand(strings.Join(fmts, ","))
	if len(fs) == 0 {
		return fmt.Errorf("nothing to inspect")
	}
	opts := append(getOptions(), config.SetDoubleUp()) // speed up by allowing sig double ups
	if *inspectMI != "" {
		id, err = mimeinfo.New(opts...)
	} else if strings.HasPrefix(fs[0], "fdd") || *inspectLOC || (*inspectFDD != "") {
		if *inspectFDD == "" && !*inspectLOC {
			opts = append(opts, config.SetLOC(""))
		}
		id, err = loc.New(opts...)
	} else {
		if !*inspectReports {
			opts = append(opts, config.SetNoReports()) // speed up by building from droid xml
		}
		id, err = pronom.New(opts...)
	}
	if err != nil {
		return err
	}
	rep, err := id.Inspect(fs...)
	if err == nil {
		fmt.Println(rep)
	}
	return err
}

func graphPriorities(typ int) error {
	var id core.Identifier
	var err error
	opts := append(getOptions(), config.SetDoubleUp()) // speed up by allowing sig double ups
	if *inspectMI != "" {
		id, err = mimeinfo.New(opts...)
	} else if *inspectLOC || (*inspectFDD != "") {
		id, err = loc.New(opts...)
	} else {
		if !*inspectReports {
			opts = append(opts, config.SetNoReports()) // speed up by building from droid xml
		}
		id, err = pronom.New(opts...)
	}
	if err == nil {
		fmt.Println(id.GraphP(typ))
	}
	return err
}

func blameSig(i int) error {
	if *inspectHome != config.Home() {
		config.SetHome(*inspectHome)
	}
	s, err := siegfried.Load(config.Signature())
	if err == nil {
		fmt.Println(s.Blame(i, *inspectCType, *inspectCName))
	}
	return err
}

func viewReleases() error {
	xm, err := pronom.LoadReleases(config.Local("release-notes.xml"))
	if err != nil {
		return err
	}
	years, fields, releases := pronom.Releases(xm)
	fmt.Println(chart.Chart("PRONOM releases",
		years,
		fields,
		map[string]bool{"number releases": true},
		releases))
	return nil
}

func getOptions() []config.Option {
	opts := []config.Option{}
	// build options
	if *droid != config.Droid() {
		opts = append(opts, config.SetDroid(*droid))
	}
	if *container != config.Container() {
		opts = append(opts, config.SetContainer(*container))
	}
	if *mi != "" {
		opts = append(opts, config.SetMIMEInfo(*mi))
	}
	if *fdd != "" {
		opts = append(opts, config.SetLOC(*fdd))
	}
	if *locfdd {
		opts = append(opts, config.SetLOC(""))
	}
	if *nopronom {
		opts = append(opts, config.SetNoPRONOM())
	}
	if *name != "" {
		opts = append(opts, config.SetName(*name))
	}
	if *details != config.Details() {
		opts = append(opts, config.SetDetails(*details))
	}
	if *extend != "" {
		opts = append(opts, config.SetExtend(sets.Expand(*extend)))
	}
	if *extendc != "" {
		if *extend == "" {
			fmt.Println(
				`roy: warning! Unless the container extension only extends formats defined in 
the DROID signature file you should also include a regular signature extension 
(-extend) that includes a FileFormatCollection element describing the new formats.`)
		}
		opts = append(opts, config.SetExtendC(sets.Expand(*extendc)))
	}
	if *include != "" {
		opts = append(opts, config.SetLimit(sets.Expand(*include)))
	}
	if *exclude != "" {
		opts = append(opts, config.SetExclude(sets.Expand(*exclude)))
	}
	if *bof != 0 {
		opts = append(opts, config.SetBOF(*bof))
	}
	if *eof != 0 {
		opts = append(opts, config.SetEOF(*eof))
	}
	if *noeof {
		opts = append(opts, config.SetNoEOF())
	}
	if *multi != "" {
		opts = append(opts, config.SetMulti(strings.ToLower(*multi)))
	}
	if *nobyte {
		opts = append(opts, config.SetNoByte())
	}
	if *nocontainer {
		opts = append(opts, config.SetNoContainer())
	}
	if *notext {
		opts = append(opts, config.SetNoText())
	}
	if *noname {
		opts = append(opts, config.SetNoName())
	}
	if *nomime {
		opts = append(opts, config.SetNoMIME())
	}
	if *noxml {
		opts = append(opts, config.SetNoXML())
	}
	if *noriff {
		opts = append(opts, config.SetNoRIFF())
	}
	if *noreports {
		opts = append(opts, config.SetNoReports())
	}
	if *doubleup {
		opts = append(opts, config.SetDoubleUp())
	}
	if *rng != config.Range() {
		opts = append(opts, config.SetRange(*rng))
	}
	if *distance != config.Distance() {
		opts = append(opts, config.SetDistance(*distance))
	}
	if *choices != config.Choices() {
		opts = append(opts, config.SetChoices(*choices))
	}
	// inspect options
	if *inspectMI != "" {
		opts = append(opts, config.SetMIMEInfo(*inspectMI))
	}
	if *inspectFDD != "" {
		opts = append(opts, config.SetLOC(*fdd))
	}
	if *inspectLOC {
		opts = append(opts, config.SetLOC(""))
	}
	if *inspectInclude != "" {
		opts = append(opts, config.SetLimit(sets.Expand(*inspectInclude)))
	}
	if *inspectExclude != "" {
		opts = append(opts, config.SetExclude(sets.Expand(*inspectExclude)))
	}
	if *inspectExtend != "" {
		opts = append(opts, config.SetExtend(sets.Expand(*inspectExtend)))
	}
	if *inspectExtendc != "" {
		if *inspectExtend == "" {
			fmt.Println(
				`roy: warning! Unless the container extension only extends formats defined in 
the DROID signature file you should also include a regular signature extension 
(-extend) that includes a FileFormatCollection element describing the new formats.`)
		}
		opts = append(opts, config.SetExtendC(sets.Expand(*inspectExtendc)))
	}
	// set home
	if *home != config.Home() {
		config.SetHome(*home)
	} else if *inspectHome != config.Home() {
		config.SetHome(*inspectHome)
	}
	return opts
}

func setHarvestOptions() {
	if *harvestDroid != config.Droid() {
		config.SetDroid(*harvestDroid)()
	}
	if *harvestHome != config.Home() {
		config.SetHome(*harvestHome)
	}
	if *timeout != htimeout {
		config.SetHarvestTimeout(*timeout)
	}
	if *throttlef > 0 {
		config.SetHarvestThrottle(*throttlef)
	}
}

func setSetsOptions() {
	if *setsDroid != config.Droid() {
		config.SetDroid(*setsDroid)()
	}
	if *setsHome != config.Home() {
		config.SetHome(*setsHome)
	}
}

func main() {
	var err error
	if len(os.Args) < 2 {
		log.Fatal(usage)
	}

	switch os.Args[1] {
	case "build":
		err = build.Parse(os.Args[2:])
		if err == nil {
			if build.Arg(0) != "" {
				config.SetSignature(build.Arg(0))
			}
			s := siegfried.New()
			err = makegob(s, getOptions())
		}
	case "add":
		err = build.Parse(os.Args[2:])
		if err == nil {
			if build.Arg(0) != "" {
				config.SetSignature(build.Arg(0))
			}
			var s *siegfried.Siegfried
			s, err = siegfried.Load(config.Signature())
			if err == nil {
				err = makegob(s, getOptions())
			}
		}
	case "harvest":
		err = harvest.Parse(os.Args[2:])
		if err == nil {
			setHarvestOptions()
			if *harvestChanges {
				err = pronom.GetReleases(config.Local("release-notes.xml"))
			} else {
				err = savereps()
			}
		}
	case "inspect":
		inspect.Usage = func() { fmt.Print(inspectUsage) }
		err = inspect.Parse(os.Args[2:])
		if err == nil {
			input := inspect.Arg(0)
			switch {
			case input == "":
				err = inspectSig(-1)
			case input == "bytematcher", input == "bm":
				err = inspectSig(core.ByteMatcher)
			case input == "containermatcher", input == "cm":
				err = inspectSig(core.ContainerMatcher)
			case input == "namematcher", input == "nm":
				err = inspectSig(core.NameMatcher)
			case input == "mimematcher", input == "mm":
				err = inspectSig(core.MIMEMatcher)
			case input == "riffmatcher", input == "rm":
				err = inspectSig(core.RIFFMatcher)
			case input == "xmlmatcher", input == "xm":
				err = inspectSig(core.XMLMatcher)
			case input == "textmatcher", input == "tm":
				err = inspectSig(core.TextMatcher)
			case input == "priorities", input == "p":
				err = graphPriorities(0)
			case input == "missing-priorities", input == "mp":
				err = graphPriorities(1)
			case input == "implicit-priorities", input == "ip":
				err = graphPriorities(2)
			case input == "releases":
				err = viewReleases()
			case filepath.Ext(input) == ".sig":
				config.SetSignature(input)
				err = inspectSig(-1)
			default:
				var i int
				i, err = strconv.Atoi(input)
				if err == nil {
					err = blameSig(i)
				} else {
					err = inspectFmts(inspect.Args())
				}
			}
		}
		if err != nil {
			err = fmt.Errorf("%s\nUsage: `roy inspect -help`", err.Error())
		}
	case "sets":
		err = setsf.Parse(os.Args[2:])
		if err == nil {
			setSetsOptions()
			if *setsList != "" {
				fmt.Println(strings.Join(sets.Expand(*setsList), "\n"))
			} else if *setsChanges {
				releases, err := pronom.LoadReleases(config.Local("release-notes.xml"))
				if err == nil {
					err = pronom.ReleaseSet("pronom-changes.json", releases)
				}
			} else {
				err = pronom.TypeSets("pronom-all.json", "pronom-families.json", "pronom-types.json")
			}
		}
	case "compare":
		err = comparef.Parse(os.Args[2:])
		if err == nil {
			err = reader.Compare(os.Stdout, *compareJoin, comparef.Args()...)
		}
	default:
		log.Fatal(usage)
	}
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
