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
func _droid(p string) (map[string]string, error) {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	if len(b) < 3 {
		return nil, fmt.Errorf("Empty results")
	}
	switch string(b[:3]) {
	case `"ID`:
		return droidCSV(b)
	case "DRO":
		return droidRaw(b)
	}
	return nil, fmt.Errorf("Not a valid droid result set")
}

func droidRaw(b []byte) (map[string]string, error) {
	sep := []byte("Open archives: False")
	i := bytes.Index(b, sep)
	b = bytes.TrimSpace(b[i+len(sep):])
	s := bufio.NewScanner(bytes.NewReader(b))
	out := make(map[string]string)
	addFunc := resultAdder("Unknown", *droot)
	for ok := s.Scan(); ok; ok = s.Scan() {
		line := s.Bytes()
		idx := bytes.LastIndex(line, []byte{','})
		addFunc(out, string(line[idx+1:]), string(line[:idx]), string(line[idx+1:]))
	}
	return out, nil
}

func droidCSV(b []byte) (map[string]string, error) {
	rdr := csv.NewReader(bytes.NewReader(b))
	rdr.FieldsPerRecord = -1
	entries, err := rdr.ReadAll()
	if err != nil || len(entries) < 2 {
		return nil, fmt.Errorf("DROID error: either no results, or bad CSV. CSV err: %v", err)
	}
	out := make(map[string]string)
	addFunc := resultAdder("0", *droot)

	for _, v := range entries[1:] {
		switch v[8] {
		case "File", "Container":
			addFunc(out, v[13], v[3], v[14])
		}
		num, err := strconv.Atoi(v[13])
		// if we've got more than one PUID, grab it here
		if err == nil && num > 1 {
			for i := 1; i < num; i++ {
				addFunc(out, v[13], v[3], v[14+i*4])
			}
		}
	}
	return out, nil
}
*/
