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

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/containermatcher"
	"github.com/richardlehane/siegfried/pkg/core/extensionmatcher"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/pronom/mappings"
)

type pronom struct {
	*Identifier
	j  parseable
	c  *mappings.Container
	pm priority.Map
}

// Pronom creates a pronom object
func newPronom() (*pronom, error) {
	p := &pronom{}
	if err := p.setParseables(); err != nil {
		return nil, err
	}
	if config.Inspect() {
		return p, nil
	}
	// apply noPriority rule
	if !config.NoPriority() {
		p.pm = p.j.priorities()
		p.pm.Complete()
	}
	// apply no container rule
	if !config.NoContainer() {
		if err := p.setContainers(); err != nil {
			return nil, fmt.Errorf("Pronom: error loading containers; got %s\nUnless you have set `-nocontainer` you need to download a container signature file", err)
		}
	}
	return p, nil
}

func (p *pronom) identifier() *Identifier {
	i := &Identifier{p: p}
	i.Name = config.Name()
	i.Details = config.Details()
	i.NoPriority = config.NoPriority()
	i.Infos = p.j.infos()
	p.Identifier = i
	return i
}

// set parseables joins signatures in the DROID signature file with any extra reports and adds that to the pronom object
func (p *pronom) setParseables() error {
	d, err := newDroid(config.Droid())
	if err != nil {
		return fmt.Errorf("Pronom: error loading Droid file; got %s\nYou must have a Droid file to build a signature", err)
	}
	// if we are just inspecting a single report file
	if config.Inspect() {
		r, err := newReports(config.Limit(d.puids()), nil)
		if err != nil {
			return fmt.Errorf("Pronom: error loading reports; got %s\nYou must download PRONOM reports to build a signature (unless you use the -noreports flag). You can use `roy harvest` to download reports", err)
		}
		infos := r.infos()
		sigs, puids, err := r.signatures()
		if err != nil {
			return fmt.Errorf("Pronom: parsing signatures; got %s", err)
		}
		var puid string
		for i, sig := range sigs {
			if puids[i] != puid {
				puid = puids[i]
				fmt.Printf("%s: \n", infos[puid].Name)
			}
			fmt.Println(sig)
		}
		fmt.Println()
		p.j = r
		return nil
	}
	// apply limit or exclude filters (only one can be applied)
	puids := d.puids()
	if config.HasLimit() {
		puids = config.Limit(puids)
	} else if config.HasExclude() {
		puids = config.Exclude(puids)
	}
	// if noreports set
	if config.Reports() == "" {
		p.j = d
		// apply filter
		if config.HasLimit() || config.HasExclude() {
			p.j = applyFilter(puids, p.j)
		}
	} else { // otherwise build from reports
		r, err := newReports(puids, d.idsPuids())
		if err != nil {
			return fmt.Errorf("Pronom: error loading reports; got %s\nYou must download PRONOM reports to build a signature (unless you use the -noreports flag). You can use `roy harvest` to download reports", err)
		}
		p.j = r
	}
	// add extensions
	for _, v := range config.Extend() {
		e, err := newDroid(v)
		if err != nil {
			return fmt.Errorf("Pronom: error loading extension file; got %s", err)
		}
		p.j = join(p.j, e)
	}
	// mirror PREV wild segments into EOF if maxBof and maxEOF set
	if config.MaxBOF() > 0 && config.MaxEOF() > 0 {
		p.j = &mirror{p.j}
	}
	return nil
}

func newDroid(path string) (*droid, error) {
	d := &mappings.Droid{}
	if err := openXML(path, d); err != nil {
		return nil, err
	}
	return &droid{d}, nil
}

