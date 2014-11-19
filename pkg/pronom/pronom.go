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
			return nil, err
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
		return err
	}
	// if we are just inspecting a single report file
	if config.Inspect() {
		r, err := newReports(config.Include(d.puids()), nil)
		if err != nil {
			return err
		}
		sigs, puids, err := r.signatures()
		if err != nil {
			return err
		}
		var puid string
		fmt.Println("BYTE SIGNATURES")
		for i, sig := range sigs {
			if puids[i] != puid {
				puid = puids[i]
				fmt.Printf("For %s: \n", puids[i])
			}
			fmt.Println(sig)
		}
		fmt.Println()
		p.j = r
		return nil
	}
	// if noreports set
	if config.Reports() == "" {
		p.j = d
		for _, v := range config.Extend() {
			e, err := newDroid(v)
			if err != nil {
				return err
			}
			p.j = join(p.j, e)
		}
		return nil
	}
	puids := d.puids()
	if config.HasInclude() {
		puids = config.Include(puids)
	} else if config.HasExclude() {
		puids = config.Exclude(puids)
	}
	r, err := newReports(puids, d.idsPuids())
	if err != nil {
		return err
	}
	p.j = r
	for _, v := range config.Extend() {
		e, err := newDroid(v)
		if err != nil {
			return err
		}
		p.j = join(p.j, e)
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
		return nil, nil
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
	errs := applyAll(reps, apply)
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
	return openXML(config.Container(), p.c)
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
		p.PuidsB = puidsB(p.BPuids)
		l, err := m.Add(bytematcher.SignatureSet(sigs), p.pm.List(p.BPuids))
		if err != nil {
			return err
		}
		p.BStart = l - len(p.BPuids)
	}
	return nil
}

func puidsB(puids []string) map[string][]int {
	pb := make(map[string][]int)
	for i, v := range puids {
		_, ok := pb[v]
		if ok {
			pb[v] = append(pb[v], i)
		} else {
			pb[v] = []int{i}
		}
	}
	return pb
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
		typ := c.ContainerType
		names := make([]string, 0, 1)
		sigs := make([]frames.Signature, 0, 1)
		for _, f := range c.Files {
			names = append(names, f.Path)
			sig, err := parseContainerSig(puid, f.Signature)
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
	_, err := m.Add(containermatcher.SignatureSet{containermatcher.Zip, znames, zsigs}, p.pm.List(zpuids))
	if err != nil {
		return err
	}
	l, err := m.Add(containermatcher.SignatureSet{containermatcher.Mscfb, mnames, msigs}, p.pm.List(mpuids))
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
	return applyAll(d.puids(), apply)
}

func openXML(path string, els interface{}) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return xml.Unmarshal(buf, els)
}

func applyAll(reps []string, apply func(puid string) error) []error {
	ch := make(chan error, len(reps))
	wg := sync.WaitGroup{}
	queue := make(chan struct{}, 200) // to prevent exhausting all the file descriptors
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
