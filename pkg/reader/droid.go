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
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/richardlehane/siegfried/internal/checksum"
)

const droidTime = "2006-01-02T15:04:05"

var (
	droidIDs      = [][2]string{{"pronom", ""}}
	droidFields   = [][]string{{"ns", "id", "format", "version", "mime", "basis", "warning"}}
	droidNpFields = [][]string{{"ns", "id", "warning"}}
)

type droid struct {
	rdr  *csv.Reader
	hh   string
	path string
	peek []string
	err  error
}

func newDroid(r io.Reader, path string) (Reader, error) {
	rdr := csv.NewReader(r)
	rdr.FieldsPerRecord = -1
	//rdr.LazyQuotes = true
	rec, err := rdr.Read()
	if err != nil || rec[0] != "ID" || len(rec) < 17 {
		return nil, fmt.Errorf("bad or invalid DROID CSV: %v", err)
	}
	dr := &droid{
		rdr:  rdr,
		path: path,
	}
	cs := checksum.GetHash(strings.TrimSuffix(rec[12], "_HASH"))
	if cs >= 0 {
		dr.hh = cs.String()
	}
	return dr, dr.nextFile()
}

func (dr *droid) nextFile() error {
	for {
		dr.peek, dr.err = dr.rdr.Read()
		if dr.err != nil {
			return fmt.Errorf("bad or invalid DROID CSV: %v", dr.err)
		}
		if len(dr.peek) > 8 && dr.peek[8] != "Folder" {
			return nil
		}
	}
}

func (dr *droid) Head() Head {
	return Head{
		ResultsPath: dr.path,
		Identifiers: droidIDs,
		Fields:      droidFields,
		HashHeader:  dr.hh,
	}
}

func didVals(puid, format, version, mime, basis, mismatch string) []string {
	var warn string
	if mismatch == "true" {
		warn = extMismatch
	} else if basis == "Extension" {
		warn = extWarn
	} else if puid == "" {
		warn = unknownWarn
	}
	if puid == "" {
		puid = "UNKNOWN"
	}
	return []string{droidIDs[0][0], puid, format, version, mime, strings.ToLower(basis), warn}
}

func (dr *droid) Next() (File, error) {
	if dr.peek == nil || dr.err != nil {
		return File{}, dr.err
	}
	file, err := newFile(dr.peek[3], dr.peek[7], dr.peek[10], dr.peek[12], "")
	fn := dr.peek[3]
	for {
		file.IDs = append(file.IDs, newDefaultID(droidFields[0],
			didVals(dr.peek[14], dr.peek[16], dr.peek[17], dr.peek[15], dr.peek[5], dr.peek[11])))
		// single line multi ids
		if len(dr.peek) > 18 {
			num, err := strconv.Atoi(dr.peek[13])
			if err == nil && num > 1 {
				for i := 1; i < num; i++ {
					file.IDs = append(file.IDs, newDefaultID(droidFields[0],
						didVals(dr.peek[14+i*4], dr.peek[16+i*4], dr.peek[17+i*4], dr.peek[15+i*4], dr.peek[5], dr.peek[11])))
				}
			}
		}
		// multi line multi ids
		err := dr.nextFile()
		if err != nil || fn != dr.peek[3] {
			break
		}
	}
	return file, err
}

type droidNp struct {
	buf  *bufio.Reader
	path string
	ids  [][2]string
	peek []string
	err  error
}

func newDroidNp(r io.Reader, path string) (Reader, error) {
	dnp := &droidNp{
		buf:  bufio.NewReader(r),
		path: path,
		ids:  make([][2]string, 1),
	}
	dnp.ids[0][0] = droidIDs[0][0]
	var (
		sigs []string
		byts []byte
		err  error
	)
	for {
		byts, err = dnp.buf.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		if bytes.HasPrefix(byts, []byte("Binary signature file: ")) {
			sigs = append(sigs, string(byts))
		} else if bytes.HasPrefix(byts, []byte("Container signature file: ")) {
			sigs = append(sigs, string(byts))
		}
		if !bytes.Contains(byts, []byte(": ")) {
			break
		}
	}
	dnp.ids[0][1] = strings.Join(sigs, "; ")
	return dnp, dnp.setPeek(byts)
}

func (dnp *droidNp) advance() {
	byts, err := dnp.buf.ReadBytes('\n')
	if err != nil {
		dnp.err = err
		return
	}
	dnp.err = dnp.setPeek(byts)
}

func (dnp *droidNp) setPeek(byts []byte) error {
	idx := bytes.LastIndex(byts, []byte{','})
	if idx < 0 {
		if strings.TrimSpace(string(byts)) == "" {
			return io.EOF
		}
		return fmt.Errorf("bad droid no profile file; line without comma separator: %v", byts)
	}
	var fn, puid string
	fn = string(byts[:idx])
	if idx < len(byts)-2 {
		puid = strings.TrimSpace(string(byts[idx+1:]))
	}
	dnp.peek = []string{fn, puid}
	return nil
}

func (dnp *droidNp) Head() Head {
	return Head{
		ResultsPath: dnp.path,
		Identifiers: dnp.ids,
		Fields:      droidNpFields,
	}
}

func (dnp *droidNp) Next() (File, error) {
	if dnp.peek == nil || dnp.err != nil {
		return File{}, dnp.err
	}
	file, err := newFile(dnp.peek[0], "", "", "", "")
	fn := dnp.peek[0]
	for {
		var puid, warn string
		puid = dnp.peek[1]
		if puid == "Unknown" {
			puid = "UNKNOWN"
			warn = unknownWarn
		}
		file.IDs = append(file.IDs, newDefaultID(droidNpFields[0],
			[]string{droidIDs[0][0], puid, warn}))
		// multi line multi ids
		dnp.advance()
		if dnp.err != nil || fn != dnp.peek[0] {
			break
		}
	}
	return file, err
}
