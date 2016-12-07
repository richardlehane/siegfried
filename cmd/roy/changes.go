package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var SFHome = "../../siegfried/cmd/roy/data"

type Releases struct {
	XMLName  xml.Name  `xml:"release_notes"`
	Releases []Release `xml:"release_note"`
}

type Release struct {
	ReleaseDate   string    `xml:"release_date"`
	SignatureName string    `xml:"signature_filename"`
	Outlines      []Outline `xml:"release_outline"`
}

type Outline struct {
	Typ   string `xml:"name,attr"`
	Puids []Puid `xml:"format>puid"`
}

type Puid struct {
	Typ string `xml:"type,attr"`
	Val string `xml:",chardata"`
}

type Droid struct {
	XMLName     xml.Name     `xml:"FFSignatureFile"`
	FileFormats []FileFormat `xml:"FileFormatCollection>FileFormat"`
}

type FileFormat struct {
	XMLName  xml.Name `xml:"FileFormat"`
	Puid     string   `xml:"PUID,attr"`
	Name     string   `xml:",attr"`
	Version  string   `xml:",attr"`
	families string
	types    string
}

func word(w string) string {
	w = strings.TrimSpace(w)
	w = strings.ToLower(w)
	ws := strings.Split(w, " ")
	w = ws[0]
	if len(ws) > 1 {
		for _, s := range ws[1:] {
			s = strings.TrimSuffix(strings.TrimPrefix(s, "("), ")")
			s = strings.Replace(s, "-", "", 1)
			w += strings.Title(s)
		}
	}
	return w
}

func normalise(ws string) []string {
	ss := strings.Split(ws, ",")
	for i, s := range ss {
		ss[i] = word(s)
	}
	if len(ss) == 1 && ss[0] == "" {
		return nil
	}
	return ss
}

func getTypes(puid string) ([]string, []string) {
	var (
		family, typ             string
		insideFamily, insideTyp bool
	)
	name := strings.Join(strings.Split(puid, "/"), "") + ".xml"
	name = filepath.Join(SFHome, "pronom", name)
	f, err := os.Open(name)
	if err != nil {
		return nil, nil
	}
	defer f.Close()
	dec := xml.NewDecoder(f)
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, nil
		}
		switch el := tok.(type) {
		default:
			continue
		case xml.StartElement:
			switch el.Name.Local {
			case "FormatFamilies":
				insideFamily = true
			case "FormatTypes":
				insideTyp = true
			}
		case xml.EndElement:
			switch el.Name.Local {
			case "FormatFamilies":
				insideFamily = false
			case "FormatTypes":
				return normalise(family), normalise(typ)
			}
		case xml.CharData:
			if insideFamily {
				family = string(el)
			} else if insideTyp {
				typ = string(el)
			}
		}
	}
	return normalise(family), normalise(typ) // shouldn't get here for well-formed xml
}

func makePuids(in []Puid) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = v.Typ + "/" + v.Val
	}
	return out
}

type KeyVal struct {
	Key string
	Val []string
}

// Define an ordered map
type OrderedMap []KeyVal

