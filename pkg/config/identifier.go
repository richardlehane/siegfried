// Copyright 2014 Richard Lehane. All rights reserved.
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

package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Name of the default identifier as well as settings for how a new identifer will be built
var identifier = struct {
	name        string   // Name of the default identifier
	details     string   // a short string describing the signature e.g. with what DROID and container file versions was it built?
	maxBOF      int      // maximum offset from beginning of file to scan
	maxEOF      int      // maximum offset from end of file to scan
	noEOF       bool     // trim end of file segments from signatures
	noByte      bool     // don't build with byte signatures
	noContainer bool     // don't build with container signatures
	multi       Multi    // define how many results identifiers should return
	noText      bool     // don't build with text signatures
	noName      bool     // don't build with filename signatures
	noMIME      bool     // don't build with MIME signatures
	noXML       bool     // don't build with XML signatures
	noRIFF      bool     // don't build with RIFF signatures
	limit       []string // limit signature to a set of included PRONOM reports
	exclude     []string // exclude a set of PRONOM reports from the signature
	extensions  string   // directory where custom signature extensions are stored
	extend      []string // list of custom signature extensions
	verbose     bool     // verbose output when building signatures
}{
	multi:      Conclusive,
	extensions: "custom",
}

// GETTERS
const emptyNamespace = ""

// Name returns the name of the identifier.
func Name() string {
	switch {
	case identifier.name != emptyNamespace:
		return identifier.name
	case mimeinfo.mi != emptyNamespace:
		return mimeinfo.name
	case loc.fdd != emptyNamespace:
		return loc.name
	case GetWikidataNamespace() != emptyNamespace:
		return GetWikidataNamespace()
	default:
		return pronom.name
	}
}

// Details returns a description of the identifier. This is auto-populated if not set directly.
// Extra information from signatures such as date last modified can be given to this function.
func Details(extra ...string) string {
	// if the details string has been explicitly set, return it
	if len(identifier.details) > 0 {
		return identifier.details
	}
	// ... otherwise create a default string based on the identifier settings chosen
	var str string
	if len(mimeinfo.mi) > 0 {
		str = mimeinfo.mi
	} else if len(loc.fdd) > 0 {
		str = loc.fdd
		if !loc.nopronom {
			extra = append(extra, DroidBase())
			if !identifier.noContainer {
				extra = append(extra, ContainerBase())
			}
		}
	} else if wikidata.namespace != "" {
		str = wikidata.definitions
		if !wikidata.nopronom {
			extra = append(extra, DroidBase())
			if !identifier.noContainer {
				extra = append(extra, ContainerBase())
			}
		}
	} else {
		str = DroidBase()
		if !identifier.noContainer {
			str += "; " + ContainerBase()
		}
	}
	if len(extra) > 0 {
		str += " (" + strings.Join(extra, ", ") + ")"
	}
	if identifier.maxBOF > 0 {
		str += fmt.Sprintf("; max BOF %d", identifier.maxBOF)
	}
	if identifier.maxEOF > 0 {
		str += fmt.Sprintf("; max EOF %d", identifier.maxEOF)
	}
	if identifier.noEOF {
		str += "; no EOF signature parts"
	}
	if identifier.noByte {
		str += "; no byte signatures"
	}
	if identifier.noContainer {
		str += "; no container signatures"
	}
	if identifier.multi != Conclusive {
		str += "; multi set to " + identifier.multi.String()
	}
	if identifier.noText {
		str += "; no text matcher"
	}
	if identifier.noName {
		str += "; no filename matcher"
	}
	if identifier.noMIME {
		str += "; no MIME matcher"
	}
	if identifier.noXML {
		str += "; no XML matcher"
	}
	if identifier.noRIFF {
		str += "; no RIFF matcher"
	}
	if pronom.reports == "" {
		str += "; built without reports"
	}
	if pronom.doubleup {
		str += "; byte signatures included for formats that also have container signatures"
	}
	if HasLimit() {
		str += "; limited to ids: " + strings.Join(identifier.limit, ", ")
	}
	if HasExclude() {
		str += "; excluding ids: " + strings.Join(identifier.exclude, ", ")
	}
	if len(identifier.extend) > 0 {
		str += "; extensions: " + strings.Join(identifier.extend, ", ")
	}
	if len(pronom.extendc) > 0 {
		str += "; container extensions: " + strings.Join(pronom.extendc, ", ")
	}
	return str
}

