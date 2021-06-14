// Copyright 2020 Ross Spencer, Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

// Fixtures for testing converter capability.

package converter

// testPatterns for Converter. The list is created from the Wikidata
// sources at the time of writing and minimized to a sensible number.
//
// Custom patterns have also been added as required.
//
// IRIs are followed by un-exported note fields, e.g. encodingNote or
// newEncodingNote to identify clearly for the reader what the IRI
// represents.
var testPatterns = `[
{
  "Signature": "DONOTCONVERT",
  "Encoding": "http://www.wikidata.org/entity/Q98056596",
  "encodingNote": "PERL",
  "Comment": "Do not convert because for now PERL conversion should be limited...",
  "Fail": true,
  "Converted": false
}, {
  "Signature": "ACDC1",
  "Encoding": "http://www.wikidata.org/entity/Q82828",
  "encodingNote": "Hexadecimal",
  "Comment": "This should fail as it is an uneven length HEX string",
  "Fail": true,
  "Converted": false
}, {
  "Signature": "♕♖♗♘♙♚♛♜♝♞♟",
  "Encoding": "http://www.wikidata.org/entity/Q8815",
  "encodingNote": "ASCII",
  "Comment": "A decent chunk of Unicode to test (Chess 1 emoji)",
  "Fail": false,
  "Converted": true
}, {
  "Signature": "\u2655\u2656\u2657\u2658\u2659\u265A\u265B\u265C\u265D\u265E\u265F",
  "Encoding": "http://www.wikidata.org/entity/Q8815",
  "encodingNote": "ASCII",
  "Comment": "A decent chunk of Unicode to test (Chess 2 code-points)",
  "Fail": false,
  "Converted": true
}, {
  "Signature": "424D{4}00000000{4}28000000{8}0100(01|04|08|18|20)00(00|01|02)000000",
  "Encoding": "http://www.wikidata.org/entity/Q35432091",
  "encodingNote": "PRONOM",
  "Fail": false,
  "Converted": false
}, {
  "Signature": "00 61 73 6d",
  "Encoding": "",
  "NewEncoding": "http://www.wikidata.org/entity/Q82828",
  "newEncodingNote": "Hexadecimal",
  "NewSignature": "0061736D",
  "Converted": true
}, {
  "Signature": "7573746172",
  "Encoding": "http://www.wikidata.org/entity/Q82828",
  "encodingNote": "Hexadecimal",
  "Converted": false
}, {
  "Signature": "786D6C6E733A7064666169643D(22|27)687474703A2F2F7777772E6169696D2E6F72672F706466612F6E732F6964*7064666169643A70617274(3D22|3D27|3E)33(22|27|3C2F7064666169643A706172743E){0-11}7064666169643A636F6E666F726D616E6365(3E|3D22|3D27)42(22|27|3C2F7064666169643A636F6E666F726D616E63653E)",
  "Encoding": "http://www.wikidata.org/entity/Q35432091",
  "encodingNote": "PRONOM",
  "Converted": false
}, {
  "Signature": "255044462D312E[30:37]",
  "Encoding": "http://www.wikidata.org/entity/Q35432091",
  "encodingNote": "PRONOM",
  "Converted": false
}, {
  "Signature": "63616666",
  "Encoding": "http://www.wikidata.org/entity/Q82828",
  "encodingNote": "Hexadecimal",
  "Converted": false
}, {
  "Signature": "424D{4}00000000{4}6C000000{8}0100(01|04|08|10|18|20)00(00|01|02|03)00000000",
  "Encoding": "http://www.wikidata.org/entity/Q35432091",
  "encodingNote": "PRONOM",
  "Converted": false
}, {
  "Signature": "GIF89a",
  "Encoding": "http://www.wikidata.org/entity/Q8815",
  "encodingNote": "ASCII",
  "NewEncoding": "http://www.wikidata.org/entity/Q82828",
  "newEncodingNote": "Hexadecimal",
  "NewSignature": "474946383961",
  "Converted": true
}, {
  "Signature": "BLENDER_",
  "Encoding": "http://www.wikidata.org/entity/Q8815",
  "encodingNote": "ASCII",
  "NewEncoding": "http://www.wikidata.org/entity/Q82828",
  "newEncodingNote": "Hexadecimal",
  "NewSignature": "424C454E4445525F",
  "Converted": true
}, {
  "Signature": "ý7zXZ",
  "Encoding": "http://www.wikidata.org/entity/Q8815",
  "encodingNote": "ASCII",
  "NewEncoding": "http://www.wikidata.org/entity/Q82828",
  "newEncodingNote": "Hexadecimal",
  "NewSignature": "C3BD377A585A",
  "Converted": true
}, {
  "Signature": "FD 37 7A 58 5A 00",
  "Encoding": "http://www.wikidata.org/entity/Q82828",
  "encodingNote": "Hexadecimal",
  "NewEncoding": "http://www.wikidata.org/entity/Q82828",
  "newEncodingNote": "Hexadecimal",
  "NewSignature": "FD377A585A00",
  "Comment": "Semantics! Hexadecimal sequences with spaces aren't converted per se, but are normalized.",
  "Converted": false
}, {
  "Signature": "325E1010",
  "Encoding": "",
  "NewEncoding": "http://www.wikidata.org/entity/Q82828",
  "newEncodingNote": "Hexadecimal",
  "NewSignature": "325E1010",
  "Converted": true
}, {
  "Signature": "B297E169",
  "Encoding": "",
  "NewEncoding": "http://www.wikidata.org/entity/Q82828",
  "newEncodingNote": "Hexadecimal",
  "NewSignature": "B297E169",
  "Converted": true
}, {
  "Signature": "RIFF.{4}WEBPVP8\\x20",
  "Encoding": "http://www.wikidata.org/entity/Q98056596",
  "encodingNote": "PERL",
  "NewEncoding": "http://www.wikidata.org/entity/Q35432091",
  "newEncodingNote": "PRONOM",
  "NewSignature": "52494646{4}5745425056503820",
  "Converted": true
}, {
  "Signature": "RIFF.{4}WEBPVP8L",
  "Encoding": "http://www.wikidata.org/entity/Q98056596",
  "encodingNote": "PERL",
  "NewEncoding": "http://www.wikidata.org/entity/Q35432091",
  "newEncodingNote": "PRONOM",
  "NewSignature": "52494646{4}574542505650384C",
  "Converted": true
}, {
  "Signature": "RIFF.{4}WEBPVP8X",
  "Encoding": "http://www.wikidata.org/entity/Q98056596",
  "encodingNote": "PERL",
  "NewEncoding": "http://www.wikidata.org/entity/Q35432091",
  "newEncodingNote": "PRONOM",
  "NewSignature": "52494646{4}5745425056503858",
  "Converted": true
}, {
  "Signature": "RIFF.{4}WEBP",
  "Encoding": "http://www.wikidata.org/entity/Q98056596",
  "encodingNote": "PERL",
  "NewEncoding": "http://www.wikidata.org/entity/Q35432091",
  "newEncodingNote": "PRONOM",
  "NewSignature": "52494646{4}57454250",
  "Converted": true
}, {
  "Signature": "00 61 73 6d",
  "Encoding": "",
  "NewEncoding": "http://www.wikidata.org/entity/Q82828",
  "newEncodingNote": "Hexadecimal",
  "NewSignature": "0061736D",
  "Converted": true
}, {
  "Signature": "badf00d1",
  "Encoding": "Nonsense encoding...",
  "NewEncoding": "http://www.wikidata.org/entity/Q82828",
  "newEncodingNote": "Hexadecimal",
  "NewSignature": "BADF00D1",
  "Converted": true
}, {
  "Signature": "00021401-0000-0000-c000-000000000046",
  "Encoding": "http://www.wikidata.org/entity/Q254972",
  "encodingNote": "GUID",
  "Fail": true,
  "Converted": false
}]`
