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

import "fmt"

func ExampleChart() {
	fmt.Print(Chart("Census",
		[]string{"1950", "1951", "1952"},
		[]string{"deaths", "births", "marriages"},
		map[string]bool{},
		map[string]map[string]int{"1950": {"births": 11, "deaths": 49}, "1951": {"deaths": 200, "births": 9}},
	))
	// Output:
	// CENSUS
	// 1950
	// deaths:   ■ ■ (49)
	// births:   ■ (11)
	//
	// 1951
	// deaths:   ■ ■ ■ ■ ■ ■ ■ ■ ■ ■ (200)
	// births:   ■ (9)
}
