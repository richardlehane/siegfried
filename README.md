# Siegfried

[Siegfried](http://www.itforarchivists.com/siegfried) is a signature-based file identification tool.

## Version

0.6.1

[![Build Status](https://travis-ci.org/richardlehane/siegfried.png?branch=master)](https://travis-ci.org/richardlehane/siegfried) [![GoDoc](https://godoc.org/github.com/richardlehane/siegfried/pkg/core?status.svg)](https://godoc.org/github.com/richardlehane/siegfried/pkg/core)

## Usage

### Command line

    sf file.ext
    sf /DIR

![Usage](usage.gif)

By default, siegfried uses latest DROID and container signature with no buffer limits. You can customise your signature file by using the [roy tool](https://github.com/richardlehane/siegfried/wiki/Building-a-signature-file-with-ROY).

## Install

### With go installed: 

    go get github.com/richardlehane/siegfried/cmd/sf

    sf -update


### Or, without go installed:

For OS X:

    brew install mistydemeo/digipres/siegfried

For Ubuntu:

    sudo add-apt-repository ppa:archivematica/externals-dev
    sudo apt-get update
    sudo apt-get install siegfried

For Win:

Download a pre-built binary from the [releases page](https://github.com/richardlehane/siegfried/releases). Unzip to a location in your system path. Then run:

	sf -update

## Roadmap

### Road to 1.0 (early 2015)

- optimisations (load time, mmap, multi-thread)
- additional documentation & tests
- server mode

## Rights

Copyright 2014 Richard Lehane 

Licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

## Recent Changes
### Version 0.6.1 (21/11/2014)
- roy (r2d2 rename) signature customisation
- parse Droid signature (not just PRONOM reports)
- support extension signatures

### Version 0.6.0 (11/11/2014)
- support multiple identifiers
- config package
- mscfb bugfixes
- license info in srcs (no change to license; this allows for attributing authorship for non-Richard contribs)
- default home change to "$HOME/siegfried" (no longer ".siegfried")

## Contributing

Like Siegfried and want to get involved? That'd be fantastic! I've started to jot some notes about contributing on the [wiki](https://github.com/richardlehane/siegfried/wiki).

## Thanks

Thanks TNA for http://www.nationalarchives.gov.uk/pronom/ and http://www.nationalarchives.gov.uk/information-management/projects-and-work/droid.htm

Thanks Ross for https://github.com/exponential-decay/skeleton-test-suite-generator, very handy!

Thanks Misty for the brew and ubuntu packaging
