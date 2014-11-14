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

	. "github.com/richardlehane/siegfried/pkg/pronom/mappings"
)

type pronom struct {
	*Identifier
	droid     *Droid
	container *Container
	reports   []*Report
	puids     map[string]int // map of puids to File Format indexes
	ids       map[int]string // map of droid FileFormatIDs to puids
	ps        priority.Map
}

func (p pronom) String() string {
	return p.droid.String()
}

// newPronom creates a pronom object.
func NewPronom() (*pronom, error) {
	p := new(pronom)
	if err := p.setDroid(); err != nil {
		return p, err
	}
	if err := p.setContainers(); err != nil {
		return p, err
	}
	errs := p.setReports()
	if len(errs) > 0 {
		var str string
		for _, e := range errs {
			str += fmt.Sprintln(e)
		}
		return p, fmt.Errorf(str)
	}
	p.ps = p.priorities()
	return p, nil
}

func (p *pronom) identifier() *Identifier {
	i := &Identifier{p: p}
	i.Name = config.Name()
	i.Details = config.Details()
	i.NoPriority = config.NoPriority()
	i.Infos = p.GetInfos()
	i.BPuids, i.PuidsB = p.GetPuids()
	p.Identifier = i
	return i
}

func (p *pronom) add(m core.Matcher) error {
	switch t := m.(type) {
	default:
		return fmt.Errorf("Pronom: unknown matcher type %T", t)
	case extensionmatcher.Matcher:
		return p.extMatcher(m)
	case containermatcher.Matcher:
		return p.contMatcher(m)
	case *bytematcher.Matcher:
		/*
			sigs, _, _, _, _, _, err := p.Parse()
			if err != nil {
				return err
			}
			l, err := m.Add(bytematcher.SignatureSet(sigs), p.ps.List(p.BPuids))
			if err != nil {
				return err
			}
			p.BStart = l - len(p.BPuids)*/
	}
	return nil
}

func (p pronom) signatures() []Signature {
	sigs := make([]Signature, 0, 1000)
	//for _, f := range p.droid.FileFormats {
	//	sigs = append(sigs, f.Signatures...)
	//}
	return sigs
}

func (p pronom) GetInfos() map[string]FormatInfo {
	infos := make(map[string]FormatInfo)
	for _, f := range p.droid.FileFormats {
		infos[f.Puid] = FormatInfo{f.Name, f.Version, f.MIMEType}
	}
	return infos
}

// returns a slice of puid strings that corresponds to indexes of byte signatures
func (p pronom) GetPuids() ([]string, map[string][]int) {
	var iter int
	puids := make([]string, len(p.signatures()))
	bids := make(map[string][]int)
	for _, f := range p.droid.FileFormats {
		rng := iter + len(f.Signatures)
		for iter < rng {
			puids[iter] = f.Puid
			_, ok := bids[f.Puid]
			if ok {
				bids[f.Puid] = append(bids[f.Puid], iter)
			} else {
				bids[f.Puid] = []int{iter}
			}
			iter++
		}
	}
	return puids, bids
}

func (p pronom) extMatcher(m core.Matcher) error {
	p.EPuids = make([]string, len(p.droid.FileFormats))
	es := make(extensionmatcher.SignatureSet, len(p.droid.FileFormats))
	for i, f := range p.droid.FileFormats {
		p.EPuids[i] = f.Puid
		es[i] = f.Extensions
	}
	l, err := m.Add(es, nil)
	if err != nil {
		return err
	}
	p.EStart = l - len(p.EPuids)
	return nil
}

func (p pronom) contMatcher(m core.Matcher) error {
	var zpuids, mpuids []string
	var zsigs, msigs [][]frames.Signature
	var znames, mnames [][]string
	cpuids := make(map[int]string)
	for _, fm := range p.container.FormatMappings {
		cpuids[fm.Id] = fm.Puid
	}
	for _, c := range p.container.ContainerSignatures {
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
	_, err := m.Add(containermatcher.SignatureSet{containermatcher.Zip, znames, zsigs}, p.ps.List(zpuids))
	if err != nil {
		return err
	}
	l, err := m.Add(containermatcher.SignatureSet{containermatcher.Mscfb, mnames, msigs}, p.ps.List(mpuids))
	if err != nil {
		return err
	}
	p.CPuids = append(zpuids, mpuids...)
	p.CStart = l - len(p.CPuids)
	return nil
}

// SaveReports fetches pronom reports listed in the given droid file.
func FetchReports() []error {
	p := new(pronom)
	if err := p.setDroid(); err != nil {
		return []error{err}
	}
	apply := func(p *pronom, puid string) error {
		url, _, _ := config.HarvestOptions()
		return save(puid, url, config.Reports())
	}
	return p.applyAll(apply)
}

// SaveReport fetches and saves a given puid from the base URL and writes to disk at the given path.
func FetchReport(puid, url, path string) error {
	return save(puid, url, path)
}

// setDroid adds a Droid file to a pronom object and sets the list of puids.
func (p *pronom) setDroid() error {
	p.droid = new(Droid)
	if err := openXML(config.Droid(), p.droid); err != nil {
		return err
	}
	p.puids = make(map[string]int)
	p.ids = make(map[int]string)
	for i, v := range p.droid.FileFormats {
		p.puids[v.Puid] = i
		p.ids[v.Id] = v.Puid
	}
	return nil
}

// setContainers adds containers to a pronom object. It takes as an argument the path to a container signature file
func (p *pronom) setContainers() error {
	p.container = new(Container)
	return openXML(config.Container(), p.container)
}

func (p *pronom) setReports() []error {
	p.reports = make([]*Report, len(p.puids))
	apply := func(p *pronom, puid string) error {
		idx := p.puids[puid]
		buf, err := get(puid)
		if err != nil {
			return err
		}
		p.reports[idx] = &Report{}
		return xml.Unmarshal(buf, p.reports[idx])
	}
	return p.applyAll(apply)
}

func openXML(path string, els interface{}) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return xml.Unmarshal(buf, els)
}

func (p *pronom) applyAll(apply func(p *pronom, puid string) error) []error {
	ch := make(chan error, len(p.puids))
	wg := sync.WaitGroup{}
	queue := make(chan struct{}, 200)
	for puid := range p.puids {
		wg.Add(1)
		go func(puid string) {
			queue <- struct{}{}
			defer wg.Done()
			if err := apply(p, puid); err != nil {
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

func get(puid string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(config.Reports(), strings.Replace(puid, "/", "", 1)+".xml"))
}

func save(puid, url, path string) error {
	b, err := getHttp(url + puid + ".xml")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(path, strings.Replace(puid, "/", "", 1)+".xml"), b, os.ModePerm)
}
