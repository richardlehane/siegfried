package wikidata

import (
	"fmt"
	"os"

	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/identifier"
	"github.com/richardlehane/siegfried/pkg/wikidata/internal/mappings"
)

// wikidataNotProcessed is a signature we can use as a placeholder when there
// is an error in processing, but the ID and Signature slices still need to
// align for provenance reasons.
const wikidataNotProcessed = "57494B49444154413A204E6F742050726F63657373656420627920526F79"

// Globs are filename signatures. In this instance, File format extensions.
func (wdd wikidataFDDs) Globs() ([]string, []string) {
	fmt.Fprintf(os.Stderr, "WD: Adding Glob signatures to identifier....\n")
	globs, ids := make([]string, 0, len(wdd.formats)), make([]string, 0, len(wdd.formats))
	for _, v := range wdd.formats {
		for _, w := range v.Extension {
			globs, ids = append(globs, "*."+w), append(ids, v.ID)
		}
	}
	return globs, ids
}

// WIKIDATA TODO: Add in the PRONOM compatible pieces here again.
func (wdd wikidataFDDs) Signatures() ([]frames.Signature, []string, error) {
	frames, ids, err := SignaturesPRONOMandWikidata(wdd)
	// frames, ids, err := SignaturesPRONOMOnly(wdd)
	// frames, ids, err := SignaturesNativeWikidata(wdd)
	return frames, ids, err
}

func collectPUIDs(puidsIDs map[string][]string, v mappings.Wikidata) map[string][]string {
	if puidsIDs != nil {
		for _, puid := range v.PUIDs() {
			puidsIDs[puid] = append(puidsIDs[puid], v.ID)
		}
	}
	return puidsIDs
}

func SignaturesPRONOMOnly(wdd wikidataFDDs) ([]frames.Signature, []string, error) {
	// WIKIDATA TODO Siegfried uses an errors slice to return cumulative errors
	// here... e.g. var errs []error, we can do the same...

	var errs []error

	// WIKIDATA TODO do we need a Parseable object to take along with us while
	// building a signature file.

	fmt.Fprintf(os.Stderr, "WD: Adding Byte signatures to identifier...\n")

	sigs, ids := make([]frames.Signature, 0, len(wdd.formats)), make([]string, 0, len(wdd.formats))
	for _, wd := range wdd.formats {
		if len(wd.Signatures) > 0 {
			sig := formatSig(wd.Signatures[0].Signature)
			fmt.Fprintf(os.Stderr, "WD byte signature: %s %s\n", sig, wd.ID)
			frames, err := magics(sig)
			if err != nil {
				errs = append(errs, err)
			}
			for _, fr := range frames {
				sigs = append(sigs, fr)
				ids = append(ids, wd.ID)
			}
		}
	}
	return sigs, ids, nil
}

func SignaturesPRONOMandWikidata(wdd wikidataFDDs) ([]frames.Signature, []string, error) {

	// WIKIDATA TODO Siegfried uses an errors slice to return cumulative errors
	// here... e.g. var errs []error, we can do the same...

	var errs []error

	var puidsIDs map[string][]string
	if len(wdd.parseable.IDs()) > 0 {
		puidsIDs = make(map[string][]string)
	}

	// WIKIDATA TODO do we need a Parseable object to take along with us while
	// building a signature file.

	fmt.Fprintf(os.Stderr, "WD: Adding Byte signatures to identifier...\n")

	sigs, ids := make([]frames.Signature, 0, len(wdd.formats)), make([]string, 0, len(wdd.formats))
	for _, wd := range wdd.formats {

		// Collect PUIDs to supercharge the native identifier by including a
		// PRONOM identifier alongside the Wikidata one.
		puidsIDs = collectPUIDs(puidsIDs, wd)

		if len(wd.Signatures) > 0 {
			sig := formatSig(wd.Signatures[0].Signature)
			fmt.Fprintf(os.Stderr, "WD byte signature: %s %s\n", sig, wd.ID)
			frames, err := magics(sig)
			if err != nil {
				errs = append(errs, err)
			}
			for _, fr := range frames {
				sigs = append(sigs, fr)
				ids = append(ids, wd.ID)
			}
		}
	}

	// WIKIDATA TODO: can we combine with puidsIDs above... this is used to
	// determine if we add PRONOM proper to another identifier...
	if puidsIDs != nil {
		puids := make([]string, 0, len(puidsIDs))
		for p := range puidsIDs {
			puids = append(puids, p)
		}
		newParseable := identifier.Filter(puids, wdd.parseable)
		pronomSignatures, pronomIdentifiers, err := newParseable.Signatures()
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
	return sigs, ids, nil
}

// WIKIDATA TODO: This is a hack to match LOC style signatures. We might move
// this to the load signature module. Also, should we convert all signatures to
// PRONOM?
func formatSig(sig string) []string {
	var formatted []string
	newsig := ""
	for s := 0; s < len(sig); s += 2 {
		newsig = newsig + sig[s:s+2] + " "
	}
	newsig = "Hex: " + newsig
	formatted = append(formatted, newsig)
	return formatted
}

func SignaturesNativeWikidata(wdd wikidataFDDs) ([]frames.Signature, []string, error) {

	// WIKIDATA TODO Siegfried uses an errors slice to return cumulative errors
	// here... e.g. var errs []error, we can do the same...

	var errs []error

	// WIKIDATA TODO do we need a Parseable object to take along with us while
	// building a signature file.

	fmt.Fprintf(os.Stderr, "WD: Adding Wikidata Byte signatures to identifier...\n")

	// WIKIDATA TODO: We start piecing signatures together here...

	sigs := make([]frames.Signature, 0, len(wdd.formats))
	ids := make([]string, 0, len(wdd.formats))
	// prov := make([]string, 0, len(wdd.formats))

	for _, wd := range wdd.formats {
		if len(wd.Signatures) > 0 {
			// WIKIDATA TODO: Only doing index[0] here, so this needs to be
			// fixed.
			sig := formatSig(wd.Signatures[0].Signature)
			// fmt.Fprintf(os.Stderr, "WD byte signature: %s %s %s\n", sig, wd.ID, wd.Signatures[0].Provenance)
			frames, err := magics(sig)
			if err != nil {
				errs = append(errs, err)
			}
			for _, fr := range frames {
				sigs = append(sigs, fr)
				ids = append(ids, wd.ID)
				// prov = append(prov, wd.Signatures[0].Provenance)
			}
		}
	}
	return sigs, ids, nil
}

// WIKIDATA TODO WIKIDATA containers are currently disabled...

// Zips adds ZIP based container signatures to the identifier.
func (wdd wikidataFDDs) Zips() ([][]string, [][]frames.Signature, []string, error) {
	return wdd.containers("ZIP")
}

// MSCFBs adds OLE2 based container signatures to the identifier.
func (wdd wikidataFDDs) MSCFBs() ([][]string, [][]frames.Signature, []string, error) {
	return wdd.containers("OLE2")
}

// WIKIDATA TODO: Wikidata doesn't have its own concept of container format
// identification just yet and so we do this via PRONOM's in-build methods.
// This mimics that of the Library of Congress identifier. Wikidata container
// modelling is in-progress.
func (wdd wikidataFDDs) containers(typ string) ([][]string, [][]frames.Signature, []string, error) {
	fmt.Fprintf(os.Stderr, "WD: Adding container signatures to identifier...\n")
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

	names, sigs, ids := make([][]string, 0, len(wdd.formats)), make([][]frames.Signature, 0, len(wdd.formats)), make([]string, 0, len(wdd.formats))
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
