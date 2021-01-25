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

// Pre-process and process Wikidata signatures and enable the return of
// linting information that describe whether the information retrieved
// from the service can be processed correctly.

// WIKIDATA TODO: preValidateSignatures, updateSequences are two
// functions in need of some decent testing because they direct the
// logic of how we build the identifier. They are also responsible for
// making sure we get the bulk of the linting messages out of this
// package. The more accurate and precise we can make these functions
// (and probably a handful of others in this file) the better we can
// make the Wikidata identifier as well as the Wikidata sources.

package wikidata

import (
	"github.com/richardlehane/siegfried/pkg/wikidata/internal/converter"
	"github.com/richardlehane/siegfried/pkg/wikidata/internal/mappings"

	"github.com/ross-spencer/wikiprov/pkg/spargo"
)

// ByteSequence provides an alias for the mappings.ByteSequence object.
type ByteSequence = mappings.ByteSequence

// handleLinting ensures that our sequence error arrays are added to
// following the validation of the information.
func handleLinting(uri string, lint linting) {
	if lint != nle {
		addLinting(uri, lint)
	}
}

// newSignature will parse signature information from the Spargo Item
// structure and create a new Signature structure to be returned. If
// there is an error we log it out with the format identifier so that
// more work can be done on the source data.
func newByteSequence(wikidataItem map[string]spargo.Item) ByteSequence {

	tmpSequence := ByteSequence{}

	uri := wikidataItem[uriField].Value

	// Add relativity to sequence.
	relativity, lint, _ := validateAndReturnRelativity(
		wikidataItem[relativityField].Value)
	handleLinting(uri, lint)
	tmpSequence.Relativity = relativity

	// Add offset to sequence.
	offset, lint := validateAndReturnOffset(
		wikidataItem[offsetField].Value, wikidataItem[offsetField].Type)
	handleLinting(uri, lint)
	tmpSequence.Offset = offset

	// Add encoding to sequence.
	encoding, lint := validateAndReturnEncoding(
		wikidataItem[encodingField].Value)
	handleLinting(uri, lint)
	tmpSequence.Encoding = encoding

	// Add the signature to the sequence.
	signature, lint, _ := validateAndReturnSignature(
		wikidataItem[signatureField].Value, encoding)
	handleLinting(uri, lint)
	tmpSequence.Signature = signature

	return tmpSequence
}

// updateSignatures will create a new ByteSequence and associate it
// with either an existing Signature or create a brand new Signature.
// If there is a problem processing that means the sequence shouldn't be
// added to the identifier for the sake of consistency then a linting
// error is returned and we should stop processing.
func updateSequences(wikidataItem map[string]spargo.Item, wd *wikidataRecord) linting {

	// Pre-process the encoding.
	encoding, lint := validateAndReturnEncoding(
		wikidataItem[encodingField].Value)
	handleLinting(wd.URI, lint)

	// Pre-process the relativity.
	relativity, lint, _ := validateAndReturnRelativity(
		wikidataItem[relativityField].Value)
	handleLinting(wd.URI, lint)

	// Pre-process the sequence.
	signature, lint, _ := validateAndReturnSignature(
		wikidataItem[signatureField].Value, encoding)
	handleLinting(wd.URI, lint)

	// WIKIDATA FUTURE it's nearly impossible to tease apart sequences
	// in Wikidata right now to determine which duplicate sequences are
	// new signatures or which belong to the same group. Provenance
	// could differ but three can be multiple provenances, different
	// sequences which they're returned from the service, etc.
	if !sequenceInSignatures(wd.Signatures, signature) {
		if relativityAlreadyInSignatures(wd.Signatures, relativity) {
			if relativity == relativeBOF {
				// Create a new record...
				sig := Signature{}
				bs := newByteSequence(wikidataItem)
				sig.ByteSequences = append(sig.ByteSequences, bs)
				prov, lint := validateAndReturnProvenance(wikidataItem[referenceField].Value)
				handleLinting(wd.URI, lint)
				sig.Source = parseProvenance(prov)
				sig.Date, lint = validateAndReturnDate(wikidataItem[dateField].Value)
				handleLinting(wd.URI, lint)
				wd.Signatures = append(wd.Signatures, sig)
				return nle
			}
			// We've a bad heuristic and can't piece together a
			// valid signature.
			return heuWDE01
		}
		// Append to record...
		idx := len(wd.Signatures)
		sig := &wd.Signatures[idx-1]
		if checkEncodingCompatibility(wd.Signatures[idx-1], encoding) {
			bs := newByteSequence(wikidataItem)
			sig.ByteSequences = append(sig.ByteSequences, bs)
			return nle
		}
		// We've a bad heuristic and can't piece together a
		// valid signature.
		return heuWDE01
	}
	// Sequence already in signatures, no need to process, no errors of
	// note.
	return nle
}

// sequenceInSignatures will tell us if there are any duplicate byte
// sequences. At which point we can stop processing.
func sequenceInSignatures(signatures []Signature, signature string) bool {
	for _, sig := range signatures {
		for _, seq := range sig.ByteSequences {
			if signature == seq.Signature {
				return true
			}
		}
	}
	return false
}

