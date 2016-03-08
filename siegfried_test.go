package siegfried

import (
	"bytes"
	"testing"

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/persist"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

func TestLoad(t *testing.T) {
	s := New()
	config.SetHome("./cmd/roy/data")
	p, err := pronom.New()
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(p)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIdentify(t *testing.T) {
	s := New()
	s.nm = testEMatcher{}
	s.bm = testBMatcher{}
	s.cm = nil
	s.ids = append(s.ids, testIdentifier{})
	c, err := s.Identify(bytes.NewBufferString("test"), "test.doc", "")
	if err != nil {
		t.Error(err)
	}
	i := <-c
	if i.String() != "fmt/3" {
		t.Error("expecting fmt/3")
	}
}

// extension matcher test stub

type testEMatcher struct{}

func (t testEMatcher) Identify(n string, sb *siegreader.Buffer, exclude ...int) (chan core.Result, error) {
	ret := make(chan core.Result)
	go func() {
		ret <- testResult(0)
		close(ret)
	}()
	return ret, nil
}

func (t testEMatcher) String() string                                         { return "" }
func (t testEMatcher) Save(l *persist.LoadSaver)                              {}
func (t testEMatcher) Add(ss core.SignatureSet, l priority.List) (int, error) { return 0, nil }

// byte matcher test stub

type testBMatcher struct{}

func (t testBMatcher) Identify(nm string, sb *siegreader.Buffer, exclude ...int) (chan core.Result, error) {
	ret := make(chan core.Result)
	go func() {
		ret <- testResult(1)
		ret <- testResult(2)
		close(ret)
	}()
	return ret, nil
}

func (t testBMatcher) String() string                                         { return "" }
func (t testBMatcher) Save(l *persist.LoadSaver)                              {}
func (t testBMatcher) Add(ss core.SignatureSet, l priority.List) (int, error) { return 0, nil }

type testResult int

func (tr testResult) Index() int    { return int(tr) }
func (tr testResult) Basis() string { return "" }

// identifier test stub

type testIdentifier struct{}

func (t testIdentifier) YAML() string                                       { return "" }
func (t testIdentifier) Name() string                                       { return "a" }
func (t testIdentifier) Details() string                                    { return "b" }
func (t testIdentifier) Fields() []string                                   { return nil }
func (t testIdentifier) Save(l *persist.LoadSaver)                          {}
func (t testIdentifier) Recorder() core.Recorder                            { return testRecorder{} }
func (t testIdentifier) Recognise(m core.MatcherType, i int) (bool, string) { return false, "" }
func (t testIdentifier) String() string                                     { return "" }

// recorder test stub

type testRecorder struct{}

func (t testRecorder) Active(m core.MatcherType)                     {}
func (t testRecorder) Record(m core.MatcherType, r core.Result) bool { return true }
func (t testRecorder) Satisfied(m core.MatcherType) (bool, int)      { return false, 0 }
func (t testRecorder) Report(c chan core.Identification)             { c <- testIdentification{} }

// identification test stub

type testIdentification struct{}

func (t testIdentification) String() string          { return "fmt/3" }
func (t testIdentification) Warn() string            { return "" }
func (t testIdentification) Known() bool             { return true }
func (t testIdentification) YAML() string            { return "" }
func (t testIdentification) JSON() string            { return "" }
func (t testIdentification) CSV() []string           { return nil }
func (t testIdentification) Archive() config.Archive { return 0 }
