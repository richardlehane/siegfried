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

package mimeinfo

import (
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns"
)

type Big16 uint16

type Big32 uint32

type Little16 uint16

type Little32 uint32

type Host16 uint16 // Implement as: Big OR Little

type Host32 uint32 // Implement as: Big OR Little

type IgnoreCase []byte // @book has 16 possible values 1*2*2*2*2

type Mask struct {
	pat patterns.Pattern
	val []byte // masks for numerical types can be any number; masks for strings must be in base16 and start with 0x
}

func (m Mask) Test(b []byte) (bool, int) {
	if len(b) < len(m.val) {
		return false, 0
	}
	t := make([]byte, len(m.val))
	for i := range t {
		t[i] = t[i] & m.val[i]
	}
	return m.pat.Test(t)
}
