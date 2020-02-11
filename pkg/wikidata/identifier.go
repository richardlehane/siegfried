package wikidata

import (
	"fmt"
	"strings"
	"time"

	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

const unknown = "UNKNOWN"

// I can't remember what the definition of Golang init is...
func init() {
	core.RegisterIdentifier(core.Wikidata, Load)
}

// Identifier comment...
type Identifier struct {
	infos map[string]formatInfo
	*identifier.Base
}

// WIKIDATA TODO: We may eventually delete this, but right now what it allows
// us to do is keep track of the PUIDs going to be output in the identifier
// which need provenance. At least it felt needed at the time, but need to
// look at in more detail.
var tmpPuids []string

// New is the entry point for an Identifier when it is compiled by the Roy tool
// to a brand new signature file.
//
// New will read a Wikidata report, and parse its information into structures
// suitable for compilation by Roy.
//
// New will also update its identification information with provenance-like
// info. It will enable signature extensions to be added by the utility, and
// enables configuration to be applied as well.
//
// Examples of extensions include: ...
//
// Examples of configuration include: ...
//
func New(opts ...config.Option) (core.Identifier, error) {

	fmt.Println("WD Roy: congratulations: doing something with the Wikidata identifier package!")

	wikidata, puids, err := newWikidata()
	if err != nil {
		return nil, fmt.Errorf("WD Roy: error in New Wikidata: %s", err)
	}

	// WIKIDATA TODO: Refer to purpose of this above.
	tmpPuids = puids

	// WIKIDATA TODO: What does date versioning look like from the Wikidata
	// sources?
	updatedDate := time.Now().Format("2006-01-02")

	// WIKIDATA TODO: Refer to LoC identifier for more information here...
	// WIKIDATA TODO: Add extensions.

	/*  type Base struct {
		p                                        Parseable
		name                                     string
		details                                  string
		multi                                    config.Multi
		zipDefault                               bool
		gids, mids, cids, xids, bids, rids, tids *indexes
	}
	*/

	wikidata = identifier.ApplyConfig(wikidata)

	base := identifier.New(
		wikidata,
		"Wikidata Name: I don't think this field is used...",
		updatedDate,
	)

	infos := infos(wikidata.Infos())

	return &Identifier{
		infos: infos,
		Base:  base,
	}, nil
}

// Recorder comment that belongs to an identifier...
func (i *Identifier) Recorder() core.Recorder {
	return &Recorder{
		Identifier: i,
		ids:        make(matchIDs, 0, 1),
	}
}

// Identification contains the result of a single ID for a file. There may be
// multiple, per file. The identification to the user looks something like as
// follows:
//
//  - ns      : 'wikidata'
//    id      : 'Q136218'
//    format  : 'ZIP'
//    URI     : 'http://www.wikidata.org/entity/Q136218'
//    puid    : 'x-fmt/263' <-- NB. we have removed this, can  it be in basis?
//    mime    : 'application/zip'
//    basis   : 'byte match at [[0 4] [75 3] [129 4]]'
//    warning :
//
type Identification struct {
	Namespace  string
	ID         string
	Name       string
	LongName   string
	PUID       string
	MIME       string
	Basis      []string
	Source     []string
	Warning    string
	archive    config.Archive
	confidence int
}

// String creates a human readable representation of an identifier for output
// by fmt-like functions.
func (id Identification) String() string {
	return fmt.Sprintf(
		"%s %s\n",
		id.ID,
		id.Name,
	)
}

// Fields describes a portion of YAML that will be output by Siegfried's
// identifier for an individual match. E.g.
//
//      matches  :
//        - ns      : 'wikidata'
//          id      : 'Q475488'
//          format  : 'EPUB'
//          etc.    : '...'
//          etc.    : '...'
//
//        - ns      : repeated per match...
//
// siegfried/pkg/writer/writer.go normalizes the output of this field grouping
// so that if it sees certain fields, e.g. namespace, then it can convert that
// to something anticipated by the consumer,
//
//      e.g. namespace => becomes => ns
//
func (i *Identifier) Fields() []string {
	// WIKIDATA TODO: What fields do we want come the end of this project. NB.
	// PUID is available in the identifier. We can potentially return that in
	// the basis field if PRONOM identification is being triggered.
	if config.GetWikidataSourceField() {
		return []string{
			"namespace",
			"id",
			"format",
			"URI",
			"mime",
			"basis",
			"source",
			"warning",
		}
	}
	return []string{
		"namespace",
		"id",
		"format",
		"URI",
		"mime",
		"basis",
		"warning",
	}
}

// Archive comment... [Mandatory]
func (id Identification) Archive() config.Archive {
	return id.archive
}

// Known comment... [Mandatory]
func (id Identification) Known() bool {
	return id.ID != unknown
}

// Warn comment... [Mandatory]
func (id Identification) Warn() string {
	return id.Warning
}

// Values returns a string slice containing each of the identifier segments.
func (id Identification) Values() []string {
	var basis string
	var source string
	if len(id.Basis) > 0 {
		basis = strings.Join(id.Basis, "; ")
	}
	if config.GetWikidataSourceField() {
		if len(id.Source) > 0 {
			// WIKIDATA TOOD: This is a bit of a hack to enable provenance to
			// play nicely with container identification.
			if id.Source[0] != "" {
				source = strings.Join(id.Source, "; ")
			}
			source = strings.TrimSpace(strings.Join(id.Source, " "))
		}
		return []string{
			id.Namespace,
			id.ID,
			id.Name,
			id.LongName,
			id.MIME,
			basis,
			source,
			id.Warning,
		}
	}
	return []string{
		id.Namespace,
		id.ID,
		id.Name,
		id.LongName,
		id.MIME,
		basis,
		id.Warning,
	}
}
