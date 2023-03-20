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

package writer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

var testValues = []string{"pronom",
	"fmt/43",
	"JPEG File Interchange Format",
	"1.01",
	"image/jpeg",
	"extension match jpg; byte match at [[[0 14]] [[75201 2]]]",
	""}

type testErr struct{}

func (t testErr) Error() string { return "mscfb: bad OLE" }

type testID struct{}

func (t testID) String() string          { return testValues[1] }
func (t testID) Known() bool             { return true }
func (t testID) Warn() string            { return "" }
func (t testID) Values() []string        { return testValues }
func (t testID) Archive() config.Archive { return 0 }

func makeFields() []string {
	return []string{"namespace",
		"id",
		"format",
		"version",
		"mime",
		"basis",
		"warning"}
}

var controlCharacters = []string{"\u0000", "\u0001", "\u0002", "\u0003",
	"\u0004", "\u0005", "\u0006", "\u0007", "\u0008", "\u0009", "\u000A",
	"\u000B", "\u000C", "\u000D", "\u000E", "\u000F", "\u0010", "\u0011",
	"\u0012", "\u0013", "\u0014", "\u0015", "\u0016", "\u0017", "\u0018",
	"\u0019",
}
var nonControlCharacters = []string{"\u0020", "\u1F5A4", "\u265B", "\u1F0A1",
	"\u262F",
}

// TestJSONControlCharacters tests control characters that are valid but
// need special treatment from the writer and makes sure that they
// create invalid JSON.
func TestControlCharacters(t *testing.T) {
	buf := &bytes.Buffer{}
	js := JSON(buf)
	js.Head("", time.Time{}, time.Time{}, [3]int{}, [][2]string{{"pronom", ""}}, [][]string{makeFields()}, "")
	// Loop through the control characters to make sure the JSON output
	// is valid.
	for _, val := range controlCharacters {
		js.File(fmt.Sprintf("path/%sto/file", val), 1, "2015-05-24T16:59:13+10:00", nil, testErr{}, []core.Identification{testID{}})
	}
	js.Tail()
	if !json.Valid(buf.Bytes()) {
		t.Errorf("Invalid JSON:\n%s", buf.String())
	}
}

// TestNonControlCharacters tests valid characters and simply makes sure
// that the JSON output is correct.
func TestNonControlCharacters(t *testing.T) {
	buf := &bytes.Buffer{}
	js := JSON(buf)
	js.Head("", time.Time{}, time.Time{}, [3]int{}, [][2]string{{"pronom", ""}}, [][]string{makeFields()}, "")
	// Loop through the non control characters to make sure the JSON output
	// is valid.
	for _, val := range nonControlCharacters {
		js.File(fmt.Sprintf("path/%sto/file", val), 1, "2015-05-24T16:59:13+10:00", nil, testErr{}, []core.Identification{testID{}})
	}
	js.Tail()
	if !json.Valid(buf.Bytes()) {
		t.Errorf("Invalid JSON:\n%s", buf.String())
	}
}

func TestYAMLHeader(t *testing.T) {
	expect := "  - ns      : %v\n    id      : %v\n    format  : %v\n    version : %v\n    mime    : %v\n    basis   : %v\n    warning : %v\n"
	ret := header(makeFields())
	if expect != ret {
		t.Errorf("Expecting header to return %s\nGot: %s", expect, ret)
	}
}

