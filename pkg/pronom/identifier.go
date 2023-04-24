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

package pronom

import (
	"fmt"
	"sort"
	"strings"

	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

func init() {
	core.RegisterIdentifier(core.Pronom, Load)
}

// Identifier is the PRONOM implementation of the Identifier interface
// It wraps the base Identifier implementation with a formatinfo map
type Identifier struct {
	*identifier.Base      // embedded base identifier
	hasClass         bool // has this PRONOM identifier been built from reports & without the NoClass config option set?
	infos            map[string]formatInfo
	priorities       priority.Map
}

// Save persists the PRONOM Identifier to disk
func (i *Identifier) Save(ls *persist.LoadSaver) {
	ls.SaveByte(core.Pronom)
	i.Base.Save(ls)
	ls.SaveBool(i.hasClass)
	multi := i.Multi() == config.DROID
	ls.SaveSmallInt(len(i.infos))
	for k, v := range i.infos {
		ls.SaveString(k)
		ls.SaveString(v.name)
		ls.SaveString(v.version)
		ls.SaveString(v.mimeType)
		if i.hasClass {
			ls.SaveString(v.class)
		}
		if multi {
			ls.SaveStrings(i.priorities[k])
		}
	}

}

// Load reads PRONOM information back from the persist byte format into
// the Identifier's FormatInfo struct.
func Load(ls *persist.LoadSaver) core.Identifier {
	i := &Identifier{}
	i.Base = identifier.Load(ls)
	i.hasClass = ls.LoadBool()
	multi := i.Multi() == config.DROID
	if multi {
		i.priorities = make(priority.Map)
	}
	i.infos = make(map[string]formatInfo)
	le := ls.LoadSmallInt()
	if i.hasClass {
		for j := 0; j < le; j++ {
			k := ls.LoadString()
			i.infos[k] = formatInfo{
				name:     ls.LoadString(),
				version:  ls.LoadString(),
				mimeType: ls.LoadString(),
				class:    ls.LoadString(),
			}
			if multi {
				i.priorities[k] = ls.LoadStrings()
			}
		}
	} else {
		for j := 0; j < le; j++ {
			k := ls.LoadString()
			i.infos[k] = formatInfo{
				name:     ls.LoadString(),
				version:  ls.LoadString(),
				mimeType: ls.LoadString(),
			}
			if multi {
				i.priorities[k] = ls.LoadStrings()
			}
		}
	}

	return i
}

// New creates a new PRONOM Identifier
func New(opts ...config.Option) (core.Identifier, error) {
	for _, v := range opts {
		v()
	}
	pronom, err := raw()
	if err != nil {
		return nil, err
	}
	var pmap priority.Map
	if config.GetMulti() == config.DROID {
		pmap = pronom.Priorities()
	}
	pronom = identifier.ApplyConfig(pronom)
	id := &Identifier{
		Base:     identifier.New(pronom, config.ZipPuid()),
		hasClass: config.Reports() != "" && !config.NoClass(),
		infos:    infos(pronom.Infos()),
	}
	if id.Multi() == config.DROID {
		id.priorities = pmap
	}
	return id, nil
}

// Fields returns the user-facing fields used in the Identifier's
// reports.
func (i *Identifier) Fields() []string {
	if i.hasClass {
		return []string{
			"namespace",
			"id",
			"format",
			"version",
			"mime",
			"class",
			"basis",
			"warning",
		}
	} else {
		return []string{
			"namespace",
			"id",
			"format",
			"version",
			"mime",
			"basis",
			"warning",
		}
	}
}

// Recorder provides a new recorder for identification results.
func (i *Identifier) Recorder() core.Recorder {
	return &Recorder{
		Identifier: i,
		ids:        make(pids, 0, 1),
	}
}

// Recorder stores information about match results.
type Recorder struct {
	*Identifier
	ids        pids
	cscore     int
	satisfied  bool
	extActive  bool
	mimeActive bool
	textActive bool
}

const (
	extScore = 1 << iota
	mimeScore
	textScore
	incScore
)

// Active sets the flags for the recorder's active matchers.
func (r *Recorder) Active(m core.MatcherType) {
	if r.Identifier.Active(m) {
		switch m {
		case core.NameMatcher:
			r.extActive = true
		case core.MIMEMatcher:
			r.mimeActive = true
		case core.TextMatcher:
			r.textActive = true
		}
	}
}

// Record builds possible results sets associated with an identification.
func (r *Recorder) Record(m core.MatcherType, res core.Result) bool {
	switch m {
	default:
		return false
	case core.NameMatcher:
		if hit, id := r.Hit(m, res.Index()); hit {
			r.ids = add(r.ids, r.Name(), id, r.infos[id], res.Basis(), extScore)
			return true
		}
		return false
	case core.MIMEMatcher:
		if hit, id := r.Hit(m, res.Index()); hit {
			r.ids = add(r.ids, r.Name(), id, r.infos[id], res.Basis(), mimeScore)
			return true
		}
		return false
	case core.ContainerMatcher:
		// add zip default
		if res.Index() < 0 {
			if r.ZipDefault() {
				r.cscore += incScore
				r.ids = add(r.ids, r.Name(), config.ZipPuid(), r.infos[config.ZipPuid()], res.Basis(), r.cscore)
			}
			return false
		}
		if hit, id := r.Hit(m, res.Index()); hit {
			r.cscore += incScore
			basis := res.Basis()
			p, t := r.Place(core.ContainerMatcher, res.Index())
			if t > 1 {
				basis = basis + fmt.Sprintf(" (signature %d/%d)", p, t)
			}
			r.ids = add(r.ids, r.Name(), id, r.infos[id], basis, r.cscore)
			return true
		}
		return false
	case core.ByteMatcher:
		if hit, id := r.Hit(m, res.Index()); hit {
			if r.satisfied {
				return true
			}
			r.cscore += incScore
			basis := res.Basis()
			p, t := r.Place(core.ByteMatcher, res.Index())
			if t > 1 {
				basis = basis + fmt.Sprintf(" (signature %d/%d)", p, t)
			}
			r.ids = add(r.ids, r.Name(), id, r.infos[id], basis, r.cscore)
			return true
		}
		return false
	case core.TextMatcher:
		if hit, id := r.Hit(m, res.Index()); hit {
			if r.satisfied {
				return true
			}
			r.ids = add(r.ids, r.Name(), id, r.infos[id], res.Basis(), textScore)
			return true
		}
		return false
	}
}

// Satisfied determines whether we should continue running identification
// with a given matcher type.
func (r *Recorder) Satisfied(mt core.MatcherType) (bool, core.Hint) {
	if r.NoPriority() && r.Multi() != config.DROID { // config.DROID is higher than Comprehensive, nevertheless we may want to halt e.g. no byte matching if successful container match
		return false, core.Hint{}
	}
	if r.cscore < incScore {
		if len(r.ids) == 0 {
			return false, core.Hint{}
		}
		if mt == core.ContainerMatcher || mt == core.ByteMatcher || mt == core.XMLMatcher || mt == core.RIFFMatcher {
			if mt == core.ByteMatcher || mt == core.ContainerMatcher {
				keys := make([]string, len(r.ids))
				for i, v := range r.ids {
					keys[i] = v.String()
				}
				return false, core.Hint{
					Exclude: r.Start(mt),
					Pivot:   r.Lookup(mt, keys),
				}
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
		return true, core.Hint{
			Exclude: r.Start(mt),
			Pivot:   nil,
		}
	}
	return true, core.Hint{}
}

func lowConfidence(conf int) string {
	var ls = make([]string, 0, 1)
	if conf&extScore == extScore {
		ls = append(ls, "extension")
	}
	if conf&mimeScore == mimeScore {
		ls = append(ls, "MIME")
	}
	if conf&textScore == textScore {
		ls = append(ls, "text")
	}
	switch len(ls) {
	case 0:
		return ""
	case 1:
		return ls[0]
	case 2:
		return ls[0] + " and " + ls[1]
	default:
		return strings.Join(ls[:len(ls)-1], ", ") + " and " + ls[len(ls)-1]
	}
}

// Report organizes the results output and lists the highest priority
// results first.
func (r *Recorder) Report() []core.Identification {
	// no results
	if len(r.ids) == 0 {
		if r.hasClass {
			return []core.Identification{Identification{
				Namespace: r.Name(),
				ID:        "UNKNOWN",
				Warning:   "no match",
			}}
		}
		return []core.Identification{NoClassIdentification{
			Identification{
				Namespace: r.Name(),
				ID:        "UNKNOWN",
				Warning:   "no match",
			},
		}}
	}
	sort.Sort(r.ids)
	// exhaustive
	if r.Multi() == config.Exhaustive {
		ret := make([]core.Identification, len(r.ids))
		for i, v := range r.ids {
			ret[i] = r.updateWarning(v)
		}
		return ret
	}
	conf := r.ids[0].confidence
	// if we've only got extension / mime matches, check if those matches are ruled out by lack of byte match
	// only permit a single extension or mime only match
	// add warnings too
	if conf <= textScore {
		nids := make([]Identification, 0, 1)
		for _, v := range r.ids {
			// if overall confidence is greater than mime or ext only, then rule out any lesser confident matches
			if conf > mimeScore && v.confidence != conf {
				break
			}
			// if we have plain text result that is based on ext or mime only,
			// and not on a text match, and if text matcher is on for this identifier,
			// then don't report a text match
			if v.ID == config.TextPuid() && conf < textScore && r.textActive {
				continue
			}
			// if the match has no corresponding byte or container signature...
			if ok := r.HasSig(v.ID, core.ContainerMatcher, core.ByteMatcher); !ok {
				// break immediately if more than one match
				if len(nids) > 0 {
					nids = nids[:0]
					break
				}
				nids = append(nids, v)
			}
		}
		if len(nids) != 1 {
			poss := make([]string, len(r.ids))
			for i, v := range r.ids {
				poss[i] = v.ID
				conf = conf | v.confidence
			}
			if r.hasClass {
				return []core.Identification{Identification{
					Namespace: r.Name(),
					ID:        "UNKNOWN",
					Warning:   fmt.Sprintf("no match; possibilities based on %v are %v", lowConfidence(conf), strings.Join(poss, ", ")),
				}}
			}
			return []core.Identification{NoClassIdentification{
				Identification{
					Namespace: r.Name(),
					ID:        "UNKNOWN",
					Warning:   fmt.Sprintf("no match; possibilities based on %v are %v", lowConfidence(conf), strings.Join(poss, ", ")),
				},
			}}
		}
		r.ids = nids
	}
	// handle multi-mode single where there isn't a definitive match
	if r.Multi() == config.Single && len(r.ids) > 1 && r.ids[0].confidence == r.ids[1].confidence {
		poss := make([]string, 0, len(r.ids))
		for _, v := range r.ids {
			if v.confidence < conf {
				break
			}
			poss = append(poss, v.ID)
		}
		if r.hasClass {
			return []core.Identification{Identification{
				Namespace: r.Name(),
				ID:        "UNKNOWN",
				Warning:   fmt.Sprintf("multiple matches %v", strings.Join(poss, ", ")),
			}}
		}
		return []core.Identification{NoClassIdentification{
			Identification{
				Namespace: r.Name(),
				ID:        "UNKNOWN",
				Warning:   fmt.Sprintf("multiple matches %v", strings.Join(poss, ", ")),
			},
		}}
	}
	ret := make([]core.Identification, len(r.ids))
	for i, v := range r.ids {
		if i > 0 {
			switch r.Multi() {
			case config.Single:
				return ret[:i]
			case config.Conclusive:
				if v.confidence < conf {
					return ret[:i]
				}
			case config.DROID:
				if v.confidence < incScore {
					if i > 1 {
						return applyPriorities(r.priorities, ret[:i])
					}
					return ret[:i] // don't bother applying priorities if only one strong match
				}
			default:
				if v.confidence < incScore {
					return ret[:i]
				}
			}
		}
		ret[i] = r.updateWarning(v)
	}
	if len(ret) > 1 && r.Multi() == config.DROID {
		return applyPriorities(r.priorities, ret)
	}
	return ret
}

func applyPriorities(pmap priority.Map, ids []core.Identification) []core.Identification {
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = id.String()
	}
	sups := pmap.Apply(keys)
	ret := make([]core.Identification, len(sups))
	for i, s := range sups {
		for _, id := range ids {
			if s == id.String() {
				ret[i] = id
				break
			}
		}
	}
	return ret
}

func (r *Recorder) updateWarning(i Identification) core.Identification {
	// apply low confidence
	if i.confidence <= textScore {
		if len(i.Warning) > 0 {
			i.Warning += "; " + "match on " + lowConfidence(i.confidence) + " only"
		} else {
			i.Warning = "match on " + lowConfidence(i.confidence) + " only"
		}
	}
	// apply mismatches
	if r.extActive && (i.confidence&extScore != extScore) {
		for _, v := range r.IDs(core.NameMatcher) {
			if i.ID == v {
				if len(i.Warning) > 0 {
					i.Warning += "; extension mismatch"
				} else {
					i.Warning = "extension mismatch"
				}
				break
			}
		}
	}
	if r.mimeActive && (i.confidence&mimeScore != mimeScore) {
		for _, v := range r.IDs(core.MIMEMatcher) {
			if i.ID == v {
				if len(i.Warning) > 0 {
					i.Warning += "; MIME mismatch"
				} else {
					i.Warning = "MIME mismatch"
				}
				break
			}
		}
	}
	if !r.hasClass {
		return NoClassIdentification{i}
	}
	return i
}

// Identification records format related metadata.
type Identification struct {
	Namespace  string
	ID         string
	Name       string
	Version    string
	MIME       string
	Class      string
	Basis      []string
	Warning    string
	archive    config.Archive
	confidence int
}

func (id Identification) String() string {
	return id.ID
}

// Known outputs true if the identification result isn't UNKNOWN.
func (id Identification) Known() bool {
	return id.ID != "UNKNOWN"
}

// Warn returns the associated warning for a given identification.
func (id Identification) Warn() string {
	return id.Warning
}

// Values returns the Identification result information to the caller.
func (id Identification) Values() []string {
	var basis string
	if len(id.Basis) > 0 {
		basis = strings.Join(id.Basis, "; ")
	}
	return []string{
		id.Namespace,
		id.ID,
		id.Name,
		id.Version,
		id.MIME,
		id.Class,
		basis,
		id.Warning,
	}
}

// Archive returns the archive value for a given identification.
func (id Identification) Archive() config.Archive {
	return id.archive
}

type pids []Identification

func (p pids) Len() int { return len(p) }

func (p pids) Less(i, j int) bool { return p[j].confidence < p[i].confidence }

func (p pids) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func add(p pids, id string, f string, info formatInfo, basis string, c int) pids {
	for i, v := range p {
		if v.ID == f {
			p[i].confidence += c
			p[i].Basis = append(p[i].Basis, basis)
			return p
		}
	}
	return append(p,
		Identification{
			Namespace:  id,
			ID:         f,
			Name:       info.name,
			Version:    info.version,
			MIME:       info.mimeType,
			Class:      info.class,
			Basis:      []string{basis},
			Warning:    "",
			archive:    config.IsArchive(f),
			confidence: c,
		},
	)
}

// NoClassIdentification wraps Identification to implement the noclass option
type NoClassIdentification struct {
	Identification
}

func (nc NoClassIdentification) Values() []string {
	var basis string
	if len(nc.Basis) > 0 {
		basis = strings.Join(nc.Basis, "; ")
	}
	return []string{
		nc.Namespace,
		nc.ID,
		nc.Name,
		nc.Version,
		nc.MIME,
		basis,
		nc.Warning,
	}
}
