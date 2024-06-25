//go:build ignore
// +build ignore

// Copyright 2020 Richard Lehane. All rights reserved.
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

// gen.go updates signature and sets files
// invoke using `go generate`
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/loc"
	"github.com/richardlehane/siegfried/pkg/mimeinfo"
	"github.com/richardlehane/siegfried/pkg/pronom"
	"github.com/richardlehane/siegfried/pkg/sets"
	"github.com/richardlehane/siegfried/pkg/wikidata"
)

var genhome = flag.String("home", "data", "override the default home directory")

type job func() error

func main() {
	jobs := []job{
		makeDefault,
		makeLoc,
		makeTika,
		makeFreedesktop,
		makeDeluxe,
		makeArchivematica,
		makeSets,
		makeWikidata,
	}
	for i, j := range jobs {
		fmt.Printf("Running job %d\n", i)
		if err := j(); err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
}

func writeSigFile(name string, identifiers ...core.Identifier) error {
	s := siegfried.New()
	for _, id := range identifiers {
		if err := s.Add(id); err != nil {
			return err
		}
	}
	return s.Save(filepath.Join("data", name))
}

func makeDefault() error {
	config.SetHome(*genhome)
	p, err := pronom.New()
	if err != nil {
		return err
	}
	return writeSigFile(config.SignatureBase(), p)
}

func makeLoc() error {
	config.SetHome(*genhome)
	l, err := loc.New(config.SetLOC(""))
	if err != nil {
		return err
	}
	return writeSigFile("loc.sig", l)
}

func makeTika() error {
	config.SetHome(*genhome)
	m, err := mimeinfo.New(config.SetMIMEInfo("tika"))
	if err != nil {
		return err
	}
	return writeSigFile("tika.sig", m)
}

func makeFreedesktop() error {
	config.SetHome(*genhome)
	m, err := mimeinfo.New(config.SetMIMEInfo("freedesktop"))
	if err != nil {
		return err
	}
	return writeSigFile("freedesktop.sig", m)
}

func makeArchivematica() error {
	config.SetHome(*genhome)
	p, err := pronom.New(
		config.SetName("archivematica"),
		config.SetExtend(sets.Expand("archivematica-fmt2.xml,archivematica-fmt3.xml,archivematica-fmt4.xml,archivematica-fmt5.xml")))
	if err != nil {
		return err
	}
	return writeSigFile("archivematica.sig", p)
}

func makeWikidata() error {
	config.SetHome(*genhome)
	wikidataOpts := []config.Option{
		config.Clear(),
		config.SetWikidataNamespace(),
		config.SetWikidataNoPRONOM(),
	}
	w, err := wikidata.New(wikidataOpts...)
	if err != nil {
		return err
	}
	return writeSigFile("wikidata.sig", w)
}

func makeDeluxe() error {
	config.SetHome(*genhome)
	p, err := pronom.New(config.Clear())
	if err != nil {
		return err
	}
	m, err := mimeinfo.New(config.SetMIMEInfo("tika"))
	if err != nil {
		return err
	}
	f, err := mimeinfo.New(config.SetMIMEInfo("freedesktop"))
	if err != nil {
		return err
	}
	l, err := loc.New(config.SetLOC(""))
	if err != nil {
		return err
	}
	wikidataOpts := []config.Option{config.SetWikidataNamespace()}
	wikidataOpts = append(wikidataOpts, config.SetWikidataNoPRONOM())
	w, err := wikidata.New(wikidataOpts...)
	if err != nil {
		return err
	}
	return writeSigFile("deluxe.sig", p, m, f, l, w)
}

func makeSets() error {
	config.SetHome(*genhome)
	releases, err := pronom.LoadReleases(config.Local("release-notes.xml"))
	if err == nil {
		err = pronom.ReleaseSet("pronom-changes.json", releases)
	}
	if err == nil {
		err = pronom.TypeSets("pronom-all.json", "pronom-families.json", "pronom-types.json")
	}
	if err == nil {
		err = pronom.ExtensionSet("pronom-extensions.json")
	}
	return err
}
