package main

import (
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/pkg/config"
)

var testResp = `
[
	{
		"sf":[1,11,1],
		"created":"2023-05-12T09:10:13Z",
		"hash":"2ad9b1cb28370add9320473676e3a1ba9e0311fc22a058c0862f4fb68f890582",
		"size":204689,
		"path":"https://www.itforarchivists.com/siegfried/latest"
	},
	{
		"sf":[1,10,1],
		"created":"2023-05-12T09:10:13Z",
		"hash":"2ad9b1cb28370add9320473676e3a1ba9e0311fc22a058c0862f4fb68f890582",
		"size":204689,
		"path":"https://www.itforarchivists.com/siegfried/latest"
	}
]
`

func TestUpdate(t *testing.T) {
	config.SetHome("../roy/data")
	var us Updates
	if err := json.Unmarshal([]byte(testResp), &us); err != nil {
		t.Fatal(err)
	}
	// Edit the version, size, hash, and created time of the first response so that
	// it always matches and we are always up-to-date
	us[0].Version = config.Version()
	fbuf, err := os.ReadFile(config.Signature())
	if err != nil {
		t.Fatal(err)
	}
	us[0].Size = len(fbuf)
	h := sha256.New()
	h.Write(fbuf)
	us[0].Hash = hex.EncodeToString(h.Sum(nil))
	byt, err := json.Marshal(us)
	if err != nil {
		t.Fatal(err)
	}
	if len(fbuf) < len(config.Magic())+2+15 {
		t.Fatal("signature file too short!")
	}
	rc := flate.NewReader(bytes.NewBuffer(fbuf[len(config.Magic())+2:]))
	nbuf := make([]byte, 15)
	if n, _ := rc.Read(nbuf); n < 15 {
		t.Fatal("bad signature")
	}
	rc.Close()
	ls := persist.NewLoadSaver(nbuf)
	tt := ls.LoadTime()
	if ls.Err != nil {
		t.Fatal(ls.Err)
	}
	us[0].Created = tt.Format(time.RFC3339)
	tgf := func(url string) ([]byte, error) {
		return byt, nil
	}
	_, str, err := updateSigsDo("http://www.example.com", nil, tgf)
	if str != "You are already up to date!" {
		t.Fatalf("got: %s; error: %v", str, err)
	}
}
