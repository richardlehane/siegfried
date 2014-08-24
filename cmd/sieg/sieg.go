package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var (
	thisVersion = [2]int{0, 4}
	sigfile     string
	update      = flag.Bool("update", false, "update or install a Siegfried signature file")
	version     = flag.Bool("version", false, "display version information")
	defaultSigs = "pronom.gob"
	updateUrl   = "http://www.itforarchivists.com/siegfried/update"
	latestUrl   = "http://www.itforarchivists.com/siegfried/latest"
	timeout     = 30 * time.Second
	transport   = http.Transport{Proxy: http.ProxyFromEnvironment}
)

func init() {
	current, err := user.Current()
	if err != nil {
		log.Fatalf("Sieg error: can't obtain a current user %v", err)
	}
	defaultSigs = filepath.Join(current.HomeDir, ".siegfried", defaultSigs)

	flag.StringVar(&sigfile, "sigs", defaultSigs, "path to Siegfried signature file")
}

type Update struct {
	SiegVersion   [2]int
	PronomVersion int
	GobSize       int
}

func getHttp(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "siegfried/siegbot (+https://github.com/richardlehane/siegfried)")
	timer := time.AfterFunc(timeout, func() {
		transport.CancelRequest(req)
	})
	defer timer.Stop()
	client := http.Client{
		Transport: &transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func updateSigs() (string, error) {
	response, err := getHttp(updateUrl)
	if err != nil {
		return "", err
	}
	var u Update
	if err := json.Unmarshal(response, &u); err != nil {
		return "", err
	}
	if thisVersion[0] < u.SiegVersion[0] || (u.SiegVersion[0] == thisVersion[0] && thisVersion[1] < u.SiegVersion[1]) {
		return "Your version of Siegfried is out of date; please install latest from http://www.itforarchivists.com/siegfried before continuing.", nil
	}
	p, err := pronom.Load(sigfile)
	if err == nil {
		if !p.Update(u.PronomVersion) {
			return "You are already up to date!", nil
		}
	} else {
		err = os.MkdirAll(filepath.Dir(sigfile), os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	fmt.Println("... downloading latest signature file ...")
	response, err = getHttp(latestUrl)
	if err != nil {
		return "", err
	}
	if len(response) != u.GobSize {
		return "", fmt.Errorf("Error retrieving pronom.gob; expecting %d bytes, got %d bytes", u.GobSize, len(response))
	}
	err = ioutil.WriteFile(sigfile, response, os.ModePerm)
	if err != nil {
		return "", err
	}
	fmt.Printf("... writing %s ...\n", sigfile)
	return "Your signature file has been updated", nil
}

func load(sigs string) (*core.Siegfried, error) {
	s := core.NewSiegfried()
	p, err := pronom.Load(sigs)
	if err != nil {
		return nil, err
	}
	s.AddIdentifier(p)
	return s, nil
}

func identify(s *core.Siegfried, p string) ([]string, error) {
	ids := make([]string, 0)
	file, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("failed to open %v, got: %v", p, err)
	}
	c, err := s.Identify(file, p)
	if err != nil {
		return nil, fmt.Errorf("failed to identify %v, got: %v", p, err)
	}
	for i := range c {
		ids = append(ids, i.String())
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func multiIdentify(s *core.Siegfried, r string) ([][]string, error) {
	set := make([][]string, 0)
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		ids, err := identify(s, path)
		if err != nil {
			return err
		}
		set = append(set, ids)
		return nil
	}
	err := filepath.Walk(r, wf)
	return set, err
}

func multiIdentifyP(s *core.Siegfried, r string) error {
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open %v, got: %v", path, err)
		}
		c, err := s.Identify(file, path)
		if err != nil {
			return fmt.Errorf("failed to identify %v, got: %v", path, err)
		}
		fmt.Println(path)
		for i := range c {
			fmt.Println(i)
		}
		fmt.Println()
		file.Close()
		return nil
	}
	return filepath.Walk(r, wf)
}

func main() {

	flag.Parse()

	if *version {
		p, err := pronom.Load(sigfile)
		if err != nil {
			log.Fatalf("Sieg error: error loading signature file, got: %v\nIf you haven't installed a signature file yet, run sieg -update.", err)
		}
		fmt.Printf("Siegfried version: %d.%d; %s\n", thisVersion[0], thisVersion[1], p.Version())
		return
	}

	if *update {
		msg, err := updateSigs()
		if err != nil {
			log.Fatalf("Sieg error: error updating signature file, %v", err)
		}
		fmt.Println(msg)
		return
	}

	if flag.NArg() != 1 {
		log.Fatal("Sieg error: expecting a single file or directory argument")
	}

	var err error
	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatalf("Sieg error: error opening %v, got: %v", flag.Arg(0), err)
	}
	info, err := file.Stat()
	if err != nil {
		log.Fatalf("Sieg error: error getting info for %v, got: %v", flag.Arg(0), err)
	}

	s, err := load(sigfile)
	if err != nil {
		log.Fatalf("Sieg error: error loading signature file, got: %v", err)

	}

	if info.IsDir() {
		file.Close()
		err = multiIdentifyP(s, flag.Arg(0))
		if err != nil {
			log.Fatalf("Sieg error: %v", err)
		}
		os.Exit(0)
	}

	c, err := s.Identify(file, flag.Arg(0))
	if err != nil {
		log.Fatalf("Sieg error: %v", err)
	}
	fmt.Println(flag.Arg(0))
	for i := range c {
		fmt.Println(i)
	}
	file.Close()

	os.Exit(0)
}
