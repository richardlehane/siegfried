package pronom

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type FormatInfo struct {
	Name     string
	Version  string
	MIMEType string
}

type PronomIdentifier struct {
	SigVersion SigVersion
	Infos      map[string]FormatInfo
	BPuids     []string         // slice of puids that corresponds to the bytematcher's int signatures
	PuidsB     map[string][]int // map of puids to slices of bytematcher int signatures
	CPuids     []string
	EPuids     []string     // slice of puids that corresponds to the extension matcher's int signatures
	em         core.Matcher // extensionmatcher
	cm         core.Matcher // containermatcher
	bm         core.Matcher // bytematcher
	ids        pids
}

func (pi *PronomIdentifier) String() string {
	return pi.bm.String() + "\n---------\n\nContainer matcher\n" + pi.cm.String()
}

func (pi *PronomIdentifier) Details() string {
	return pi.SigVersion.String()
}

func (pi *PronomIdentifier) Version() string {
	return fmt.Sprintf("Signature version: %d; based on droid sig: %s; and container sig: %s", pi.SigVersion.Gob, pi.SigVersion.Droid, pi.SigVersion.Containers)
}

func (pi *PronomIdentifier) Update(i int) bool {
	return i > pi.SigVersion.Gob
}

func (pi *PronomIdentifier) Identify(b *siegreader.Buffer, n string, c chan core.Identification, wg *sync.WaitGroup) {
	pi.ids = pi.ids[:0]
	// Extension Matcher
	if len(n) > 0 {
		ems := pi.em.Identify(n, nil)
		for v := range ems {
			pi.ids = add(pi.ids, pi.EPuids[v.Index()], pi.Infos[pi.EPuids[v.Index()]], v.Basis(), 0.1)
		}
	}
	var cscore float64 = 0.1 //confidence score

	// Container Matcher
	if pi.cm != nil {
		cms := pi.cm.Identify(n, b)
		for v := range cms {
			cscore *= 1.1
			pi.ids = add(pi.ids, pi.CPuids[v.Index()], pi.Infos[pi.CPuids[v.Index()]], v.Basis(), cscore)
		}
	}
	// Byte Matcher (skip if a container matched and cscore changed)
	if cscore == 0.1 {
		ids := pi.bm.Identify("", b)
		for r := range ids {
			i := r.Index()
			cscore *= 1.1
			pi.ids = add(pi.ids, pi.BPuids[i], pi.Infos[pi.BPuids[i]], r.Basis(), cscore)
		}
	}
	if len(pi.ids) > 0 {
		sort.Sort(pi.ids)
		conf := pi.ids[0].confidence
		// if we've only got extension matches, check if those matches are ruled out by lack of byte match
		// add warnings too
		if conf == 0.1 {
			nids := make([]PronomIdentification, 0, len(pi.ids))
			for _, v := range pi.ids {
				if _, ok := pi.PuidsB[v.puid]; !ok {
					v.warning = "match on extension only"
					nids = append(nids, v)
				}
			}
			if len(nids) == 0 {
				poss := make([]string, len(pi.ids))
				for i, v := range pi.ids {
					poss[i] = v.puid
				}
				nids = []PronomIdentification{PronomIdentification{"UNKNOWN", "", "", "", nil, fmt.Sprintf("no match; possibilities based on extension are %v", strings.Join(poss, ", ")), 0}}
			}
			pi.ids = nids
		}

		c <- pi.ids[0]
		if len(pi.ids) > 1 {
			for i, v := range pi.ids[1:] {
				if v.confidence == conf {
					c <- pi.ids[i+1]
				} else {
					break
				}
			}
		}
	}
	wg.Done()
}

type PronomIdentification struct {
	puid       string
	name       string
	version    string
	mime       string
	basis      []string
	warning    string
	confidence float64
}

func (pid PronomIdentification) String() string {
	return pid.puid
}

func quoteText(s string) string {
	if len(s) == 0 {
		return s
	}
	return "\"" + s + "\""
}

func (pid PronomIdentification) Details() string {
	var basis string
	if len(pid.basis) > 0 {
		basis = quoteText(strings.Join(pid.basis, "; "))
	}
	return fmt.Sprintf("  - puid    : %v\n    format  : %v\n    version : %v\n    mime    : %v\n    basis   : %v\n    warning : %v\n",
		pid.puid, quoteText(pid.name), quoteText(pid.version), quoteText(pid.mime), basis, quoteText(pid.warning))
}

func (pid PronomIdentification) Json() string {
	type jsonid struct {
		Puid    string `json:"puid"`
		Name    string `json:"name"`
		Version string `json:"version"`
		Mime    string `json:"mime"`
		Basis   string `json:"basis"`
		Warning string `json:"warning"`
	}
	var basis string
	if len(pid.basis) > 0 {
		basis = strings.Join(pid.basis, "; ")
	}
	b, err := json.Marshal(jsonid{pid.puid, pid.name, pid.version, pid.mime, basis, pid.warning})
	if err != nil {
		return `{
			"puid": "",
			"name": "",
			"version": "",
			"mime": "",
			"basis": "",
			"warning": "json encoding error"
			}`
	}
	return string(b)
}

func (pid PronomIdentification) Confidence() float64 {
	return pid.confidence
}

type pids []PronomIdentification

func (p pids) Len() int { return len(p) }

func (p pids) Less(i, j int) bool { return p[j].confidence < p[i].confidence }

func (p pids) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func add(p pids, f string, info FormatInfo, basis string, c float64) pids {
	for i, v := range p {
		if v.puid == f {
			p[i].confidence += c
			p[i].basis = append(p[i].basis, basis)
			return p
		}
	}
	return append(p, PronomIdentification{f, info.Name, info.Version, info.MIMEType, []string{basis}, "", c})
}
