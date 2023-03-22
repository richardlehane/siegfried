// Copyright 2014 Richard Lehane. All rights reserved.
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

// Package pronom implements the TNA's PRONOM signatures as a siegfried identifier
package pronom

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/pronom/internal/mappings"
)

type pronom struct {
	identifier.Parseable
	c identifier.Parseable
}

// add container IDs to the DROID IDs (this ensures container extensions register)
func (p *pronom) IDs() []string {
	ids := make([]string, len(p.Parseable.IDs()), len(p.Parseable.IDs())+len(p.c.IDs()))
	copy(ids, p.Parseable.IDs())
	for _, id := range p.c.IDs() {
		var present bool
		for _, ida := range p.Parseable.IDs() {
			if id == ida {
				present = true
				break
			}
		}
		if !present {
			ids = append(ids, id)
		}
	}
	return ids
}

func (p *pronom) Zips() ([][]string, [][]frames.Signature, []string, error) {
	return p.c.Zips()
}

func (p *pronom) MSCFBs() ([][]string, [][]frames.Signature, []string, error) {
	return p.c.MSCFBs()
}

// return a PRONOM object without applying the config
func raw() (identifier.Parseable, error) {
	p := &pronom{
		c: identifier.Blank{},
	}
	// apply no container rule
	if !config.NoContainer() {
		if err := p.setContainers(); err != nil {
			return nil, fmt.Errorf("pronom: error loading containers; got %s\nUnless you have set `-nocontainer` you need to download a container signature file", err)
		}
	}
	if err := p.setParseables(); err != nil {
		return nil, err
	}
	return p, nil
}

// Pronom creates a pronom object
func NewPronom() (identifier.Parseable, error) {
	p, err := raw()
	if err != nil {
		return nil, err
	}
	return identifier.ApplyConfig(p), nil
}

// set identifiers joins signatures in the DROID signature file with any extra reports and adds that to the pronom object
func (p *pronom) setParseables() error {
	d, err := newDroid(config.Droid())
	if err != nil {
		return fmt.Errorf("Pronom: error loading Droid file; got %s\nYou must have a Droid file to build a signature", err)
	}

	// if noreports set
	if config.Reports() == "" {
		p.Parseable = d
	} else { // otherwise build from reports
		// get list of puids that applies limit or exclude filters (actual filtering of Parseable delegated to core/identifier)
		puids := d.IDs()
		if config.HasLimit() {
			puids = config.Limit(puids)
		} else if config.HasExclude() {
			puids = config.Exclude(puids)
		}
		r, err := newReports(puids, d.idsPuids())
		if err != nil {
			return fmt.Errorf("Pronom: error loading reports; got %s\nYou must download PRONOM reports to build a signature (unless you use the -noreports flag). You can use `roy harvest` to download reports", err)
		}
		p.Parseable = r
	}
	// add extensions
	for _, v := range config.Extend() {
		e, err := newDroid(v)
		if err != nil {
			return fmt.Errorf("Pronom: error loading extension file; got %s", err)
		}
		p.Parseable = identifier.Join(p.Parseable, e)
	}
	// exclude byte signatures where also have container signatures, unless doubleup set
	if !config.DoubleUp() {
		p.Parseable = doublesFilter{
			config.ExcludeDoubles(p.IDs(), p.c.IDs()),
			p.Parseable,
		}
	}
	return nil
}

func newDroid(path string) (*droid, error) {
	d := &mappings.Droid{}
	if err := openXML(path, d); err != nil {
		return nil, err
	}
	return &droid{d, identifier.Blank{}}, nil
}

func newReports(reps []string, idsPuids map[int]string) (*reports, error) {
	r := &reports{reps, make([]*mappings.Report, len(reps)), idsPuids, identifier.Blank{}}
	if len(reps) == 0 {
		return r, nil // empty signatures
	}
	indexes := make(map[string]int)
	for i, v := range reps {
		indexes[v] = i
	}
	apply := func(puid string) error {
		idx := indexes[puid]
		r.r[idx] = &mappings.Report{}
		return openXML(reportPath(puid), r.r[idx])
	}
	errs := applyAll(200, reps, apply)
	if len(errs) > 0 {
		strs := make([]string, len(errs))
		for i, v := range errs {
			strs[i] = v.Error()
		}
		return nil, fmt.Errorf(strings.Join(strs, "\n"))
	}
	return r, nil
}

func reportPath(puid string) string {
	return filepath.Join(config.Reports(), strings.Replace(puid, "/", "", 1)+".xml")
}

// setContainers adds containers to a pronom object. It takes as an argument the path to a container signature file
func (p *pronom) setContainers() error {
	c := &mappings.Container{}
	err := openXML(config.Container(), c)
	if err != nil {
		return err
	}
	for _, ex := range config.ExtendC() {
		c1 := &mappings.Container{}
		err = openXML(ex, c1)
		if err != nil {
			return err
		}
		c.ContainerSignatures = append(c.ContainerSignatures, c1.ContainerSignatures...)
		c.FormatMappings = append(c.FormatMappings, c1.FormatMappings...)
	}
	p.c = &container{c, identifier.Blank{}}
	return nil
}

// UTILS
// Harvest fetches PRONOM reports listed in the DROID file
func Harvest() []error {
	d, err := newDroid(config.Droid())
	if err != nil {
		return []error{err}
	}
	apply := func(puid string) error {
		url, _, _, _ := config.HarvestOptions()
		return save(puid, url, config.Reports())
	}
	return applyAll(5, d.IDs(), apply)
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
	return in
}

