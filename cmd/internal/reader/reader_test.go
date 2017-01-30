package reader

import (
	"testing"
)

func TestCSVvsYAML(t *testing.T) {
	sfc, err := Open("examples/multi/multi.csv")
	if err != nil {
		t.Fatal(err)
	}
	sfy, err := Open("examples/multi/multi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for f, e := sfc.Next(); e == nil; f, e = sfc.Next() {
		y, e1 := sfy.Next()
		if e1 != nil {
			t.Errorf("got a YAML error for a valid CSV %s; %v", f.Path, e1)
		}
		if len(f.IDs) != len(y.IDs) {
			t.Errorf("YAML and CSV IDs don't match for %s; got %d and %d", f.Path, len(f.IDs), len(y.IDs))
		}
	}
}

/*
func testOutput(tool string, output map[string]string) error {
	if len(output) != 26124 && len(output) != 13289 {
		return fmt.Errorf("%s: length of output does not meet expectation %d", tool, len(output))
	}
	p, ok := output[clean(`/home/richard/local/bench/govdocs_selected/PDF_449/498897.pdf`)]
	if !ok {
		return fmt.Errorf("%s: missing /home/richard/local/bench/govdocs_selected/PDF_449/README", tool)
	}
	if p != "fmt/17" {
		return fmt.Errorf("%s: expecting fmt/17, got %s", tool, p)
	}
	return nil
}

func TestSf(t *testing.T) {
	output, err := sf("examples/govdocs/sf.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := testOutput("sf", output); err != nil {
		t.Error(err)
	}
}

func TestDroid(t *testing.T) {
	output, err := _droid("examples/govdocs/droid.csv")
	if err != nil {
		t.Fatal(err)
	}
	if err := testOutput("droid", output); err != nil {
		t.Error(err)
	}
}

func TestFido(t *testing.T) {
	output, err := fido("examples/govdocs/fido.csv")
	if err != nil {
		t.Fatal(err)
	}
	if err := testOutput("fido", output); err != nil {
		t.Error(err)
	}
}
*/
