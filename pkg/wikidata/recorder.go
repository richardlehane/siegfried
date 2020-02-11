package wikidata

import (
	"fmt"
	"os"
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

// Active comment...
func (r *Recorder) Active(m core.MatcherType) {
	// WIKIDATA TODO: Active isn't implemented...
}

// Record will build possible results sets associated with an identification.
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

func add(matches matchIDs,
	id string,
	wddID string,
	info formatInfo,
	basis string,
	source string,
	confidence int,
) matchIDs {
	for i, v := range matches {
		// WIKIDATA TODO: This function is looping too much, we need to find
		// out why. Given no extension whatsoever, this will loop through
		// all identifiers I believe...At least for now.
		//
		// fmt.Fprintf(os.Stderr, "LOOPING: %+v\n", v)
		//
		if v.ID == wddID {
			matches[i].confidence += confidence
			matches[i].Basis = append(matches[i].Basis, basis)
			matches[i].Source = append(matches[i].Source, source)
			return matches
		}
	}
	if config.GetWikidataSourceField() {
		return append(
			matches, Identification{
				Namespace:  id,
				ID:         wddID,
				Name:       info.name,
				LongName:   info.uri,
				MIME:       info.mime,
				Basis:      []string{basis},
				Source:     []string{source},
				Warning:    "",
				archive:    config.IsArchive(wddID),
				confidence: confidence,
			})
	}
	return append(
		matches, Identification{
			Namespace:  id,
			ID:         wddID,
			Name:       info.name,
			LongName:   info.uri,
			MIME:       info.mime,
			Basis:      []string{basis},
			Warning:    "",
			archive:    config.IsArchive(wddID),
			confidence: confidence,
		})
}

func recordNameMatcher(recorder *Recorder, matcher core.MatcherType, result core.Result) bool {
	if hit, id := recorder.Hit(matcher, result.Index()); hit {
		recorder.ids = add(
			recorder.ids,
			recorder.Name(),
			id,
			recorder.infos[id],
			result.Basis(),
			"",
			extScore,
		)
		return true
	}
	return false
}

// WIKIDATA TODO: I wonder how much recordByteMatcher and containerByteMatcher
// can be compressed, or wrapped to call one or the other.
func recordByteMatcher(recorder *Recorder, matcher core.MatcherType, result core.Result) bool {
	if hit, id := recorder.Hit(matcher, result.Index()); !hit {
		return false
	} else {
		if recorder.satisfied {
			// WIKIDATA TODO: This is never set for this identifier yet.
			return true
		}
		recorder.cscore += incScore
		basis := result.Basis()

		position, total := recorder.Place(core.ByteMatcher, result.Index())

		source := ""
		if position-1 >= len(recorder.infos[id].sources) {
			// WIKIDATA TODO: It will be possible to remove this IF once
			// sources work.
			fmt.Fprintf(
				os.Stderr,
				"WARNING: Problem with identifier index position: '%d' source length: '%d' info: %s\n",
				position-1,
				len(recorder.infos[id].sources),
				recorder.infos[id],
			)
		} else {
			source = fmt.Sprintf(
				"%s", recorder.infos[id].sources[position-1],
			)
		}
		if total > 1 {
			basis = fmt.Sprintf("%s (signature %d/%d)", basis, position, total)
		} else {
			if source != "" {
				basis = fmt.Sprintf("%s", basis)
			}
		}
		// WIKIDATA TODO: sourceField toggle exists for demonstration purposes
		// so we need to remove it once we've settled on an approach.
		if !config.GetWikidataSourceField() {
			basis = fmt.Sprintf("%s (%s)", basis, source)
			source = ""
		}
		recorder.ids = add(
			recorder.ids,
			recorder.Name(),
			id,
			recorder.infos[id],
			basis,
			source,
			recorder.cscore,
		)
	}
	return true
}

// recordContainerMatcher comes from PRONOM.
//
// WIKIDATA TODO: Could this be exported from PRONOM and called directly?
//
func recordContainerMatcher(recorder *Recorder, matcher core.MatcherType, result core.Result) bool {
	if result.Index() < 0 {
		if recorder.ZipDefault() {
			recorder.cscore += incScore
			recorder.ids = add(
				recorder.ids,
				recorder.Name(),
				config.ZipPuid(), // WIKIDATA TODO: Add to pkg/config/wikidata.go.
				recorder.infos[config.ZipPuid()],
				result.Basis(),
				"",
				recorder.cscore,
			)
		}
		return false
	}
	if hit, id := recorder.Hit(matcher, result.Index()); hit {
		recorder.cscore += incScore
		basis := result.Basis()
		position, total := recorder.Place(core.ContainerMatcher, result.Index())

		source := ""
		if position-1 >= len(recorder.infos[id].sources) {
			// WIKIDATA TODO: WARNING the container provenance component isn't
			// working as anticipated. We need to figure this out at the source
			// so to speak, pun partially intended.
			source = pronomOfficialContainer
		} else {
			source = fmt.Sprintf(
				"%s", recorder.infos[id].sources[position-1],
			)
		}
		if total > 1 {
			basis = fmt.Sprintf("%s (signature %d/%d)", basis, position, total)
		} else {
			if source != "" {
				basis = fmt.Sprintf("%s", basis)
			}
		}
		// WIKIDATA TODO: sourceField toggle exists for demonstration purposes
		// so we need to remove it once we've settled on an approach.
		if !config.GetWikidataSourceField() {
			basis = fmt.Sprintf("%s (%s)", basis, source)
			source = ""
		}
		recorder.ids = add(
			recorder.ids,
			recorder.Name(),
			id,
			recorder.infos[id],
			basis,
			source,
			recorder.cscore,
		)
		return true
	}
	return false
}

// Satisfied is drawn from the PRONOM identifier and tells us whether or not
// we should continue with any particular matcher...
func (r *Recorder) Satisfied(mt core.MatcherType) (bool, core.Hint) {
	if r.NoPriority() {
		return false, core.Hint{}
	}
	if r.cscore < incScore {
		if len(r.ids) == 0 {
			return false, core.Hint{}
		}
		if mt == core.ContainerMatcher ||
			mt == core.ByteMatcher ||
			mt == core.XMLMatcher ||
			mt == core.RIFFMatcher {
			if mt == core.ByteMatcher ||
				mt == core.ContainerMatcher {
				keys := make([]string, len(r.ids))
				for i, v := range r.ids {
					keys[i] = v.String()
				}
				return false, core.Hint{r.Start(mt), r.Lookup(mt, keys)}
			}
			return false, core.Hint{}
		}
		for _, res := range r.ids {
			if res.ID == config.TextPuid() {
				return false, core.Hint{}
			}
		}
	}
	r.satisfied = true
	if mt == core.ByteMatcher {
		return true, core.Hint{r.Start(mt), nil}
	}
	return true, core.Hint{}
}

/* Report comment...

WIKIDATA TODO: Add each...

	Single
	Conclusive		<-- Default...
	Positive
	Comprehensive
	Exhaustive
*/
func (r *Recorder) Report() []core.Identification {
	// Happy path for zero results...
	if len(r.ids) == 0 {
		return []core.Identification{Identification{
			Namespace: r.Name(),
			ID:        "UNKNOWN",
			Warning:   "no match",
		}}
	}
	// Sort IDs by confidence to return highest first.
	sort.Sort(r.ids)
	confidence := r.ids[0].confidence
	ret := make([]core.Identification, len(r.ids))
	for i, v := range r.ids {
		if i > 0 {
			switch r.Multi() {
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

		ret[i] = v
	}
	return ret
}
