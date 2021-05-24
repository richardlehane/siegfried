// Copyright 2017 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chart

import (
	"testing"
)

// TestExampleChart will compare the chart output with a test string to
// ensure consistency of results.
func TestExampleChart(t *testing.T) {

	const res string = `CENSUS
1950
deaths:    ■ ■ (49)
births:    ■ (11)

1951
deaths:    ■ ■ ■ ■ ■ ■ ■ ■ ■ ■ (200)
births:    ■ (9)
`

	chart := Chart("Census",
		[]string{"1950", "1951", "1952"},
		[]string{"deaths", "births", "marriages"},
		map[string]bool{},
		map[string]map[string]int{"1950": {"births": 11, "deaths": 49}, "1951": {"deaths": 200, "births": 9}},
	)

	// Chart returned from Chart() call is terminated with \n, so this
	// must be accounted for here with strings.Trim, or in the test
	// string above.
	if chart != res {
		t.Errorf(
			"Expected (@@ denotes BOF/EOF string):\n@@%s@@\n\nReceived:\n@@%s@@",
			res,
			chart,
		)
	}
}