func TestYAMLMultilineString(t *testing.T) {
	buf := &bytes.Buffer{}
	yml := YAML(buf)
	yml.Head("", time.Time{}, time.Time{}, [3]int{}, [][2]string{{"pronom", ""}}, [][]string{makeFields()}, "")
	yml.File("example.\ndoc", 1, "2015-05-24T16:59:13+10:00", nil, testErr{}, []core.Identification{testID{}})
	yml.Tail()
	expect :=
		`---
siegfried   : 0.0.0
scandate    : 0001-01-01T00:00:00Z
signature   : 
created     : 0001-01-01T00:00:00Z
identifiers : 
  - name    : 'pronom'
    details : ''
---
filename : "example.\ndoc"
filesize : 1
modified : 2015-05-24T16:59:13+10:00
errors   : 'mscfb: bad OLE'
matches  :
  - ns      : 'pronom'
    id      : 'fmt/43'
    format  : 'JPEG File Interchange Format'
    version : '1.01'
    mime    : 'image/jpeg'
    basis   : 'extension match jpg; byte match at [[[0 14]] [[75201 2]]]'
    warning : 
`
	ret := buf.String()
	if expect != ret {
		var detail string
		if len(expect) != len(ret) {
			detail = fmt.Sprintf("Strings differ in length, %d vs %d", len(expect), len(ret))
		} else {
			for i := range expect {
				if expect[i] != ret[i] {
					detail = fmt.Sprintf("Strings differ at index %d, %d vs %d", i, expect[i], ret[i])
				}
			}
		}
		t.Errorf("Expecting return: %s\nGot: %s\n%s", expect, ret, detail)
	}
}

// TestDroidHeader ensures that the DROID header is output consistently
// and matches the DROID CSV specification.
func TestDroidHeader(t *testing.T) {
	buf := &bytes.Buffer{}
	droid := Droid(buf)
	droid.Head("", time.Time{}, time.Time{}, [3]int{}, [][2]string{{"pronom", ""}}, [][]string{makeFields()}, "md5")
	droid.Tail()
	// DROID identification result isn't tested here as the paths output
	// are absolute and require a bit of finessing in SF to get right.
	expected := `ID,PARENT_ID,URI,FILE_PATH,NAME,METHOD,STATUS,SIZE,TYPE,EXT,LAST_MODIFIED,EXTENSION_MISMATCH,MD5_HASH,FORMAT_COUNT,PUID,MIME_TYPE,FORMAT_NAME,FORMAT_VERSION`
	res := strings.Trim(buf.String(), "\n")
	if res != expected {
		t.Errorf("DROID output didn't output as expected: \n'%s' got: \n'%s'", res, expected)
	}
}

func ExampleYAML() {
	yml := YAML(ioutil.Discard)
	yml.Head("", time.Time{}, time.Time{}, [3]int{}, [][2]string{{"pronom", ""}}, [][]string{makeFields()}, "")
	yml.(*yamlWriter).w = bufio.NewWriter(os.Stdout)
	yml.File("example.doc", 1, "2015-05-24T16:59:13+10:00", nil, testErr{}, []core.Identification{testID{}})
	yml.Tail()
	// Output:
	// ---
	// filename : 'example.doc'
	// filesize : 1
	// modified : 2015-05-24T16:59:13+10:00
	// errors   : 'mscfb: bad OLE'
	// matches  :
	//   - ns      : 'pronom'
	//     id      : 'fmt/43'
	//     format  : 'JPEG File Interchange Format'
	//     version : '1.01'
	//     mime    : 'image/jpeg'
	//     basis   : 'extension match jpg; byte match at [[[0 14]] [[75201 2]]]'
	//     warning :
}

func ExampleJSON() {
	js := JSON(ioutil.Discard)
	js.Head("", time.Time{}, time.Time{}, [3]int{}, [][2]string{{"pronom", ""}}, [][]string{makeFields()}, "")
	js.(*jsonWriter).w = bufio.NewWriter(os.Stdout)
	js.File("example.doc", 1, "2015-05-24T16:59:13+10:00", nil, testErr{}, []core.Identification{testID{}})
	js.Tail()
	// Output:
	// {"filename":"example.doc","filesize": 1,"modified":"2015-05-24T16:59:13+10:00","errors": "mscfb: bad OLE","matches": [{"ns":"pronom","id":"fmt/43","format":"JPEG File Interchange Format","version":"1.01","mime":"image/jpeg","basis":"extension match jpg; byte match at [[[0 14]] [[75201 2]]]","warning":""}]}]}
}
