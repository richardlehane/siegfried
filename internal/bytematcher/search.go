// Copyright 2018 Richard Lehane. All rights reserved.
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

// This file implements Aho Corasick searching for the bytematcher package
package bytematcher

import (
	//"fmt"
	//"io"
	//"sort"
	//"strings"
	//"sync"

	"github.com/richardlehane/siegfried/internal/siegreader"
)

type searcher struct {
	bofWac        ac
	eofWac        ac
	maxBof        int
	maxEof        int
	wildSeqs      []seq          // separate out wild sequences to create a dynamic searcher for wildcard matching
	entanglements []entanglement // same len as wildSeqs
}

func newSearcher(bofSeqs []seq, eofSeqs []seq, entanglements map[int]entanglement) *searcher {
	return nil
}

func (s *searcher) search(buf *siegreader.Buffer) chan result {
	output := make(chan result)
	// check bof
	// check eof
	// build wild matcher
	// check wild bof
	return output
}
