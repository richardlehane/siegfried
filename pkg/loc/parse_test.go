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

import "testing"

var magicTests = [][]string{
	{
		`Hex: 52 49 46 46 xx xx xx xx 57 41 56 45 66 6D 74 20`,
		`ASCII: RIFF....WAVEfmt`,
	},
	{
		`Hex: 46 4F 52 4D 00`,
		`ASCII: FORM`,
	},
	{
		`Hex: 06 0E 2B 34 02 05 01 01 0D 01 02 01 01 02`,
		`ASCII: ..+4`,
	},
	{
		`Hex: 0xFF 0xD8`,
	},
	{
		`Hex: FF D8 FF E0 xx xx 4A 46 49 46 00`,
		`ASCII: ÿØÿè..JFIF.`,
	},
	{
		`Hex: FF D8 FF E8 xx xx 53 50 49 46 46 00  `,
		`ASCII: ÿØÿè..SPIFF`,
	},
	{
		`Hex: 49 49 2A 00`,
		`Hex: 49 49`,
		`ASCII: II`,
		`Hex: 4D 4D 00 2A`,
		`ASCII: MM`,
	},
	{
		`Hex: 4F 67 67 53 00 02 00 00 00 00 00 00 00 00`,
		`ASCII: OggS`,
	},
	{
		`Hex: 30 26 B2 75 8E 66 CF 11 A6 D9 00 AA 00 62 CE 6C`,
		`ASCII: 0&²u.fÏ.¦Ù.ª.bÎl`,
	},
	{
		`Hex: 00 00 01 Bx`,
		`ASCII: ....`,
	},
	{
		`Hex: 25 50 44 46`,
		`ASCII: %PDF`,
	},
	{
		`ASCII: msid`,
	},
	{
		`Hex: 00 00 01 Bx`,
		`ASCII: ....`,
	},
	{
		`Hex: 2E 52 4D 46`,
		`ASCII: .RMF`,
		`Hex: 2E 52 4D 46 00 00 00 12 00`,
		`ASCII: .RMF`,
	},
	{
		`Hex: 2E 52 4D 46`,
		`ASCII: .RMF`,
		`Hex: 2E 52 4D 46 00 00 00 12 00`,
		`ASCII: .RMF`,
	},
	{
		`Hex: xx xx xx xx 6D 6F 6F 76`,
		`ASCII: ....moov`,
	},
	{
		`Hex: 52 49 46 46 xx xx xx xx 41 56 49 20 4C 49 53 54`,
		`ASCII: RIFF....AVILIST`,
	},
	{
		`Hex: 30 26 B2 75 8E 66 CF 11 A6 D9 00 AA 00 62 CE 6C`,
		`ASCII: 0&²u.fÏ.¦Ù.ª.bÎl`,
	},
	{
		`Hex: 30 26 B2 75 8E 66 CF 11 A6 D9 00 AA 00 62 CE 6C`,
		`ASCII: 0&²u.fÏ.¦Ù.ª.bÎl`,
	},
	{
		`Hex: FF FB 30`,
	},
	{
		`Hex: 46 4F 52 4D 00`,
		`ASCII: FORM`,
	},
	{
		`Hex: 52 49 46 46`,
		`ASCII: RIFF`,
	},
	{
		`Hex: 4D 54 68 64`,
		`ASCII: MThd`,
	},
	{
		`Hex: 52 49 46 46 xx xx xx xx 52 4D 49 44 64 61 74 61`,
		`ASCII: RIFF....RMIDdata`,
	},
	{
		`Hex: 58 4D 46 5F`,
		`ASCII: XMF_`,
	},
	{
		`Hex: 4A 4E`,
		`ASCII: JN`,
		` Hex: 69 66`,
		`ASCII: if`,
		`Hex: 44 44 4D 46`,
		`ASCII: DDMF`,
		`Hex: 46 41 52 FE`,
		`ASCII: FAR`,
		`Hex: 49 4D 50 4D`,
		`ASCII: IMPM`,
		`Hex: 4D 4D 44`,
		`ASCII: MMD`,
		`Hex: 4D 54 4D`,
		`ASCII: MTM`,
		`Hex: 4F 4B 54 41 53 4F 4E 47 43 4D 4F 44`,
		` ASCII: OKTASONGCMOD`,
		`Hex (position 25): 00 00 00 1A 10 00 00`,
		`Hex: 45 78 74 65 6E 64 65 64 20 6D 6F 64 75 6C 65 3A 20`,
		`ASCII: Extended module:`,
	},
	{
		`12 byte string: X'0000 000C 6A50 2020 0D0A 870A'`,
	},
	{
		`Hex: 46 57 53`,
		`ASCII: FWS`,
		`Hex: 43 57 53`,
		`ASCII: CWS`,
	},
	{
		`Hex: 46 4C 56`,
		`ASCII: FLV`,
	},
	{
		`Hex: D0 CF 11 E0 A1 B1 1A E1 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00`,
	},
	{
		`Hex: 47 49 46 38 39 61`,
		`ASCII: GIF89a`,
	},
	{
		`Hex: 00 00 00 0C 6A 50 20 20 0D 0A 87 0A 00 00 00 14 66 74 79 70 6A 70 32`,
	},
	{
		`Hex: 49 20 49`,
		`ASCII: I<space>I`,
	},
	{
		`HEX:  FF D8 FF E1 xx xx 45 78 69 66 00`,
		`	ASCII: ÿØÿà..EXIF.`,
	},
	{
		`Hex: 0xFF 0xD8`,
	},
	{
		`Hex: 0xFF 0xD8`,
	},
	{
		`Hex: 0xFF 0xD8`,
	},
	{
		`Hex: 89 50 4e 47 0d 0a 1a 0a`,
		`ASCII: \211 P N G \r \n \032 \n`,
	},
	{
		`Hex: 0x53445058`,
		`ASCII: SDPX`,
		`Hex: 0x58504453`,
		`ASCII: XPDS`,
	},
	{
		`BM`,
	},
	{
		`Hex: 66 4C 61 43 00 00 00 22`,
		`ASCII: fLaC`,
	},
	{
		`Hex: 61 6A 6B 67 02 FB`,
		`ASCII: ajkg`,
	},
	{
		`Hex: 0E 03 13 01`,
	},
	{
		`Hex: 89 48 44 46 0d 0a 1a 0a`,
		`ASCII: \211 HDF \r \n \032 \n`,
	},
	{
		`Not applicable`,
	},
	{
		`Hex: 49 49 1A 00 00 00 48 45 41 50 43 43 44 52 02 00 01`,
		`ASCII: II [null] HEAPCCDR`,
		`Hex: 00 4D 52 4D`,
		`ASCII: .MRM`,
		`Hex: 46 4F 56 62`,
		`ASCII: FOVb`,
	},
	{
		`Hex: 49 49 BC`,
		`ASCII: II.`,
	},
	{
		`Hex: 46 57 53`,
		`ASCII: FWS`,
		`Hex: 43 57 53`,
		`ASCII: CWS`,
	},
	{
		`Hex: 0x2321414d520a`,
		`ASCII: #!AMR\n `,
		`Hex: 00 00 00`,
	},
	{
		`Hex: 0x2321414d522d57420a`,
		`ASCII: #!AMR-WB\n`,
	},
	{
		`Hex: 4F 67 67 53 00 02 00 00 00 00 00 00 00 00`,
		`ASCII: OggS`,
	},
	{
		` Hex: 00 00 27 0A 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00`,
		`ASCII: &apos; `,
	},
	{
		`HEX: 53 49 4d 50 4c 45`,
		`ASCII: SIMPLE`,
		`HEX: 53 49 4d 50 4c 45 20 20 3d {20 bytes of Hex 20} 54`,
		`ASCII: SIMPLE {2 spaces} = {20 spaces} T`,
	},
	{
		`Hex: 43 44 46 01`,
		`ASCII: CDF \x01`,
		`Hex: 43 44 46 02`,
		`ASCII: CDF \x02`,
	},
	{
		`Hex: 0xFF 0xD8`,
	},
	{
		`Hex: 0xFF 0xD8`,
	},
	{
		`Hex: 00 00 01 Bx`,
		`ASCII: ....`,
	},
	{
		`1A 45 DF A3 93 42 82 88
	6D 61 74 72 6F 73 6B 61`,
	},
	{
		`Hex: 43 44 30 30 31`,
		`ASCII: CD001`,
	},
	{
		`Hex: 21 42 44 4E`,
		`ASCII: !BDN`,
	},
	{
		`00 0D BB A0`,
	},
	{
		`Hex: 45 56 46 09 0D 0A FF 00`,
		`ASCII: EVF...ÿ.`,
	},
	{
		`Hex: 45 56 46 09 0D 0A FF 00`,
		`ASCII: EVF...ÿ.`,
	},
	{
		`Hex: 4C 56 46 09 0D 0A FF 00`,
		`ASCII: LVF...ÿ.`,
	},
	{
		`Hex: 45 56 46 32 0D 0A 81 00`,
		`ASCII: EVF2....`,
	},
	{
		`Hex: 4C 45 46 32 0D 0A 81 00`,
		`ASCII: LEF2....`,
	},
	{
		`Hex: 41 46 46`,
		`ASCII: AFF`,
	},
	{
		`LASF`,
	},
	{
		`Hex:  53 51 4c 69 74 65 20 66 6f 72 6d 61 74 20 33 00`,
		`ASCII:  SQLite format 3`,
	},
	{
		`ASCII: EHFA_HEADER_TAG`,
		`ASCII: ERDAS_IMG_EXTERNAL_RASTER`,
	},
	{
		`1A 00 00 04 00 00`,
	},
	{
		`Hex: 4D 41 54 4C 41 42 20 35 2E 30 20 4D 41 54 2D 66 69 6C 65 2C 20 50 6C 61 74 66 6F 72 6D 3A 20`,
		`ASCII: MATLAB 5.0 MAT-file, Platform:`,
	},
}

func TestMagic(t *testing.T) {
	for _, v := range magicTests {
		_, err := magics(v)
		if err != nil {
			t.Fatalf("Error parsing: %v; got %v", v, err)
		}
	}
}
