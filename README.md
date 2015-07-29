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

1.2.0

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
    sf -z file.zip | DIR                       // Decompress and scan zip, tar, gzip
    sf -hash md5                               // Calculate md5, sha1, sha256, sha512, or crc hash
    sf -sig custom.sig file.ext                // Use a custom signature file
    sf -home c:\junk -sig custom.sig file.ext  // Use a custom home directory
    sf -debug file.ext                         // Scan in debug mode
    sf -version                                // Display version information
    sf -serve hostname:port                    // Server mode


![Usage](usage.gif)

By default, siegfried uses the latest PRONOM and container signatures with no buffer limits. You can customise your signature file by using the [roy tool](https://github.com/richardlehane/siegfried/wiki/Building-a-signature-file-with-ROY).

## Install

### With go installed: 

    go get github.com/richardlehane/siegfried/cmd/sf

    sf -update


### Or, without go installed:

For OS X:

    brew install mistydemeo/digipres/siegfried

For Ubuntu/Debian:

    wget -qO - https://bintray.com/user/downloadSubjectPublicKey?username=bintray | sudo apt-key add -
    echo "deb http://dl.bintray.com/siegfried/debian wheezy main" | sudo tee -a /etc/apt/sources.list
    sudo apt-get update && sudo apt-get install siegfried

For Win:

Download a pre-built binary from the [releases page](https://github.com/richardlehane/siegfried/releases). Unzip to a location in your system path. Then run:

	sf -update

## Recent Changes
### Version 1.2.0 (31/7/2015)
- text matcher (i.e. sf README will now report a 'Plain Text File' result)
- -notext flag to suppress text matcher (roy build -notext)
- all outputs now include file last modified time
- -hash flag with choice of md5, sha1, sha256, sha512, crc (e.g. sf -hash md5 FILE)
- -droid flag to mimic droid output (sf -droid FILE)
- bugfix: [detect encoding of zip filenames](https://github.com/richardlehane/siegfried/commit/0c92c52d3d709e1a9b2822fa182ebd1847a6c394) reported by [Dragan Espenschied](https://github.com/richardlehane/siegfried/issues/42)
- bugfix: [mscfb](https://github.com/richardlehane/mscfb/commit/f790430b648469e862b40f599171e361e30442e7) reported by [Dragan Espenschied](https://github.com/richardlehane/siegfried/issues/41)

### Version 1.1.0 (17/5/2015)
- scan within archive formats (zip, tar, gzip) with -z flag
- format sets (e.g. roy build -exclude @pdfa)
- leaner, faster signature format
- support bitmask patterns
- mirror bof patterns as eof patterns where both roy -bof and -eof limits set
- 'sf -' reads files piped to stdin
- bugfix: [mscfb](https://github.com/richardlehane/mscfb/commit/22552265cefc80b400ff64156155f53a5d5751e6) reported by [Pascal Aantz](https://github.com/richardlehane/siegfried/issues/32)
- bugfix: race condition in scorer (affected tip golang)
- archivematica build: fpr server

### Version 1.0.0 (22/3/2015)
- [user documentation](http://github.com/richardlehane/siegfried/wiki)
- bugfixes (mscfb, match/wac and sf)
- QA using [comparator](http://github.com/richardlehane/comparator)

[Full change history](https://github.com/richardlehane/siegfried/wiki/Change-history)

## Rights

Copyright 2015 Richard Lehane 

Licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

## Contributing

Like siegfried and want to get involved in its development? That'd be wonderful! There are some notes on the [wiki](https://github.com/richardlehane/siegfried/wiki) to get you started, and please get in touch.

## Thanks

Thanks TNA for http://www.nationalarchives.gov.uk/pronom/ and http://www.nationalarchives.gov.uk/information-management/projects-and-work/droid.htm

Thanks Ross for https://github.com/exponential-decay/skeleton-test-suite-generator and http://exponentialdecay.co.uk/sd/index.htm, both are very handy!

Thanks Misty for the brew and ubuntu packaging
