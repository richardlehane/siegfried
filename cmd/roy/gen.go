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
	"log"
	"path/filepath"
	"testing"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/loc"
	"github.com/richardlehane/siegfried/pkg/mimeinfo"
	"github.com/richardlehane/siegfried/pkg/pronom"
	"github.com/richardlehane/siegfried/pkg/sets"
)

var genhome = flag.String("home", "data", "override the default home directory")

type job func() error

func main() {
	jobs := []job{
		makeDefault,
		makeLoc,
		makeTika,
		makeFreedesktop,
		makePronomTikaLoc,
		makeDeluxe,
		makeArchivematica,
		makeSets,
	}
	for _, j := range jobs {
		if err := j(); err != nil {
			log.Fatal(err)
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

func makePronomTikaLoc() error {
	config.SetHome(*genhome)
	p, err := pronom.New(config.Clear())
	if err != nil {
		return err
	}
	m, err := mimeinfo.New(config.SetMIMEInfo("tika"))
	if err != nil {
		return err
	}
	l, err := loc.New(config.SetLOC(""))
	if err != nil {
		return err
	}
	return writeSigFile("pronom-tika-loc.sig", p, m, l)
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
	return writeSigFile("deluxe.sig", p, m, f, l)
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
