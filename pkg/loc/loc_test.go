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

	"github.com/richardlehane/siegfried/config"
)

func TestLOC(t *testing.T) {
	var dump = false // set to true to print out LOC sigs

	config.SetHome(filepath.Join("..", "..", "cmd", "roy", "data"))
	l, err := newLOC(config.LOC())
	if l != nil {
		if dump {
			fdd := l.(fdds)
			for _, v := range fdd {
				fmt.Println(v)
				fmt.Println("****************")
			}
		}
	} else {
		t.Fatalf("Expecting a LOC, got nothing! Error: %v", err)
	}
}

var hexTests = []string{"52 49 46 46 xx xx xx xx 57 41 56 45 66 6D 74 20",
	"46 4F 52 4D 00",
	"06 0E 2B 34 02 05 01 01 0D 01 02 01 01 02",
	"0xFF 0xD8",
	"FF D8 FF E0 xx xx 4A 46 49 46 00",
	"FF D8 FF E8 xx xx 53 50 49 46 46 00 ",
	"49 49 2A 00",
	"49 49",
	"4D 4D 00 2A",
	"4F 67 67 53 00 02 00 00 00 00 00 00 00 00",
	"30 26 B2 75 8E 66 CF 11 A6 D9 00 AA 00 62 CE 6C",
	"00 00 01 Bx",
	"25 50 44 46",
	//"00 00 01 Bx",
	"2E 52 4D 46",
	"2E 52 4D 46 00 00 00 12 00",
	"2E 52 4D 46",
	"2E 52 4D 46 00 00 00 12 00",
	"xx xx xx xx 6D 6F 6F 76",
	"52 49 46 46 xx xx xx xx 41 56 49 20 4C 49 53 54",
	"30 26 B2 75 8E 66 CF 11 A6 D9 00 AA 00 62 CE 6C",
	"30 26 B2 75 8E 66 CF 11 A6 D9 00 AA 00 62 CE 6C",
	"FF FB 30",
	"46 4F 52 4D 00",
	"52 49 46 46",
	"4D 54 68 64",
	"52 49 46 46 xx xx xx xx 52 4D 49 44 64 61 74 61",
	"58 4D 46 5F",
	"4A 4E",
	"69 66",
	"44 44 4D 46",
	"46 41 52 FE",
	"49 4D 50 4D",
	"4D 4D 44",
	"4D 54 4D",
	"4F 4B 54 41 53 4F 4E 47 43 4D 4F 44",
	//Hex (position 25): 00 00 00 1A 10 00 00
	"45 78 74 65 6E 64 65 64 20 6D 6F 64 75 6C 65 3A 20",
	//12 byte string: X'0000 000C 6A50 2020 0D0A 870A'
	"46 57 53",
	"43 57 53",
	"46 4C 56",
	"D0 CF 11 E0 A1 B1 1A E1 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00",
	"47 49 46 38 39 61",
	"00 00 00 0C 6A 50 20 20 0D 0A 87 0A 00 00 00 14 66 74 79 70 6A 70 32",
	"49 20 49",
	"FF D8 FF E1 xx xx 45 78 69 66 00",
	"0xFF 0xD8",
	"0xFF 0xD8",
	"0xFF 0xD8",
	"89 50 4e 47 0d 0a 1a 0a",
	"0x53445058",
	"0x58504453",
	"66 4C 61 43 00 00 00 22",
	"61 6A 6B 67 02 FB",
	"0E 03 13 01",
	"89 48 44 46 0d 0a 1a 0a",
	"49 49 1A 00 00 00 48 45 41 50 43 43 44 52 02 00 01",
	"00 4D 52 4D",
	"46 4F 56 62",
	"49 49 BC",
	"46 57 53",
	"43 57 53",
	"0x2321414d520a",
	"00 00 00",
	"0x2321414d522d57420a",
	"4F 67 67 53 00 02 00 00 00 00 00 00 00 00",
	"00 00 27 0A 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00",
	"53 49 4d 50 4c 45",
	"53 49 4d 50 4c 45 20 20 3d {20 bytes of Hex 20} 54",
	"43 44 46 01",
	"43 44 46 02",
	"0xFF 0xD8",
	"0xFF 0xD8",
	"00 00 01 Bx",
	//1A 45 DF A3 93 42 82 88
	//6D 61 74 72 6F 73 6B 61
	"43 44 30 30 31",
	"21 42 44 4E",
	"00 0D BB A0",
	"45 56 46 09 0D 0A FF 00",
	"45 56 46 09 0D 0A FF 00",
	"4C 56 46 09 0D 0A FF 00",
	"45 56 46 32 0D 0A 81 00",
	"4C 45 46 32 0D 0A 81 00",
	"41 46 46",
	"53 51 4c 69 74 65 20 66 6f 72 6d 61 74 20 33 00",
	"1A 00 00 04 00 00",
	"4D 41 54 4C 41 42 20 35 2E 30 20 4D 41 54 2D 66 69 6C 65 2C 20 50 6C 61 74 66 6F 72 6D 3A 20"}

func TestHex(t *testing.T) {
	for _, v := range hexTests {
		_, err := hexMagic(v)
		if err != nil {
			t.Fatalf("Error parsing hex: %s; got %v", v, err)
		}
	}
}
