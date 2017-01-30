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

package checksum

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"hash/crc32"
)

const HashChoices = "'md5', 'sha1', 'sha256', 'sha512', 'crc'"

type HashTyp int

const (
	md5Hash HashTyp = iota
	sha1Hash
	sha256Hash
	sha512Hash
	crcHash
)

func GetHash(typ string) HashTyp {
	switch typ {
	case "", "false":
	case "md5", "MD5":
		return md5Hash
	case "sha1", "SHA1":
		return sha1Hash
	case "sha256", "SHA256":
		return sha256Hash
	case "sha512", "SHA512":
		return sha512Hash
	case "crc", "CRC":
		return crcHash
	}
	return -1
}

func MakeHash(typ HashTyp) hash.Hash {
	switch typ {
	case md5Hash:
		return md5.New()
	case sha1Hash:
		return sha1.New()
	case sha256Hash:
		return sha256.New()
	case sha512Hash:
		return sha512.New()
	case crcHash:
		return crc32.NewIEEE()
	}
	return nil
}

func (typ HashTyp) String() string {
	switch typ {
	case md5Hash:
		return "md5"
	case sha1Hash:
		return "sha1"
	case sha256Hash:
		return "sha256"
	case sha512Hash:
		return "sha512"
	case crcHash:
		return "crc"
	}
	return ""
}
