package containermatcher

import (
	"errors"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type Matcher interface {
	Identify(*siegreader.Buffer) (core.Result, bool)
	String() string
	Save(io.Writer) (int, error)
}

type CTest struct {
	Satisfied   []int
	Unsatisfied []int
	buffer      []frames.Signature
	BM          *bytematcher.Bytematcher
}

func newCTest() *CTest {
	return &CTest{
		make([]int, 0),
		make([]int, 0),
		make([]frames.Signature, 0),
		nil,
	}
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
func (ct *CTest) commit() error {
	bm, err := bytematcher.Signatures(buffer)
	if err != nil {
		return err
	}
	ct.BM = bm
	return nil
}

type ContainerMatcher struct {
	NameCTest  map[string]*CTest
	Parts      []int
	Names      [][]string // what names must match? used for resolving priorities
	Priorities [][]int
	tally      int
	// temp stuff used during identification
	entryBuf     *siegreader.Buffer // shared buffer used by each entry in a container
	nameMatches  []string           // names that have matched, used for resolving priorities
	partsMatched []int              // parts that have matched
}

func New() *ContainerMatcher {
	return &ContainerMatcher{
		make(map[string]*CTest),
		make([]int, 0),
		0,
	}
}

func (c *ContainerMatcher) AddSignature(nameParts []string, sigParts []frames.Signature) error {
	if len(nameParts) != len(sigParts) {
		return errors.New("Container matcher: nameParts and sigParts should be equal")
	}
	c.Parts = append(c.Parts, len(nameParts))
	for i, nm := range nameParts {
		ct, ok := c.NameCTest[nm]
		if !ok {
			ct = newCTest()
			c.NameCTest[nm] = ct
		}
		ct.add(sigParts[i], c.tally)
	}
	c.tally++
	return nil
}

func (c *ContainerMatcher) Commit() error {
	for _, v := range c.NameCTest {
		err := v.commit()
		if err != nil {
			return err
		}
	}
	return nil
}
