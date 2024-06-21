# Change Log
## v1.11.1 (2024-06-28)
### Added
- WASM build. See pkg/wasm/README.md for more details. Feature sponsored by Archives New Zealand. Inspired by [Andy Jackson](https://siegfried-js.glitch.me/)
- `-sym` flag enables following symbolic links to files during scanning. Requested by [Max Moser](https://github.com/richardlehane/siegfried/issues/245) 

### Changed
- XDG_DATA_DIRS checked when determining siegfried home location. Requested by [Michał Górny](https://github.com/mgorny)
- Windows 7 build on [releases page](https://github.com/richardlehane/siegfried/releases) (built with go 1.20). Requested by [Aleksandr Sergeev](https://github.com/richardlehane/siegfried/issues/240)
- update PRONOM to v118
- update LOC to 2024-06-14

### Fixed
- zips piped into STDIN are decompressed with `-z` flag. Reported by [Max Moser](https://github.com/richardlehane/siegfried/issues/244)
- panics from OS calls in init functions. Reported by [Jürgen Enge](https://github.com/richardlehane/siegfried/issues/247)

## v1.11.0 (2023-12-17)
### Added
- glob-matching for container signatures; see [digital-preservation/pronom#10](https://github.com/digital-preservation/pronom/issues/10)
- `sf -update` requires less updating of siegfried; see [#231](https://github.com/richardlehane/siegfried/issues/231)

### Changed
- default location for siegfried HOME now follows XDG Base Directory Specification; see [#216](https://github.com/richardlehane/siegfried/issues/216). Implemented by [Bernhard Hampel-Waffenthal](https://github.com/richardlehane/siegfried/pull/221)
- siegfried prints version before erroring with failed signature load; requested by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/228)
- update PRONOM to v116
- update LOC to 2023-12-14
- update tika-mimetypes to v3.0.0-BETA
- update freedesktop.org to v2.4

### Fixed
- panic on malformed zip file during container matching; reported by [James Mooney](https://github.com/richardlehane/siegfried/issues/238)

## v1.10.2 (2023-12-17)
### Changed
- update PRONOM to v116
- update LOC to 2023-12-14
- update tika-mimetypes to v3.0.0-BETA
- update freedesktop.org to v2.4

## v1.10.1 (2023-04-24)
### Fixed
- glob expansion now only on Windows & when no explicit path match. Implemented by [Bernhard Hampel-Waffenthal](https://github.com/richardlehane/siegfried/pull/229)
- compression algorithm for debian packages changed back to xz. Implemented by [Paul Millar](https://github.com/richardlehane/siegfried/pull/230)
- `-multi droid` setting returned empty results when priority lists contained self-references. See [#218](https://github.com/richardlehane/siegfried/issues/218)
- CGO disabled for debian package and linux binaries. See [#219](https://github.com/richardlehane/siegfried/issues/219)

## v1.10.0 (2023-03-25)
### Added
- format classification included as "class" field in PRONOM results. Requested by [Robin François](https://github.com/richardlehane/siegfried/discussions/207). Implemented by [Ross Spencer](https://github.com/richardlehane/siegfried/commit/7f695720a752ac5fca3e1de8ba034b92ab6da1d9)
- `-noclass` flag added to roy build command. Use this flag to build signatures that omit the new "class" field from results.
- glob paths can be used in place of file or directory paths for identification (e.g. `sf *.jpg`). Implemented by [Ross Spencer](https://github.com/richardlehane/siegfried/commit/54bf6596c5fe7d1c9858348f0170d0dd7365fc8f)
- `-multi droid` setting for roy build command. Applies priorities after rather than during identificaiton for more DROID-like results. Reported by [David Clipsham](https://github.com/richardlehane/siegfried/issues/146)
- `/update` command for server mode. Requested by [Luis Faria](https://github.com/richardlehane/siegfried/issues/208)

### Changed
- new algorithm for dynamic multi-sequence matching for improved wildcard performance
- update PRONOM to v111
- update LOC to 2023-01-27
- update tika-mimetypes to v2.7.0
- minimum go version to build siegfried is now 1.18 

### Fixed
- archivematica extensions built into wikidata signatures. Reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/210)
- trailing slash for folder paths in URI field in droid output. Reported by Philipp Wittwer
- crash when using `sf -replay` with droid output

## v1.9.6 (2022-11-06)
### Changed
- update PRONOM to v109

## v1.9.5 (2022-09-12)
### Added
- `roy inspect` now takes a `-droid` flag to allow easier inspection of old or custom DROID files
- github action to update siegfried docker deployment [https://github.com/keeps/siegfried-docker]. Implemented by [Keep Solutions](https://github.com/richardlehane/siegfried/pull/201)

### Changed
- update PRONOM to v108
- update tika-mimetype signatures to v2.4.1
- update LOC signatures to 2022-09-01

### Fixed
- incorrect encoding of YAML strings containing line endings; [#202](https://github.com/richardlehane/siegfried/issues/202).
- parse signatures with offsets and offsets in patterns e.g. fmt/1741; [#203](https://github.com/richardlehane/siegfried/issues/203)

## v1.9.4 (2022-07-18)
### Added
- new pkg/static and static builds. This allows direct use of sf API and self-contained binaries without needing separate signature files. 

### Changed
- update PRONOM to v106

### Fixed
- inconsistent output for `roy inspect priorities`. Reported by [Dave Clipsham](https://github.com/richardlehane/siegfried/issues/192)

## v1.9.3 (2022-05-23)
### Added
- JS/WASM build support contributed by [Andy Jackson](https://github.com/richardlehane/siegfried/pull/188)
- wikidata signature added to `-update`. Contributed by [Ross Spencer](https://github.com/richardlehane/siegfried/pull/178)
- `-nopronom` flag added to `roy inspect` subcommand. Contributed by [Ross Spencer](https://github.com/richardlehane/siegfried/pull/185)

### Changed
- update PRONOM to v104
- update LOC signatures to 2022-05-09
- update Wikidata to 2022-05-20
- update tika-mimetypes signatures to v2.4.0
- update freedesktop.org signatures to v2.2

### Fixed
- invalid JSON output for fmt/1472 due to tab in MIME field. Reported by [Robert Schultz](https://github.com/richardlehane/siegfried/issues/186)
- panic on corrupt Zip containers. Reported by [A. Diamond](https://github.com/richardlehane/siegfried/issues/181)

## v1.9.2 (2022-02-07)
### Added
- Wikidata definition file specification has been updated and now includes endpoint (users will need to harvest Wikidata again)
- Custom Wikibase endpoint can now be specified for harvesting when paired with a custom SPARQL query and property mappings
- Wikidata identifier includes permalinks in results
- Wikidata revision history visible using `roy inspect`
- roy inspect returns format ID with name

### Changed
- update PRONOM to v100
- update LOC signatures to 2022-02-01
- update tika-mimetypes signatures to v2.2.1
- update freedesktop.org signatures to v2.1

### Fixed
- parse issues for container files where zero indexing used for Position. Spotted by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/175)
- sf -droid output can't be read by sf (e.g. for comparing results). Reported by [ostnatalie](https://github.com/richardlehane/siegfried/issues/174)
- panic when running in server mode due to race condition. Reported by [Miguel Guimarães](https://github.com/richardlehane/siegfried/issues/172)
- panic when reading malformed MSCFB files. Reported by [Greg Lepore](https://github.com/richardlehane/siegfried/issues/171)
- unescaped control characters in JSON output. Reported by [Sebastian Lange](https://github.com/richardlehane/siegfried/issues/165)
- zip file names with null terminated strings prevent ID of Serif formats. Reported by [Tyler Thorsted](https://github.com/richardlehane/siegfried/issues/150)

## v1.9.1 (2020-10-11)
### Changed
- update PRONOM to v97
- zs flag now activates -z flag

### Fixed
- details text in PRONOM identifier
- `roy` panic when building signatures with empty sequences. Reported by [Greg Lepore](https://github.com/richardlehane/siegfried/issues/149)

## v1.9.0 (2020-09-22)
### Added
- a new Wikidata identifier, harvesting information from the Wikidata Query Service. Implemented by [Ross Spencer](https://github.com/richardlehane/siegfried/commit/dfb579b4ae46ae6daa814fc3fc74271d768f2f9c). 
- select which archive types (zip, tar, gzip, warc, or arc) are unpacked using the -zs flag (sf -zs tar,zip). Implemented by [Ross Spencer](https://github.com/richardlehane/siegfried/commit/88dd43b55e5f83304705f6bcd439d502ef08cd38).

### Changed
- update LOC signatures to 2020-09-21
- update tika-mimetypes signatures to v1.24
- update freedesktop.org signatures to v2.0

### Fixed
- incorrect basis for some signatures with multiple patterns. Reported and fixed by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/142).

## v1.8.0 (2020-01-22)
### Added
- utc flag returns file modified dates in UTC e.g. `sf -utc FILE | DIR`. Requested by [Dragan Espenschied](https://github.com/richardlehane/siegfried/issues/136)
- new cost and repetition flags to control segmentation when building signatures

### Changed
- update PRONOM to v96
- update LOC signatures to 2019-12-18
- update tika-mimetypes signatures to v1.23
- update freedesktop.org signatures to v1.15

### Fixed
- XML namespaces detected by prefix on root tag, as well as default namespace (for mime-info spec)
- panic when scanning certain MS-CFB files. Reported separately by Mike Shallcross and Euan Cochrane
- file with many FF xx sequences grinds to a halt. Reported by [Andy Foster](https://github.com/richardlehane/siegfried/issues/128)

## v1.7.13 (2019-08-18)
### Added
- the `-f` flag now scans directories, as well as files. Requested by [Harry Moss](https://github.com/richardlehane/siegfried/issues/130)

### Changed
- update LOC signatures to 2019-06-16
- update tika-mimetypes signatures to v1.22

### Fixed
- filenames with "?" were parsed as URLs; reported by [workflowsguy](https://github.com/richardlehane/siegfried/issues/129)

## v1.7.12 (2019-06-15)
### Changed
- update PRONOM to v95
- update LOC signatures to 2019-05-20
- update tika-mimetypes signatures to v1.21

### Fixed
- .docx files with .doc extensions panic due to bug in division of hints in container matcher. Thanks to Jean-Séverin Lair for [reporting and sharing samples](https://github.com/richardlehane/siegfried/issues/126) and to VAIarchief for [additional report with example](https://github.com/richardlehane/siegfried/issues/127).
- mime-info signatures panic on some files due to duplicate entries in the freedesktop and tika signature files; spotted during an attempt at pair coding with Ross Spencer... thanks Ross and sorry for hogging the laptop! [#125](https://github.com/richardlehane/siegfried/issues/125)

## v1.7.11 (2019-02-16)
### Changed
- update LOC signatures to 2019-01-06
- update tika-mimetypes signatures to v1.20

### Fixed
- container matching can now match against directory names. Thanks Ross Spencer for [reporting](https://github.com/richardlehane/siegfried/issues/123) and for the sample SIARD signature file. Thanks Dave Clipsham, Martin Hoppenheit and Phillip Tommerholt for contributions on the ticket.
- fixes to travis.yml for auto-deploy of debian release; [#124](https://github.com/richardlehane/siegfried/issues/124)

## v1.7.10 (2018-09-19)
### Added
- print configuration defaults with `sf -version`

### Changed
- update PRONOM to v94

### Fixed
- LOC identifier fixed after regression in v1.7.9
- remove skeleton-suite files triggering malware warnings by adding to .gitignore; reported by [Dave Rice](https://github.com/richardlehane/siegfried/issues/118)
- release built with Go version 11, which includes a fix for a CIFS error that caused files to be skipped during file walk; reported by [Maarten Savels](https://github.com/richardlehane/siegfried/issues/115)

## v1.7.9 (2018-08-30)
### Added
- save defaults in a configuration file: use the -setconf flag to record any other flags used into a config file. These defaults will be loaded each time you run sf. E.g. `sf -multi 16 -setconf` then `sf DIR` (loads the new multi default)
- use `-conf filename` to save or load from a named config file. E.g. `sf -multi 16 -serve :5138 -conf srv.conf -setconf` and then `sf -conf srv.conf` 
- added `-yaml` flag so, if you set json/csv in default config :(, you can override with YAML instead. Choose the YAML!

### Changed
- the `roy compare -join` options that join on filepath now work better when comparing results with mixed windows and unix paths
- exported decompress package to give more functionality for users of the golang API; requested by [Byron Ruth](https://github.com/richardlehane/siegfried/issues/119)
- update LOC signatures to 2018-06-14
- update freedesktop.org signatures to v1.10
- update tika-mimetype signatures to v1.18

### Fixed
- misidentifications of some files e.g. ODF presentation due to sf quitting early on strong matches. Have adjusted this algorithm to make sf wait longer if there is evidence (e.g. from filename) that the file might be something else. Reported by [Jean-Séverin Lair](https://github.com/richardlehane/siegfried/issues/112)
- read and other file errors caused sf to hang; reports by [Greg Lepore and Andy Foster](https://github.com/richardlehane/siegfried/issues/113); fix contributed by [Ross Spencer](https://github.com/richardlehane/siegfried/commit/ea5300d3639d741a451522958e8b99912f7d639d)
- bug reading streams where EOF returned for reads exactly adjacent the end of file
- bug in mscfb library ([race condition for concurrent access to a global variable](https://github.com/richardlehane/siegfried/issues/117))
- some matches result in extremely verbose basis fields; reported by [Nick Krabbenhoeft](https://github.com/richardlehane/siegfried/issues/111). Partly fixed: basis field now reports a single basis for a match but work remains to speed up matching for these cases.

## v1.7.8 (2017-12-02)
### Changed
- update LOC signatures to 2017-09-28
- update PRONOM signatures to v93

## v1.7.7 (2017-11-30)
### Added
- version information for MIME-info signatures (freedesktop.org and tika-mimetypes) now recorded in mime-info.json file and presented in results
- new sets file for PRONOM extensions. This creates sets like @.doc and @.txt (i.e. all PUIDs with those extensions). Allows you to do commands like `roy build -limit @.doc,@.docx`, `roy inspect @.txt` and `sf -log @.pdf,o DIR`

### Changed
- update freedesktop.org signatures to v1.9

### Fixed
- out of memory error when using `sf -z` on compressed files that contain very large files; reported by [Terry Jolliffe](https://github.com/richardlehane/siegfried/issues/109)
- report errors that occur during file decompression. Previously, only fatal errors encountered when a compressed file is first opened were reported. Now errors that are encountered while attempting to walk the contents of a compressed file are also reported. 
- report errors for 'roy inspect' when roy can't find anything to inspect; reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/108)

## v1.7.6 (2017-10-04)
### Added
- continue on error flag (-coe) can now be used to continue scans despite fatal file errors that would normally cause scanning to halt. This may be useful e.g. for big directory scans over unreliable networks. Usage: `sf -coe DIR`

### Changed
- update PRONOM signatures to v92

### Fixed
- file scanning is now restricted to regular files (i.e. not symlinks, sockets, devices etc.). Reported by [Henk Vanstappen](https://github.com/richardlehane/siegfried/issues/107).
- windows longpath fix now works for paths that appear short

## v1.7.5 (2017-08-12)
### Added
- `sf -update` flag can now be used to download/update non-PRONOM signatures. Options are "loc", "tika", "freedesktop", "pronom-tika-loc", "deluxe" and "archivematica". To update a non-PRONOM signature, include the signature name as an argument after the flags e.g. `sf -update freedesktop`. This command will overwrite 'default.sig' (the default signature file that sf loads). You can preserve your default signature file by providing an alternative `-sig` target e.g. `sf -sig notdefault.sig -update loc`. If you use one of the signature options as a filename (with or without a .sig extension), you can omit the signature argument i.e. `sf -update -sig loc.sig` is equivalent to `sf -sig loc.sig -update loc`. Feature requested by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/103).
- `sf -update` now does SHA-256 hash verification of updates and communication with the update server is via HTTPS.

### Changed
- update PRONOM signatures to v91

### Fixed
- fixes to config package where global variables are polluted with subsquent calls to the Add(Identifier) function
- fix to reader package where panic triggered by illegal slice access in some cases

## v1.7.4 (2017-07-14)
### Added
- `roy build` and `roy add` now take a `-nobyte` flag to omit byte signatures from the identifier; requested by [Nick Krabbenhoeft](https://github.com/richardlehane/siegfried/issues/102) 

### Changed
- update Tika MIMEInfo signatures to 1.16
- update LOC to 2017-06-10

## v1.7.3-(x) (2017-05-30)
### Fixed
- no changes since v1.7.3, repairing Travis-CI auto-deploy of Debian packages

## v1.7.3 (2017-05-20)
### Added
- sf now accepts multiple files or directories as input e.g. `sf myfile1.doc mydir myfile3.txt`
- LOC signature update

### Changed
- code re-organisation to export reader and writer packages
- `sf -replay` can now take lists of results files with `-f` flag e.g. `sf -replay -f list-of-results.txt`

### Fixed
- the command `sf -replay -` now works on Windows as expected e.g. `sf myfiles | sf -replay -json -`
- text matcher not allocating hits to correct identifiers; fixes [#101](https://github.com/richardlehane/siegfried/issues/101)
- unescaped YAML field contains quote; reported by [Ross Spencer](https://github.com/richardlehane/siegfried/issues/100)

## v1.7.2 (2017-04-4)
### Added
- PRONOM v90 update

### Fixed
- the -home flag was being overriden for roy subcommands due to interaction other flags

## v1.7.1 (2017-03-12)
### Added
- signature updates for PRONOM, LOC and tika-mimetypes

### Changed
- `roy inspect` accepts space as well as comma-separated lists of formats e.g. `roy inspect fmt/1 fmt/2`

## v1.7.0 (2017-02-17)
### Added
- log files that match particular formats with `-log fmt/1,@set2` (comma separated list of format IDs/format sets). These can be mixed with regular log options e.g. `-log unknown,fmt/1,chart`
- generate a summary view of formats matched during a scan with `-log chart` (or just `-log c`)
- replay scans from results files with `sf -replay`: load one or more results files to replay logging or to convert to a different output format e.g. `sf -replay -csv results.yaml` or `sf -replay -log unknown,chart,stdout results1.yaml results2.csv`
- compare results with `roy compare` subcommand: view the difference between two or more results e.g. `roy compare results1.yaml results2.csv droid.csv ...`
- `roy sets` subcommand: `roy sets` creates pronom-all.json, pronom-families.json, and pronom-types.json sets files;
`roy sets -changes` creates a pronom-changes.json sets file from a PRONOM release-notes.xml file; `roy sets -list @set1,@set2` lists contents of a comma-separated list of format sets
- `roy inspect releases` provides a summary view of a PRONOM release-notes.xml file

### Changed
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
