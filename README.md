# Siegfried

[Siegfried](http://www.itforarchivists.com/siegfried) is a signature-based file identification tool.

## Version

0.4.2

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

Download a pre-built binary from the [releases page](https://github.com/richardlehane/siegfried/releases). Unzip to a location in your system path. Then run:

	sf -update

## Roadmap

### Next up: version 0.5 (September 2014)

- container matching (see [https://www.github.com/richardlehane/mscfb](https://www.github.com/richardlehane/mscfb))

### Thereafter...

- server mode

## Rights

Copyright 2014 Richard Lehane. 

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

## Changes
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

## Hacking

Have a peek at upcoming features, planned optimisations and known bugs on the [trello list](https://trello.com/b/ABXkGk6T/siegfried).

## Thanks

Thanks TNA for http://www.nationalarchives.gov.uk/pronom/ and http://www.nationalarchives.gov.uk/information-management/projects-and-work/droid.htm

Thanks Ross for https://github.com/exponential-decay/skeleton-test-suite-generator, very handy!