// Copyright 2016 Richard Lehane. All rights reserved.
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
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/persist"
)

func init() {
	core.RegisterIdentifier(core.MIMEInfo, Load)
}

type Identifier struct {
}

func (i *Identifier) Save(ls *persist.LoadSaver) {
}

func Load(ls *persist.LoadSaver) core.Identifier {
	return nil
}

func New(opts ...config.Option) (*Identifier, error) {
	for _, v := range opts {
		v()
	}
	mi, err := newMIMEInfo()
	if err != nil {
		return nil, err
	}
	return mi.identifier(), nil
}

func (i *Identifier) Add(m core.Matcher, t core.MatcherType) error {
	return nil
}

func (i *Identifier) Describe() [2]string {
	return [2]string{}
}

func (i *Identifier) String() string {
	return ""
}

func (i *Identifier) Recognise(m core.MatcherType, idx int) (bool, string) {
	return false, ""
}

func (i *Identifier) Recorder() core.Recorder {
	return nil
}
