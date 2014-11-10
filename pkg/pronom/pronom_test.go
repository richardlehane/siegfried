package pronom

import (
	"path/filepath"
	"testing"

	"github.com/richardlehane/siegfried/config"
)

var p *pronom

func TestNew(t *testing.T) {
	config.SetHome(filepath.Join("..", "..", "cmd", "r2d2", "data"))()
	_, err := NewPronom()
	if err != nil {
		t.Error(err)
	}
}

// These work but take a while, so left out of routine testing
/*
func TestSaveReports(t *testing.T) {
	errs := SaveReports(config.Droid()), "http://www.nationalarchives.gov.uk/pronom/", filepath.Join(Config.Data, Config.Reports))
	if len(errs) != 0 {
		for _, err := range errs {
			t.Error("SaveReports fail", err)
		}

	}
}

func TestSaveReport(t *testing.T) {
	for _, puid := range []string{"x-fmt/365", "x-fmt/128"} {
		err := SaveReport(puid, "http://www.nationalarchives.gov.uk/pronom/", filepath.Join(Config.Data, Config.Reports))
		if err != nil {
			t.Error("SaveReport fail ", err)
		}
	}
}
*/
