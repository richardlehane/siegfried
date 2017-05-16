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

var (
	fidoIDs    = [][2]string{{"fido", ""}}
	fidoFields = [][]string{{"ns", "id", "format", "full", "mime", "basis", "warning", "time"}}
)

type fido struct {
	rdr  *csv.Reader
	path string
	peek []string
	err  error
}

func newFido(r io.Reader, path string) (Reader, error) {
	fi := &fido{
		rdr:  csv.NewReader(r),
		path: path,
	}
	fi.peek, fi.err = fi.rdr.Read()
	if fi.err == nil && len(fi.peek) < 9 {
		fi.err = fmt.Errorf("not a valid fido results file, need 9 fields, got %d", len(fi.peek))
	}
	return fi, fi.err
}

func (fi *fido) Head() Head {
	return Head{
		ResultsPath: fi.path,
		Identifiers: fidoIDs,
		Fields:      fidoFields,
	}
}

func idVals(known, puid, format, full, mime, basis, time string) []string {
	var warn string
	if known == "KO" {
		puid = "UNKNOWN"
		warn = unknownWarn
	} else if basis == "extension" {
		warn = extWarn
	}
	if mime == "None" {
		mime = ""
	}
	return []string{"fido", puid, format, full, mime, basis, warn, time}
}

func (fi *fido) Next() (File, error) {
	if fi.peek == nil || fi.err != nil {
		return File{}, fi.err
	}
	file, err := newFile(fi.peek[6], fi.peek[5], "", "", "")
	fn := fi.peek[6]
	for {
		file.IDs = append(file.IDs, newDefaultID(fidoFields[0],
			idVals(fi.peek[0], fi.peek[2], fi.peek[3], fi.peek[4], fi.peek[7], fi.peek[8], fi.peek[1])))
		fi.peek, fi.err = fi.rdr.Read()
		if fi.peek == nil || fi.err != nil || fn != fi.peek[6] {
			break
		}
	}
	return file, err
}
