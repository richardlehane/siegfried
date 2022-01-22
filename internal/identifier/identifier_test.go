package identifier

import (
	"reflect"
	"testing"

	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core"

	"github.com/richardlehane/siegfried/internal/bytematcher/patterns"
)

// Globals to enable testing and comparison of Parseable results.
var sigs []frames.Signature
var ids []string
var f0, f1, f2, f3, f4, f5, f6 frames.Signature

func init() {
	sigs = make([]frames.Signature, 0, 7)
	ids = make([]string, 0, 7)

	hx := "Hex: 4D 4D 00 2A"

	var pat = patterns.Sequence(hx)

	f0 = frames.Signature{frames.NewFrame(frames.BOF, pat, 0, 0)}
	f1 = frames.Signature{frames.NewFrame(frames.BOF, pat, 1, 1)}
	f2 = frames.Signature{frames.NewFrame(frames.BOF, pat, 2, 2)}
	f3 = frames.Signature{frames.NewFrame(frames.BOF, pat, 3, 3)}
	f4 = frames.Signature{frames.NewFrame(frames.BOF, pat, 4, 4)}
	f5 = frames.Signature{frames.NewFrame(frames.BOF, pat, 5, 5)}
	f6 = frames.Signature{frames.NewFrame(frames.BOF, pat, 6, 6)}

	sigs = append(sigs, f6)
	sigs = append(sigs, f2)
	sigs = append(sigs, f1)
	sigs = append(sigs, f4)
	sigs = append(sigs, f5)
	sigs = append(sigs, f0)
	sigs = append(sigs, f3)

	// IDs deliberately out of order so that they are reordered during
	// Parseable's sort.
	ids = append(ids, "text/x-go")
	ids = append(ids, "fdd000002")
	ids = append(ids, "fdd000001")
	ids = append(ids, "fmt/1")
	ids = append(ids, "fmt/2")
	ids = append(ids, "application/x-elf")
	ids = append(ids, "fdd000002")
}

func TestFind(t *testing.T) {
	testBase := &Base{
		gids: &indexes{
			start: 50,
			ids: []string{
				"fmt/1",
				"fmt/2",
				"fmt/3",
				"fmt/4",
				"fmt/1",
				"fmt/5",
			},
		},
	}
	expect := []int{50, 54, 52}
	lookup := testBase.Lookup(core.NameMatcher, []string{"fmt/1", "fmt/3"})
	if len(lookup) != len(expect) || lookup[0] != expect[0] || lookup[1] != expect[1] || lookup[2] != expect[2] {
		t.Fatalf("Failed lookup: got %v, expected %v", lookup, expect)
	}
}

// Utilize Parseable's Blank identifier so that we can override
// Signatures() for the purposes of testing.
type testParseable struct{ Blank }

func (b testParseable) Signatures() ([]frames.Signature, []string, error) {
	return sigs, ids, nil
}

// TestSorted tests the Parseable sort mechanism that will be shared
// across identifiers. Identifiers each contain a Parseable.
func TestSorted(t *testing.T) {

	sigsBeforeSort := []frames.Signature{f6, f2, f1, f4, f5, f0, f3}
	sigsAfterSort := []frames.Signature{f0, f1, f2, f3, f4, f5, f6}

	idsBeforeSort := []string{
		"text/x-go",
		"fdd000002",
		"fdd000001",
		"fmt/1",
		"fmt/2",
		"application/x-elf",
		"fdd000002",
	}
	idsAfterSort := []string{
		"application/x-elf",
		"fdd000001",
		"fdd000002",
		"fdd000002",
		"fmt/1",
		"fmt/2",
		"text/x-go",
	}

	identifier := &Base{}
	identifier.p = testParseable{}

	sigs, ids, err := identifier.p.Signatures()

	if err != nil {
		t.Error("Signatures() should not have returned an error", err)
	}

	for idx, val := range sigs {
		if !reflect.DeepEqual(sigsBeforeSort[idx], val) {
			t.Error("Results should not have been sorted")
			t.Errorf("Returned: %+v expected: %+v", sigs, sigsBeforeSort)
		}
	}

	if !reflect.DeepEqual(ids, idsBeforeSort) {
		t.Error("Results should not have been sorted")
		t.Errorf("Returned: %s expected: %s", ids, idsBeforeSort)
	}

	identifier.p = ApplyConfig(identifier.p)
	sigs, ids, err = identifier.p.Signatures()

	if err != nil {
		t.Error("Signatures() should not have returned an error", err)
	}

	for idx, val := range sigs {
		if !reflect.DeepEqual(sigsAfterSort[idx], val) {
			t.Error("Results should have been sorted")
			t.Errorf("Returned: %+v expected: %+v", sigs, sigsAfterSort)
		}
	}

	if !reflect.DeepEqual(ids, idsAfterSort) {
		t.Error("Results should not have been sorted")
		t.Errorf("Returned: %s expected: %s", ids, idsAfterSort)
	}
}
