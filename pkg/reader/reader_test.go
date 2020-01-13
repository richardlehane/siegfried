package reader

import (
	"bytes"
	"os"
	"testing"
)

const (
	ipresFiles      = 2190
	ipresFidoIDs    = 2984
	ipresDroidIDs   = 2451
	ipresDroidNpIDs = 2192
)

func TestSF(t *testing.T) {
	f1, err := os.Open("examples/multi/multi.csv")
	defer f1.Close()
	sfc, err := New(f1, "examples/multi/multi.csv")
	if err != nil {
		t.Fatal(err)
	}
	f2, err := os.Open("examples/multi/multi.yaml")
	defer f2.Close()
	sfy, err := New(f2, "examples/multi/multi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	f3, err := os.Open("examples/multi/multi.json")
	defer f3.Close()
	sfj, err := New(f3, "examples/multi/multi.json")
	if err != nil {
		t.Fatal(err)
	}
	for f, e := sfc.Next(); e == nil; f, e = sfc.Next() {
		y, e1 := sfy.Next()
		if e1 != nil {
			t.Errorf("got a YAML error for a valid CSV %s; %v", f.Path, e1)
		}
		j, e2 := sfj.Next()
		if e2 != nil {
			t.Errorf("got a JSON error for a valid CSV %s; %v", f.Path, e2)
		}
		if len(f.IDs) != len(y.IDs) || len(f.IDs) != len(j.IDs) {
			t.Errorf("JSON, YAML and CSV IDs don't match for %s; got %d, %d and %d", f.Path, len(j.IDs), len(y.IDs), len(f.IDs))
		}
	}
}

func testRdr(t *testing.T, path string, expectFiles, expectIDs int) {
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rdr, err := New(f, path)
	if err != nil {
		t.Fatal(err)
	}
	var i, j int
	var ff File
	var e error
	for ff, e = rdr.Next(); e == nil; ff, e = rdr.Next() {
		i++
		j += len(ff.IDs)
	}
	if i != expectFiles || j != expectIDs {
		t.Errorf("Expecting %d files and %d IDs, got %d files and %d IDs; error: %v", expectFiles, expectIDs, i, j, e)
	}
}

func TestFido(t *testing.T) {
	testRdr(t, "examples/ipresShowcase/fido.csv", ipresFiles, ipresFidoIDs)
}

func TestDroid(t *testing.T) {
	testRdr(t, "examples/ipresShowcase/droid-gui-m.csv", ipresFiles, ipresDroidIDs)
	testRdr(t, "examples/ipresShowcase/droid-gui-s.csv", ipresFiles, ipresDroidIDs)
	testRdr(t, "examples/ipresShowcase/droid-np.csv", ipresFiles, ipresDroidNpIDs)
}

func TestCompare(t *testing.T) {
	w := &bytes.Buffer{}
	if err := Compare(w, 0, "examples/ipresShowcase/droid-gui-m.csv", "examples/ipresShowcase/droid-gui-s.csv"); err != nil {
		t.Fatal(err)
	}
	if string(w.Bytes()) != "COMPLETE MATCH\n" {
		t.Fatalf("expecting a complete match; got %s", string(w.Bytes()))
	}
}
