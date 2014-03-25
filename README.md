# Siegfried

Siegfried is a signature-based file identification tool.

## Version

0.2

[![Build Status](https://travis-ci.org/richardlehane/siegfried.png?branch=master)](https://travis-ci.org/richardlehane/siegfried)

## Usage

### Command line

    ./sieg file.ext
    ./sieg /DIR

## Install

With go installed: 

    go get github.com/richardlehane/siegfried/cmd/sieg

Or download a pre-built binary and signature file:

- [Windows (64)](https://dl.dropboxusercontent.com/u/48160346/Releases/Win64/Siegfried_Win64_0_1.zip)

- [OSX (64)](https://dl.dropboxusercontent.com/u/48160346/Releases/Darwin/Siegfried_OSX64_0_1.zip)

- [Linux (64)](https://dl.dropboxusercontent.com/u/48160346/Releases/Linux/Siegfried_Linux64_0_1.zip)


### Signature file

To run Siegfried, you need to have:

-  an up-to-date signature file (for a description of the Siegfried signature format, see [this wiki page](https://github.com/richardlehane/siegfried/wiki/Siegfried-signature-format)). The pre-built packages come with a signature file built against Droid v73, "pronom.gob". This file should be copied to a "siegfried" directory within your computer's home directory (e.g. ~/siegfried or c:\users\richardl\siegfried). Alternatively, you can provide a different file location as a flag when using the sieg executable.

You can also build your own signature file using the R2D2 tool (it talks to Droid, get it!... and, yes, a protocol droid would have been more appropriate but that name was already taken :().

To do this:

- install r2d2 (go get github.com/richardlehane/cmd/r2d2)

- get recent Droid signature and container files. These are available for download from this page: [http://www.nationalarchives.gov.uk/aboutapps/pronom/droid-signature-files.html](http://www.nationalarchives.gov.uk/aboutapps/pronom/droid-signature-files.html). If these aren't the same as the defaults in the R2D2 tool (use the r2d2 -defaults command to check), you will need to supply the names of your updated files with additional flags to the relevant R2D2 commands

- harvest Pronom reports with R2D2 (./r2d2 -harvest)

- run the R2D2 build command (./r2d2 -build)

## Package Documentation

- [Siegfried package](http://godoc.org/github.com/richardlehane/siegfried/pkg/core)

- [Namematcher package](http://godoc.org/github.com/richardlehane/siegfried/pkg/core/namematcher)

- [Bytematcher package](http://godoc.org/github.com/richardlehane/siegfried/pkg/core/bytematcher)

	- [Patterns package](http://godoc.org/github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns)

	- [Frames package](http://godoc.org/github.com/richardlehane/siegfried/pkg/core/bytematcher/frames)

- [Siegreader package](http://godoc.org/github.com/richardlehane/siegfried/pkg/core/siegreader)

- [Pronom package](http://godoc.org/github.com/richardlehane/siegfried/pkg/pronom)

## Roadmap

### Version 0.3 (April 2014)

- container matching (see github.com/richardlehane/mscfb)

### Version 0.4 (May 2014)

- server mode

### Version 0.5 (July 2014)

- HTML GUI

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
### Version 0.2 (25/03/2014)

- benchmarks (cmd/sieg/testdata)
- an Identifier type that controls the matching process and stops on best possible match (i.e. no longer require a full file scan for all files)
- name/extension matching
- a custom reader (pkg/core/siegreader)
- simplifications to the sieg command and signature file
- optimisations that have boosted performance (see cmd/sieg/testdata/bench_results.txt). But still too slow!

### Version 0.1 (28/02/2014)

First release. Parses PRONOM signatures and performs byte matching. Bare bones CLI. Glacially slow!

## Hacking

Have a peek at upcoming features, planned optimisations and known bugs on the [trello list](https://trello.com/b/ABXkGk6T/siegfried). Please get in touch if you'd like to take any of it on.

## Thanks

Thanks TNA for http://www.nationalarchives.gov.uk/pronom/ and http://www.nationalarchives.gov.uk/information-management/projects-and-work/droid.htm

Thanks Ross for https://github.com/exponential-decay/skeleton-test-suite-generator, very handy!