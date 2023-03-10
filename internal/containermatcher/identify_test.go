package containermatcher

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/bytematcher/frames/tests"
	"github.com/richardlehane/siegfried/internal/siegreader"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

func TestIdentify(t *testing.T) {
	// testTrigger defined in container_test and just returns true
	// newTestReader defined in reader_test and returns a test reader
	// This test works because the outside file (bytes 0..8) is just a
	// meaningless wrapper. The matcher always detects container (due to testTrigger)
	// and newTestReader has the content we actually test against.
	config.SetOut(os.Stderr)
	config.SetDebug()
	ctypes = []ctype{{testTrigger, newTestReader}}
	// test adding
	count++
	testMatcher, _, err := Add(Matcher{testContainerMatcher},
		SignatureSet{
			0,
			[][]string{
				{"one", "two"},
				{"one"},
			},
			[][]frames.Signature{
				{tests.TestSignatures[3], tests.TestSignatures[4]}, // {[BOF 0:test], [P 10-20:TESTY|YNESS], [BOF *:test]}, {[BOF *:junk]}
				{tests.TestSignatures[2]},                          // {[BOF 0-5:a|b|c..j], [P *:test]}
			},
		},
		nil,
	)
	/*
		// The file content is always: "test12345678910YNESSjunktestyjunktestytest12345678910111223"
		   The signatures are processed as these sequences:
		   "one": [ {Offsets: 0; Choices: [test]} {Offsets: -1; Choices: [test]}
		     {Offsets: 5, -1; Choices: [a | b | c | d | e | f | g | h | i | j], [test]}
		   ]
		   "two" [{Offsets: -1; Choices: [junk]}]
		   // Issue (1) we have a signature of *junk - this signature is never sent to resume as no hits indicated it. Need to add a list of non-anchored wildcard signatures.
		   // Issue (2) the signature at 2[1] isn't being sent on the resume channel. Or it is being sent but ends up with wrong test index.
	*/
	if err != nil {
		t.Fatal(err)
	}
	r := bytes.NewBuffer([]byte("012345678"))
	bufs := siegreader.New()
	b, err := bufs.Get(r)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	res, _ := testMatcher.Identify("example.tt", b)
	var collect []core.Result
	for r := range res {
		collect = append(collect, r)
	}
	expect := count * 2
	if len(collect) != expect {
		t.Errorf("Expecting %d results, got %d", expect, len(collect))
		for _, r := range collect {
			t.Error(r.Basis())
		}
	}
}
