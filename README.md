# Siegfried

[Siegfried](http://www.itforarchivists.com/siegfried) is a signature-based file format identification tool, implementing:

  - the National Archives UK's [PRONOM](http://www.nationalarchives.gov.uk/pronom) file format signatures
  - freedesktop.org's [MIME-info](https://freedesktop.org/wiki/Software/shared-mime-info/) file format signatures
  - the Library of Congress's [FDD](http://www.digitalpreservation.gov/formats/fdd/descriptions.shtml) file format signatures (*beta*).
  - Wikidata (*beta*).

### Version

1.9.2

[![GoDoc](https://godoc.org/github.com/richardlehane/siegfried?status.svg)](https://godoc.org/github.com/richardlehane/siegfried) [![Go Report Card](https://goreportcard.com/badge/github.com/richardlehane/siegfried)](https://goreportcard.com/report/github.com/richardlehane/siegfried)

## Usage

### Command line

    sf file.ext
    sf DIR

#### Options

    sf -csv file.ext | DIR                     // Output CSV rather than YAML
    sf -json file.ext | DIR                    // Output JSON rather than YAML
    sf -droid file.ext | DIR                   // Output DROID CSV rather than YAML
    sf -nr DIR                                 // Don't scan subdirectories
    sf -z file.zip | DIR                       // Decompress and scan zip, tar, gzip, warc, arc
    sf -zs gzip,tar file.tar.gz | DIR          // Selectively decompress and scan 
    sf -hash md5 file.ext | DIR                // Calculate md5, sha1, sha256, sha512, or crc hash
    sf -sig custom.sig file.ext                // Use a custom signature file
    sf -                                       // Scan stream piped to stdin
    sf -name file.ext -                        // Provide filename when scanning stream 
    sf -f myfiles.txt                          // Scan list of files and directories
    sf -v | -version                           // Display version information
    sf -home c:\junk -sig custom.sig file.ext  // Use a custom home directory
    sf -serve hostname:port                    // Server mode
    sf -throttle 10ms DIR                      // Pause for duration (e.g. 1s) between file scans
    sf -multi 256 DIR                          // Scan multiple (e.g. 256) files in parallel 
    sf -log [comma-sep opts] file.ext | DIR    // Log errors etc. to stderr (default) or stdout
    sf -log e,w file.ext | DIR                 // Log errors and warnings to stderr
    sf -log u,o file.ext | DIR                 // Log unknowns to stdout
    sf -log d,s file.ext | DIR                 // Log debugging and slow messages to stderr
    sf -log p,t DIR > results.yaml             // Log progress and time while redirecting results
    sf -log fmt/1,c DIR > results.yaml         // Log instances of fmt/1 and chart results
    sf -replay -log u -csv results.yaml        // Replay results file, convert to csv, log unknowns
    sf -setconf -multi 32 -hash sha1           // Save flag defaults in a config file
    sf -setconf -serve :5138 -conf srv.conf    // Save/load named config file with '-conf filename' 

#### Example

[![asciicast](https://asciinema.org/a/ernm49loq5ofuj48ywlvg7xq6.png)](https://asciinema.org/a/ernm49loq5ofuj48ywlvg7xq6)

#### Signature files

By default, siegfried uses the latest PRONOM signatures without buffer limits (i.e. it may do full file scans). To use MIME-info or LOC signatures, or to add buffer limits or other customisations, use the [roy tool](https://github.com/richardlehane/siegfried/wiki/Building-a-signature-file-with-ROY) to build your own signature file.

## Install
### With go installed: 

    go install github.com/richardlehane/siegfried/cmd/sf@latest

    sf -update


### Or, without go installed:
#### Win:

Download a pre-built binary from the [releases page](https://github.com/richardlehane/siegfried/releases). Unzip to a location in your system path. Then run:

    sf -update

#### Mac Homebrew (or [Linuxbrew](http://brew.sh/linuxbrew/)):

    brew install mistydemeo/digipres/siegfried

Or, for the most recent updates, you can install from this fork:

    brew install richardlehane/digipres/siegfried

#### Ubuntu/Debian (64 bit):

    wget -qO - https://bintray.com/user/downloadSubjectPublicKey?username=bintray | sudo apt-key add -
    echo "deb http://dl.bintray.com/siegfried/debian wheezy main" | sudo tee -a /etc/apt/sources.list
    sudo apt-get update && sudo apt-get install siegfried

#### FreeBSD:

    pkg install siegfried

#### Arch Linux: 

    git clone https://aur.archlinux.org/siegfried.git
    cd siegfried
    makepkg -si

## Changes
### v1.9.2 (2022-02-07)
### Added
- Wikidata definition file specification has been updated and now includes endpoint (users will need to harvest Wikidata again)
- Custom Wikibase endpoint can now be specified for harvesting when paired with a custom SPARQL query and property mappings
- Wikidata identifier includes permalinks in results
- Wikidata revision history visible using `roy inspect`

### Changed
- update PRONOM to v100
- update LOC signatures to 2022-02-01
- update tika-mimetypes signatures to v2.1
- update freedesktop.org signatures to v2.2.1
- roy inspect returns format ID with name

### Fixed
- parse issues for container files where zero indexing used for Position. Spotted by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/175)
- sf -droid output can't be read by sf (e.g. for comparing results). Reported by [ostnatalie](https://github.com/richardlehane/siegfried/issues/174)
- panic when running in server mode due to race condition. Reported by [Miguel Guimarães](https://github.com/richardlehane/siegfried/issues/172)
- panic when reading malformed MSCFB files. Reported by [Greg Lepore](https://github.com/richardlehane/siegfried/issues/171)
- unescaped control characters in JSON output. Reported by [Sebastian Lange](https://github.com/richardlehane/siegfried/issues/165)
- zip file names with null terminated strings prevent ID of Serif formats. Reported by [Tyler Thorsted](https://github.com/richardlehane/siegfried/issues/150)

### v1.9.1 (2020-10-11)
### Changed
- update PRONOM to v97
- zs flag now activates -z flag

### Fixed
- details text in PRONOM identifier
- `roy` panic when building signatures with empty sequences. Reported by [Greg Lepore](https://github.com/richardlehane/siegfried/issues/149)

### v1.9.0 (2020-09-22)
### Added
- a new Wikidata identifier, harvesting information from the Wikidata Query Service. Implemented by [Ross Spencer](https://github.com/richardlehane/siegfried/commit/dfb579b4ae46ae6daa814fc3fc74271d768f2f9c). 
- select which archive types (zip, tar, gzip, warc, or arc) are unpacked using the -zs flag (sf -zs tar,zip). Implemented by [Ross Spencer](https://github.com/richardlehane/siegfried/commit/88dd43b55e5f83304705f6bcd439d502ef08cd38).

### Changed
- update LOC signatures to 2020-09-21
- update tika-mimetypes signatures to v1.24
- update freedesktop.org signatures to v2.0

### Fixed
- incorrect basis for some signatures with multiple patterns. Reported and fixed by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/142).

### v1.8.0 (2020-01-22)
### Added
- utc flag returns file modified dates in UTC e.g. `sf -utc FILE | DIR`. Requested by [Dragan Espenschied](https://github.com/richardlehane/siegfried/issues/136)
- new cost and repetition flags to control segmentation when building signatures

### Changed
- update PRONOM to v96
- update LOC signatures to 2019-12-18
- update tika-mimetypes signatures to v1.23
- update freedesktop.org signatures to v1.15

### Fixed
- XML namespaces detected by prefix on root tag, as well as default namespace (for mime-info spec)
- panic when scanning certain MS-CFB files. Reported separately by Mike Shallcross and Euan Cochrane
- file with many FF xx sequences grinds to a halt. Reported by [Andy Foster](https://github.com/richardlehane/siegfried/issues/128)

See the [CHANGELOG](CHANGELOG.md) for the full history.

## Rights

Copyright 2020 Richard Lehane, Ross Spencer 

Licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

## Announcements

Join the [Google Group](https://groups.google.com/d/forum/sf-roy) for updates, signature releases, and help.

## Contributing

Like siegfried and want to get involved in its development? That'd be wonderful! There are some notes on the [wiki](https://github.com/richardlehane/siegfried/wiki) to get you started, and please get in touch.

## Thanks

Thanks TNA for http://www.nationalarchives.gov.uk/pronom/ and http://www.nationalarchives.gov.uk/information-management/projects-and-work/droid.htm

Thanks Ross for https://github.com/exponential-decay/skeleton-test-suite-generator and http://exponentialdecay.co.uk/sd/index.htm, both are very handy!

Thanks Misty for the brew and ubuntu packaging

Thanks Steffen for the FreeBSD and Arch Linux packaging
