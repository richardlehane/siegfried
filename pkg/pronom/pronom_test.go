package pronom

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/pkg/config"
)

var dataPath string = filepath.Join("..", "..", "cmd", "roy", "data")

var minimalPronom = []string{"fmt/1", "fmt/3", "fmt/5", "fmt/11", "fmt/14"}

func TestNew(t *testing.T) {
	config.SetHome(dataPath)
	_, err := NewPronom()
	if err != nil {
		t.Error(err)
	}
}

// verifyIdentifier provides a number of tests that can be run against
// a PRONOM identifier.
func verifyIdentifier(i identifier.Parseable, t *testing.T) {
	const minReports int = 5
	if len(i.Infos()) != minReports {
		t.Error("Unexpected number of reports for PRONOM minimal tests")
	}
	expectedPuids := []string{
		"fmt/1",
		"fmt/3",
		"fmt/5",
		"fmt/11",
		"fmt/14",
	}
	expectedNames := []string{
		"Broadcast WAVE",
		"Graphics Interchange Format",
		"Audio/Video Interleaved Format",
		"Portable Network Graphics",
		"Acrobat PDF 1.0 - Portable Document Format",
	}
	expectedVersions := []string{
		"0 Generic",
		"87a",
		"",
		"1.0",
		"1.0",
	}
	expectedMimes := []string{
		"image/gif",
		"video/x-msvideo",
		"image/png",
		"application/pdf",
		"audio/x-wav",
	}
	expectedTypes := []string{
		"Audio",
		"Image (Raster)",
		"Audio, Video",
		"Image (Raster)",
		"Page Description",
	}
	puids := make([]string, 0)
	names := make([]string, 0)
	versions := make([]string, 0)
	mimes := make([]string, 0)
	types := make([]string, 0)
	for puid := range i.Infos() {
		puids = append(puids, puid)
		names = append(names, i.Infos()[puid].(formatInfo).name)
		versions = append(versions, i.Infos()[puid].(formatInfo).version)
		mimes = append(mimes, i.Infos()[puid].(formatInfo).mimeType)
		types = append(types, i.Infos()[puid].(formatInfo).class)
	}
	sort.Strings(puids)
	sort.Strings(expectedPuids)
	if !reflect.DeepEqual(puids, expectedPuids) {
		t.Errorf("PUIDs from minimal PRONOM set do not match expected values;  expected %v; got %v", puids, expectedPuids)
	}
	sort.Strings(names)
	sort.Strings(expectedNames)
	if !reflect.DeepEqual(names, expectedNames) {
		t.Errorf("Format names from minimal PRONOM set do not match expected values; expected %v, got %v", expectedNames, names)
	}
	sort.Strings(versions)
	sort.Strings(expectedVersions)
	if !reflect.DeepEqual(versions, expectedVersions) {
		t.Errorf("Format versions from minimal PRONOM set do not match expected values;  expected %v; got %v", versions, expectedVersions)
	}
	sort.Strings(mimes)
	sort.Strings(expectedMimes)
	if !reflect.DeepEqual(mimes, expectedMimes) {
		t.Errorf("MIMETypes from minimal PRONOM set do not match expected values; expected %v; got %v", mimes, expectedMimes)
	}
	sort.Strings(types)
	sort.Strings(expectedTypes)
	if !reflect.DeepEqual(types, expectedTypes) {
		t.Errorf("Format types from minimal PRONOM set do not match expected values; expected %v; got %v", types, expectedTypes)
	}
}

// TestFormatInfosDefault inspects the values loaded into a PRONOM
// identifier from a minimal PRONOM dataset, i.e. fewer than loading
// all of PRONOM.
func TestFormatInfosDefault(t *testing.T) {
	config.SetHome(dataPath)
	config.SetLimit(minimalPronom)()
	i, err := NewPronom()
	if err != nil {
		t.Error(err)
	}
	verifyIdentifier(i, t)
	config.Clear()()
}

// TestFormatInfosNoReports performs the same tests as TestFormatInfosDefault
// but does so without loading PRONOM reports, preferring to create an
// identifier using a signature file only.
func TestFormatInfosNoReports(t *testing.T) {
	config.SetHome(dataPath)
	config.SetLimit(minimalPronom)()
	config.SetNoContainer()()
	config.SetNoReports()()
	if config.Reports() != "" {
		t.Errorf("pronon.reports should be unset, not: %s", config.Reports())
	}
	i, err := NewPronom()
	if err != nil {
		t.Error(err)
	}
	verifyIdentifier(i, t)
	config.Clear()()
}
