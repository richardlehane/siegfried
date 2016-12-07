package sets

import (
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/richardlehane/siegfried/pkg/config"
)

var testhome = flag.String("testhome", "../../cmd/roy/data", "override the default home directory")

func TestSets(t *testing.T) {
	config.SetHome(*testhome)
	list := "fmt/1,fmt/2,@pdfa,x-fmt/19"
	expect := "fmt/1,fmt/2,fmt/95,fmt/354,fmt/476,fmt/477,fmt/478,fmt/479,fmt/480,fmt/481,x-fmt/19"
	res := strings.Join(Expand(list), ",")
	if res != expect {
		t.Errorf("expecting %s, got %s", expect, res)
	}
	pdfs := strings.Join(Expand("@pdf"), ",")
	expect = "fmt/14,fmt/15,fmt/16,fmt/17,fmt/18,fmt/19,fmt/20,fmt/95,fmt/144,fmt/145,fmt/146,fmt/147,fmt/148,fmt/157,fmt/158,fmt/276,fmt/354,fmt/476,fmt/477,fmt/478,fmt/479,fmt/480,fmt/481,fmt/488,fmt/489,fmt/490,fmt/491,fmt/492,fmt/493"
	if pdfs != expect {
		t.Errorf("expecting %s, got %s", expect, pdfs)
	}
	compression := strings.Join(Expand("@compression"), ",")
	expect = "fmt/626,x-fmt/266,x-fmt/267,x-fmt/268"
	if compression != expect {
		t.Errorf("expecting %s, got %s", expect, compression)
	}
}

var testSet = map[string][]string{
	"t": {"a", "a", "b", "c"},
	"u": {"b", "d"},
	"v": {"@t", "@u"},
}

func TestDupeSets(t *testing.T) {
	orig := sets
	sets = testSet
	expect := "a,b,c,d"
	res := strings.Join(Expand("@v"), ",")
	if res != expect {
		t.Errorf("expecting %s, got %s", expect, res)
	}
	sets = orig
}

var (
	tika          = []string{"x-fmt/111", "@pdf", "@msword"}
	fnName string = "IsText"
)

func ExampleSets() {
	fmt.Printf("func %s(puid string) bool {\n  switch puid {\n  case \"%s\":\n    return true\n  }\n  return false\n}", fnName, strings.Join(Sets(tika...), "\",\""))
	// Output:
	//func IsText(puid string) bool {
	//   switch puid {
	//   case "fmt/14","fmt/15","fmt/16","fmt/17","fmt/18","fmt/19","fmt/20","fmt/37","fmt/38","fmt/39","fmt/40","fmt/95","fmt/144","fmt/145","fmt/146","fmt/147","fmt/148","fmt/157","fmt/158","fmt/276","fmt/354","fmt/412","fmt/473","fmt/476","fmt/477","fmt/478","fmt/479","fmt/480","fmt/481","fmt/488","fmt/489","fmt/490","fmt/491","fmt/492","fmt/493","fmt/523","fmt/597","fmt/599","fmt/609","fmt/754","x-fmt/45","x-fmt/111","x-fmt/273","x-fmt/274","x-fmt/275","x-fmt/276":
	//     return true
	//   }
	//   return false
	//}
}
