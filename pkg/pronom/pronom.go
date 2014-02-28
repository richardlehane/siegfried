package pronom

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
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

type pronom struct {
	droid     *Droid
	container *Container
	puids     map[string]int
}

func (p pronom) Signatures() []Signature {
	sigs := make([]Signature, 0, 1000)
	for _, f := range p.droid.FileFormats {
		sigs = append(sigs, f.Signatures...)
	}
	return sigs
}

func (p pronom) Puids() []string {
	var iter int
	puids := make([]string, len(p.Signatures()))
	for _, f := range p.droid.FileFormats {
		rng := iter + len(f.Signatures)
		for iter < rng {
			puids[iter] = f.Puid
			iter++
		}
	}
	return puids
}

func PuidsFromDroid(droid, reports string) ([]string, error) {
	p := new(pronom)
	if err := p.setDroid(droid); err != nil {
		return nil, err
	}
	errs := p.setReports(reports)
	if len(errs) > 0 {
		var str string
		for _, e := range errs {
			str += fmt.Sprintln(e)
		}
		return nil, fmt.Errorf(str)
	}
	return p.Puids(), nil
}

func (p pronom) String() string {
	return p.droid.String()
}

// New creates a pronom object. It takes as arguments the paths to a Droid signature file, a container file, and a base directory or base url for Pronom reports.
func New(droid, container, reports string) (*pronom, error) {
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
	for i, v := range p.droid.FileFormats {
		p.puids[v.Puid] = i
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
