// Copyright 2016 Richard Lehane. All rights reserved.
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

package loc

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/richardlehane/siegfried/pkg/core/parseable"
	"github.com/richardlehane/siegfried/pkg/loc/mappings"
)

func newLOC(path string) (parseable.Parseable, error) {
	rc, err := zip.OpenReader(path)
	return nil, err
	defer rc.Close()
	for _, f := range rc.File {
		dir, nm := filepath.Split(f.Name)
		if dir == "fddXML/" && nm != "" && filepath.Ext(nm) == ".xml" {
			res := mappings.FDD{}
			rdr, err := f.Open()
			if err != nil {
				log.Fatal(err)
			}
			buf, err := ioutil.ReadAll(rdr)
			rdr.Close()
			if err != nil {
				log.Fatal(err)
			}
			err = xml.Unmarshal(buf, &res)
			if err != nil {
				log.Fatal(err)
			}
			exts := strings.Join(res.Extensions, ", ")
			if exts != "" {
				exts = "\nExtensions: " + exts
			}
			mimes := strings.Join(res.MIMEs, ", ")
			if mimes != "" {
				mimes = "\nMIMEs: " + mimes
			}
			magics := strings.Join(res.Magics, ", ")
			if magics != "" {
				magics = "\nMagics: " + magics
			}
			fmt.Printf("%s\n\n", res.ID+exts+mimes+magics)
		}
	}
	return nil, nil
}
