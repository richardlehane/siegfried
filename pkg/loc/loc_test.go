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

package loc

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/richardlehane/siegfried/config"
)

func TestLOC(t *testing.T) {
	var dump, dumpmagic bool // set to true to print out LOC sigs

	config.SetHome(filepath.Join("..", "..", "cmd", "roy", "data"))
	l, err := newLOC(config.LOC())
	if l != nil {
		if dump {
			fdd := l.(fdds)
			for _, v := range fdd {
				fmt.Println(v)
				fmt.Println("****************")
			}
		} else if dumpmagic {
			fdd := l.(fdds)
			for _, v := range fdd {
				if len(v.Magics) > 0 {
					fmt.Println("{")
					for _, m := range v.Magics {
						fmt.Println("`" + m + "`,")
					}
					fmt.Println("},")
				}
			}
		}
		if _, _, err = l.Signatures(); err != nil {
			t.Fatal(err)
		}
	} else {
		t.Fatalf("Expecting a LOC, got nothing! Error: %v", err)
	}
}

func TestUpdated(t *testing.T) {
	config.SetHome(filepath.Join("..", "..", "cmd", "roy", "data"))
	l, err := newLOC(config.LOC())
	if err != nil || l == nil {
		t.Fatalf("couldn't parse LOC file: %v", err)
	}
	expect, _ := time.Parse(dateFmt, "2016-01-01")
	f := l.(fdds)
	if !f.Updated().After(expect) {
		t.Fatalf("expected %v, got %v", expect, f.Updated())
	}
}
