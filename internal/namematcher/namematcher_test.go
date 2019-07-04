package namematcher

import (
	"testing"

	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/pkg/core"
)

var fmts = SignatureSet{"*.wav", "*.doc", "*.xls", "*.pdf", "*.ppt", "*.adoc.txt", "README"}

var sm core.Matcher

func init() {
	sm, _, _ = Add(nil, fmts, nil)
}

func TestWavMatch(t *testing.T) {
	res, _ := sm.Identify("hello/apple.wav", nil)
	e := <-res
	if e.Index() != 0 {
		t.Errorf("Expecting 0, got %v", e)
	}
	e, ok := <-res
	if ok {
		t.Error("Expecting a length of 1")
	}
}

func TestAdocMatch(t *testing.T) {
	res, _ := sm.Identify("hello/apple.adoc.txt", nil)
	e := <-res
	if e.Index() != 5 {
		t.Errorf("Expecting 5, got %v", e)
	}
	e, ok := <-res
	if ok {
		t.Error("Expecting a length of 1")
	}
}

func TestREADMEMatch(t *testing.T) {
	res, _ := sm.Identify("hello/README", nil)
	e, ok := <-res
	if ok {
		if e.Index() != 6 {
			t.Errorf("Expecting 6, got %v", e)
		}
	} else {
		t.Error("Expecting 5, got nothing")
	}
	e, ok = <-res
	if ok {
		t.Error("Expecting a length of 1")
	}
}

func TestNoMatch(t *testing.T) {
	res, _ := sm.Identify("hello/apple.tty", nil)
	_, ok := <-res
	if ok {
		t.Error("Should not match")
	}
}

func TestNoExt(t *testing.T) {
	res, _ := sm.Identify("hello/apple", nil)
	_, ok := <-res
	if ok {
		t.Error("Should not match")
	}
}

func TestIO(t *testing.T) {
	sm, _, _ = Add(nil, SignatureSet{"*.bla", "*.doc", "*.ppt"}, nil)
	str := sm.String()
	saver := persist.NewLoadSaver(nil)
	Save(sm, saver)
	if len(saver.Bytes()) < 10 {
		t.Errorf("Save string matcher: too small, only got %v", saver.Bytes())
	}
	loader := persist.NewLoadSaver(saver.Bytes())
	newsm := Load(loader)
	str2 := newsm.String()
	if str != str2 {
		t.Errorf("Load string matcher: expecting first matcher (%v), to equal second matcher (%v)", str, str2)
	}
}

var fnames = []string{
	"README",
	"README",
	"",
	"\\this\\directory\\file.txt",
	"file.txt",
	"txt",
	"c:\\docs\\SONG.MP3",
	"SONG.MP3",
	"mp3",
	"Climate/Existential.pdf",
	"Existential.pdf",
	"pdf",
	"/Volumes/Public/bearbeiten/Dateien/ermitteln Dateityp/Salzburger Nachtstudio.2019-06-19 - Kulturkampf im Klassenzimmer?.mp3",
	"Salzburger Nachtstudio.2019-06-19 - Kulturkampf im Klassenzimmer?.mp3",
	"mp3",
	"http://www.archive.org/about/faq.php?faq_id=243 207.241.229.39",
	"faq.php",
	"php",
	"http://www.archive.org/images/wayback-election2000.gif",
	"wayback-election2000.gif",
	"gif",
	"http://www.example.org/foo.html#bar",
	"foo.html",
	"html",
}

func TestNormalise(t *testing.T) {
	for i := 0; i < len(fnames); i += 3 {
		fname, ext := normalise(fnames[i])
		if fname != fnames[i+1] {
			t.Errorf("normalise filename error\ninput: %s\nexpect: %s\ngot: %s", fnames[i], fnames[i+1], fname)
		}
		if ext != fnames[i+2] {
			t.Errorf("normalise ext error\ninput: %s\nexpect: %s\ngot: %s", fnames[i], fnames[i+2], ext)
		}
	}
}
