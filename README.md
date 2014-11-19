# Siegfried

[Siegfried](http://www.itforarchivists.com/siegfried) is a signature-based file identification tool.

## Version

0.6.0

[![Build Status](https://travis-ci.org/richardlehane/siegfried.png?branch=master)](https://travis-ci.org/richardlehane/siegfried) [![GoDoc](https://godoc.org/github.com/richardlehane/siegfried/pkg/core?status.svg)](https://godoc.org/github.com/richardlehane/siegfried/pkg/core)

## Usage

### Command line

    sf file.ext
    sf /DIR

![Usage](usage.gif)

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

## Changes
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

### Version 0.5.0 (1/10/2014)
- container matching

### Version 0.4.2 (23/09/2014)
- cross-compile was broken (because of use of os/user). Now doing native builds on the three platforms so the download binaries should all work now.

### Version 0.4.2 (16/09/2014)
- bug in processing code caused really bad matching profile for MP3 sigs. No need to update the tool for this, but please do a sieg -update to get the latest signature file.

### Version 0.4.1 (14/09/2014)
- sf command line: descriptive output in YAML, including basis for matches
- optimisations inc. initial BOF loop before main matching loop

### Version 0.4 (24/08/2014)
- sf command line changes: -version and -update flags now enabled
- over-the-wire updates of signature files from www.itforarchivists.com/siegfried

### Version 0.3 (19/08/2014)
- replaced ac matcher with wac matcher
- re-write of bytematcher code
- some benchmarks slower but fewer really poor edge cases (see cmd/sieg/testdata/bench_results.txt)... so a win!
- but still too slow!

### Version 0.2 (26/03/2014)

- benchmarks (cmd/sieg/testdata)
- an Identifier type that controls the matching process and stops on best possible match (i.e. no longer require a full file scan for all files)
- name/extension matching
- a custom reader (pkg/core/siegreader)
- simplifications to the sieg command and signature file
- optimisations that have boosted performance (see cmd/sieg/testdata/bench_results.txt). But still too slow!

### Version 0.1 (28/02/2014)

First release. Parses PRONOM signatures and performs byte matching. Bare bones CLI. Glacially slow!

## Contributing

Like Siegfried and want to get involved? That'd be fantastic! I've started to jot some notes about contributing on the [wiki](https://github.com/richardlehane/siegfried/wiki).

## Thanks

Thanks TNA for http://www.nationalarchives.gov.uk/pronom/ and http://www.nationalarchives.gov.uk/information-management/projects-and-work/droid.htm

Thanks Ross for https://github.com/exponential-decay/skeleton-test-suite-generator, very handy!

Thanks Misty for the brew and ubuntu packaging