// relativityInSlice helps us to identify if the needle: relativity is
// in the slice which we use for validation.
func relativityAlreadyInSignatures(signatures []Signature, relativity string) bool {
	for _, sig := range signatures {
		for _, seq := range sig.ByteSequences {
			if relativity == seq.Relativity {
				return true
			}
		}
	}
	return false
}

// checkEncodingCompatibility should work for now and just makes sure
// we're not trying to combine encodings that don't match, i.e. anything
// not PRONOM or HEX. ASCII should work too because we'll have encoded
// it as hex by now 🤞.
func checkEncodingCompatibility(signature Signature, givenEncoding int) bool {
	for _, seq := range signature.ByteSequences {
		if (seq.Encoding == converter.GUIDEncoding &&
			givenEncoding != converter.GUIDEncoding) ||
			(seq.Encoding == converter.PerlEncoding &&
				givenEncoding != converter.PerlEncoding) {
			return false
		}
	}
	return true
}

// preValidateSignatures performs some rudimentary validation of the
// sequences belonging to a Wikidata record. The sequences are stepped
// through as logically as possible to provide a sensible filter
// heuristic.
func preValidateSignatures(preProcessedSequences []preProcessedSequence) bool {

	// Map our values into slices to analyze cross-sectionally.
	var encoding []string
	var relativity []string
	var offset []string
	var signature []string
	for _, value := range preProcessedSequences {
		encoding = append(encoding, value.encoding)
		if value.relativity != "" {
			relativity = append(relativity, value.relativity)
		}
		offset = append(offset, value.offset)
		signature = append(signature, value.signature)
		_, _, err := validateAndReturnRelativity(value.relativity)
		if err != nil {
			return false
		}
		_, _, err = validateAndReturnSignature(
			value.signature, converter.LookupEncoding(value.encoding))
		if err != nil {
			return false
		}
	}
	// Maps act like sets when we're only interested in the keys. We
	// want to use sets to understand more about the unique values in
	// each of the records.
	var relativityMap = make(map[string]bool)
	var signatureMap = make(map[string]bool)
	var encodingMap = make(map[string]bool)
	for _, value := range signature {
		signatureMap[value] = true
	}
	for _, value := range relativity {
		relativityMap[value] = true
	}
	for _, value := range encoding {
		encodingMap[value] = true
	}
	if len(preProcessedSequences) == 2 {
		// The most simple validation we can do. If both we have two
		// values and two different relativities we can let the
		// signature through.
		if len(relativityMap) == 2 {
			return true
		}
		// If the relativities don't differ or aren't available then we
		// can then check to see if the signatures are different
		// because we will create two new records the the sequences.
		// They will both be beginning of file sequences.
		if len(signatureMap) == 2 {
			return true
		}
	}
	// We are going to start wrestling with a sensible heuristic with
	// sequences over 2 in length. Validate those.
	if len(preProcessedSequences) > 2 {
		// Processing starts to get too complicated if we have to work
		// out whether multiple encodings are valid when combined.
		if len(encodingMap) != 1 && len(encodingMap) != 0 {
			return false
		}
		// If we haven't a uniform relativity then we can't easily
		// guess how to combine signatures, e.g. how do we pair a single
		// EOF with one of three BOF sequences? Albeit an unlikely
		// scenario. but also, What if the EOF was not meant to be
		// paired?
		if len(relativityMap) != 1 && len(relativityMap) != 0 {
			return false
		}

	}
	// We should have enough information in these records to be able to
	// write a signature that is reliable.
	if len(signature) == len(encoding) && len(offset) == len(signature) {
		if len(relativity) == 0 || len(relativity) == len(signature) {
			return true
		}

	}
	// Anything else, we can't guarantee enough about the sequences to
	// write a signature. We may still have issues with the one's we've
	// pre-processed even, but we can give ourselves a chance.
	return false
}

// addSignatures tells us whether a signature can be added to the
// wikidata identifier after some level of validation.
func addSignatures(wikidataItems []map[string]spargo.Item, id string) bool {
	var preProcessedSequences []preProcessedSequence
	for _, wikidataItem := range wikidataItems {
		if getID(wikidataItem[uriField].Value) == id {
			if wikidataItem[signatureField].Value != "" {
				preProcessed := preProcessedSequence{}
				preProcessed.signature = wikidataItem[signatureField].Value
				preProcessed.offset = wikidataItem[offsetField].Value
				preProcessed.encoding = wikidataItem[encodingField].Value
				preProcessed.relativity = wikidataItem[relativityField].Value
				if len(preProcessedSequences) == 0 {
					preProcessedSequences =
						append(preProcessedSequences, preProcessed)
				}
				found := false
				for _, value := range preProcessedSequences {
					if preProcessed == value {
						found = true
						break
					}
				}
				if !found {
					preProcessedSequences =
						append(preProcessedSequences, preProcessed)
				}
			}
		}
	}
	var add bool
	if len(preProcessedSequences) > 0 {
		add = preValidateSignatures(preProcessedSequences)
	}
	return add
}
