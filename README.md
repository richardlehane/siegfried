# Siegfried

[Siegfried](http://www.itforarchivists.com/siegfried) is a signature-based file identification tool.

Key features are:

  - implements [PRONOM signatures](http://www.nationalarchives.gov.uk/aboutapps/pronom/)
  - simple command line interface
  - decent speeds without limiting the number of bytes scanned

## Version

0.8.2

[![Build Status](https://travis-ci.org/richardlehane/siegfried.png?branch=master)](https://travis-ci.org/richardlehane/siegfried) [![GoDoc](https://godoc.org/github.com/richardlehane/siegfried/pkg/core?status.svg)](https://godoc.org/github.com/richardlehane/siegfried/pkg/core)

## Usage

### Command line

    sf file.ext
    sf DIR

#### Options

    sf -csv file.ext                           // Output CSV rather than YAML
    sf -json file.ext                          // Output JSON rather than YAML
    sf -nr DIR                                 // Don't scan subdirectories
    sf -sig custom.gob file.ext                // Use a custom signature file
    sf -home c:\junk -sig custom.gob file.ext  // Use a custom home directory
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

For Ubuntu:

    sudo add-apt-repository ppa:archivematica/externals-dev
    sudo apt-get update
    sudo apt-get install siegfried

For Win:

Download a pre-built binary from the [releases page](https://github.com/richardlehane/siegfried/releases). Unzip to a location in your system path. Then run:

	sf -update

## Roadmap

### Road to 1.0 (March 2015)

- additional documentation & tests

## Recent Changes
### Version 0.8.2 (22/2/2015)
- json output
- server mode

### Version 0.8.1 (1/2/2015)
- bugfix: single quote YAML output

### Version 0.8.0 (26/1/2015)
- optimisations (mmap, multithread, etc.)

### Version 0.7.1 (9/12/2014)
- csv output
- periodic priority checking to stop searches earlier
- range/distance/choices bugfix

### Version 0.7.0 (24/11/2014)
- change to signature file format

[Full change history](https://github.com/richardlehane/siegfried/wiki/Change-history)

## Rights

Copyright 2014 Richard Lehane 

Licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

## Contributing

Like Siegfried and want to be involved in its development? That'd be wonderful! There are some notes on the [wiki](https://github.com/richardlehane/siegfried/wiki) to get you started, and please get in touch.

## Thanks

Thanks TNA for http://www.nationalarchives.gov.uk/pronom/ and http://www.nationalarchives.gov.uk/information-management/projects-and-work/droid.htm

Thanks Ross for https://github.com/exponential-decay/skeleton-test-suite-generator and http://exponentialdecay.co.uk/sd/index.htm, both are very handy!

Thanks Misty for the brew and ubuntu packaging
