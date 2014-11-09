// Copyright 2014 Richard Lehane. All rights reserved.
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

package pronom

import (
	"github.com/richardlehane/siegfried/pkg/core/priority"
)

// returns a map of puids and the indexes of byte signatures that those puids should give priority to
func (p pronom) priorities() priority.Map {
	pMap := make(priority.Map)
	for _, f := range p.droid.FileFormats {
		superior := f.Puid
		for _, v := range f.Priorities {
			subordinate := p.ids[v]
			pMap.Add(subordinate, superior)
		}
	}
	pMap.Complete()
	return pMap
}
