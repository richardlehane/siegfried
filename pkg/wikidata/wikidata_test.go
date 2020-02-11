package wikidata

import (
	"testing"
)

type idTestsStruct struct {
	uri string
	res string
}

var idTests = []idTestsStruct{
	idTestsStruct{"http://www.wikidata.org/entity/Q1023647", "Q1023647"},
	idTestsStruct{"http://www.wikidata.org/entity/Q336284", "Q336284"},
	idTestsStruct{"http://www.wikidata.org/entity/Q9296340", "Q9296340"},
}

// TestGetID is a rudimentary test to make sure that we can retrieve
// QUIDs reliably from a Wikidata URI.
func TestGetID(t *testing.T) {
	for _, v := range idTests {
		res := getID(v.uri)
		if res != v.res {
			t.Errorf("Expected to generate QID '%s' but received '%s'", v.res, res)
		}
	}
}
