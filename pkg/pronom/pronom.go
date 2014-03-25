package pronom

import (
	"bytes"
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/core/namematcher"

	. "github.com/richardlehane/siegfried/pkg/pronom/mappings"
)

var Config = struct {
	Droid     string
	Container string
	Reports   string
	Data      string

	Timeout   time.Duration
	Transport http.Transport
}{
	"DROID_SignatureFile_V73.xml",
	"container-signature-20140227.xml",
	"pronom",
	filepath.Join("..", "..", "cmd", "r2d2", "data"),

	120 * time.Second,
	http.Transport{Proxy: http.ProxyFromEnvironment},
}

func ConfigPaths() (string, string, string) {
	return filepath.Join(Config.Data, Config.Droid),
		filepath.Join(Config.Data, Config.Container),
		filepath.Join(Config.Data, Config.Reports)
}

func NewIdentifier(droid, container, reports string) (*PronomIdentifier, error) {
	pronom, err := newPronom(droid, container, reports)
	if err != nil {
		return nil, err
	}
	return pronom.identifier()
}

func Load(path string) (*PronomIdentifier, error) {
	c, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(c)
	dec := gob.NewDecoder(buf)
	var p PronomIdentifier
	err = dec.Decode(&p)
	if err != nil {
		return nil, err
	}
	p.Bm.Start()
	p.ids = make(pids, 20)
	return &p, nil
}

func (p *PronomIdentifier) Save(path string) error {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(p)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, buf.Bytes(), os.ModeExclusive)
}

type pronom struct {
	droid     *Droid
	container *Container
	puids     map[string]int
	ids       map[int]string
}

func (p pronom) String() string {
	return p.droid.String()
}

func (p *pronom) identifier() (*PronomIdentifier, error) {
	pi := new(PronomIdentifier)
	pi.ids = make(pids, 20)
	pi.BPuids = p.getPuids()
	pi.Priorities = p.priorities()
	pi.Em, pi.EPuids = p.extensionMatcher()
	sigs, err := p.parse()
	if err != nil {
		return nil, err
	}
	pi.Bm, err = bytematcher.Signatures(sigs)
	return pi, err
}

func (p pronom) signatures() []Signature {
	sigs := make([]Signature, 0, 1000)
	for _, f := range p.droid.FileFormats {
		sigs = append(sigs, f.Signatures...)
	}
	return sigs
}

func (p pronom) getPuids() []string {
	var iter int
	puids := make([]string, len(p.signatures()))
	for _, f := range p.droid.FileFormats {
		rng := iter + len(f.Signatures)
		for iter < rng {
			puids[iter] = f.Puid
			iter++
		}
	}
	return puids
}

func (p pronom) extensionMatcher() (namematcher.ExtensionMatcher, []string) {
	em := namematcher.NewExtensionMatcher()
	epuids := make([]string, len(p.droid.FileFormats))
	for i, f := range p.droid.FileFormats {
		epuids[i] = f.Puid
		for _, v := range f.Extensions {
			em.Add(v, i)
		}
	}
	return em, epuids
}

func (p pronom) priorities() map[string][]int {
	var iter int
	priorities := make(map[string][]int)
	for _, f := range p.droid.FileFormats {
		for _ = range f.Signatures {
			for _, v := range f.Priorities {
				puid := p.ids[v]
				_, ok := priorities[puid]
				if ok {
					priorities[puid] = append(priorities[puid], iter)
				} else {
					priorities[puid] = []int{iter}
				}
			}
			iter++
		}
	}
	for k := range priorities {
		sort.Ints(priorities[k])
	}
	return priorities
}

// newPronom creates a pronom object. It takes as arguments the paths to a Droid signature file, a container file, and a base directory or base url for Pronom reports.
func newPronom(droid, container, reports string) (*pronom, error) {
	p := new(pronom)
	if err := p.setDroid(droid); err != nil {
		return p, err
	}
	if err := p.setContainers(container); err != nil {
		return p, err
	}
	errs := p.setReports(reports)
	if len(errs) > 0 {
		var str string
		for _, e := range errs {
			str += fmt.Sprintln(e)
		}
		return p, fmt.Errorf(str)
	}
	return p, nil
}

// SaveReports fetches pronom reports listed in the given droid file. It fetches over http (from the given base url) and writes them to disk (at the path argument).
func SaveReports(droid, url, path string) []error {
	p := new(pronom)
	if err := p.setDroid(droid); err != nil {
		return []error{err}
	}
	apply := func(p *pronom, puid string) error {
		return save(puid, url, path)
	}
	return p.applyAll(apply)
}

// SaveReport fetches and saves a given puid from the base URL and writes to disk at the given path.
func SaveReport(puid, url, path string) error {
	return save(puid, url, path)
}

// setDroid adds a Droid file to a pronom object and sets the list of puids.
func (p *pronom) setDroid(path string) error {
	p.droid = new(Droid)
	if err := openXML(path, p.droid); err != nil {
		return err
	}
	p.puids = make(map[string]int)
	p.ids = make(map[int]string)
	for i, v := range p.droid.FileFormats {
		p.puids[v.Puid] = i
		p.ids[v.ID] = v.Puid
	}
	return nil
}

// setContainers adds containers to a pronom object. It takes as an argument the path to a container signature file
func (p *pronom) setContainers(path string) error {
	p.container = new(Container)
	return openXML(path, p.container)
}

// setReports adds pronom reports to a pronom object.
// These reports are either fetched over http or from a local directory, depending on whether the path given is prefixed with 'http'.
func (p *pronom) setReports(path string) []error {
	var local bool
	if !strings.HasPrefix(path, "http") {
		local = true
	}
	apply := func(p *pronom, puid string) error {
		idx := p.puids[puid]
		buf, err := get(path, puid, local)
		if err != nil {
			return err
		}
		p.droid.FileFormats[idx].Report = new(Report)
		return xml.Unmarshal(buf, p.droid.FileFormats[idx].Report)
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
	for puid := range p.puids {
		wg.Add(1)
		go func(puid string) {
			defer wg.Done()
			if err := apply(p, puid); err != nil {
				ch <- err
			}
		}(puid)
	}
	wg.Wait()
	close(ch)
	errors := make([]error, 0)
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
	timer := time.AfterFunc(Config.Timeout, func() {
		Config.Transport.CancelRequest(req)
	})
	defer timer.Stop()
	client := http.Client{
		Transport: &Config.Transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func get(path string, puid string, local bool) ([]byte, error) {
	if local {
		return ioutil.ReadFile(filepath.Join(path, strings.Replace(puid, "/", ".", 1)+".xml"))
	}
	return getHttp(path + puid + ".xml")
}

func save(puid, url, path string) error {
	b, err := getHttp(url + puid + ".xml")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(path, strings.Replace(puid, "/", ".", 1)+".xml"), b, 0644)
}
