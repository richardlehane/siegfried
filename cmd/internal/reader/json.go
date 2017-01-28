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

/*
type sfjresults struct {
	Files []*sfjres `json:"files"`
}

type sfjres struct {
	Filename string      `json:"filename"`
	Matches  []*sfjmatch `json:"matches"`
}

type sfjmatch struct {
	Puid string `json:"puid"`
}

func (m *sfjmatch) puid() string {
	return m.Puid
}

func sfJson(b []byte) (map[string]string, error) {
	doc := &sfjresults{}
	if err := json.Unmarshal(b, doc); err != nil {
		return nil, fmt.Errorf("JSON parsing error: %v", err)
	}
	out := make(map[string]string)
	convertJ := func(s []*sfjmatch) []puidMatch {
		ret := make([]puidMatch, len(s))
		for i := range s {
			ret[i] = puidMatch(s[i])
		}
		return ret
	}
	for _, v := range doc.Files {
		addMatches(out, prefix(v.Filename, *sroot), convertJ(v.Matches))
	}
	return out, nil
}
*/
