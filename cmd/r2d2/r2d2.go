package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var (
	view     = flag.String("view", "", "view a Pronom signature e.g. fmt/161")
	harvest  = flag.Bool("harvest", false, "harvest Pronom reports")
	build    = flag.Bool("build", false, "build a Siegfried signature file")
	inspect  = flag.Bool("inspect", false, "describe a Siegfried signature file")
	defaults = flag.Bool("defaults", false, "print the default paths and settings")
	timeout  = flag.Duration("timeout", 120*time.Second, "set duration before timing-out harvesting requests e.g. 120s")
)

var pronom_url = "http://apps.nationalarchives.gov.uk/pronom/"

var (
	sigfile   string
	droid     string
	container string
	reports   string

	defaultSigPath       = "pronom.gob"
	defaultDroidPath     = "DROID_SignatureFile_V77.xml"
	defaultContainerPath = "container-signature-20140717.xml"
	defaultReportsPath   = "pronom"
)

func init() {
	current, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	defaultSigs := filepath.Join(current.HomeDir, "siegfried", defaultSigPath)
	defaultDroid := filepath.Join(current.HomeDir, "siegfried", defaultDroidPath)
	defaultContainer := filepath.Join(current.HomeDir, "siegfried", defaultContainerPath)
	defaultReports := filepath.Join(current.HomeDir, "siegfried", defaultReportsPath)

	flag.StringVar(&sigfile, "sigs", defaultSigs, "set path to Siegfried signature file")
	flag.StringVar(&droid, "droid", defaultDroid, "set path to Droid signature file")
	flag.StringVar(&container, "container", defaultContainer, "set path to Droid Container signature file")
	flag.StringVar(&reports, "reports", defaultReports, "set path to Pronom reports directory")
}

func savereps() error {
	file, err := os.Open(reports)
	if err != nil {
		err = os.Mkdir(reports, os.ModeDir)
		if err != nil {
			log.Fatal(err)
		}
	}
	file.Close()
	errs := pronom.SaveReports(droid, pronom_url, reports)
	for _, e := range errs {
		fmt.Print(e)
	}
	if len(errs) > 0 {
		return fmt.Errorf("Errors saving reports to disk")
	}
	return nil
}

func makegob() error {
	p, err := pronom.NewIdentifier(droid, container, reports)
	if err != nil {
		return err
	}
	return p.Save(sigfile)
}

func inspectPronom() error {
	p, err := pronom.Load(sigfile)
	if err != nil {
		return err
	}
	fmt.Print(p)
	return nil
}

func viewSig(f string) error {
	sigs, err := pronom.ParsePuid(f, reports)
	if err != nil {
		return err
	}
	fmt.Println("Signatures:")
	for _, s := range sigs {
		fmt.Println(s)
	}
	bm, err := bytematcher.Signatures(sigs)
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

	pronom.Config.Timeout = *timeout
	var err error
	switch {
	case *harvest:
		err = savereps()
	case *build:
		err = makegob()
	case *inspect:
		err = inspectPronom()
	case *view != "":
		err = viewSig(*view)
	case *defaults:
		fmt.Println(droid)
		fmt.Println(container)
		fmt.Println(reports)
		fmt.Println(sigfile)
		fmt.Println(*timeout)
	}
	if err != nil {
		log.Fatal(err)
	}
}
