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

// Organizes the format identification results from the Wikidata
// package.

// WIKIDATA TODO: This part of an identifier is still somewhat
// unfamiliar to me so I need to spend a bit longer on it at some point.

package wikidata

import (
	"fmt"
	"sort"

	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

const (
	extScore = 1 << iota
	mimeScore
	textScore
	incScore
)

type matchIDs []Identification

// Len needed to satisfy the sort interface for sorting the slice during
// reporting.
func (matches matchIDs) Len() int { return len(matches) }

// Less needed to satisfy the sort interface for sorting the slice during
// reporting.
func (matches matchIDs) Less(i, j int) bool { return matches[j].confidence < matches[i].confidence }

// Swap needed to satisfy the sort interface for sorting the slice during
// reporting.
func (matches matchIDs) Swap(i, j int) { matches[i], matches[j] = matches[j], matches[i] }

// Recorder comment...
type Recorder struct {
	*Identifier
	ids        matchIDs
	cscore     int
	satisfied  bool
	extActive  bool
	mimeActive bool
	textActive bool
}

// Active tells the recorder what matchers are active which helps when
// providing a detailed response to the caller.
func (recorder *Recorder) Active(matcher core.MatcherType) {
	if recorder.Identifier.Active(matcher) {
		switch matcher {
		case core.NameMatcher:
			recorder.extActive = true
		case core.MIMEMatcher:
			recorder.mimeActive = true
		case core.TextMatcher:
			recorder.textActive = true
		}
	}
}

// Record will build possible results sets associated with an
// identification.
func (recorder *Recorder) Record(matcher core.MatcherType, result core.Result) bool {
	switch matcher {
	default:
		return false
	case core.NameMatcher:
		return recordNameMatcher(recorder, matcher, result)
	case core.ContainerMatcher:
		return recordContainerMatcher(recorder, matcher, result)
	case core.ByteMatcher:
		return recordByteMatcher(recorder, matcher, result)
	}
}

// add appends identifications to a matchIDs slice.
func add(matches matchIDs, id string, wikidataID string, info formatInfo, basis string, confidence int) matchIDs {
	for idx, match := range matches {
		// WIKIDATA TODO: This function is looping too much, especially
		// with extension matches which might point to a part of this
		// implementation running sub-optimally. Or it might be expected
		// of extension matches.
		//
		// Run with: fmt.Fprintf(os.Stderr, "LOOPING: %#v\n", v)
		//
		if match.ID == wikidataID {
			matches[idx].confidence += confidence
			matches[idx].Basis = append(matches[idx].Basis, basis)
			return matches
		}
	}
	return append(
		matches, Identification{
			Namespace:  id,
			ID:         wikidataID,
			Name:       info.name,
			LongName:   info.uri,
			Permalink:  info.permalink,
			MIME:       info.mime,
			Basis:      []string{basis},
			Warning:    "",
			archive:    config.IsArchive(wikidataID),
			confidence: confidence,
		})
}

// recordNameMatcher ...
func recordNameMatcher(recorder *Recorder, matcher core.MatcherType, result core.Result) bool {
	if hit, id := recorder.Hit(matcher, result.Index()); hit {
		recorder.ids = add(
			recorder.ids,
			recorder.Name(),
			id,
			recorder.infos[id],
			result.Basis(),
			extScore,
		)
		return true
	}
	return false
}

// recordByteMatcher ...
func recordByteMatcher(recorder *Recorder, matcher core.MatcherType, result core.Result) bool {
	var hit bool
	var id string
	if hit, id = recorder.Hit(matcher, result.Index()); !hit {
		return false
	}
	if recorder.satisfied {
		// This is never set for this identifier yet which might be
		// related to the issue we're seeing in add(...)
		return true
	}
	recorder.cscore += incScore
	basis := result.Basis()
	position, total := recorder.Place(core.ByteMatcher, result.Index())
	// Depending on how defensive we are being, we might check:
	//
	//    position-1 >= len(recorder.infos[id].sources)
	//
	// where identifiers and "source" slices need to align 1:1 to output
	// correctly. See: richardlehane/siegfried#142
	source := fmt.Sprintf("%s", recorder.infos[id].sources[position-1])
	if total > 1 {
		basis = fmt.Sprintf(
			"%s (signature %d/%d)", basis, position, total,
		)
	} else {
		if source != "" {
			basis = fmt.Sprintf("%s", basis)
		}
	}
	basis = fmt.Sprintf("%s (%s)", basis, source)
	recorder.ids = add(
		recorder.ids,
		recorder.Name(),
		id,
		recorder.infos[id],
		basis,
		recorder.cscore,
	)
	return true
}

// recordContainerMatcher ...
func recordContainerMatcher(recorder *Recorder, matcher core.MatcherType, result core.Result) bool {
	if result.Index() < 0 {
		if recorder.ZipDefault() {
			recorder.cscore += incScore
			recorder.ids = add(
				recorder.ids,
				recorder.Name(),
				config.ZipPuid(),
				recorder.infos[config.ZipPuid()],
				result.Basis(),
				recorder.cscore,
			)
		}
		return false
	}
	if hit, id := recorder.Hit(matcher, result.Index()); hit {
		recorder.cscore += incScore
		basis := result.Basis()
		position, total := recorder.Place(
			core.ContainerMatcher, result.Index(),
		)

		source := ""
		if position-1 >= len(recorder.infos[id].sources) {
			// Container provenance isn't working as anticipated in the
			// Wikidata identifier yet. We use a placeholder here in
			// their place.
			source = pronomOfficialContainer
		} else {
			source = fmt.Sprintf(
				"%s", recorder.infos[id].sources[position-1],
			)
		}
		if total > 1 {
			basis = fmt.Sprintf(
				"%s (signature %d/%d)", basis, position, total,
			)
		} else {
			if source != "" {
				basis = fmt.Sprintf("%s", basis)
			}
		}
		basis = fmt.Sprintf("%s (%s)", basis, source)
		recorder.ids = add(
			recorder.ids,
			recorder.Name(),
			id,
			recorder.infos[id],
			basis,
			recorder.cscore,
		)
		return true
	}
	return false
}

// Satisfied is drawn from the PRONOM identifier and tells us whether or not
// we should continue with any particular matcher...
func (recorder *Recorder) Satisfied(mt core.MatcherType) (bool, core.Hint) {
	if recorder.NoPriority() {
		return false, core.Hint{}
	}
	if recorder.cscore < incScore {
		if len(recorder.ids) == 0 {
			return false, core.Hint{}
		}
		if mt == core.ContainerMatcher ||
			mt == core.ByteMatcher ||
			mt == core.XMLMatcher ||
			mt == core.RIFFMatcher {
			if mt == core.ByteMatcher ||
				mt == core.ContainerMatcher {
				keys := make([]string, len(recorder.ids))
				for i, v := range recorder.ids {
					keys[i] = v.String()
				}
				return false, core.Hint{recorder.Start(mt), recorder.Lookup(mt, keys)}
			}
			return false, core.Hint{}
		}
		for _, res := range recorder.ids {
			if res.ID == config.TextPuid() {
				return false, core.Hint{}
			}
		}
	}
	recorder.satisfied = true
	if mt == core.ByteMatcher {
		return true, core.Hint{recorder.Start(mt), nil}
	}
	return true, core.Hint{}
}

// Report organizes the identification output so that the highest
// priority results are output first.
func (recorder *Recorder) Report() []core.Identification {
	// Happy path for zero results...
	if len(recorder.ids) == 0 {
		return []core.Identification{Identification{
			Namespace: recorder.Name(),
			ID:        "UNKNOWN",
			Warning:   "no match",
		}}
	}
	// Sort IDs by confidence to return highest first.
	sort.Sort(recorder.ids)
	confidence := recorder.ids[0].confidence
	ret := make([]core.Identification, len(recorder.ids))
	for i, v := range recorder.ids {
		if i > 0 {
			switch recorder.Multi() {
			case config.Single:
				return ret[:i]
			case config.Conclusive:
				if v.confidence < confidence {
					return ret[:i]
				}
			default:
				if v.confidence < incScore {
					return ret[:i]
				}
			}
		}
		ret[i] = recorder.updateWarning(v)
	}
	return ret
}

// updateWarning is used to add precision to the identification. A
// classic example is the application of an extension mismatch to an
// identification when the binary match works but the extension is
// known to be in all likelihood incorrect.
func (recorder *Recorder) updateWarning(identification Identification) Identification {
	// Apply mismatches to the identification result.
	if recorder.extActive && (identification.confidence&extScore != extScore) {
		for _, v := range recorder.IDs(core.NameMatcher) {
			if identification.ID == v {
				if len(identification.Warning) > 0 {
					identification.Warning += "; extension mismatch"
				} else {
					identification.Warning = "extension mismatch"
				}
				break
			}
		}
	}
	return identification
}
