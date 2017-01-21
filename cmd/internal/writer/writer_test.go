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
	"io/ioutil"
	"os"
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

func TestYAMLHeader(t *testing.T) {
	expect := "  - ns      : %v\n    id      : %v\n    format  : %v\n    version : %v\n    mime    : %v\n    basis   : %v\n    warning : %v\n"
	ret := header(makeFields())
	if expect != ret {
		t.Errorf("Expecting header to return %s\nGot: %s", expect, ret)
	}
}

func ExampleYAML() {
	yml := YAML(ioutil.Discard)
	yml.Head("", time.Time{}, [][2]string{{"pronom", ""}}, [][]string{makeFields()}, "")
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
	js.Head("", time.Time{}, [][2]string{{"pronom", ""}}, [][]string{makeFields()}, "")
	js.(*jsonWriter).w = bufio.NewWriter(os.Stdout)
	js.File("example.doc", 1, "2015-05-24T16:59:13+10:00", nil, testErr{}, []core.Identification{testID{}})
	js.Tail()
	// Output:
	// {"filename":"example.doc","filesize": 1,"modified":"2015-05-24T16:59:13+10:00","errors": "mscfb: bad OLE","matches": [{"ns":"pronom","id":"fmt/43","format":"JPEG File Interchange Format","version":"1.01","mime":"image/jpeg","basis":"extension match jpg; byte match at [[[0 14]] [[75201 2]]]","warning":""}]}]}
}
