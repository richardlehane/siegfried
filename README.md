# Siegfried

Siegfried is a signature-based file identification tool.

## Version

0.1

## Install

With go installed: 

    go get github.com/richardlehane/siegfried/cmd/siegfried

    go get github.com/richardlehane/siegfried/cmd/rd2d

Or:

Download this pre-built Win64 package <https://dl.dropboxusercontent.com/u/48160346/Siegfried-0_1-Win64.zip>

Before using Siegfried, you must download current Pronom reports and generate a signature file.

To do this:

- Get Droid signature and container files. These are available for download from this page: http://www.nationalarchives.gov.uk/aboutapps/pronom/droid-signature-files.htm. If these aren't the same as the defaults in the r2d2 tool (use the r2d2 -defaults command to check), you will need to supply the names of your updated files with additional flags to the relevant R2D2 commands

- Harvest Pronom reports with the R2D2 tool (./r2d2 -harvest)

- Build a Siegfried signature file with the R2D2 build command (./r2d2 -build)

## Usage

### Command line

    ./siegfried file.ext
    ./siegfried /DIR

    ./r2d2 -harvest
    ./r2d2 -build 
    ./r2d2 -stats
    ./r2d2 -printdroid
    ./r2d2 -defaults

## TODO

### Version 0.2 (March 2014)

- begin optimising. e.g. replace stop the world, full buffer, recursive frame matching with goroutines, channels and limited buffers
- implement an Identifier type that controls the matching process and stops on best possible match (i.e. no longer require a full file scan for all files)
- name/extension matching

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

### Version 0.1 (28/02/2014)

First release. Parses PRONOM signatures and performs byte matching. Bare bones CLI. Glacially slow!


## Thanks

Thanks TNA for http://www.nationalarchives.gov.uk/pronom/ and http://www.nationalarchives.gov.uk/information-management/projects-and-work/droid.htm

Thanks Ross for https://github.com/exponential-decay/skeleton-test-suite-generator, very handy!