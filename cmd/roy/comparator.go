// A simple script to compare output of siegfried against DROID and FIDO

package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	UNKNOWN = "UNKNOWN"
	SEP     = "; "
)

var (
	sfF    = flag.String("sf", "", "path to siegfried result")
	sf2F   = flag.String("sf2", "", "path to second siegfried result")
	droidF = flag.String("droid", "", "path to droid result")
	fidoF  = flag.String("fido", "", "path to fido result")
	sroot  = flag.String("sr", "", "base path for reconciling file names between results")
	droot  = flag.String("dr", "", "base path for reconciling file names between results")
	froot  = flag.String("fr", "", "base path for reconciling file names between results")
)

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

func sf(p string) (map[string]string, error) {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	if len(b) < 1 {
		return nil, fmt.Errorf("Empty results")
	}
	switch b[0] {
	case '-':
		return sfYaml(b)
	case 'f':
		return sfCsv(b)
	case '{':
		return sfJson(b)
	}
	return nil, fmt.Errorf("Not a valid sf result set")
}

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

type sfyres struct {
	Filename string      `yaml:"filename"`
	Matches  []*sfymatch `yaml:"matches"`
}

type sfymatch struct {
	Puid string `yaml:"puid"`
	Id   string `yaml:"id"`
}

func (m *sfymatch) puid() string {
	if m.Puid == "" {
		return m.Id
	}
	return m.Puid
}

func sfYaml(b []byte) (map[string]string, error) {
	bs := bytes.Split(b, []byte("---\n"))
	if len(bs) < 3 {
		return nil, fmt.Errorf("Yaml error: not enough records to parse")
	}
	out := make(map[string]string)
	convertY := func(s []*sfymatch) []puidMatch {
		ret := make([]puidMatch, len(s))
		for i := range s {
			ret[i] = puidMatch(s[i])
		}
		return ret
	}
	for _, v := range bs[2:] {
		doc := &sfyres{}
		if err := yaml.Unmarshal(v, doc); err != nil {
			return nil, fmt.Errorf("Parsing error: %v; record is %s", err, v)
		}
		addMatches(out, prefix(doc.Filename, *sroot), convertY(doc.Matches))
	}
	return out, nil
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

func sfCsv(b []byte) (map[string]string, error) {
	rdr := csv.NewReader(bytes.NewReader(b))
	entries, err := rdr.ReadAll()
	if err != nil || len(entries) < 2 {
		return nil, fmt.Errorf("SF error: either no results, or bad YAML/CSV. CSV err: %v", err)
	}
	puidCol := 5
	// an extra column if we've done a hash
	if entries[0][puidCol] == "id" {
		puidCol = 6
	}
	out := make(map[string]string)
	addFunc := resultAdder(UNKNOWN, *sroot)

	for _, v := range entries[1:] {
		addFunc(out, v[puidCol], v[0], v[puidCol])
	}
	return out, nil
}

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
