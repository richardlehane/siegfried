package siegfried

import (
	"bytes"
	"io"
	"testing"

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

type testNMatcher struct{}

func (t testNMatcher) Identify(n string, sb *siegreader.Buffer) chan core.Result {
	ret := make(chan core.Result)
	go func() {
		ret <- testResult(0)
		close(ret)
	}()
	return ret
}

func (t testNMatcher) String() string { return "" }

func (t testNMatcher) Save(w io.Writer) (int, error) { return 0, nil }

func (t testNMatcher) Add(ss core.SignatureSet, l priority.List) (int, error) { return 0, nil }

type testBMatcher struct{}

func (t testBMatcher) Identify(nm string, sb *siegreader.Buffer) chan core.Result {
	ret := make(chan core.Result)
	go func() {
		ret <- testResult(1)
		ret <- testResult(2)
		close(ret)
	}()
	return ret
}

func (t testBMatcher) String() string { return "" }

func (t testBMatcher) Save(w io.Writer) (int, error) { return 0, nil }

func (t testBMatcher) Add(ss core.SignatureSet, l priority.List) (int, error) { return 0, nil }

type testResult int

func (tr testResult) Index() int { return int(tr) }

func (tr testResult) Basis() string { return "" }

type testIdentifier struct{}

func (t testIdentifier) Yaml() string { return "" }

func (t testIdentifier) Save(w io.Writer) (int, error) { return 0, nil }

func (t testIdentifier) Recorder() core.Recorder { return testRecorder{} }

type testRecorder struct{}

func (t testRecorder) Record(m core.MatcherType, r core.Result) bool { return true }

func (t testRecorder) Satisfied() bool { return false }

func (t testRecorder) Report(c chan core.Identification) { c <- testIdentification{} }

type testIdentification struct{}

func (t testIdentification) String() string { return "fmt/3" }

func (t testIdentification) Yaml() string { return "" }

func (t testIdentification) Json() string { return "" }

func TestIdentify(t *testing.T) {
	s := New()
	s.em = testNMatcher{}
	s.bm = testBMatcher{}
	s.cm = nil
	s.ids = append(s.ids, testIdentifier{})
	c, err := s.Identify("test.doc", bytes.NewBufferString("test"))
	if err != nil {
		t.Error(err)
	}
	i := <-c
	if i.String() != "fmt/3" {
		t.Error("expecting fmt/3")
	}
}

func TestLoad(t *testing.T) {
	s := New()
	p, err := pronom.New(config.SetHome("./cmd/r2d2/data"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(p)
	if err != nil {
		t.Fatal(err)
	}
}
