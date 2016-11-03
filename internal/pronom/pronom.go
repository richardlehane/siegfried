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
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/richardlehane/siegfried/internal/config"
	"github.com/richardlehane/siegfried/internal/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/core/identifier"
	"github.com/richardlehane/siegfried/internal/pronom/mappings"
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

// Pronom creates a pronom object
func NewPronom() (identifier.Parseable, error) {
	p := &pronom{
		c: identifier.Blank{},
	}
	// apply no container rule
	if !config.NoContainer() {
		if err := p.setContainers(); err != nil {
			return nil, fmt.Errorf("Pronom: error loading containers; got %s\nUnless you have set `-nocontainer` you need to download a container signature file", err)
		}
	}
	if err := p.setParseables(); err != nil {
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
