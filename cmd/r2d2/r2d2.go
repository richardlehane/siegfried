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
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var (
	blame     = flag.Int("blame", -1, "identify a signature from an initial test tree index")
	compile   = flag.String("compile", "", "compile a single Pronom signature (for testing)")
	view      = flag.String("view", "", "view a Pronom signature e.g. fmt/161")
	harvest   = flag.Bool("harvest", false, "harvest Pronom reports")
	build     = flag.String("build", "false", "build a Siegfried signature file; give a label for the identifier e.g. 'pronom'")
	inspect   = flag.Bool("inspect", false, "describe a Siegfried signature file")
	timeout   = flag.Duration("timeout", config.Pronom.HarvestTimeout, "set duration before timing-out harvesting requests e.g. 120s")
	sigfile   = flag.String("sig", config.Siegfried.Signature, "set path to Siegfried signature file")
	droid     = flag.String("droid", config.Pronom.Droid, "set path to Droid signature file")
	container = flag.String("container", config.Pronom.Container, "set path to Droid Container signature file")
	reports   = flag.String("reports", config.Pronom.Reports, "set path to Pronom reports directory")
)

func savereps() error {
	file, err := os.Open(config.Pronom.Reports)
	if err != nil {
		err = os.Mkdir(config.Pronom.Reports, os.ModeDir)
		if err != nil {
			log.Fatal(err)
		}
	}
	file.Close()
	errs := pronom.SaveReports()
	for _, e := range errs {
		fmt.Print(e)
	}
	if len(errs) > 0 {
		return fmt.Errorf("Errors saving reports to disk")
	}
	return nil
}

func makegob() error {
	s := siegfried.New()
	err := s.Add(siegfried.Pronom)
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
	return nil
}

func viewSig(f string) error {
	sigs, err := pronom.ParsePuid(f)
	if err != nil {
		return err
	}
	fmt.Println("Signatures:")
	for _, s := range sigs {
		fmt.Println(s)
	}
	bm := bytematcher.New()
	_, err = bm.Add(bytematcher.SignatureSet(sigs), nil)
	if err != nil {
		return err
	}
	fmt.Println("\nKeyframes:")
	for _, kf := range bm.KeyFrames {
		fmt.Println(kf)
	}
	fmt.Println("\nTests:")
	for _, test := range bm.Tests {
		fmt.Println(test)
	}
	if len(bm.BOFSeq.Set) > 0 {
		fmt.Println("\nBOF seqs:")
		for _, seq := range bm.BOFSeq.Set {
			fmt.Println(seq)
		}
	}
	if len(bm.EOFSeq.Set) > 0 {
		fmt.Println("\nEOF seqs:")
		for _, seq := range bm.EOFSeq.Set {
			fmt.Println(seq)
		}
	}
	fmt.Println("\nBytematcher:")

	fmt.Println(bm)
	return nil
}

func main() {
	flag.Parse()

	config.Pronom.HarvestTimeout = *timeout
	config.SetLatest()
	var err error
	switch {
	case *harvest:
		err = savereps()
	case *build != "false":
		err = makegob()
	case *inspect:
		err = inspectPronom()
	case *view != "":
		err = viewSig(*view)
	case *blame > -1:
		err = blameSig(*blame)
	}
	if err != nil {
		log.Fatal(err)
	}
}
