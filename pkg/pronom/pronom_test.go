package pronom

import (
	//"path/filepath"
	"testing"
)

var p *pronom

func TestNew(t *testing.T) {
	var err error
	p, err = New(ConfigPaths())
	if err != nil {
		t.Error("New fail: ", err)
	}
}

/*These seems to work but takes a while, so left out of routine testing
func TestSaveReports(t *testing.T) {
	errs := SaveReports(filepath.Join(Config.Data, Config.Droid), "http://www.nationalarchives.gov.uk/pronom/", filepath.Join(Config.Data, Config.Reports))
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
