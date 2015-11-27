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

type Big16 int16

type Big32 int32

type Little16 int16

type Little32 int32

type Host16 int16

type Host32 int32

type Mask struct {
	pat patterns.Pattern
	val []byte // masks for numerical types can be any number; masks for strings must be in base16 and start with 0x
}
