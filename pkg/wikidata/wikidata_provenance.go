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

// Helper functions for cleaning up provenance "source" information in
// the Wikidata identifier.

package wikidata

import ()

// Source strings from Wikidata. We may have a use-case to normalize
// them, but also, we can do this via ShEx on a long-enough time-line.
const (
	// Wikidata source strings are those returned by Wikidata
	// specifically.
	prePronom  = "PRONOM"
	preKessler = "Gary Kessler's File Signature Table"

	// Normalized source strings are those that we want to return from
	// the Wikidata identifier to the user so that they can be parsed
	// consistently by the consumer.
	tnaPronom = "PRONOM (TNA)"
	wdPronom  = "PRONOM (Wikidata)"
	wdNone    = "Wikidata reference is empty"

	// Provenance to include in source information when the PRONOM
	// signatures are being used to compliment those in the Wikidata
	// identifier.
	pronomOfficial          = "PRONOM (Official (%s))"
	pronomOfficialContainer = "PRONOM (Official container ID)"
)

// parseProvenance normalizes the provenance string and let's us  know
// if the value is in-fact empty.
func parseProvenance(prov string) string {
	if prov == prePronom {
		prov = wdPronom
	}
	if prov == "" {
		prov = wdNone
	}
	return prov
}
