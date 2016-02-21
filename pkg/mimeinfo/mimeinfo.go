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
	"encoding/xml"
	"io/ioutil"

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/parseable"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/mimeinfo/mappings"
)

type mimeinfo []mappings.MIMEType

func (mi mimeinfo) identifier() *Identifier { return nil }

func newMIMEInfo() (mimeinfo, error) {
	buf, err := ioutil.ReadFile(config.MIMEInfo())
	if err != nil {
		return nil, err
	}
	mi := &mappings.MIMEInfo{}
	err = xml.Unmarshal(buf, mi)
	if err != nil {
		return nil, err
	}
	return mi.MIMETypes, nil
}

func (mi mimeinfo) IDs() []string                                     { return nil }
func (mi mimeinfo) Infos() map[string]parseable.FormatInfo            { return nil }
func (mi mimeinfo) Globs() ([][]string, []string)                     { return nil, nil }
func (mi mimeinfo) MIMEs() ([][]string, []string)                     { return nil, nil }
func (mi mimeinfo) XMLs() ([][][2]string, []string)                   { return nil, nil }
func (mi mimeinfo) Signatures() ([]frames.Signature, []string, error) { return nil, nil, nil }
func (mi mimeinfo) Priorities() priority.Map                          { return nil }
