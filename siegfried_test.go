package siegfried

import (
	"bytes"
	"testing"

	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/internal/siegreader"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
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
	i := c[0]
	if i.String() != "fmt/3" {
		t.Error("expecting fmt/3")
	}
}

func TestLabel(t *testing.T) {
	s := &Siegfried{ids: []core.Identifier{testIdentifier{}}}
	res := s.Label(testIdentification{})
	if len(res) != 2 ||
		res[0][0] != "namespace" ||
		res[0][1] != "a" ||
		res[1][0] != "id" ||
		res[1][1] != "fmt/3" {
		t.Errorf("bad label, got %v", res)
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
func (t testEMatcher) String() string { return "" }

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
func (t testBMatcher) String() string { return "" }

type testResult int

func (tr testResult) Index() int    { return int(tr) }
func (tr testResult) Basis() string { return "" }

// identifier test stub

type testIdentifier struct{}

func (t testIdentifier) Add(m core.Matcher, mt core.MatcherType) (core.Matcher, error) {
	return nil, nil
}
func (t testIdentifier) Recorder() core.Recorder                            { return testRecorder{} }
func (t testIdentifier) Name() string                                       { return "a" }
func (t testIdentifier) Details() string                                    { return "b" }
func (t testIdentifier) Fields() []string                                   { return []string{"namespace", "id"} }
func (t testIdentifier) Save(l *persist.LoadSaver)                          {}
func (t testIdentifier) String() string                                     { return "" }
func (t testIdentifier) Inspect(s ...string) (string, error)                { return "", nil }
func (t testIdentifier) GraphP(i int) string                                { return "" }
func (t testIdentifier) Recognise(m core.MatcherType, i int) (bool, string) { return false, "" }

// recorder test stub

type testRecorder struct{}

func (t testRecorder) Active(m core.MatcherType)                     {}
func (t testRecorder) Record(m core.MatcherType, r core.Result) bool { return true }
func (t testRecorder) Satisfied(m core.MatcherType) (bool, int)      { return false, 0 }
func (t testRecorder) Report() []core.Identification {
	return []core.Identification{testIdentification{}}
}

// identification test stub

type testIdentification struct{}

func (t testIdentification) String() string          { return "fmt/3" }
func (t testIdentification) Warn() string            { return "" }
func (t testIdentification) Known() bool             { return true }
func (t testIdentification) Values() []string        { return []string{"a", "fmt/3"} }
func (t testIdentification) Archive() config.Archive { return 0 }