func newReports(reps []string, idsPuids map[int]string) (*reports, error) {
	if len(reps) == 0 {
		return nil, fmt.Errorf("no valid PRONOM reports given")
	}
	indexes := make(map[string]int)
	for i, v := range reps {
		indexes[v] = i
	}
	r := &reports{reps, make([]*mappings.Report, len(reps)), idsPuids}
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
	p.c = &mappings.Container{}
	err := openXML(config.Container(), p.c)
	if err != nil {
		return err
	}
	for _, ex := range config.ExtendC() {
		c := &mappings.Container{}
		err = openXML(ex, c)
		if err != nil {
			return err
		}
		p.c.ContainerSignatures = append(p.c.ContainerSignatures, c.ContainerSignatures...)
		p.c.FormatMappings = append(p.c.FormatMappings, c.FormatMappings...)
	}
	return nil
}

// add adds extension, bytematcher or containermatcher signatures to the identifier
func (p *pronom) add(m core.Matcher) error {
	switch t := m.(type) {
	default:
		return fmt.Errorf("Pronom: unknown matcher type %T", t)
	case extensionmatcher.Matcher:
		var exts [][]string
		exts, p.EPuids = p.j.extensions()
		l, err := m.Add(extensionmatcher.SignatureSet(exts), nil)
		if err != nil {
			return err
		}
		p.EStart = l - len(p.EPuids)
		return nil
	case containermatcher.Matcher:
		return p.contMatcher(m)
	case *bytematcher.Matcher:
		var sigs []frames.Signature
		var err error
		sigs, p.BPuids, err = p.j.signatures()
		if err != nil {
			return err
		}
		var plist priority.List
		if !config.NoPriority() {
			plist = p.pm.List(p.BPuids)
		}
		l, err := m.Add(bytematcher.SignatureSet(sigs), plist)
		if err != nil {
			return err
		}
		p.BStart = l - len(p.BPuids)
	}
	return nil
}

// for limit/exclude filtering of containers
func (p pronom) hasPuid(puid string) bool {
	for _, v := range p.j.puids() {
		if puid == v {
			return true
		}
	}
	return false
}

func (p pronom) contMatcher(m core.Matcher) error {
	// when no container is set
	if p.c == nil {
		return nil
	}
	var zpuids, mpuids []string
	var zsigs, msigs [][]frames.Signature
	var znames, mnames [][]string
	cpuids := make(map[int]string)
	for _, fm := range p.c.FormatMappings {
		cpuids[fm.Id] = fm.Puid
	}
	for _, c := range p.c.ContainerSignatures {
		puid := cpuids[c.Id]
		// only include the included fmts
		if !p.hasPuid(puid) {
			continue
		}
		typ := c.ContainerType
		names := make([]string, 0, 1)
		sigs := make([]frames.Signature, 0, 1)
		for _, f := range c.Files {
			names = append(names, f.Path)
			sig, err := processDROID(puid, f.Signature.ByteSequences)
			if err != nil {
				return err
			}
			sigs = append(sigs, sig)
		}
		switch typ {
		case "ZIP":
			zpuids = append(zpuids, puid)
			znames = append(znames, names)
			zsigs = append(zsigs, sigs)
		case "OLE2":
			mpuids = append(mpuids, puid)
			mnames = append(mnames, names)
			msigs = append(msigs, sigs)
		default:
			return fmt.Errorf("Pronom: container parsing - unknown type %s", typ)
		}
	}
	// apply no priority config
	var zplist, mplist priority.List
	if !config.NoPriority() {
		zplist, mplist = p.pm.List(zpuids), p.pm.List(mpuids)
	}
	_, err := m.Add(containermatcher.SignatureSet{containermatcher.Zip, znames, zsigs}, zplist)
	if err != nil {
		return err
	}
	l, err := m.Add(containermatcher.SignatureSet{containermatcher.Mscfb, mnames, msigs}, mplist)
	if err != nil {
		return err
	}
	p.CPuids = append(zpuids, mpuids...)
	p.CStart = l - len(p.CPuids)
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
		url, _, _ := config.HarvestOptions()
		return save(puid, url, config.Reports())
	}
	return applyAll(5, d.puids(), apply)
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
	for _, puid := range reps {
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
	_, timeout, transport := config.HarvestOptions()
	req.Header.Add("User-Agent", "siegfried/r2d2bot (+https://github.com/richardlehane/siegfried)")
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
