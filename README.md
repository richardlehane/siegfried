# Siegfried

[Siegfried](http://www.itforarchivists.com/siegfried) is a signature-based file format identification tool, implementing:

  - the National Archives UK's [PRONOM](http://www.nationalarchives.gov.uk/pronom) file format signatures
  - freedesktop.org's [MIME-info](https://freedesktop.org/wiki/Software/shared-mime-info/) file format signatures
  - the Library of Congress's [FDD](http://www.digitalpreservation.gov/formats/fdd/descriptions.shtml) file format signatures (*beta*).
  - Wikidata (*beta*).

### Version

1.10.1

[![GoDoc](https://godoc.org/github.com/richardlehane/siegfried?status.svg)](https://godoc.org/github.com/richardlehane/siegfried) [![Go Report Card](https://goreportcard.com/badge/github.com/richardlehane/siegfried)](https://goreportcard.com/report/github.com/richardlehane/siegfried)

## Usage

### Command line

    sf file.ext
    sf *.ext
    sf DIR

#### Options

    sf -csv file.ext | *.ext | DIR             // Output CSV rather than YAML
    sf -json file.ext | *.ext | DIR            // Output JSON rather than YAML
    sf -droid file.ext | *.ext | DIR           // Output DROID CSV rather than YAML
    sf -nr DIR                                 // Don't scan subdirectories
    sf -z file.zip | *.ext | DIR               // Decompress and scan zip, tar, gzip, warc, arc
    sf -zs gzip,tar file.tar.gz | *.ext | DIR  // Selectively decompress and scan 
    sf -hash md5 file.ext | *.ext | DIR        // Calculate md5, sha1, sha256, sha512, or crc hash
    sf -sig custom.sig *.ext | DIR             // Use a custom signature file
    sf -                                       // Scan stream piped to stdin
    sf -name file.ext -                        // Provide filename when scanning stream 
    sf -f myfiles.txt                          // Scan list of files and directories
    sf -v | -version                           // Display version information
    sf -home c:\junk -sig custom.sig file.ext  // Use a custom home directory
    sf -serve hostname:port                    // Server mode
    sf -throttle 10ms DIR                      // Pause for duration (e.g. 1s) between file scans
    sf -multi 256 DIR                          // Scan multiple (e.g. 256) files in parallel 
    sf -log [comma-sep opts] file.ext          // Log errors etc. to stderr (default) or stdout
    sf -log e,w file.ext | *.ext | DIR         // Log errors and warnings to stderr
    sf -log u,o file.ext | *.ext | DIR         // Log unknowns to stdout
    sf -log d,s file.ext | *.ext | DIR         // Log debugging and slow messages to stderr
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

    curl -sL "http://keyserver.ubuntu.com/pks/lookup?op=get&search=0x20F802FE798E6857" | gpg --dearmor | sudo tee /usr/share/keyrings/siegfried-archive-keyring.gpg
    echo "deb [signed-by=/usr/share/keyrings/siegfried-archive-keyring.gpg] https://www.itforarchivists.com/ buster main" | sudo tee -a /etc/apt/sources.list.d/siegfried.list
    sudo apt-get update && sudo apt-get install siegfried

#### FreeBSD:

    pkg install siegfried

#### Arch Linux: 

    git clone https://aur.archlinux.org/siegfried.git
    cd siegfried
    makepkg -si

## Changes
### v1.10.1 (2023-04-24)
### Fixed
- glob expansion now only on Windows & when no explicit path match. Implemented by [Bernhard Hampel-Waffenthal](https://github.com/richardlehane/siegfried/pull/229)
- compression algorithm for debian packages changed back to xz. Implemented by [Paul Millar](https://github.com/richardlehane/siegfried/pull/230)
- `-multi droid` setting returned empty results when priority lists contained self-references. See [#218](https://github.com/richardlehane/siegfried/issues/218)
- CGO disabled for debian package and linux binaries. See [#219](https://github.com/richardlehane/siegfried/issues/219)

### v1.10.0 (2023-03-25)
### Added
- format classification included as "class" field in PRONOM results. Requested by [Robin François](https://github.com/richardlehane/siegfried/discussions/207). Implemented by [Ross Spencer](https://github.com/richardlehane/siegfried/commit/7f695720a752ac5fca3e1de8ba034b92ab6da1d9)
- `-noclass` flag added to roy build command. Use this flag to build signatures that omit the new "class" field from results.
- glob paths can be used in place of file or directory paths for identification (e.g. `sf *.jpg`). Implemented by [Ross Spencer](https://github.com/richardlehane/siegfried/commit/54bf6596c5fe7d1c9858348f0170d0dd7365fc8f)
- `-multi droid` setting for roy build command. Applies priorities after rather than during identification for more DROID-like results. Reported by [David Clipsham](https://github.com/richardlehane/siegfried/issues/146)
- `/update` command for server mode. Requested by [Luis Faria](https://github.com/richardlehane/siegfried/issues/208)

### Changed
- new algorithm for dynamic multi-sequence matching for improved wildcard performance
- update PRONOM to v111
- update LOC to 2023-01-27
- update tika-mimetypes to v2.7.0
- minimum go version to build siegfried is now 1.18 

### Fixed
- archivematica extensions built into wikidata signatures. Reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/210)
- trailing slash for folder paths in URI field in droid output. Reported by Philipp Wittwer
- crash when using `sf -replay` with droid output

### v1.9.6 (2022-11-06)
### Changed
- update PRONOM to v109

### v1.9.5 (2022-09-13)
### Added
- `roy inspect` now takes a `-droid` flag to allow easier inspection of old or custom DROID files
- github action to update siegfried docker deployment [https://github.com/keeps/siegfried-docker]. Implemented by [Keep Solutions](https://github.com/richardlehane/siegfried/pull/201)

### Changed
- update PRONOM to v108
- update tika-mimetype signatures to v2.4.1
- update LOC signatures to 2022-09-01

### Fixed
- incorrect encoding of YAML strings containing line endings; [#202](https://github.com/richardlehane/siegfried/issues/202).
- parse signatures with offsets and offsets in patterns e.g. fmt/1741; [#203](https://github.com/richardlehane/siegfried/issues/203)

### v1.9.4 (2022-07-18)
### Added
- new pkg/static and static builds. This allows direct use of sf API and self-contained binaries without needing separate signature files. 

### Changed
- update PRONOM to v106

### Fixed
- inconsistent output for `roy inspect priorities`. Reported by [Dave Clipsham](https://github.com/richardlehane/siegfried/issues/192)

### v1.9.3 (2022-05-23)
### Added
- JS/WASM build support contributed by [Andy Jackson](https://github.com/richardlehane/siegfried/pull/188)
- wikidata signature added to `-update`. Contributed by [Ross Spencer](https://github.com/richardlehane/siegfried/pull/178)
- `-nopronom` flag added to `roy inspect` subcommand. Contributed by [Ross Spencer](https://github.com/richardlehane/siegfried/pull/185)

### Changed
- update PRONOM to v104
- update LOC signatures to 2022-05-09
- update Wikidata to 2022-05-20
- update tika-mimetypes signatures to v2.4.0
- update freedesktop.org signatures to v2.2

### Fixed
- invalid JSON output for fmt/1472 due to tab in MIME field. Reported by [Robert Schultz](https://github.com/richardlehane/siegfried/issues/186)
- panic on corrupt Zip containers. Reported by [A. Diamond](https://github.com/richardlehane/siegfried/issues/181)

### v1.9.2 (2022-02-07)
### Added
- Wikidata definition file specification has been updated and now includes endpoint (users will need to harvest Wikidata again)
- Custom Wikibase endpoint can now be specified for harvesting when paired with a custom SPARQL query and property mappings
- Wikidata identifier includes permalinks in results
- Wikidata revision history visible using `roy inspect`
- roy inspect returns format ID with name

### Changed
- update PRONOM to v100
- update LOC signatures to 2022-02-01
- update tika-mimetypes signatures to v2.1
- update freedesktop.org signatures to v2.2.1

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
