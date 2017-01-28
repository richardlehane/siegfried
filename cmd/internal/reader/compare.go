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
	"io"
)

func Compare(w io.Writer, readers ...Reader) error {
	return nil
}

/*

type puidMatch interface {
	puid() string
}

func manyToOne(puids []string) string {
	if len(puids) == 0 {
		return UNKNOWN
	}
	sort.Strings(puids)
	return strings.Join(puids, SEP)
}

func addMatches(m map[string]string, name string, matches []puidMatch) {
	var l int
	for _, p := range matches {
		if p.puid() != "UNKNOWN" {
			l++
		}
	}
	puids := make([]string, l)
	var i int
	for _, p := range matches {
		if p.puid() != "UNKNOWN" {
			puids[i] = p.puid()
			i++
		}
	}
	m[clean(name)] = manyToOne(puids)
}

func oneToOne(puids, puid string) string {
	return manyToOne(append(strings.Split(puids, SEP), puid))
}

func resultAdder(unknown, p string) func(map[string]string, string, string, string) {
	return func(m map[string]string, u string, file string, puid string) {
		file = prefix(file, p)
		if u == unknown {
			m[clean(file)] = UNKNOWN
			return
		}
		if _, ok := m[clean(file)]; ok {
			m[clean(file)] = oneToOne(m[clean(file)], puid)
		} else {
			m[clean(file)] = puid
		}
	}
}


var replacer = strings.NewReplacer("''", "'")

func clean(path string) string {
	vol := filepath.VolumeName(path)
	return strings.TrimPrefix(filepath.Clean(replacer.Replace(path)), vol)
}

func prefix(n, p string) string {
	if p == "" {
		return n
	}
	return strings.TrimPrefix(n, p)
}

func run(comp map[string][4]string, fg *string, fn func(string) (map[string]string, error), idx, no int, members [4]bool) ([4]bool, int) {
	if *fg != "" {
		m, err := fn(*fg)
		if err != nil {
			log.Fatalf("Error reading %s: %v", *fg, err)
		}
		for k, v := range m {
			if w, ok := comp[k]; ok {
				w[idx] = v
				comp[k] = w
			} else {
				switch idx {
				case 0:
					comp[k] = [4]string{v, "", "", ""}
				case 1:
					comp[k] = [4]string{"", v, "", ""}
				case 2:
					comp[k] = [4]string{"", "", v, ""}
				case 3:
					comp[k] = [4]string{"", "", "", v}
				}
			}
		}
		members[idx] = true
		no++
		return members, no
	}
	return members, no
}

func _main() {
	flag.Parse()
	// collect results for each tool invoked. Results in form: filename -> matches
	comp := make(map[string][4]string)
	names := []string{"sf", "second sf", "droid", "fido"}
	var members [4]bool
	var no int // number of members
	members, no = run(comp, sfF, sf, 0, no, members)
	members, no = run(comp, sf2F, sf, 1, no, members)
	members, no = run(comp, droidF, _droid, 2, no, members)
	members, no = run(comp, fidoF, fido, 3, no, members)
	if no <= 1 {
		log.Fatalf("Must have more than one output to compare")
	}
	// now filter results for just the mismatches. Rearrange in form: matches -> filename (so sorted in output)
	results := make(map[[4]string][]string)
	matches := func(outs [4]string, members [4]bool) bool {
		var sub bool
		var this string
		for i, v := range members {
			if v {
				if !sub {
					this = outs[i]
					sub = true
					continue
				}
				if this != outs[i] {
					return false
				}
			}
		}
		return true
	}
	for k, v := range comp {
		if !matches(v, members) {
			if _, ok := results[v]; ok {
				results[v] = append(results[v], k)
			} else {
				results[v] = []string{k}
			}
		}
	}
	if len(results) == 0 {
		log.Println("COMPLETE MATCH")
		os.Exit(0)
	}
	// write output
	writer := csv.NewWriter(os.Stdout)
	header := func(names []string, members [4]bool, l int) []string {
		row := make([]string, l+1)
		row[0] = "filename"
		var i int
		for j, v := range members {
			if v {
				i++
				row[i] = names[j]
			}
		}
		return row
	}
	err := writer.Write(header(names, members, no))
	if err != nil {
		log.Fatal(err)
	}
	body := func(filename string, outs [4]string, members [4]bool, l int) []string {
		row := make([]string, l+1)
		row[0] = filename
		var i int
		for j, v := range members {
			if v {
				i++
				row[i] = outs[j]
				if row[i] == "" {
					row[i] = "MISSING"
				}
			}
		}
		return row
	}
	for k, v := range results {
		for _, w := range v {
			err := writer.Write(body(w, k, members, no))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	writer.Flush()
	log.Println("COMPLETE")
}
*/
