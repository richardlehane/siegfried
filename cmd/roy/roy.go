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
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/loc"
	"github.com/richardlehane/siegfried/pkg/mimeinfo"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var (
	// BUILD, ADD flag sets
	build       = flag.NewFlagSet("build | add", flag.ExitOnError)
	home        = build.String("home", config.Home(), "override the default home directory")
	droid       = build.String("droid", config.Droid(), "set name/path for DROID signature file")
	mi          = build.String("mi", "", "set name/path for MIMEInfo signature file")
	fdd         = build.String("fdd", "", "set name/path for LOC FDD signature file")
	locfdd      = build.Bool("loc", false, "build a LOC FDD signature file")
	container   = build.String("container", config.Container(), "set name/path for Droid Container signature file")
	reports     = build.String("reports", config.Reports(), "set path for PRONOM reports directory")
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
	harvest        = flag.NewFlagSet("harvest", flag.ExitOnError)
	harvestHome    = harvest.String("home", config.Home(), "override the default home directory")
	harvestDroid   = harvest.String("droid", config.Droid(), "set name/path for DROID signature file")
	harvestReports = harvest.String("reports", config.Reports(), "set path for PRONOM reports directory")
	_, htimeout, _ = config.HarvestOptions()
	timeout        = flag.Duration("timeout", htimeout, "set duration before timing-out harvesting requests e.g. 120s")

	// INSPECT (roy inspect | roy inspect fmt/121 | roy inspect usr/local/mysig.gob | roy inspect 10)
	inspect        = flag.NewFlagSet("inspect", flag.ExitOnError)
	inspectHome    = inspect.String("home", config.Home(), "override the default home directory")
	inspectReports = inspect.String("reports", config.Reports(), "set path for PRONOM reports directory")
	inspectMI      = inspect.String("mi", "", "set name/path for MIMEInfo signature file")
	inspectCType   = inspect.Int("ct", 0, "provide container type to inspect container hits")
	inspectCName   = inspect.String("cn", "", "provide container name to inspect container hits")
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
	s, err := siegfried.Load(config.Signature())
	if err != nil {
		return err
	}
	fmt.Print(s.Inspect(t))
	return nil
}

func inspectFmt(f string) error {
	if *inspectMI != "" {
		_, err := mimeinfo.New(config.SetMIMEInfo(*inspectMI), config.SetLimit(expandSets(f)), config.SetInspect())
		return err
	}
	_, err := pronom.New(config.SetLimit(expandSets(f)), config.SetInspect(), config.SetNoContainer())
	return err
}

func blameSig(i int) error {
	s, err := siegfried.Load(config.Signature())
	if err != nil {
		return err
	}
	fmt.Println(s.Blame(i, *inspectCType, *inspectCName))
	return nil
}

func buildOptions() []config.Option {
	if *home != config.Home() {
		config.SetHome(*home)
	}
	opts := []config.Option{}
	if *droid != config.Droid() {
		opts = append(opts, config.SetDroid(*droid))
	}
	if *container != config.Container() {
		opts = append(opts, config.SetContainer(*container))
	}
	if *reports != config.Reports() {
		opts = append(opts, config.SetReports(*reports))
	}
	if *mi != "" {
		opts = append(opts, config.SetMIMEInfo(*mi))
	}
	if *fdd != "" {
		opts = append(opts, config.SetLOC(*fdd))
	}
	if *name != "" {
		opts = append(opts, config.SetName(*name))
	}
	if *details != config.Details() {
		opts = append(opts, config.SetDetails(*details))
	}
	if *extend != "" {
		opts = append(opts, config.SetExtend(expandSets(*extend)))
	}
	if *extendc != "" {
		if *extend == "" {
			fmt.Println(
				`roy: warning! Unless the container extension only extends formats defined in 
the DROID signature file you should also include a regular signature extension 
(-extend) that includes a FileFormatCollection element defining the new formats.`)
		}
		opts = append(opts, config.SetExtendC(expandSets(*extendc)))
	}
	if *include != "" {
		opts = append(opts, config.SetLimit(expandSets(*include)))
	}
	if *exclude != "" {
		opts = append(opts, config.SetExclude(expandSets(*exclude)))
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
	return opts
}

func setHarvestOptions() {
	if *harvestHome != config.Home() {
		config.SetHome(*harvestHome)
	}
	if *harvestDroid != config.Droid() {
		config.SetDroid(*harvestDroid)()
	}
	if *harvestReports != config.Reports() {
		config.SetReports(*harvestReports)()
	}
	if *timeout != htimeout {
		config.SetHarvestTimeout(*timeout)
	}
}

func setInspectOptions() {
	if *inspectHome != config.Home() {
		config.SetHome(*inspectHome)
	}
	if *inspectReports != config.Reports() {
		config.SetReports(*inspectReports)()
	}
}

var usage = `
Usage
   roy build -help
   roy add -help 
   roy harvest -help
   roy inspect -help
`

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
			err = makegob(s, buildOptions())
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
				err = makegob(s, buildOptions())
			}
		}
	case "harvest":
		err = harvest.Parse(os.Args[2:])
		if err == nil {
			setHarvestOptions()
			err = savereps()
		}
	case "inspect":
		err = inspect.Parse(os.Args[2:])
		if err == nil {
			setInspectOptions()
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
			case filepath.Ext(input) == ".sig":
				config.SetSignature(input)
				err = inspectSig(-1)
			case strings.Contains(input, "fmt"), *inspectMI != "":
				err = inspectFmt(input)
			default:
				var i int
				i, err = strconv.Atoi(input)
				if err != nil {
					log.Fatal(err)
				}
				err = blameSig(i)
			}
		}
	default:
		log.Fatal(usage)
	}
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
