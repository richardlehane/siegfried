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

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var (
	// BUILD, ADD flag sets
	build       = flag.NewFlagSet("build | add", flag.ExitOnError)
	home        = build.String("home", config.Home(), "override the default home directory")
	droid       = build.String("droid", config.Droid(), "set name/path for DROID signature file")
	container   = build.String("container", config.Container(), "set name/path for Droid Container signature file")
	reports     = build.String("reports", config.Reports(), "set path for PRONOM reports directory")
	name        = build.String("name", config.Name(), "set identifier name")
	details     = build.String("details", config.Details(), "set identifier details")
	extend      = build.String("extend", "", "comma separated list of additional signatures")
	bof         = build.Int("bof", 0, "define a maximum BOF offset")
	eof         = build.Int("eof", 0, "define a maximum EOF offset")
	noeof       = build.Bool("noeof", false, "ignore EOF segments in signatures")
	nopriority  = build.Bool("nopriority", false, "ignore priority rules when recording results")
	nocontainer = build.Bool("nocontainer", false, "skip container signatures")
	rng         = build.Int("range", config.Range(), "define a maximum range for segmentation")
	distance    = build.Int("distance", config.Distance(), "define a maximum distance for segmentation")
	varLength   = build.Int("varlen", config.VarLength(), "define a maximum length for variable offset search sequences")
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
	p, err := pronom.New(opts...)
	if err != nil {
		return err
	}
	err = s.Add(p)
	if err != nil {
		return err
	}
	return s.Save(config.Signature())
}

func inspectPronom() error {
	s, err := siegfried.Load(config.Signature())
	if err != nil {
		return err
	}
	fmt.Print(s)
	return nil
}

func blameSig(i int) error {
	s, err := siegfried.Load(config.Signature())
	if err != nil {
		return err
	}
	fmt.Println(s.InspectTTI(i))
	/*
		p, err := pronom.NewPronom()
		if err != nil {
			return err
		}
		sigs, err := p.Parse()
		if err != nil {
			return err
		}
		bm := bytematcher.New()
		_, err = bm.Add(bytematcher.SignatureSet(sigs), nil)
		if err != nil {
			return err
		}
		if i > len(bm.Tests)-1 {
			return fmt.Errorf("Test index out of range: got %d, have %d tests", i, len(bm.Tests))
		}
		puids, _ := p.GetPuids()
		tn := bm.Tests[i]
		if len(tn.Complete) > 0 {
			fmt.Println("Completes:")
		}
		for _, v := range tn.Complete {
			fmt.Println(puids[v[0]])
		}
		if len(tn.Incomplete) > 0 {
			fmt.Println("Incompletes:")
		}
		for _, v := range tn.Incomplete {
			fmt.Println(puids[v.Kf[0]])
		}
	*/
	return nil
}

func viewSig(f string) error {
	p, err := pronom.New(config.SetSingle(f))
	if err != nil {
		return err
	}
	s := siegfried.New()
	err = s.Add(p)
	if err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

func buildOptions() []config.Option {
	opts := []config.Option{}
	if *home != config.Home() {
		opts = append(opts, config.SetHome(*home))
	}
	if *droid != config.Droid() {
		opts = append(opts, config.SetDroid(*droid))
	}
	if *container != config.Container() {
		opts = append(opts, config.SetContainer(*container))
	}
	if *reports != config.Reports() {
		opts = append(opts, config.SetReports(*reports))
	}
	if *name != config.Name() {
		opts = append(opts, config.SetName(*name))
	}
	if *details != config.Details() {
		opts = append(opts, config.SetDetails(*details))
	}
	if *extend != "" {
		opts = append(opts, config.SetExtend(*extend))
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
	if *nopriority {
		opts = append(opts, config.SetNoPriority())
	}
	if *nocontainer {
		opts = append(opts, config.SetNoContainer())
	}
	if *rng != config.Range() {
		opts = append(opts, config.SetRange(*rng))
	}
	if *distance != config.Distance() {
		opts = append(opts, config.SetDistance(*distance))
	}
	if *varLength != config.VarLength() {
		opts = append(opts, config.SetVarLength(*varLength))
	}
	if *choices != config.Choices() {
		opts = append(opts, config.SetChoices(*choices))
	}
	return opts
}

func setHarvestOptions() {
	if *harvestHome != config.Home() {
		config.SetHome(*harvestHome)()
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
		config.SetHome(*inspectHome)()
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
			switch inspect.Arg(0) {
			case "":
				err = inspectPronom()
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
