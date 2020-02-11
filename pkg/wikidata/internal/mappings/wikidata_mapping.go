package mappings

// Wikidata stores information about something which constitutes a format
// resource in Wikidata. I.e. Anything which has a URI and describes a
// file-format.
type Wikidata struct {
	ID         string      // Wikidata short name, e.g. Q12345 can be appended to a URI to be dereferenced.
	Name       string      // Name of the format as described in Wikidata.
	URI        string      // URI is the absolute URL in Wikidata terms that can be dereferenced.
	PRONOM     []string    // 1:1 mapping to PRONOM wherever possible.
	LOC        []string    // Library of Congress identifiers.
	Extension  []string    // Extension returned by Wikidata.
	Mimetype   []string    // Mimetype as recorded by Wikidata.
	Signatures []Signature // Signature associated with a record which we will convert to a new Type.
}

// Signature describes a complete signature resource, i.e. a way to identify
// a file format using Wikidata information.
type Signature struct {
	Signature  string // Signature byte sequence.
	Provenance string // Provenance of the signature.
	Date       string // Date the signature was submitted.
	Encoding   string // Signature encoding, e.g. Hexadecimal, ASCII, PRONOM.
	Relativity string // Position relative to beginning or end of file, or elsewhere.
}

// WikidataMapping provides a way to persist Wikidata resources in memory.
var WikidataMapping = make(map[string]Wikidata)

// PUIDs enables the Wikidata format records to be mapped to existing PRONOM
// records when run in PRONOM mode, i.e. not just with Wikidata signatures.
func (wdd Wikidata) PUIDs() []string {
	var puids []string
	for _, puid := range wdd.PRONOM {
		puids = append(puids, puid)
	}
	return puids
}

// Temporary: We don't need this in time.
func DeleteSignatures(wd *Wikidata) Wikidata {
	return Wikidata{
		ID:         wd.ID,
		Name:       wd.Name,
		URI:        wd.URI,
		PRONOM:     wd.PRONOM,
		LOC:        wd.LOC,
		Extension:  wd.Extension,
		Mimetype:   wd.Mimetype,
		Signatures: []Signature{},
	}
}
