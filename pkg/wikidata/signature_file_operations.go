package wikidata

import (
	"fmt"
	"os"
	"time"

	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"

	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/internal/persist"

	"github.com/richardlehane/siegfried/pkg/wikidata/internal/mappings"
)

// Save will write a Wikidata identifier to the Siegfried signature file.
func (i *Identifier) Save(ls *persist.LoadSaver) {

	// Save the Wikidata magic enum from core.
	ls.SaveByte(core.Wikidata)

	// Save the no. formatInfo entries to read.
	ls.SaveSmallInt(len(i.infos))

	// Save the information in the formatInfo records.
	for key, value := range i.infos {
		ls.SaveString(key)
		ls.SaveString(value.name)
		ls.SaveString(value.uri)
		ls.SaveString(value.mime)
		ls.SaveString(value.puid)
		ls.SaveStrings(value.sources)
	}
	i.Base.Save(ls)
}

// Load will read a Wikidata identifier from the Siegfried signature file.
func Load(ls *persist.LoadSaver) core.Identifier {
	i := &Identifier{}
	le := ls.LoadSmallInt()
	i.infos = make(map[string]formatInfo)

	// WD error is happening here because???
	for j := 0; j < le; j++ {
		i.infos[ls.LoadString()] = formatInfo{
			ls.LoadString(),
			ls.LoadString(),
			ls.LoadString(),
			ls.LoadString(),
			ls.LoadStrings(),
		}
	}
	i.Base = identifier.Load(ls)
	return i
}

// WIKIDATA TODO: I think this can hold pretty much anything, so what else is
// important and that characterizes a Wikidata signature?

// WIKIDATA TODO look at other format infos!!! This is looking promising, and
// for other information we might create lists, e.g. for mimetype... (string)

// ---
// items: [ 1, 2, 3, 4, 5 ]
// names: [ "one", "two", "three", "four" ]

// how does this impact how we return other serializations?

// formatInfo is our top-level structure describing an identification. An
// identification, e.g. PNG 1.1 might have multiple signatures, but it would
// only have one identifier for the record.
type formatInfo struct {
	name    string
	uri     string
	mime    string
	puid    string
	sources []string
}

func infos(formatInfoMap map[string]identifier.FormatInfo) map[string]formatInfo {
	// Turn the generic FormatInfo into a formatInfo structure that will be
	// written to the Siegfried identifier (signature file).
	idx := make(map[string]formatInfo, len(formatInfoMap))
	for key, value := range formatInfoMap {
		idx[key] = value.(formatInfo)
	}
	return idx
}

// Serialize formatInfo as a string for debugging.
func (f formatInfo) String() string {
	return fmt.Sprintf(
		"Format info: Name: '%s'; MIMEType: '%s'; PUID: '%s'; Sources: %s ",
		f.name,
		f.mime,
		f.puid,
		f.sources,
	)
}

// Infos arranges summary information about formats within an Identifier into a
// structure suitable for output in a Siegfried signature file.
//
// Infos provides a mechanism for a placing any other information about formats
// that you'd like to talk about in an identifier.
//
func (wdd wikidataFDDs) Infos() map[string]identifier.FormatInfo {
	fmt.Fprintf(os.Stderr, "WD: in infos... %d %t\n", len(wdd.formats), config.NoPRONOM())
	formatInfoMap := make(map[string]identifier.FormatInfo, len(wdd.formats))
	for _, value := range wdd.formats {
		if len(value.PRONOM) > 1 {
			// WIKIDATA TODO: I think we might want to duplicate a record
			// and/or trigger so that a result is captured per PUID.
		}

		sources := prepareSources(value)
		fi := formatInfo{
			name:    value.Name,
			uri:     value.URI,
			mime:    value.Mimetype[0],
			puid:    value.PRONOM[0],
			sources: sources,
		}
		formatInfoMap[value.ID] = fi
	}
	return formatInfoMap
}

func prepareSources(wdMapping mappings.Wikidata) []string {
	// Prepare a slice of sources that we will use to return some sort of
	// provenance information about positive matches returned from the
	// Wikidata identifier.
	//
	// The reason that this is not straightforward is that the identifier may
	// also be using PRONOM identifiers as a baseline set of identifiers and
	// we have to map those values from elsewhere to create a complete picture
	// of what the Wikidata identifier will eventually do.
	//

	// WIKIDATA TODO: We currently return a []string slice array from this
	// function, but a struct could actually be more extensible, and encode
	// greater complexity, for example, date in its own field could be nice.
	// But what is the view for Persist (the functions that encode the
	// signature file as a binary object. Is it purely for primitives? Perhaps
	// those primitives could be streamed back into an object of type sources?

	// WIKIDATA TODO: Is this the right approach for provenance?
	sources := []string{}
	if len(wdMapping.Signatures) > 0 {
		// WIKDIATA TODO: The next problem! This actually happens before the
		// signature is processed... and SO we need a different signal, and,
		// likely the option to remove a value from the index. But I'm
		// feeling positive...

		// WIKIDATA TODO: Using index[0] here is incorrect, we need to
		// incorporate all sources that were processed.

		prov := wdMapping.Signatures[0].Provenance
		date := wdMapping.Signatures[0].Date
		if date != "" {
			date, _ := time.Parse(time.RFC3339, date)
			prov = fmt.Sprintf("%s (source date: %s)", prov, date.Format("2006-01-02"))
		}
		sources = append(sources, prov)
	}
	if !config.WDNoPRONOM() {
		// Bring PRONOM sources into the identifier.
		for _, exid := range wdMapping.PRONOM {
			// We have PRONOM identifiers to work with so we need to extend the
			// slice further.
			for _, value := range tmpPuids {
				if exid == value {
					sources = append(sources, fmt.Sprintf(pronomOfficial, exid))
				}
			}
		}
	}
	return sources
}
