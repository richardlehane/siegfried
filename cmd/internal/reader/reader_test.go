package reader

import (
	"testing"
)

func TestSF(t *testing.T) {
	sfc, err := Open("examples/multi/multi.csv")
	if err != nil {
		t.Fatal(err)
	}
	sfy, err := Open("examples/multi/multi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	sfj, err := Open("examples/multi/multi.json")
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

func TestFido(t *testing.T) {
	fi, err := Open("examples/ipresShowcase/fido.csv")
	if err != nil {
		t.Fatal(err)
	}
	var i, j int
	for f, e := fi.Next(); e == nil; f, e = fi.Next() {
		i++
		j += len(f.IDs)
	}
	if i != 2190 || j != 2984 {
		t.Errorf("Expecting 2190 files and 2984 IDs, got %d files and %d IDs", i, j)
	}
}
