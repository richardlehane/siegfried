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
	//"fmt"
	"path/filepath"
	"testing"

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/internal/persist"
)

func TestNew(t *testing.T) {
	config.SetHome(filepath.Join("..", "..", "cmd", "roy", "data"))
	config.SetMIMEInfo("tika-mimetypes.xml")()
	mi, err := newMIMEInfo(config.MIMEInfo())
	if err != nil {
		t.Error(err)
	}
	sigs, ids, err := mi.Signatures()
	if err != nil {
		t.Error(err)
	}
	for i, v := range sigs {
		if len(v) == 0 {
			t.Errorf("Empty signature: %s", ids[i])
		}
	}
	id, _ := New()
	str := id.String()
	saver := persist.NewLoadSaver(nil)
	id.Save(saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	_ = loader.LoadByte()
	id2 := Load(loader)
	if str != id2.String() {
		t.Errorf("Load identifier fail: got %s, expect %s", str, id2.String())
	}
}
