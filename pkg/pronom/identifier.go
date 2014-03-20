package pronom

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type PronomIdentifier struct {
	Bm    *bytematcher.Bytematcher
	Puids []string
}

type PronomIdentification struct {
	puid       string
	confidence float64
}

func (pid PronomIdentification) String() string {
	return pid.puid
}

func (pid PronomIdentification) Confidence() float64 {
	return pid.confidence
}

func (pi *PronomIdentifier) Identify(b *siegreader.Buffer, c chan core.Identification) {
	ids, _ := pi.Bm.Identify(b)
	for i := range ids {
		c <- PronomIdentification{pi.Puids[i], 0.9}
	}
}

func (p *pronom) Identifier() (*PronomIdentifier, error) {
	pi := new(PronomIdentifier)
	pi.Puids = p.Puids()
	sigs, err := p.Parse()
	if err != nil {
		return nil, err
	}
	pi.Bm, err = bytematcher.Signatures(sigs)
	return pi, err
}

func NewIdentifier(droid, container, reports string) (*PronomIdentifier, error) {
	pronom, err := New(droid, container, reports)
	if err != nil {
		return nil, err
	}
	return pronom.Identifier()
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