// MaxBOF returns any BOF buffer limit set.
func MaxBOF() int {
	return identifier.maxBOF
}

// MaxEOF returns any EOF buffer limit set.
func MaxEOF() int {
	return identifier.maxEOF
}

// NoEOF reports whether end of file segments of signatures should be trimmed.
func NoEOF() bool {
	return identifier.noEOF
}

// NoByte reports whether byte signatures should be omitted.
func NoByte() bool {
	return identifier.noByte
}

// NoContainer reports whether container signatures should be omitted.
func NoContainer() bool {
	return identifier.noContainer
}

// NoPriority reports whether priorities between signatures should be omitted.
func NoPriority() bool {
	return identifier.multi >= Comprehensive
}

// GetMulti returns the multi setting
func GetMulti() Multi {
	return identifier.multi
}

// NoText reports whether text signatures should be omitted.
func NoText() bool {
	return identifier.noText
}

// NoName reports whether filename signatures should be omitted.
func NoName() bool {
	return identifier.noName
}

// NoMIME reports whether MIME signatures should be omitted.
func NoMIME() bool {
	return identifier.noMIME
}

// NoXML reports whether XML signatures should be omitted.
func NoXML() bool {
	return identifier.noXML
}

// NoRIFF reports whether RIFF FOURCC signatures should be omitted.
func NoRIFF() bool {
	return identifier.noRIFF
}

// HasLimit reports whether a limited set of signatures has been selected.
func HasLimit() bool {
	return len(identifier.limit) > 0
}

// Limit takes a slice of puids and returns a new slice containing only those puids in the limit set.
func Limit(ids []string) []string {
	ret := make([]string, 0, len(identifier.limit))
	for _, v := range identifier.limit {
		for _, w := range ids {
			if v == w {
				ret = append(ret, v)
			}
		}
	}
	return ret
}

// HasExclude reports whether an exlusion set of signatures has been provided.
func HasExclude() bool {
	return len(identifier.exclude) > 0
}

func exclude(ids, ex []string) []string {
	ret := make([]string, 0, len(ids))
	for _, v := range ids {
		excluded := false
		for _, w := range ex {
			if v == w {
				excluded = true
				break
			}
		}
		if !excluded {
			ret = append(ret, v)
		}
	}
	return ret
}

// Exclude takes a slice of puids and omits those that are also in the identifier.exclude slice.
func Exclude(ids []string) []string {
	return exclude(ids, identifier.exclude)
}

func extensionPaths(e []string) []string {
	ret := make([]string, len(e))
	for i, v := range e {
		if filepath.Dir(v) == "." {
			ret[i] = filepath.Join(siegfried.home, identifier.extensions, v)
		} else {
			ret[i] = v
		}
	}
	return ret
}

// Extend reports whether a set of signature extensions has been provided.
func Extend() []string {
	return extensionPaths(identifier.extend)
}

// Verbose reports whether to build signatures with verbose logging output
func Verbose() bool {
	return identifier.verbose
}

// Return true if value 'v' is contained in slice 's'.
func contains(v string, s []string) bool {
	for _, n := range s {
		if v == n {
			return true
		}
	}
	return false
}

// IsArchive returns an Archive that corresponds to the provided id (or none if no match).
func IsArchive(id string) Archive {
	if !contains(id, archiveFilterPermissive()) {
		return None
	}
	switch {
	case contains(id, ArcZipTypes()):
		return Zip
	case contains(id, ArcGzipTypes()):
		return Gzip
	case contains(id, ArcTarTypes()):
		return Tar
	case contains(id, ArcArcTypes()):
		return ARC
	case contains(id, ArcWarcTypes()):
		return WARC
	}
	return None
}

