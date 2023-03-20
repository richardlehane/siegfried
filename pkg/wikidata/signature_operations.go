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

// Satisfies the Parseable interface to enable Roy to process Wikidata
// signatures into a Siegfried compatible identifier.

package wikidata

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/pkg/wikidata/internal/mappings"

	"github.com/richardlehane/siegfried/pkg/pronom"
)

// Globs match based on some pattern in the filename of a file. For
// Wikidata this means we'll use the extensions returned by the service
// to match formats by that.
func (wdd wikidataDefinitions) Globs() ([]string, []string) {
	logln(
		"Roy (Wikidata): Adding Glob signatures to identifier...",
	)
	globs, ids := make(
		[]string, 0, len(wdd.formats)),
		make([]string, 0, len(wdd.formats))
	for _, v := range wdd.formats {
		for _, w := range v.Extension {
			globs, ids = append(globs, "*."+w), append(ids, v.ID)
		}
	}
	return globs, ids
}

// Signatures maps our standard non-container binary signatures into the
// Wikidata identifier.
func (wdd wikidataDefinitions) Signatures() ([]frames.Signature, []string, error) {
	frames, ids, err := processSIgnatures(wdd)
	return frames, ids, err
}

// collectPUIDs identifies the PUIDs we have at our disposal in the
// Wikidata report and collects them into a map to be processed into
// the identifier to augment the identifiers capabilities with PRONOM
// binary signatures.
func collectPUIDs(puidsIDs map[string][]string, v mappings.Wikidata) map[string][]string {
	if puidsIDs != nil {
		for _, puid := range v.PUIDs() {
			puidsIDs[puid] = append(puidsIDs[puid], v.ID)
		}
	}
	return puidsIDs
}

// byteSequences provides an alias for the mappings ByteSequence object.
type byteSequences = []mappings.ByteSequence

// pronomSequence provides an alias for the PRONOM compatibility object.
type pronomSequence = pronom.PROCompatSequence

// processForPronom maps Wikidata byte sequences into a slice that can
// be processed through the PRONOM identifier which is enabled by the
// PRONOM compatibility sequence, PROCompatSequence.
func processForPronom(bs byteSequences) []pronomSequence {
	var pronomSlice []pronomSequence
	for _, b := range bs {
		ps := pronomSequence{}
		switch b.Relativity {
		case relativeBOF:
			ps.Position = pronom.BeginningOfFile
		case relativeEOF:
			ps.Position = pronom.EndOfFile
		default:
			// We might otherwise return an error. I don't think there
			// is a high risk with the pre-processing work we do for the
			// identifier, and other errors will be caught below.
		}
		ps.Hex = b.Signature
		ps.Offset = strconv.Itoa(b.Offset)
		pronomSlice = append(pronomSlice, ps)
	}
	return pronomSlice
}

// processSignatures processes the Wikidata signatures into an
// identifier and returns a slice of Signature frames, IDs, and errors
// collected along the way.
func processSIgnatures(wdd wikidataDefinitions) ([]frames.Signature, []string, error) {
	logln(
		"Roy (Wikidata): Adding Wikidata Byte signatures to identifier...",
	)
	var errs []error
	var puidsIDs map[string][]string
	if len(wdd.parseable.IDs()) > 0 {
		puidsIDs = make(map[string][]string)
	}
	sigs := make([]frames.Signature, 0, len(wdd.formats))
	ids := make([]string, 0, len(wdd.formats))
	for _, wd := range wdd.formats {
		puidsIDs = collectPUIDs(puidsIDs, wd)
		for _, v := range wd.Signatures {
			ps := processForPronom(v.ByteSequences)
			frames, err := pronom.FormatPRONOM(wd.ID, ps)
			if err != nil {
				errs = append(errs, err)
			}
			sigs = append(sigs, frames)
			ids = append(ids, wd.ID)
		}
	}
	// Add PRONOM into the mix.
	if puidsIDs != nil {
		puids := make([]string, 0, len(puidsIDs))
		for p := range puidsIDs {
			puids = append(puids, p)
		}
		newParseable := identifier.Filter(puids, wdd.parseable)
		pronomSignatures, pronomIdentifiers, err :=
			newParseable.Signatures()
		if err != nil {
			errs = append(errs, err)
		}
		for i, v := range pronomIdentifiers {
			for _, id := range puidsIDs[v] {
				sigs = append(sigs, pronomSignatures[i])
				ids = append(ids, id)
			}
		}
	}
	var err error
	if len(errs) > 0 {
		errStrs := make([]string, len(errs))
		for i, e := range errs {
			errStrs[i] = e.Error()
		}
		err = errors.New(strings.Join(errStrs[:], "; "))
	}
	return sigs, ids, err
}

// Zips adds ZIP based container signatures to the identifier.
func (wdd wikidataDefinitions) Zips() ([][]string, [][]frames.Signature, []string, error) {
	return wdd.containers("ZIP")
}

// MSCFBs adds OLE2 based container signatures to the identifier.
func (wdd wikidataDefinitions) MSCFBs() ([][]string, [][]frames.Signature, []string, error) {
	return wdd.containers("OLE2")
}

// Wikidata doesn't have its own concept of container format
// identification just yet and so we do this via PRONOM's in-build
// methods. This mimics that of the Library of Congress identifier.
// Wikidata container modeling is in-progress.
func (wdd wikidataDefinitions) containers(typ string) ([][]string, [][]frames.Signature, []string, error) {
	logln(
		"Roy (Wikidata): Adding container signatures to identifier...",
	)
	if _, ok := wdd.parseable.(identifier.Blank); ok {
		return nil, nil, nil, nil
	}
	puidsIDs := make(map[string][]string)
	for _, v := range wdd.formats {
		for _, puid := range v.PUIDs() {
			puidsIDs[puid] = append(puidsIDs[puid], v.ID)
		}
	}
	puids := make([]string, 0, len(puidsIDs))
	for p := range puidsIDs {
		puids = append(puids, p)
	}

	np := identifier.Filter(puids, wdd.parseable)

	names, sigs, ids :=
		make([][]string, 0, len(wdd.formats)),
		make([][]frames.Signature, 0, len(wdd.formats)),
		make([]string, 0, len(wdd.formats))
	var (
		ns  [][]string
		ss  [][]frames.Signature
		is  []string
		err error
	)
	switch typ {
	default:
		err = fmt.Errorf("Unknown container type: %s", typ)
	case "ZIP":
		ns, ss, is, err = np.Zips()
	case "OLE2":
		ns, ss, is, err = np.MSCFBs()
	}
	if err != nil {
		return nil, nil, nil, err
	}
	for i, puid := range is {
		for _, id := range puidsIDs[puid] {
			names = append(names, ns[i])
			sigs = append(sigs, ss[i])
			ids = append(ids, id)
		}
	}
	return names, sigs, ids, nil
}