func checkType(in string) bool {
	switch in {
	case "New Records", "Updated Records", "New Signatures", "Signatures":
		return true
	}
	return false
}

func GetReleases(path string) error {
	byts, err := getHttp(config.ChangesURL())
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, byts, os.ModePerm)
}

func LoadReleases(path string) (*mappings.Releases, error) {
	releases := &mappings.Releases{}
	err := openXML(path, releases)
	return releases, err
}

func Releases(releases *mappings.Releases) ([]string, []string, map[string]map[string]int) {
	changes := make(map[string]map[string]int)
	fields := []string{"number releases", "new records", "updated records", "new signatures"}
	for _, release := range releases.Releases {
		trimdate := strings.TrimSpace(release.ReleaseDate)
		yr := trimdate[len(trimdate)-4:]
		if changes[yr] == nil {
			changes[yr] = make(map[string]int)
		}
		changes[yr][fields[0]]++
		for _, bit := range release.Outlines {
			if !checkType(bit.Typ) {
				continue
			}
			switch nameType(bit.Typ) {
			case "new":
				changes[yr][fields[1]] += len(bit.Puids)
			case "updated":
				changes[yr][fields[2]] += len(bit.Puids)
			case "signatures":
				changes[yr][fields[3]] += len(bit.Puids)
			}
		}
	}
	yrs := make([]int, 0, len(changes))
	for k := range changes {
		i, _ := strconv.Atoi(k)
		yrs = append(yrs, i)
	}
	sort.Ints(yrs)
	years := make([]string, len(yrs))
	for i, v := range yrs {
		years[i] = strconv.Itoa(v)
	}
	return years, fields, changes
}

func makePuids(in []mappings.Puid) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = v.Typ + "/" + v.Val
	}
	return out
}

// ReleaseSet writes a changes sets file based on the latest PRONOM release file
func ReleaseSet(path string, releases *mappings.Releases) error {
	output := mappings.OrderedMap{}
	for _, release := range releases.Releases {
		bits := []mappings.KeyVal{}
		name := strings.TrimSuffix(strings.TrimPrefix(release.SignatureName, "DROID_SignatureFile_V"), ".xml")
		top := mappings.KeyVal{
			Key: name,
			Val: []string{},
		}
		for _, bit := range release.Outlines {
			if !checkType(bit.Typ) {
				continue
			}
			this := name + nameType(bit.Typ)
			top.Val = append(top.Val, "@"+this)
			bits = append(bits, mappings.KeyVal{
				Key: this,
				Val: makePuids(bit.Puids),
			})
		}
		output = append(output, top)
		output = append(output, bits...)
	}
	out, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(config.Local("sets"), path), out, 0666)
}

// TypeSets writes three sets files based on PRONOM reports:
// an all sets files, with all PUIDs; a families sets file with FormatFamilies; and a types sets file with FormatTypes.
func TypeSets(p1, p2, p3 string) error {
	d, err := newDroid(config.Droid())
	if err != nil {
		return err
	}
	r, err := newReports(d.IDs(), d.idsPuids())
	if err != nil {
		return err
	}
	families, types := r.FamilyTypes()
	all := r.Labels()
	out, err := json.MarshalIndent(map[string][]string{"all": all}, "", "  ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(config.Local("sets"), p1), out, 0666); err != nil {
		return err
	}
	out, err = json.MarshalIndent(families, "", "  ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(config.Local("sets"), p2), out, 0666); err != nil {
		return err
	}
	out, err = json.MarshalIndent(types, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(config.Local("sets"), p3), out, 0666)
}

// Extension set writes a sets file that links extensions to IDs.
func ExtensionSet(path string) error {
	d, err := newDroid(config.Droid())
	if err != nil {
		return err
	}
	r, err := newReports(d.IDs(), d.idsPuids())
	if err != nil {
		return err
	}
	exts, puids := r.Globs()
	extM := make(map[string][]string)
	for i, e := range exts {
		if len(e) > 0 {
			e = strings.TrimPrefix(e, "*")
			extM[e] = append(extM[e], puids[i])
		}
	}
	out, err := json.MarshalIndent(extM, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(config.Local("sets"), path), out, 0666)
}

func openXML(path string, els interface{}) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return xml.Unmarshal(buf, els)
}

func applyAll(max int, reps []string, apply func(puid string) error) []error {
	ch := make(chan error, len(reps))
	wg := sync.WaitGroup{}
	queue := make(chan struct{}, max) // to avoid hammering TNA
	_, _, tf, _ := config.HarvestOptions()
	var throttle *time.Ticker
	if tf > 0 {
		throttle = time.NewTicker(tf)
		defer throttle.Stop()
	}
	for _, puid := range reps {
		if tf > 0 {
			<-throttle.C
		}
		wg.Add(1)
		go func(puid string) {
			queue <- struct{}{}
			defer wg.Done()
			if err := apply(puid); err != nil {
				ch <- err
			}
			<-queue
		}(puid)
	}
	wg.Wait()
	close(ch)
	var errors []error
	for err := range ch {
		errors = append(errors, err)
	}
	return errors
}

func getHttp(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	_, timeout, _, transport := config.HarvestOptions()
	req.Header.Add("User-Agent", "siegfried/roybot (+https://github.com/richardlehane/siegfried)")
	timer := time.AfterFunc(timeout, func() {
		transport.CancelRequest(req)
	})
	defer timer.Stop()
	client := http.Client{
		Transport: transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func save(puid, url, path string) error {
	b, err := getHttp(url + puid + ".xml")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(path, strings.Replace(puid, "/", "", 1)+".xml"), b, os.ModePerm)
}
