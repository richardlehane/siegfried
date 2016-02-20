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
	"fmt"
	"path/filepath"
	"testing"

	"github.com/richardlehane/siegfried/config"
)

func TestNew(t *testing.T) {
	config.SetHome(filepath.Join("..", "..", "cmd", "roy", "data"))
	mi, err := newMIMEInfo()
	if err != nil {
		t.Error(err)
	}
	//tpmap := make(map[string]struct{})
	for _, v := range mi {
		/*
			fmt.Println(v)
			if len(v.Magic) > 1 {
				fmt.Printf("Multiple magics (%d): %s\n", len(v.Magic), v.MIME)
			}
			for _, c := range v.Magic {
				for _, d := range c.Matches {
					tpmap[d.Typ] = struct{}{}
					if len(d.Mask) > 0 {
						if d.Typ == "string" {
							fmt.Println("MAGIC: " + d.Value)
						} else {
							fmt.Println("Type: " + d.Typ)
						}
						fmt.Println("MASK: " + d.Mask)
					}
				}
			}*/
		for _, c := range v.XMLPattern {
			fmt.Printf("Root: %s; Namespace: %s\n", c.Local, c.NS)
		}
	}
	/*
		for k, _ := range tpmap {
			fmt.Println(k)
		}
	*/
	if len(mi) != 1495 {
		t.Errorf("expecting %d MIMEInfos, got %d", 1495, len(mi))
	}
}
