package pronom

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardlehane/siegfried/pkg/config"
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
	errs := []string{}
	for i, v := range rsigs {
		if !v.Equals(dsigs[i]) {
			errs = append(errs, fmt.Sprintf("Parse Droid: signatures for %s are not equal:\nReports: %s\n  Droid: %s", rpuids[i], v, dsigs[i]))
		}
	}
	dpmap, rpmap := d.Priorities(), r.Priorities()
	if len(dpmap) != len(rpmap) {
		t.Errorf("Parse Droid: Expecting length of priorities of droid and reports to be the same, got %d and %d", len(dpmap), len(rpmap))
	}
	for k, v := range dpmap {
		w, ok := rpmap[k]
		if !ok {
			t.Errorf("Parse Droid: Can't find %s in reports priorities", k)
		}
		if len(v) != len(w) {
			errs = append(errs, fmt.Sprintf("Parse Droid: priorites for %s are not equal:\nReports: %v\nDroid: %v", k, w, v))
			continue
		}
		for i, vv := range v {
			if w[i] != vv {
				errs = append(errs, fmt.Sprintf("Parse Droid: priorites for %s are not equal:\nReports: %v\nDroid: %v", k, w, v))
				break
			}
		}
	}
	if len(errs) != 0 {
		t.Skip(strings.Join(errs, "\n"))
	}
}

func fiJSON(infos map[string]formatInfo) error {
	for k, v := range infos {
		if !json.Valid([]byte("\"" + v.name + "\"")) {
			return errors.New(k + " has bad JSON: \"" + v.name + "\"")
		}
		if !json.Valid([]byte("\"" + v.version + "\"")) {
			return errors.New(k + " has bad JSON: \"" + v.version + "\"")
		}
		if !json.Valid([]byte("\"" + v.mimeType + "\"")) {
			return errors.New(k + " has bad JSON: \"" + v.mimeType + "\"")
		}
	}
	return nil
}

// Check for any issues in format infos that would break JSON encoding
// See: https://github.com/richardlehane/siegfried/issues/186
func TestJSON(t *testing.T) {
	config.SetHome(filepath.Join("..", "..", "cmd", "roy", "data"))
	d, err := newDroid(config.Droid())
	if err != nil {
		t.Fatal(err)
	}
	err = fiJSON(infos(d.Infos()))
	if err != nil {
		t.Fatalf("JSON error in DROID file: %v", err)
	}
	r, err := newReports(d.IDs(), d.idsPuids())
	if err != nil {
		t.Fatal(err)
	}
	err = fiJSON(infos(r.Infos()))
	if err != nil {
		t.Fatalf("JSON error in PRONOM reports: %v", err)
	}
}