// SETTERS

// Clear clears loc and mimeinfo details to avoid pollution when creating multiple identifiers in same session
func Clear() func() private {
	return func() private {
		identifier.name = ""
		identifier.extend = nil
		identifier.limit = nil
		identifier.exclude = nil
		identifier.multi = Conclusive
		loc.fdd = ""
		mimeinfo.mi = ""
		return private{}
	}
}

// SetName sets the name of the identifier.
func SetName(n string) func() private {
	return func() private {
		identifier.name = n
		return private{}
	}
}

// SetDetails sets the identifier's description. If not provided, this description is
// automatically generated based on options set.
func SetDetails(d string) func() private {
	return func() private {
		identifier.details = d
		return private{}
	}
}

// SetBOF limits the number of bytes to scan from the beginning of file.
func SetBOF(b int) func() private {
	return func() private {
		identifier.maxBOF = b
		return private{}
	}
}

// SetEOF limits the number of bytes to scan from the end of file.
func SetEOF(e int) func() private {
	return func() private {
		identifier.maxEOF = e
		return private{}
	}
}

// SetNoEOF will cause end of file segments to be trimmed from signatures.
func SetNoEOF() func() private {
	return func() private {
		identifier.noEOF = true
		return private{}
	}
}

// SetNoByte will cause byte signatures to be omitted.
func SetNoByte() func() private {
	return func() private {
		identifier.noByte = true
		return private{}
	}
}

// SetNoContainer will cause container signatures to be omitted.
func SetNoContainer() func() private {
	return func() private {
		identifier.noContainer = true
		return private{}
	}
}

// SetMulti defines how identifiers report multiple results.
func SetMulti(m string) func() private {
	return func() private {
		switch m {
		case "0", "single", "top":
			identifier.multi = Single
		case "1", "conclusive":
			identifier.multi = Conclusive
		case "2", "positive":
			identifier.multi = Positive
		case "3", "comprehensive":
			identifier.multi = Comprehensive
		case "4", "exhaustive":
			identifier.multi = Exhaustive
		case "5", "droid":
			identifier.multi = DROID
		default:
			identifier.multi = Conclusive
		}
		return private{}
	}
}

// SetNoText will cause text signatures to be omitted.
func SetNoText() func() private {
	return func() private {
		identifier.noText = true
		return private{}
	}
}

// SetNoName will cause extension signatures to be omitted.
func SetNoName() func() private {
	return func() private {
		identifier.noName = true
		return private{}
	}
}

// SetNoMIME will cause MIME signatures to be omitted.
func SetNoMIME() func() private {
	return func() private {
		identifier.noMIME = true
		return private{}
	}
}

// SetNoXML will cause XML signatures to be omitted.
func SetNoXML() func() private {
	return func() private {
		identifier.noXML = true
		return private{}
	}
}

// SetNoRIFF will cause RIFF FOURCC signatures to be omitted.
func SetNoRIFF() func() private {
	return func() private {
		identifier.noRIFF = true
		return private{}
	}
}

// SetLimit limits the set of signatures built to the list provide.
func SetLimit(l []string) func() private {
	return func() private {
		identifier.limit = l
		return private{}
	}
}

// SetExclude excludes the provided signatures from those built.
func SetExclude(l []string) func() private {
	return func() private {
		identifier.exclude = l
		return private{}
	}
}

// SetExtend adds extension signatures to the build.
func SetExtend(l []string) func() private {
	return func() private {
		identifier.extend = l
		return private{}
	}
}

// SetVerbose controls logging verbosity when building signatures
func SetVerbose(v bool) func() private {
	return func() private {
		identifier.verbose = v
		return private{}
	}
}
