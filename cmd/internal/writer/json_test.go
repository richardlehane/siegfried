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
	"fmt"
	"testing"
)

func TestJSON(t *testing.T) {
	json := jsonizer(makeFields())
	expect := `{"ns":"pronom","id":"fmt/43","format":"JPEG File Interchange Format","version":"1.01","mime":"image/jpeg","basis":"extension match jpg; byte match at [[[0 14]] [[75201 2]]]","warning":""}`
	ret := json(values)
	if ret != expect {
		t.Errorf("Expecting jsonizer to return :\n%s\nGot:\n%s", expect, ret)
	}
}

func BenchmarkJSON(b *testing.B) {
	json := jsonizer(makeFields())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Sprintf("%s", json(values))
	}
}
