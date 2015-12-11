// Copyright 2015 Richard Lehane. All rights reserved.
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

package mimematcher

import (
	"strings"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/priority"
)

type Matcher struct {
	precise map[string][]int
	general map[string][]int
	sigsets []int // starting indexes of signature sets (so can tell if a general match from same set as a precise)
}

func (m Matcher) Add(ss core.SignatureSet, p priority.List) (int, error) {
	return 0, nil
}

func NormaliseMIME(s string) string {
	idx := strings.LastIndex(s, ";")
	if idx > 0 {
		return s[:idx]
	}
	return s
}
