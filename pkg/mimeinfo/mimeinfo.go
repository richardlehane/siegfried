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
	"github.com/richardlehane/siegfried/pkg/mimeinfo/mappings"
)

func newMIMEInfo() ([]mappings.MIMEType, error) {
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
