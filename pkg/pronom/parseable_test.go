package pronom

import (
	"path/filepath"
	"testing"

	"github.com/richardlehane/siegfried/config"
)

// DROID parsing is tested by comparing it against Report parsing
func TestParseDroid(t *testing.T) {
	config.SetHome(filepath.Join("..", "..", "cmd", "roy", "data"))
	d, err := newDroid(config.Droid())
	if err != nil {
		t.Fatal(err)
	}
	r, err := newReports(d.IDs(), d.idsPuids())
	if err != nil {
		t.Fatal(err)
	}
	dsigs, dpuids, err := d.Signatures()
	if err != nil {
		t.Fatal(err)
	}
	rsigs, rpuids, err := r.Signatures()
	if err != nil {
		t.Fatal(err)
	}
	if len(dpuids) != len(rpuids) {
		t.Errorf("Parse Droid: Expecting length of reports and droid to be same, got %d, %d, %s", len(rpuids), len(dpuids), dpuids[len(dpuids)-8])
	}
	for i, v := range rpuids {
		if v != dpuids[i] {
			t.Errorf("Parse Droid: Expecting slices of puids to be identical but at index %d, got %s for reports and %s for droid", i, v, dpuids[i])
		}
	}
	if len(dsigs) != len(rsigs) {
		t.Errorf("Parse Droid: Expecting sig length of reports and droid to be same, got %d, %d", len(rsigs), len(dsigs))
	}
	for i, v := range rsigs {
		if !v.Equals(dsigs[i]) {
			t.Errorf("Parse Droid: signatures for %s are not equal:\nReports: %s\n  Droid: %s", rpuids[i], v, dsigs[i])
		}
	}
}
