# Change Log

## [1.6.1] - YYYY-MM-DD
### Added
- Travis and Appveyor CI

### Changed

### Fixed

### Version 1.6.0 (26/6/2016)
- feature: implement Library of Congress FDD signatures (*beta*)
- feature: implement RIFF matcher
- feature: -multi flag replaces -nopriority; based on report by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/75)
- change to -z output: use hash as filepath separator (and unix slash for webarchives); requested by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/81)
- bugfix: parsing fmt/837 signature; reported by [Sarah Romkey](https://github.com/richardlehane/siegfried/issues/80)

### Version 1.5.0 (14/3/2016)
- feature: implement freedesktop.org MIME-info signatures (and the Apache Tika variant)
- feature: implement XML matcher
- feature: file name matcher now supports glob patterns as well as file extensions
- default signature file now "default.sig" (was "pronom.sig")
- changes to YAML and JSON output: "ns" (for namespace) replaces "id", and "id" replaces "puid"
- changes to CSV output: multi-identifiers now displayed in extra columns, not extra rows 

### Version 1.4.5 (6/2/2016)
- bugfix: [big file handling](https://github.com/richardlehane/siegfried/commit/b348c4628ac8edf8e93208e9100bd15616f72e41)
- bugfix: [file handle leak](https://github.com/richardlehane/siegfried/commit/47144fd33a4ddd260bdcd5dd15c132525c3bd113); reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/66)
- bugfix: [mscfb](https://github.com/richardlehane/mscfb/commit/e19fa67f7571388d3dc956f7c6b4547bfb635072); reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/68)
- summarise os errors; requested by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/65)
- code quality: vendor external packages; implemented by [Misty de Meo](https://github.com/richardlehane/siegfried/pull/71)

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

### Version 1.3.0 (27/9/2015)
- scan within warc and arc files with -z flag; reqested by [Dragan Espenschied](https://github.com/richardlehane/siegfried/issues/43)
- [quit scanning earlier on known unknowns](https://github.com/richardlehane/siegfried/commit/f7fedf6b629048e1c41a694f4428e94deeffd3ee)
- don't include byte signatures where formats have container signatures (unless -doubleup flag is given); fixes a mis-identification reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/52)
- sf -slow FILE | DIR reports slow signatures
- sf -debug output simplified
- sf -version describes signature file; requested by [Michelle Lindlar](https://github.com/richardlehane/siegfried/issues/54)
- roy -limit and -exclude now operate on text and default zip matches
- roy -nopriority re-configured to return more results
- bugfix: upgraded versions of sf panic when attempting to read old signature files; reported by [Stefan](https://github.com/richardlehane/siegfried/issues/49) 
- bugfix: panic mmap'ing files over 1GB on Win32; reported by [Duncan](https://github.com/richardlehane/siegfried/issues/50) 
- bugfix: reporting extensions for folders with "."; reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/51)

### Version 1.2.2 (15/8/2015)
- add -noext flag to roy to suppress extension matching; requested by [Greg Lepore](https://github.com/richardlehane/siegfried/issues/46)
- -known and -unknown flags for sf to output lists of recognised and unknown files respectively; requested by [Greg Lepore](https://github.com/richardlehane/siegfried/issues/47)

### Version 1.2.1 (11/8/2015)
- support annotation of sets.json files; requested by Greg Lepore
- add warning when use -extendc without -extend
- bugfix: report container extensions in details; reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/48)

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
- bugfix: ([mscfb](https://github.com/richardlehane/mscfb/commit/22552265cefc80b400ff64156155f53a5d5751e6)) reported by [Pascal Aantz](https://github.com/richardlehane/siegfried/issues/32)
- bugfix: race condition in scorer (affected tip golang)

### Version 1.0.0 (22/3/2015)
- [user documentation](http://github.com/richardlehane/siegfried/wiki)
- bugfixes (mscfb, match/wac and sf)
- QA using [comparator](http://github.com/richardlehane/comparator)

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