// Implement the json.Marshaler interface
func (omap OrderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{")
	for i, kv := range omap {
		if i != 0 {
			buf.WriteString(",")
		}
		// marshal key
		key, err := json.MarshalIndent(kv.Key, "", "  ")
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(":")
		// marshal value
		val, err := json.MarshalIndent(kv.Val, "", "  ")
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}
	buf.WriteString("}")
	return buf.Bytes(), nil
}

func nameType(in string) string {
	switch in {
	case "New Records":
		return "new"
	case "Updated Records":
		return "updated"
	case "New Signatures", "Signatures":
		return "signatures"
	}
	log.Fatalf("UNKNOWN NAME %s", in)
	return in
}

func checkType(in string) bool {
	switch in {
	case "New Records", "Updated Records", "New Signatures", "Signatures":
		return true
	}
	//log.Println(in)
	return false
}

func squares(num int) string {
	s := make([]string, num)
	for i := 0; i < num; i++ {
		s[i] = "\xE2\x96\xA0"
	}
	return strings.Join(s, " ")
}

func _amain() {
	byts, err := ioutil.ReadFile("release-notes.xml")
	if err != nil {
		log.Fatal(err)
	}
	var releases = &Releases{}
	if err := xml.Unmarshal(byts, releases); err != nil {
		log.Fatal(err)
	}
	relfrequency, newfrequency, upfrequency, sigfrequency := make(map[int]int), make(map[int]int), make(map[int]int), make(map[int]int)
	output := OrderedMap{}
	var latest string
	for i, release := range releases.Releases {
		if i == 0 {
			latest = release.SignatureName
		}
		trimdate := strings.TrimSpace(release.ReleaseDate)
		date := trimdate[len(trimdate)-4 : len(trimdate)]
		yr, _ := strconv.Atoi(date)
		relfrequency[yr]++
		bits := []KeyVal{}
		name := strings.TrimSuffix(strings.TrimPrefix(release.SignatureName, "DROID_SignatureFile_V"), ".xml")
		top := KeyVal{
			Key: name,
			Val: []string{},
		}
		for _, bit := range release.Outlines {
			if !checkType(bit.Typ) {
				continue
			}
			switch nameType(bit.Typ) {
			case "new":
				newfrequency[yr] += len(bit.Puids)
			case "updated":
				upfrequency[yr] += len(bit.Puids)
			case "signatures":
				sigfrequency[yr] += len(bit.Puids)
			}
			this := name + nameType(bit.Typ)
			top.Val = append(top.Val, "@"+this)
			bits = append(bits, KeyVal{
				Key: this,
				Val: makePuids(bit.Puids),
			})
		}
		output = append(output, top)
		output = append(output, bits...)
	}
	out, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(SFHome, "sets", "pronom-changes.json"), out, 0666); err != nil {
		log.Fatal(err)
	}
	years := make([]int, len(relfrequency))
	i := 0
	for k, _ := range relfrequency {
		years[i] = k
		i++
	}
	sort.Ints(years)
	for _, k := range years {
		fmt.Printf("%d\n", k)
		fmt.Printf("number releases: %s (%d)\n", squares(relfrequency[k]), relfrequency[k])
		fmt.Printf("new records:     %s (%d)\n", squares(newfrequency[k]/10), newfrequency[k])
		fmt.Printf("updated records: %s (%d)\n", squares(upfrequency[k]/10), upfrequency[k])
		fmt.Printf("new signatures:  %s (%d)\n\n", squares(sigfrequency[k]/10), sigfrequency[k])
	}
	byts, err = ioutil.ReadFile(filepath.Join(SFHome, latest))
	if err != nil {
		log.Fatal(err)
	}
	var droid = &Droid{}
	if err := xml.Unmarshal(byts, droid); err != nil {
		log.Fatal(err)
	}
	puids := make([]string, len(droid.FileFormats))
	families, types := make(map[string][]string), make(map[string][]string)
	for i, ff := range droid.FileFormats {
		this := ff.Puid
		fss, tss := getTypes(this)
		var nm string
		if ff.Name == "" {
			nm = " (" + ff.Version + ")"
		} else if ff.Version == "" {
			nm = " (" + ff.Name + ")"
		} else {
			nm = " (" + ff.Name + " " + ff.Version + ")"
		}
		this += nm
		for _, fs := range fss {
			families[fs] = append(families[fs], this)
		}
		for _, ts := range tss {
			types[ts] = append(types[ts], this)
		}
		puids[i] = this
	}
	out, err = json.MarshalIndent(map[string][]string{"all": puids}, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(SFHome, "sets", "pronom-all.json"), out, 0666); err != nil {
		log.Fatal(err)
	}
	out, err = json.MarshalIndent(families, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(SFHome, "sets", "pronom-families.json"), out, 0666); err != nil {
		log.Fatal(err)
	}
	out, err = json.MarshalIndent(types, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(SFHome, "sets", "pronom-types.json"), out, 0666); err != nil {
		log.Fatal(err)
	}
}
