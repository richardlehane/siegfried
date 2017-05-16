// Copyright 2017 Richard Lehane. All rights reserved.
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

package reader

import (
	"encoding/csv"
	"fmt"
	"io"
)

type sfCSV struct {
	rdr         *csv.Reader
	hh          string
	path        string
	fields      [][]string
	identifiers [][2]string
	peek        []string
	err         error
}

func newCSV(r io.Reader, path string) (Reader, error) {
	rdr := csv.NewReader(r)
	rec, err := rdr.Read()
	if err != nil || rec[0] != "filename" || len(rec) < 5 {
		return nil, fmt.Errorf("bad or invalid CSV: %v", err)
	}
	sfc := &sfCSV{
		rdr:  rdr,
		path: path,
	}
	var (
		fieldStart = 4
		fieldIdx   = -1
		fields     = make([][]string, 0, 1)
	)
	if rec[fieldStart] != "namespace" {
		sfc.hh = rec[4]
		fieldStart++
	}
	if rec[fieldStart] != "namespace" {
		return nil, fmt.Errorf("bad CSV, expecting field 'namespace' got %s", rec[fieldStart])
	}
	for _, v := range rec[fieldStart:] {
		if v == "namespace" {
			fieldIdx++
			fields = append(fields, make([]string, 0, 7))
		}
		fields[fieldIdx] = append(fields[fieldIdx], v)
	}
	sfc.fields = fields
	sfc.peek, err = rdr.Read()
	sfc.identifiers = make([][2]string, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("bad CSV, no results; got %v", err)
	}
	for i, v := range sfc.peek {
		if rec[i] == "namespace" {
			sfc.identifiers = append(sfc.identifiers, [2]string{v, ""})
		}
	}
	return sfc, nil
}

func (sfc *sfCSV) Head() Head {
	return Head{
		ResultsPath: sfc.path,
		Identifiers: sfc.identifiers,
		Fields:      sfc.fields,
		HashHeader:  sfc.hh,
	}
}

func (sfc *sfCSV) Next() (File, error) {
	if sfc.peek == nil || sfc.err != nil {
		return File{}, sfc.err
	}
	fieldStart := 4
	var hash string
	if sfc.hh != "" {
		hash = sfc.peek[fieldStart]
		fieldStart++
	}
	file, err := newFile(sfc.peek[0], sfc.peek[1], sfc.peek[2], hash, sfc.peek[3])
	if err != nil {
		return file, err
	}
	fn := sfc.peek[0]
	for {
		idStart := fieldStart
		for _, v := range sfc.fields {
			vals := sfc.peek[idStart : idStart+len(v)]
			if vals[0] != "" {
				file.IDs = append(file.IDs, newDefaultID(v, vals))
			}
			idStart += len(v)
		}
		sfc.peek, sfc.err = sfc.rdr.Read()
		if sfc.peek == nil || sfc.err != nil || fn != sfc.peek[0] {
			break
		}
	}
	return file, nil
}
