package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var _ = pronom.Droid{} // import for side effects

var (
	defaultSigs      = filepath.Join(".", "pronom.gob")
	defaultDroid     = filepath.Join(".", "DROID_SignatureFile_V73.xml")
	defaultContainer = filepath.Join(".", "container-signature-20140227.xml")
	defaultReports   = filepath.Join(".", "pronom")
)

var (
	harvest     = flag.Bool("harvest", false, "harvest Pronom reports")
	build       = flag.Bool("build", false, "build a Siegfried signature file")
	statsf      = flag.Bool("stats", false, "describe a Siegfried signature file")
	printdroidf = flag.Bool("printdroid", false, "describe a Droid signature file")
	defaults    = flag.Bool("defaults", false, "print the default paths and settings")
	droid       = flag.String("droid", defaultDroid, "set path to Droid signature file")
	container   = flag.String("container", defaultContainer, "set path to Droid Container signature file")
	reports     = flag.String("reports", defaultReports, "set path to Pronom reports directory")
	sigfile     = flag.String("sigs", defaultSigs, "set path to Siegfried signature file")
	timeout     = flag.Duration("timeout", 120*time.Second, "set duration before timing-out harvesting requests e.g. 120s")
)

var pronom_url = "http://www.nationalarchives.gov.uk/pronom/"

func savereps() error {
	file, err := os.Open(*reports)
	if err != nil {
		err = os.Mkdir(*reports, os.ModeDir)
		if err != nil {
			log.Fatal(err)
		}
	}
	file.Close()
	errs := pronom.SaveReports(*droid, pronom_url, *reports)
	for _, e := range errs {
		fmt.Print(e)
	}
	if len(errs) > 0 {
		return fmt.Errorf("Errors saving reports to disk")
	}
	return nil
}

func makegob() error {
	p, err := pronom.NewIdentifier(*droid, *container, *reports)
	if err != nil {
		return err
	}
	return p.Save(*sigfile)
}

func stats() error {
	b, err := bytematcher.Load(*sigfile)
	if err != nil {
		return err
	}
	fmt.Print(b.Stats())
	return nil
}

func printdroid() error {
	p, err := pronom.New(*droid, *container, *reports)
	if err != nil {
		return err
	}
	fmt.Print(p)
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
	case *statsf:
		err = stats()
	case *printdroidf:
		err = printdroid()
	case *defaults:
		fmt.Println(*droid)
		fmt.Println(*container)
		fmt.Println(*reports)
		fmt.Println(*sigfile)
		fmt.Println(*timeout)
	}
	if err != nil {
		log.Fatal(err)
	}
}
