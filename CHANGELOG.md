# Change Log
## v1.7.0 (2017-02-19)
### Added
- log files that match particular formats with `-log fmt/1,@set2` (comma separated list of format IDs/format sets). These can be mixed with regular log options e.g. `-log unknown,fmt/1,chart`
- generate a summary view of formats matched during a scan with `-log chart` (or just `-log c`)
- replay scans from results files with `sf -r`: load one or more results files to replay logging or to convert to a different output format e.g. `sf -r -csv results.yaml` or `sf -r -log unknown,chart,stdout results1.yaml results2.csv`
- compare results with `roy compare` subcommand: view the difference between two or more results e.g. `roy compare results1.yaml results2.csv droid.csv ...`
- `roy sets` subcommand: `roy sets` creates pronom-all.json, pronom-families.json, and pronom-types.json sets files;
`roy sets -changes` creates a pronom-changes.json sets file from a PRONOM release-notes.xml file; `roy sets -list @set1,@set2` lists contents of a comma-separated list of format sets
- `roy inspect releases` provides a summary view of a PRONOM release-notes.xml file

## Changed
- the `sf -` command now scans stdin e.g. `cat mypdf.pdf | sf -`. You can pass a filename in to supplement the analysis with the `-name` flag. E.g. `cat myfile.pdf | sf -name myfile.pdf -`. In previous versions of sf, the dash argument signified treating stdin as a newline separated list of filenames for scanning. Use the new `-f` flag for this e.g. `sf -f myfiles.txt` or `cat myfiles.txt | sf -f -`; change requested by [pm64](https://github.com/richardlehane/siegfried/issues/96)

### Fixed
- some files cause endless scanning due to large numbers of signature hits; reported by [workflowsguy](https://github.com/richardlehane/siegfried/issues/94)
- null bytes can be written to output due to bad zip filename decoding; reported by [Tim Walsh](https://github.com/richardlehane/siegfried/issues/95)

## v1.6.7 (2016-11-23)
### Added
- enable -hash, -z, and -log flags for -serve and -multi modes
- new hash, z, and sig params for -serve mode (to control per-request)
- enable droid output in -serve mode
- GET requests in -serve mode now just percent encoded (with base64 option as a param)
- -serve mode landing page now includes example forms

### Changed
- code re-organisation using /internal directory to hide internal packages
- Identify method now returns a slice rather than channel of IDs (siegfried pkg change)

## v1.6.6 (2016-10-25)
### Added
- graph implicit and missing priorities with `roy inspect implicit-priorities` and `roy inspect missing-priorities`

### Fixed
- error parsing mimeinfo signatures with double backslashes (e.g. rtf signatures)

## v1.6.5 (2016-09-28)
### Added
- new sets files (pronom-families.json and pronom-types) automatically created from PRONOM classficiations. Removed redundant sets (database, audio, etc.).

### Fixed
- debbuilder.sh fix: debian packages were copying roy data to wrong directory

### Changed
- roy inspect priorities command now includes "orphan" fmts in graphs
- update PRONOM urls from apps. to www.

## v1.6.4 (2016-09-05)
### Added
- roy inspect FMT command now inspects sets e.g. roy inspect @pdfa
- roy inspect priorities command generates graphs of priority relations

### Fixed
- [container matcher running when empty](https://github.com/richardlehane/siegfried/issues/90) (i.e. for freedesktop/tika signature files and when -nocontainer flag used with PRONOM)
- [-doubleup flag preventing signature extensions loading](https://github.com/richardlehane/siegfried/issues/92): since v1.3.0 signature extensions included with the -extend flag haven't been loading properly due to interaction with the doubles filter (which prevents byte signatures loading for formats that also have container signatures defined)

### Changed
- use fwac rather than wac package for performance
- roy inspect FMT command speed up by building without reports and without the doubles filter
- -reports flag removed for roy harvest and roy build commands
- -reports flag changed for roy inspect command, now a boolean that, if set, will cause the signature(s) to be built from the PRONOM report(s), rather than the DROID XML file. This is slower but can be a more accurate representation.

## v1.6.3 (2016-08-18)
### Added
- roy inspect FMT command now gives details of all signatures, [including container signatures](https://github.com/richardlehane/siegfried/issues/88)

### Fixed
- misidentification: [x-fmt/45 files misidentified as fmt/40](https://github.com/richardlehane/siegfried/issues/89) due to repetition of elements in container file
- roy build -noreports includes blank extensions that generate false matches; reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/87)

## v1.6.2 (2016-08-08)
### Fixed
- poor performance unknowns due to interaction of -bof/-eof flags with known BOF/EOF calculation; reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/86)
- [unnecessary warnings for mimeinfo identifier](https://github.com/richardlehane/siegfried/issues/84)
- add fddXML.zip to .gitattributes to preserve newlines
- various [Go Report Card](https://goreportcard.com/report/github.com/richardlehane/siegfried) issues

## v1.6.1 (2016-07-06)
### Added
- Travis and Appveyor CI automated deployment to Github releases and Bintray
- PRONOM v85 signatures
- LICENSE.txt, CHANGELOG.md
- [Go Report Card](https://goreportcard.com/report/github.com/richardlehane/siegfried)

### Fixed
- golang.org/x/image/riff bug (reported [here](https://github.com/golang/go/issues/16236))
- misspellings reported by Go Report Card
- ineffectual assignments reported by Go Report Card

## v1.6.0 (2016-06-26)
### Added
- implement Library of Congress FDD signatures (*beta*)
- implement RIFF matcher
- -multi flag replaces -nopriority; based on report by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/75)

### Changed
- change to -z output: use hash as filepath separator (and unix slash for webarchives); requested by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/81)

### Fixed
- parsing fmt/837 signature; reported by [Sarah Romkey](https://github.com/richardlehane/siegfried/issues/80)

## v1.5.0 (2016-03-14)
### Added
- implement freedesktop.org MIME-info signatures (and the Apache Tika variant)
- implement XML matcher
- file name matcher now supports glob patterns as well as file extensions

### Changed
- default signature file now "default.sig" (was "pronom.sig")
- changes to YAML and JSON output: "ns" (for namespace) replaces "id", and "id" replaces "puid"
- changes to CSV output: multi-identifiers now displayed in extra columns, not extra rows 

## v1.4.5 (2016-02-06)
### Added
- summarise os errors; requested by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/65)
- code quality: vendor external packages; implemented by [Misty de Meo](https://github.com/richardlehane/siegfried/pull/71)

### Fixed
- [big file handling](https://github.com/richardlehane/siegfried/commit/b348c4628ac8edf8e93208e9100bd15616f72e41)
- [file handle leak](https://github.com/richardlehane/siegfried/commit/47144fd33a4ddd260bdcd5dd15c132525c3bd113); reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/66)
- [mscfb](https://github.com/richardlehane/mscfb/commit/e19fa67f7571388d3dc956f7c6b4547bfb635072); reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/68)

## v1.4.4 (2016-01-09)
### Changed
- code quality: refactor textmatcher package
- code quality: refactor siegreader package
- code quality: documentation

### Fixed
- speed regression in TIFF mis-identification patch last release

## v1.4.3 (2015-12-19)
### Added
- measure time elapsed with -log time

### Fixed
- [percent encode file URIs in droid output](https://github.com/richardlehane/siegfried/issues/63)
- long windows directory paths (further work on bug fixed in 1.4.2); reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/58)
- mscfb panic; reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/62)
- **TIFF mis-identifications** due to an [early halt error](https://github.com/richardlehane/siegfried/commit/5f0ccd477c467186c350e762f8fddda888d987bf)

## v1.4.2 (2015-11-27)
### Added
- new -throttle flag; requested by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/61)

### Changed
- errors logged to stderr by default (to quieten use -log ""); requested by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/60)
- mscfb update: [lazy reading](https://github.com/richardlehane/mscfb/commit/f909cfa596c7880c650ed5440df90e5474f08b29) 
- webarchive update: [decode Transfer-Encoding and Content-Encoding](https://github.com/richardlehane/webarchive/commit/2f125b9bece4d7d119ea029aa8c942a41962ecf4); requested by [Dragan Espenschied](https://github.com/richardlehane/siegfried/issues/55)

### Fixed
- long windows paths; reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/58)
- 32-bit file size overflow; reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/59)

## v1.4.1 (2015-11-06)
### Changed
- **-log replaces -debug, -slow, -unknown and -known flags** (see usage above)
- highlight empty file/stream with error and warning
- negative text match overrides extension-only plain text match

## v1.4.0 (2015-10-31)
### Added
- new MIME matcher; requested by [Dragan Espenschied](https://github.com/richardlehane/siegfried/issues/55)
- support warc continuations
- add all.json and tiff.json sets

### Changed
- minor speed-up
- report less redundant basis information
- report error on empty file/stream

## v1.3.0 (2015-09-27)
### Added
- scan within warc and arc files with -z flag; reqested by [Dragan Espenschied](https://github.com/richardlehane/siegfried/issues/43)
- sf -slow FILE | DIR reports slow signatures
- sf -version describes signature file; requested by [Michelle Lindlar](https://github.com/richardlehane/siegfried/issues/54)

### Changed
- [quit scanning earlier on known unknowns](https://github.com/richardlehane/siegfried/commit/f7fedf6b629048e1c41a694f4428e94deeffd3ee)
- don't include byte signatures where formats have container signatures (unless -doubleup flag is given); fixes a mis-identification reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/52)
- sf -debug output simplified
- roy -limit and -exclude now operate on text and default zip matches
- roy -nopriority re-configured to return more results

### Fixed
- upgraded versions of sf panic when attempting to read old signature files; reported by [Stefan](https://github.com/richardlehane/siegfried/issues/49) 
- panic mmap'ing files over 1GB on Win32; reported by [Duncan](https://github.com/richardlehane/siegfried/issues/50) 
- reporting extensions for folders with "."; reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/51)

## v1.2.2 (2015-08-15)
### Added
- -noext flag to roy to suppress extension matching; requested by [Greg Lepore](https://github.com/richardlehane/siegfried/issues/46)
- -known and -unknown flags for sf to output lists of recognised and unknown files respectively; requested by [Greg Lepore](https://github.com/richardlehane/siegfried/issues/47)

## v1.2.1 (2015-08-11)
### Added
- support annotation of sets.json files; requested by Greg Lepore
- add warning when use -extendc without -extend

### Fixed
- report container extensions in details; reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/48)

## v1.2.0 (2015-07-31)
### Added
- text matcher (i.e. sf README will now report a 'Plain Text File' result)
- -notext flag to suppress text matcher (roy build -notext)
- all outputs now include file last modified time
- -hash flag with choice of md5, sha1, sha256, sha512, crc (e.g. sf -hash md5 FILE)
- -droid flag to mimic droid output (sf -droid FILE)

### Fixed
- [detect encoding of zip filenames](https://github.com/richardlehane/siegfried/commit/0c92c52d3d709e1a9b2822fa182ebd1847a6c394) reported by [Dragan Espenschied](https://github.com/richardlehane/siegfried/issues/42)
- [mscfb](https://github.com/richardlehane/mscfb/commit/f790430b648469e862b40f599171e361e30442e7) reported by [Dragan Espenschied](https://github.com/richardlehane/siegfried/issues/41)

## v1.1.0 (2015-05-17)
### Added
- scan within archive formats (zip, tar, gzip) with -z flag
- format sets (e.g. roy build -exclude @pdfa)
- support bitmask patterns

### Changed
- leaner, faster signature format
- mirror bof patterns as eof patterns where both roy -bof and -eof limits set

### Fixed
- ([mscfb](https://github.com/richardlehane/mscfb/commit/22552265cefc80b400ff64156155f53a5d5751e6)) reported by [Pascal Aantz](https://github.com/richardlehane/siegfried/issues/32)
- race condition in scorer (affected tip golang)

## v1.0.0 (2015-03-22)
### Changed
- [user documentation](http://github.com/richardlehane/siegfried/wiki)
- bugfixes (mscfb, match/wac and sf)
- QA using [comparator](http://github.com/richardlehane/comparator)

## v0.8.2 (2015-02-22)
### Added
- json output
- server mode

## v0.8.1 (2015-02-01)
### Fixed
- single quote YAML output

## v0.8.0 (2015-01-26)
### Changed
- optimisations (mmap, multithread, etc.)

## v0.7.1 (2014-12-09)
### Added
- csv output

### Changed
- periodic priority checking to stop searches earlier

### Fixed
- range/distance/choices bugfix

## v0.7.0 (2014-11-24)
### Changed
- change to signature file format

## v0.6.1 (2014-11-21)
### Added
- roy (r2d2 rename) signature customisation
- parse Droid signature (not just PRONOM reports)
- support extension signatures

## v0.6.0 (2014-11-11)
### Added
- support multiple identifiers
- config package

### Changed
- license info in srcs (no change to license; this allows for attributing authorship for non-Richard contribs)
- default home change to "$HOME/siegfried" (no longer ".siegfried")

### Fixed
- mscfb bugfixes

## v0.5.0 (2014-10-01)
### Added
- container matching

## v0.4.3 (2014-09-23)
### Fixed
- cross-compile was broken (because of use of os/user). Now doing native builds on the three platforms so the download binaries should all work now.

## v0.4.2 (2014-09-16)
### Fixed
- bug in processing code caused really bad matching profile for MP3 sigs. No need to update the tool for this, but please do a sieg -update to get the latest signature file.

## v0.4.1 (2014-09-14)
### Added
- sf command line: descriptive output in YAML, including basis for matches

### Changed
- optimisations inc. initial BOF loop before main matching loop

## v0.4.0 (2014-08-24)
### Added
- sf command line changes: -version and -update flags now enabled
- over-the-wire updates of signature files from www.itforarchivists.com/siegfried

## v0.3.0 (2014-08-19)
### Changed
- replaced ac matcher with wac matcher
- re-write of bytematcher code
- some benchmarks slower but fewer really poor edge cases (see cmd/sieg/testdata/bench_results.txt)... so a win!
- but still too slow!

## v0.2.0 (2014-03-26)
### Added
- an Identifier type that controls the matching process and stops on best possible match (i.e. no longer require a full file scan for all files)
- name/extension matching
- a custom reader (pkg/core/siegreader)

### Changed
- benchmarks (cmd/sieg/testdata)
- simplifications to the sieg command and signature file
- optimisations that have boosted performance (see cmd/sieg/testdata/bench_results.txt). But still too slow!

## v0.1.0 (2014-02-28)
### Added
- First release. Parses PRONOM signatures and performs byte matching. Bare bones CLI. Glacially slow!
