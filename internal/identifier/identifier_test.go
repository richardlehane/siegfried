package identifier

import (
	"testing"

	"github.com/richardlehane/siegfried/pkg/core"
)

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
