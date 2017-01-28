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
func fido(p string) (map[string]string, error) {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	rdr := csv.NewReader(bytes.NewReader(b))
	entries, err := rdr.ReadAll()
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	addFunc := resultAdder("KO", *froot)
	for _, v := range entries {
		addFunc(out, v[0], v[6], v[2])
	}
	return out, nil
}
*/
