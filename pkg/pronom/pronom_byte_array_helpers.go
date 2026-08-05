package pronom

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/pronom/internal/mappings"
)

// NewFromByteArray is a helpder method designed to simplify creation
// of a new identifier with in memory data.
func NewFromByteArray(c []byte, d []byte, e []byte, opts ...config.Option) (core.Identifier, error) {
	for _, v := range opts {
		v()
	}
	pronom, err := rawFromByteArray(c, d, e)
	if err != nil {
		return nil, err
	}
	var pmap priority.Map
	if config.GetMulti() == config.DROID {
		pmap = pronom.Priorities()
	}
	pronom = identifier.ApplyConfig(pronom)
	id := &Identifier{
		Base:     identifier.New(pronom, config.ZipPuid()),
		hasClass: config.Reports() != "" && !config.NoClass(),
		infos:    infos(pronom.Infos()),
	}
	if id.Multi() == config.DROID {
		id.priorities = pmap
	}
	return id, nil
}

// rawFromByteArray allows us to create a new PRONOM parseavle
// (identifier).
func rawFromByteArray(c []byte, d []byte, e []byte) (identifier.Parseable, error) {
	p := &pronom{
		c: identifier.Blank{},
	}
	if err := p.setContainersFromByteArray(c); err != nil {
		return nil, fmt.Errorf(
			"pronom: error loading containers; got %s\nUnless you have set `-nocontainer` you need to download a container signature file",
			err.Error(),
		)
	}
	if err := p.setParseablesFromByteArray(d, e); err != nil {
		return nil, err
	}
	return p, nil
}

// setContainers adds containers to a pronom object. It takes a byte
// array as an argument.
func (p *pronom) setContainersFromByteArray(cf []byte) error {
	c := &mappings.Container{}
	path := bytes.NewBuffer(cf)
	buf, err := io.ReadAll(path)
	if err != nil {
		return err
	}
	xml.Unmarshal(buf, c)
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

// set identifiers joins signatures in an in memory DROID signature
// file with any extra reports and adds that to the pronom object
func (p *pronom) setParseablesFromByteArray(dr []byte, ex []byte) error {
	d, err := newDroidFromByteArray(dr)
	if err != nil {
		return fmt.Errorf(
			"pronom: error loading Droid file; got %s\nYou must have a Droid file to build a signature",
			config.Home(),
		)
	}
	p.Parseable = d
	if ex != nil {
		e, err := newDroidFromByteArray(ex)
		if err != nil {
			return fmt.Errorf("pronom: error loading extension file; got %s", err.Error())
		}
		p.Parseable = identifier.Join(p.Parseable, e)
	}
	// exclude byte signatures where also have container signatures,
	// unless doubleup set
	if !config.DoubleUp() {
		p.Parseable = doublesFilter{
			config.ExcludeDoubles(p.IDs(), p.c.IDs()),
			p.Parseable,
		}
	}
	return nil
}

func newDroidFromByteArray(dr []byte) (*droid, error) {
	d := &mappings.Droid{}
	path := bytes.NewBuffer(dr)
	buf, err := io.ReadAll(path)
	if err != nil {
		return nil, err
	}
	xml.Unmarshal(buf, d)
	return &droid{d, identifier.Blank{}}, nil
}
