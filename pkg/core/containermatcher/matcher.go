package containermatcher

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type Matcher []*ContainerMatcher

func (m Matcher) String() string {
	var str string
	for _, c := range m {
		str += c.String()
	}
	return str
}

func (m Matcher) Priority() bool {
	for _, c := range m {
		if c.Priorities != nil {
			return true
		}
	}
	return false
}

func Load(r io.Reader) (core.Matcher, error) {
	var m Matcher
	dec := gob.NewDecoder(r)
	err := dec.Decode(&m)
	if err != nil {
		return nil, err
	}
	for _, c := range m {
		c.ctype = ctypes[c.CType]
	}
	return m, nil
}

func (m Matcher) Save(w io.Writer) (int, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(m)
	if err != nil {
		return 0, err
	}
	sz := buf.Len()
	_, err = buf.WriteTo(w)
	if err != nil {
		return 0, err
	}
	return sz, nil
}

type ContainerMatcher struct {
	ctype
	CType      int
	NameCTest  map[string]*CTest
	Parts      []int // corresponds with each signature: represents the number of CTests for each sig
	Priorities priority.List
	Default    string // the default is an extension which when matched signals that the container matcher should quit
	// this prevents delving through zip files that will be recursed anyway
	// temp stuff used during identification
	started      bool
	entryBuf     *siegreader.Buffer // shared buffer used by each entry in a container
	partsMatched [][]hit            // hits for parts
	ruledOut     []bool             // mark additional signatures as negatively matched
	waitList     []int
	hits         []hit // shared buffer of hits used when matching

}

func (c *ContainerMatcher) String() string {
	str := "\nContainer matcher:\n"
	str += fmt.Sprintf("Type: %d\n", c.CType)
	str += "Default: "
	if c.Default == "" {
		str += "none\n"
	} else {
		str += c.Default + "\n"
	}
	str += fmt.Sprintf("Parts: %v\n", c.Parts)
	for k, v := range c.NameCTest {
		str += "-----------\n"
		str += fmt.Sprintf("Name: %v\n", k)
		str += fmt.Sprintf("Satisfied: %v\n", v.Satisfied)
		str += fmt.Sprintf("Unsatisfied: %v\n", v.Unsatisfied)
		if v.BM == nil {
			str += "Bytematcher: None\n"
		} else {
			str += "Bytematcher:\n" + v.BM.String()
		}
	}
	return str
}

type ctype struct {
	trigger func([]byte) bool
	rdr     func(*siegreader.Buffer) (Reader, error)
}

var ctypes = []ctype{
	ctype{
		zipTrigger,
		newZip,
	},
	ctype{
		mscfbTrigger,
		newMscfb,
	},
}

func zipTrigger(b []byte) bool {
	return binary.LittleEndian.Uint32(b[:4]) == 0x04034B50
}

func NewZip() *ContainerMatcher {
	return &ContainerMatcher{
		ctype:     ctypes[0],
		CType:     0,
		NameCTest: make(map[string]*CTest),
	}
}

func mscfbTrigger(b []byte) bool {
	return binary.LittleEndian.Uint64(b) == 0xE11AB1A1E011CFD0
}

func NewMscfb() *ContainerMatcher {
	return &ContainerMatcher{
		ctype:     ctypes[1],
		CType:     1,
		NameCTest: make(map[string]*CTest),
	}
}

func (c *ContainerMatcher) AddSignature(nameParts []string, sigParts []frames.Signature) error {
	if len(nameParts) != len(sigParts) {
		return errors.New("Container matcher: nameParts and sigParts must be equal")
	}
	c.Parts = append(c.Parts, len(nameParts))
	for i, nm := range nameParts {
		ct, ok := c.NameCTest[nm]
		if !ok {
			ct = &CTest{}
			c.NameCTest[nm] = ct
		}
		ct.add(sigParts[i], len(c.Parts)-1)
	}
	return nil
}

func (c *ContainerMatcher) Commit(d string) error {
	if len(d) > 0 {
		// add a c.Part for the default value. It has no tests associated, but we give it one
		// so as not to confuse priority matching for other sigs
		c.Parts = append(c.Parts, 1)
		c.Default = d
	}
	for _, v := range c.NameCTest {
		err := v.commit(c.Priorities)
		if err != nil {
			return err
		}
	}
	return nil
}

// a container test is a the basic element of container matching
type CTest struct {
	Satisfied   []int              // satisfied signatures are immediately matched: i.e. a name without a required bitstream
	Unsatisfied []int              // unsatisfied signatures depend on bitstreams as well as names matching
	buffer      []frames.Signature // temporary - used while creating CTests
	BM          *bytematcher.ByteMatcher
}

func (ct *CTest) add(s frames.Signature, t int) {
	if s == nil {
		ct.Satisfied = append(ct.Satisfied, t)
		return
	}
	ct.Unsatisfied = append(ct.Unsatisfied, t)
	ct.buffer = append(ct.buffer, s)
}

// call for each key after all signatures added
func (ct *CTest) commit(p priority.List) error {
	if ct.buffer == nil {
		return nil
	}
	bm, err := bytematcher.Signatures(ct.buffer)
	if err != nil {
		return err
	}
	ct.BM = bm
	ct.BM.Priorities = p.Subset(ct.Unsatisfied)
	return nil
}
