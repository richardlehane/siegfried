# Siegfried

[Siegfried](http://www.itforarchivists.com/siegfried) is a signature-based file format identification tool, implementing:

  - the National Archives UK's [PRONOM](http://www.nationalarchives.gov.uk/pronom) file format signatures
  - freedesktop.org's [MIME-info](https://freedesktop.org/wiki/Software/shared-mime-info/) file format signatures
  - the Library of Congress's [FDD](http://www.digitalpreservation.gov/formats/fdd/descriptions.shtml) file format signatures (*beta*).
  - Wikidata (*beta*).

### Version

1.11.1

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
### v1.11.1 (2023-12-17)
### Added
- glob-matching for container signatures; see [digital-preservation/pronom#10](https://github.com/digital-preservation/pronom/issues/10)
- `sf -update` works for older versions of siegfried; see [#231](https://github.com/richardlehane/siegfried/issues/231)

### Changed
- default location for siegfried HOME now follows XDG Base Directory Specification; see [#216](https://github.com/richardlehane/siegfried/issues/216). Implemented by [Bernhard Hampel-Waffenthal](https://github.com/richardlehane/siegfried/pull/221)
- siegfried prints version before erroring with failed signature load; requested by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/228)
- update PRONOM to v116
- update LOC to 2023-12-14
- update tika-mimetypes to v3.0.0-BETA
- update freedesktop.org to v2.4

### Fixed
- panic on malformed zip file during container matching; reported by [James Mooney](https://github.com/richardlehane/siegfried/issues/238)

### v1.10.2 (2023-12-17)
### Changed
- update PRONOM to v116
- update LOC to 2023-12-14
- update tika-mimetypes to v3.0.0-BETA
- update freedesktop.org to v2.4

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

See the [CHANGELOG](CHANGELOG.md) for the full history.

## Rights

Copyright 2024 Richard Lehane, Ross Spencer 

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
