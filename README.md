# Siegfried

[Siegfried](http://www.itforarchivists.com/siegfried) is a signature-based file format identification tool.

Key features are:

  - complete implementation of [PRONOM](http://apps.nationalarchives.gov.uk/pronom) (byte and container signatures)
  - fast matching without limiting the number of bytes scanned
  - detailed information about the basis for format matches
  - simple command line interface with a choice of outputs
  - a [built-in server](https://github.com/richardlehane/siegfried/wiki/Using-the-siegfried-server) for integrating with workflows and language inter-op
  - power options including [debug mode](https://github.com/richardlehane/siegfried/wiki/Inspect-and-Debug), [signature modification](https://github.com/richardlehane/siegfried/wiki/Building-a-signature-file-with-ROY), and [multiple identifiers](https://github.com/richardlehane/siegfried/wiki/Building-a-signature-file-with-ROY#one-signature-file-multiple-identifiers)

## Version

1.4.4

[![Build Status](https://travis-ci.org/richardlehane/siegfried.png?branch=master)](https://travis-ci.org/richardlehane/siegfried) [![GoDoc](https://godoc.org/github.com/richardlehane/siegfried?status.svg)](https://godoc.org/github.com/richardlehane/siegfried)

## Usage

### Command line

    sf file.ext
    sf DIR

#### Options

    sf -csv file.ext | DIR                     // Output CSV rather than YAML
    sf -json file.ext | DIR                    // Output JSON rather than YAML
    sf -droid file.ext | DIR                   // Output DROID CSV rather than YAML
    sf -                                       // Read list of files piped to stdin
    sf -nr DIR                                 // Don't scan subdirectories
    sf -z file.zip | DIR                       // Decompress and scan zip, tar, gzip, warc, arc
    sf -hash md5 file.ext | DIR                // Calculate md5, sha1, sha256, sha512, or crc hash
    sf -sig custom.sig file.ext                // Use a custom signature file
    sf -home c:\junk -sig custom.sig file.ext  // Use a custom home directory
    sf -serve hostname:port                    // Server mode
    sf -version                                // Display version information
    sf -throttle 10ms DIR                      // Pause for duration (e.g. 1s) between file scans
    sf -log [comma-sep opts] file.ext | DIR    // Log errors etc. to stderr (default) or stdout
    sf -log e,w file.ext | DIR                 // Log errors and warnings to stderr
    sf -log u,o file.ext | DIR                 // Log unknowns to stdout
    sf -log d,s file.ext | DIR                 // Log debugging and slow messages to stderr
    sf -log p,t DIR > results.yaml             // Log progress and time while redirecting results


![Usage](usage.gif)

By default, siegfried uses the latest PRONOM and container signatures with no buffer limits. You can customise your signature file by using the [roy tool](https://github.com/richardlehane/siegfried/wiki/Building-a-signature-file-with-ROY).

## Install

### With go installed: 

    go get github.com/richardlehane/siegfried/cmd/sf

    sf -update


### Or, without go installed:
#### Win:

Download a pre-built binary from the [releases page](https://github.com/richardlehane/siegfried/releases). Unzip to a location in your system path. Then run:

    sf -update

#### Mac Homebrew (or [Linuxbrew](http://brew.sh/linuxbrew/)):

    brew install mistydemeo/digipres/siegfried

#### Ubuntu/Debian (64 bit):

    wget -qO - https://bintray.com/user/downloadSubjectPublicKey?username=bintray | sudo apt-key add -
    echo "deb http://dl.bintray.com/siegfried/debian wheezy main" | sudo tee -a /etc/apt/sources.list
    sudo apt-get update && sudo apt-get install siegfried


## Recent Changes
### Version 1.4.4 (9/1/2016)
- fix: speed regression in TIFF mis-identification patch last release
- code quality: refactor textmatcher package
- code quality: refactor siegreader package
- code quality: documentation

### Version 1.4.3 (19/12/2015)
- measure time elapsed with -log time
- bugfix: [percent encode file URIs in droid output](https://github.com/richardlehane/siegfried/issues/63)
- bugfix: long windows directory paths (further work on bug fixed in 1.4.2); reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/58)
- bugfix: mscfb panic; reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/62)
- bugfix: **TIFF mis-identifications** due to an [early halt error](https://github.com/richardlehane/siegfried/commit/5f0ccd477c467186c350e762f8fddda888d987bf)

### Version 1.4.2 (27/11/2015)
- new -throttle flag; requested by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/61)
- errors logged to stderr by default (to quieten use -log ""); requested by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/60)
- mscfb update: [lazy reading](https://github.com/richardlehane/mscfb/commit/f909cfa596c7880c650ed5440df90e5474f08b29) 
- webarchive update: [decode Transfer-Encoding and Content-Encoding](https://github.com/richardlehane/webarchive/commit/2f125b9bece4d7d119ea029aa8c942a41962ecf4); requested by [Dragan Espenschied](https://github.com/richardlehane/siegfried/issues/55)
- bugfix: long windows paths; reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/58)
- bugfix: 32-bit file size overflow; reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/59)

### Version 1.4.1 (6/11/2015)
- **-log replaces -debug, -slow, -unknown and -known flags** (see usage above)
- highlight empty file/stream with error and warning
- negative text match overrides extension-only plain text match

### Version 1.4.0 (31/10/2015)
- new MIME matcher; requested by [Dragan Espenschied](https://github.com/richardlehane/siegfried/issues/55)
- support warc continuations
- add all.json and tiff.json sets
- minor speed-up
- report less redundant basis information
- report error on empty file/stream

[Full change history](https://github.com/richardlehane/siegfried/wiki/Change-history)

## Rights

Copyright 2016 Richard Lehane 

Licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

## Contributing

Like siegfried and want to get involved in its development? That'd be wonderful! There are some notes on the [wiki](https://github.com/richardlehane/siegfried/wiki) to get you started, and please get in touch.

## Thanks

Thanks TNA for http://www.nationalarchives.gov.uk/pronom/ and http://www.nationalarchives.gov.uk/information-management/projects-and-work/droid.htm

Thanks Ross for https://github.com/exponential-decay/skeleton-test-suite-generator and http://exponentialdecay.co.uk/sd/index.htm, both are very handy!

Thanks Misty for the brew and ubuntu packaging
